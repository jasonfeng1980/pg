package micro

import (
    "context"
    "fmt"
    "github.com/go-kit/kit/sd/etcdv3"
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/database/rdb"
    "github.com/jasonfeng1980/pg/util"
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

    "github.com/oklog/oklog/pkg/group"
    stdopentracing "github.com/opentracing/opentracing-go"
    zipkin "github.com/openzipkin/zipkin-go"
    zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
    stdprometheus "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"

    "github.com/go-kit/kit/endpoint"
    "github.com/go-kit/kit/log"
    "github.com/go-kit/kit/metrics/prometheus"
    kitgrpc "github.com/go-kit/kit/transport/grpc"

    callConf "github.com/jasonfeng1980/pg/conf"
    callendpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    callservice "github.com/jasonfeng1980/pg/micro/service"
    callGrpc "github.com/jasonfeng1980/pg/micro/transport/grpc"
    callpb "github.com/jasonfeng1980/pg/micro/transport/grpc/pb"
    callHttp "github.com/jasonfeng1980/pg/micro/transport/http"
)

type Server struct {
    Conf callConf.Config
}

var ctx = context.Background()

func (s *Server) Run() {
    logger := util.LogHandle("info")
    defer util.LogClose()

    // 链接MYSQL连接池
    db.MYSQL.Conn(s.Conf.MySQLConf)
    defer db.MYSQL.Close()

    // 连接redis连接池
    rdb.Redis.Conn(s.Conf.RedisConf)
    defer rdb.Redis.Close()
    // 设置MYSQL缓存redis  - 缓存10秒
    db.MYSQL.SetCacheRedis(rdb.Redis.Client("demo"), time.Second * 10)


    // 链路跟踪
    zipkinTracer, tracer, reporter := initTraceServer(s.Conf)
    defer reporter.Close()
    // 监控统计
    calls, duration := initMetrics(s.Conf)
    var (
        // 中间件
        mdw = getMiddleware(s.Conf.LimitServer, s.Conf.BreakerServer, zipkinTracer, tracer, duration)
        // 服务
        service = callservice.New([]callservice.Middleware{
            callservice.LoggingMiddleware(),
            callservice.InstrumentingMiddleware(calls),
        })
        endpoints = callendpoint.New(service).AddMiddleware(mdw)
        httpServer = callHttp.NewServer(endpoints, s.Conf, callHttp.DefaultServerOptions(tracer, zipkinTracer))
        grpcServer  = callGrpc.NewServer(endpoints, callGrpc.DefaultServerOptions(tracer, zipkinTracer))
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
    initEtcdServer(ctx, s.Conf)

    util.LogHandle("info").Log("exit", g.Run())
}
// 链路跟踪
func initTraceServer(conf callConf.Config) (*zipkin.Tracer, stdopentracing.Tracer, reporter.Reporter){
    tracer := stdopentracing.GlobalTracer()
    if conf.ZipkinUrl == "" {
        return nil, tracer, nil
    }
    //创建zipkin上报管理器
    reporter    := zipkinhttp.NewReporter(conf.ZipkinUrl)

    //创建trace跟踪器
    zEP, _ := zipkin.NewEndpoint(conf.ServerName + ":" + conf.ServerNo, "")
    zipkinTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
    if err != nil {
        util.LogHandle("error").Log("err", err, "k", "vvvv")
        os.Exit(1)
    }
    tracer = stdopentracing.GlobalTracer()
    return zipkinTracer, tracer, reporter
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
        Help:      "Request duration in seconds.",
    }, []string{"method", "success"})
    http.DefaultServeMux.Handle("/metrics", promhttp.Handler())
    return calls, duration
}
// 获取中间件
func getMiddleware(limit *rate.Limiter, breakerSetting gobreaker.Settings, zipkinTracer *zipkin.Tracer, tracer stdopentracing.Tracer, duration *prometheus.Summary) []endpoint.Middleware{
    return []endpoint.Middleware{
        callendpoint.LimitErroring(limit),
        callendpoint.Gobreaking(breakerSetting),
        callendpoint.TraceServer(tracer),
        callendpoint.ZipkinTrace(zipkinTracer),
        callendpoint.LoggingMiddleware(),
        callendpoint.InstrumentingMiddleware(duration.With("method", "Call")),
    }
}
// 初始化etcd
func initEtcdServer(ctx context.Context, conf callConf.Config) {
    logger := util.LogHandle("etcd")
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
    logger := util.LogHandle("debug")
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
            httpServer.Shutdown(ctx)
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
            httpServer.Shutdown(ctx)
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
