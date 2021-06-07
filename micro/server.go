package micro

import (
    "context"
    "fmt"
    "github.com/go-kit/kit/sd/etcdv3"
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/database/mdb"
    "github.com/jasonfeng1980/pg/database/rdb"
    "github.com/jasonfeng1980/pg/mq/rabbitmq"
    "github.com/jasonfeng1980/pg/util"
    zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
    "github.com/openzipkin/zipkin-go/reporter"
    "github.com/sony/gobreaker"
    "golang.org/x/time/rate"
    "google.golang.org/grpc"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-kit/kit/endpoint"
    "github.com/go-kit/kit/log"
    "github.com/go-kit/kit/metrics/prometheus"
    kitgrpc "github.com/go-kit/kit/transport/grpc"
    "github.com/oklog/oklog/pkg/group"
    stdopentracing "github.com/opentracing/opentracing-go"
    zipkin "github.com/openzipkin/zipkin-go"
    zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
    stdprometheus "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"

    callConf "github.com/jasonfeng1980/pg/conf"
    callendpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    callservice "github.com/jasonfeng1980/pg/micro/service"
    callGrpc "github.com/jasonfeng1980/pg/micro/transport/grpc"
    callpb "github.com/jasonfeng1980/pg/micro/transport/grpc/pb"
    callHttp "github.com/jasonfeng1980/pg/micro/transport/http"
)

type CloseFunc func()

type Server struct {
    Conf callConf.Config
    ctx  context.Context
    closePool []CloseFunc
}

// 添加defer 关闭
func (s *Server) AddCloseFunc(f CloseFunc) {
    s.closePool = append(s.closePool, f)
}
func (s *Server) Close() {
    for _, f :=range s.closePool{
        f()
    }
}

func (s *Server) ConnDB(){
    s.ctx = context.Background()
    // 链接MYSQL连接池
    db.MYSQL.Conn(s.Conf.MySQLConf)
    s.AddCloseFunc(db.MYSQL.Close)
    // 链接MYSQL连接池
    mdb.MONGO.Conn(s.Conf.MongoConf)
    s.AddCloseFunc(mdb.MONGO.Close)
    // 连接redis连接池
    rdb.Redis.Conn(s.Conf.RedisConf)
    s.AddCloseFunc(rdb.Redis.Close)
    // 设置MYSQL缓存redis句柄
    if s.Conf.CacheRedis != "" {
        if r, err := rdb.Redis.Client(s.Conf.CacheRedis); err == nil {
            db.MYSQL.SetCacheRedis(r, time.Second * time.Duration(s.Conf.CacheSec))
        } else {
            util.Log.Errorln(err)
        }
    }
    // 链接rabbitmq
    rabbitmq.RabbitMq.Conn(s.Conf.RabbitMQConf)
    s.AddCloseFunc(rabbitmq.RabbitMq.Close)
}

type ScriptFunc func() error
func (s *Server) Script(f ScriptFunc){
    defer s.Close()

    // 链接数据库和MQ
    s.ConnDB()

    var g group.Group
    // 优雅退出
    initShutdown(&g)
    // 开始执行脚本方法
    initScriptFunc(&g, f)

    util.Log.Debugln("exit", g.Run())

}
func initScriptFunc(g *group.Group, f ScriptFunc) {
    g.Add(func() error {
        return f()
    }, func(err error) {
        if err != nil {
            util.Log.Error(err.Error())
        }
    })
}

func (s *Server) Run() {
    defer s.Close()

    logger := util.Log
    // 链接数据库和MQ
    s.ConnDB()

    // 链路跟踪
    tracer, reporter := initTraceServer(s.Conf)
    if reporter != nil {
        defer reporter.Close()
    }
    // 监控统计
    calls, duration := initMetrics(s.Conf)
    var (
        // 中间件
        mdw = getMiddleware(s.Conf.LimitServer, s.Conf.BreakerServer, tracer, duration)
        // 服务
        service = callservice.New([]callservice.Middleware{
            callservice.LoggingMiddleware(),
            callservice.InstrumentingMiddleware(calls),
        })
        endpoints = callendpoint.New(service).AddMiddleware(mdw)
        httpServer = callHttp.NewServer(endpoints, s.Conf, callHttp.DefaultServerOptions(tracer))
        grpcServer  = callGrpc.NewServer(endpoints, callGrpc.DefaultServerOptions(tracer))
    )

    var g group.Group
    // 测试服务
    initDebugServer(&g, s.Conf)
    // http服务
    initHttpServer(&g, s.Conf, httpServer, logger)
    // grpc服务
    initGrpcServer(&g, s.Conf, grpcServer, logger)
    // 优雅退出
    initShutdown(&g)

    // 服务发现-etcd
    initEtcdServer(s.ctx, s.Conf)

    util.Log.Debugf("exit", g.Run())
}
// 链路跟踪
func initTraceServer(conf callConf.Config) (stdopentracing.Tracer, reporter.Reporter){
    tracer := stdopentracing.GlobalTracer()
    if conf.ZipkinUrl == "" {
        return tracer, nil
    }
    //创建zipkin上报管理器
    reporter    := zipkinhttp.NewReporter(conf.ZipkinUrl)

    //创建trace跟踪器
    zEP, _ := zipkin.NewEndpoint(conf.ServerName + ":" + conf.ServerNo, "")
    zipkinTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
    if err != nil {
        util.Log.Error(err)
        panic(err)
    }
    tracer = zipkinot.Wrap(zipkinTracer)
    return tracer, reporter
}
// 监控统计
func initMetrics(conf callConf.Config) (*prometheus.Counter, *prometheus.Summary) {
    calls := prometheus.NewCounterFrom(stdprometheus.CounterOpts{
        Namespace: conf.ServerName,
        Subsystem: "call",
        Name:      "count",
        Help:      "统计请求次数.",
    }, []string{})
    duration := prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
        Namespace: conf.ServerName,
        Subsystem: "call",
        Name:      "request",
        Help:      "请求持续时间（秒）",
    }, []string{"method", "success"})
    http.DefaultServeMux.Handle("/metrics", promhttp.Handler())
    // 开启DebugAddr 才可以访问这个地址
    return calls, duration
}
// 获取中间件
func getMiddleware(limit *rate.Limiter, breakerSetting gobreaker.Settings, tracer stdopentracing.Tracer, duration *prometheus.Summary) []endpoint.Middleware{
    return []endpoint.Middleware{
        callendpoint.LimitErroring(limit),
        callendpoint.Gobreaking(breakerSetting),
        callendpoint.TraceServer(tracer),
        callendpoint.LoggingMiddleware(),
        callendpoint.InstrumentingMiddleware(duration.With("method", "Call")),
    }
}
// 初始化etcd
func initEtcdServer(ctx context.Context, conf callConf.Config) {
    logger := util.Log
    if conf.HttpAddr != "" {
        k := fmt.Sprintf("/%s/%s/%s", conf.ServerName, "http", conf.ServerNo)
        regEtcdHttpServer(ctx, conf, logger, k, conf.HttpAddr)
    }
    if conf.GrpcAddr != "" {
        k := fmt.Sprintf("/%s/%s/%s", conf.ServerName, "grpc", conf.ServerNo)
        regEtcdHttpServer(ctx, conf, logger, k, conf.GrpcAddr)
    }
}
func regEtcdHttpServer(ctx context.Context, conf callConf.Config, logger log.Logger, k string, v string){
    //创建etcd连接
    client, err := etcdv3.NewClient(ctx,
        []string{conf.EtcdAddr},
        etcdv3.ClientOptions{
            DialTimeout: conf.EtcdTimeout,
            DialKeepAlive: conf.EtcdKeepAlive,
        })
    if err != nil {
        panic(err)
    }
    etcdv3.NewRegistrar(client, etcdv3.Service{
        Key: k,
        Value: v,
    }, logger).Register()
}


// 测试服务
func initDebugServer(g *group.Group, conf callConf.Config) {
    logger := util.Log
    if conf.DebugAddr == "" {
        return
    }
    debugListener, err := net.Listen("tcp", conf.DebugAddr)
    if err != nil {
        logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
        os.Exit(1)
    }
    g.Add(func() error {
        logger.Log("transport", "debug/HTTP", "addr", conf.DebugAddr)
        return http.Serve(debugListener, http.DefaultServeMux)
    }, func(error) {
        debugListener.Close()
    })
}
// http服务
func initHttpServer(g *group.Group, conf callConf.Config, httpServer http.Server, logger log.Logger) {
    if conf.HttpAddr != "" {
        g.Add(func() error {
            logger.Log("transport", "HTTP", "addr", conf.HttpAddr)
            httpServer.Addr = conf.HttpAddr
            return httpServer.ListenAndServe()
        }, func(error) {
            httpServer.Shutdown(context.Background())
        })
    }
    if len(conf.HttpsInfo) == 3 && conf.HttpsInfo[0]!=""{
        if conf.HttpsInfo[1][:1] != "/" {
            conf.HttpsInfo[1] = conf.ServerRoot + "/" + conf.HttpsInfo[1]
        }
        if conf.HttpsInfo[2][:1] != "/" {
            conf.HttpsInfo[2] = conf.ServerRoot + "/" + conf.HttpsInfo[2]
        }
        g.Add(func() error {
            logger.Log("transport", "HTTPS", "addr", conf.HttpsInfo[0])
            httpServer.Addr = conf.HttpsInfo[0]

            return httpServer.ListenAndServeTLS(conf.HttpsInfo[1],conf.HttpsInfo[2])
        }, func(error) {
            httpServer.Shutdown(context.Background())
        })
    }

}
// grpc服务
func initGrpcServer(g *group.Group, conf callConf.Config, grpcServer callpb.ServiceServer, logger log.Logger) {
    if conf.GrpcAddr == "" {
        return
    }
    grpcListener, err := net.Listen("tcp", conf.GrpcAddr)
    if err != nil {
        logger.Log("transport", "gRPC", "during", "Listen", "err", err)
        os.Exit(1)
    }
    g.Add(func() error {
        logger.Log("transport", "gRPC", "addr", conf.GrpcAddr)

        baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
        callpb.RegisterServiceServer(baseServer, grpcServer)
        return baseServer.Serve(grpcListener)
    }, func(error) {
        grpcListener.Close()
    })
}

// 优雅退出
func initShutdown(g *group.Group) {
    cancelInterrupt := make(chan struct{})
    g.Add(func() error {
        c := make(chan os.Signal, 1)
        signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
        select {
        case sig := <-c:
            return fmt.Errorf("received signal %s", sig)
        case <-cancelInterrupt:
            return nil
        }
    }, func(error) {
        close(cancelInterrupt)
    })
}
