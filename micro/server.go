package micro

import (
    "context"
    "fmt"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/database/mdb"
    "github.com/jasonfeng1980/pg/database/rdb"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/micro/finder"
    "github.com/prometheus/client_golang/prometheus"
    callHttp "github.com/jasonfeng1980/pg/micro/transport/http"
    "github.com/jasonfeng1980/pg/util"
    "github.com/oklog/oklog/pkg/group"
    "github.com/opentracing/opentracing-go"
    zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
    "github.com/openzipkin/zipkin-go"
    "github.com/openzipkin/zipkin-go/reporter"
    zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/sony/gobreaker"
    "golang.org/x/time/rate"
    "google.golang.org/grpc"
    "io"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    callendpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    callservice "github.com/jasonfeng1980/pg/micro/service"
    callGrpc "github.com/jasonfeng1980/pg/micro/transport/grpc"
    callpb "github.com/jasonfeng1980/pg/micro/transport/grpc/pb"
)

type Server struct {
    ctx  context.Context
    c    *conf.Config
    closer []io.Closer				// 结算需要关闭的
}

func NewServer(ctx context.Context) *Server{
    return &Server{
        ctx: ctx,
        c:   conf.Conf,
    }
}

// 添加defer 关闭
func (s *Server) AddCloseFunc(f io.Closer) {
    s.closer = append(s.closer, f)
}
func (s *Server) Close() {
    for _, f :=range s.closer{
        f.Close()
    }
    s.closer = nil
}
// 数据库连接池
func (s *Server) ConnDB(){
    s.ctx = context.Background()
    // 链接MYSQL连接池
    db.MYSQL.Conn(s.c.MySQL)
    s.AddCloseFunc(db.MYSQL)
    // 链接MYSQL连接池
    mdb.MONGO.Conn(s.c.Mongo)
    s.AddCloseFunc(mdb.MONGO)
    // 连接redis连接池
    rdb.Redis.Conn(s.c.Redis)
    s.AddCloseFunc(rdb.Redis)
    // 设置MYSQL和MONGODB缓存redis句柄
    if s.c.CacheRedis != "" {
        if r, err := rdb.Redis.Client(s.c.CacheRedis); err == nil {
            db.MYSQL.SetCacheRedis(r, time.Second * time.Duration(s.c.CacheSec))
            mdb.MONGO.SetCacheRedis(r, time.Second * time.Duration(s.c.CacheSec))
        } else {
            util.Error(err)
        }
    }
    //// 链接rabbitmq
    //rabbitmq.RabbitMq.Conn(s.c.RabbitMQConf)
    //s.AddCloseFunc(rabbitmq.RabbitMq)
}
type ScriptFunc func() error
func (s *Server)Script(f ScriptFunc) {
    defer s.Close()
    // 链接数据库和MQ
    s.ConnDB()

    var g group.Group
    // 优雅退出
    initShutdown(&g)
    // 开始执行脚本方法
    initScriptFunc(&g, f)

    util.Info("", "Exit", g.Run())
}

func (s *Server) Run() {
    defer s.Close()

    // 链接数据库和MQ
    s.ConnDB()

    // 链路跟踪
    tracer, reporter := initTraceServer(s.c)
    if reporter != nil {
       defer reporter.Close()
    }
    // 监控统计
    calls, duration := initMetrics(s.c)
    var (
        // 中间件
        mdw = getMiddleware(s.c.LimitServer, s.c.BreakerServer, tracer)

        // 服务
        service = callservice.New([]callservice.Middleware{
            callservice.LoggingMiddleware(),
            callservice.InstrumentingMiddleware(calls, duration),
        })
        endpoints = callendpoint.New(service).AddMiddleware(mdw)
        httpServer = callHttp.NewServer(endpoints, tracer, "Call")
        grpcServer  = callGrpc.NewServer(endpoints, tracer, "Call")
    )

    var g group.Group
    // 测试服务
    initDebugServer(&g, s.c)
    // http(s)服务
    initHttpServer(&g, s.c, httpServer)
    // grpc服务
    initGrpcServer(&g, s.c, grpcServer)
    // 优雅退出
    initShutdown(&g)

    // 服务发现-etcd
    initEtcdServer(s.ctx, s.c)

    util.Info("", "Exit", g.Run())
}

func initMetrics(conf *conf.Config) (*prometheus.CounterVec, *prometheus.SummaryVec){
    calls := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: conf.Name,
            Subsystem: conf.Num,
            Name:      "count",
            Help:      "统计请求次数.",
        },
        []string{"code", "method"},
    )
    prometheus.MustRegister(calls)

    duration := prometheus.NewSummaryVec(
        prometheus.SummaryOpts{
            Namespace: conf.Name,
            Subsystem: "call",
            Name:      "request",
            Help:      "请求持续时间（秒）",
            Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
        },
        []string{"success", "dns"},
    )
    // 开启DebugAddr 才可以访问这个地址
    http.DefaultServeMux.Handle("/metrics", promhttp.Handler())
    prometheus.MustRegister(duration)
    return calls, duration
}

// 启动脚本
func initScriptFunc(g *group.Group, f ScriptFunc) {
    g.Add(func() error {
        return f()
    }, func(err error) {
        if err != nil {
            util.Error(err.Error())
        }
    })
}
// 链路跟踪
func initTraceServer(conf *conf.Config) (opentracing.Tracer, reporter.Reporter){
    if conf.ZipkinUrl == "" {
        return nil, nil
    }
    tracer := opentracing.GlobalTracer()
    if conf.ZipkinUrl == "" {
        return tracer, nil
    }
    //创建zipkin上报管理器
    reporter    := zipkinhttp.NewReporter(conf.ZipkinUrl)

    //创建trace跟踪器
    zEP, _ := zipkin.NewEndpoint(conf.Name + ":" + conf.Num, "")
    zipkinTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
    if err != nil {
        util.Error(err)
        panic(err)
    }
    tracer = zipkinot.Wrap(zipkinTracer)
    return tracer, reporter
}

// 获取中间件
func getMiddleware(limit *rate.Limiter, breakerSetting gobreaker.Settings, tracer opentracing.Tracer) []callendpoint.Middleware{
    mid := []callendpoint.Middleware{
        callendpoint.LimitErroring(limit),
        callendpoint.Gobreaking(breakerSetting),
        callendpoint.LoggingMiddleware(),
    }
    if tracer!= nil {
        mid = append(mid, callendpoint.TraceServer(tracer))
    }
    return mid
}

// 初始化etcd
func initEtcdServer(ctx context.Context, conf *conf.Config) {
    if conf.ETCD == "" { // 如果没有配置，就不走etcd
        return
    }

    etcd, err := finder.NewEtcd(ctx, conf.ETCD)
    if err != nil {
        util.Error(err.Error())
        panic(ecode.EtcdDisconnect.Error(conf.ETCD, err.Error()))
    }

    if conf.AddrHttp != "" {
        k := fmt.Sprintf("/%s/%s/%s", conf.Name, "http", conf.Name,)
        etcd.Register(k, conf.AddrHttp)
    }
    if conf.AddrGrpc != "" {
        k := fmt.Sprintf("/%s/%s/%s", conf.Name, "grpc", conf.Name,)
        etcd.Register(k, conf.AddrGrpc)
    }
}

// 测试服务
func initDebugServer(g *group.Group, conf *conf.Config) {
    if conf.AddrDebug == "" {
        return
    }
    debugListener, err := net.Listen("tcp", conf.AddrDebug)
    if err != nil {
        util.Error("", "transport", "debug/HTTP", "during", "Listen", "err", err)
        os.Exit(1)
    }
    g.Add(func() error {
        util.Info("", "transport", "debug/HTTP", "addr", conf.AddrDebug)
        return http.Serve(debugListener, http.DefaultServeMux)
    }, func(error) {
        debugListener.Close()
    })
}

// http服务
func initHttpServer(g *group.Group, conf *conf.Config, httpServer http.Server) {
    if conf.AddrHttp != "" {
        g.Add(func() error {
            util.Info("","transport", "HTTP", "addr", conf.AddrHttp)
            httpServer.Addr = conf.AddrHttp
            return httpServer.ListenAndServe()
        }, func(error) {
            httpServer.Shutdown(context.Background())
        })
    }
    if len(conf.AddrHttps) == 3 && conf.AddrHttps[0] != "" {
       if conf.AddrHttps[1][:1] != "/" {
           conf.AddrHttps[1] = conf.Root + "/" + conf.AddrHttps[1]
       }
       if conf.AddrHttps[2][:1] != "/" {
           conf.AddrHttps[2] = conf.Root + "/" + conf.AddrHttps[2]
       }
       g.Add(func() error {
           util.Info("", "transport", "HTTPS", "addr", conf.AddrHttps[0])
           httpServer.Addr = conf.AddrHttps[0]
           return httpServer.ListenAndServeTLS(conf.AddrHttps[1], conf.AddrHttps[2])
       }, func(error) {
           httpServer.Shutdown(context.Background())
       })
    }
}
// grpc服务
func initGrpcServer(g *group.Group, conf *conf.Config, grpcServer callpb.ServiceServer) {
    if conf.AddrGrpc == "" {
        return
    }
    grpcListener, err := net.Listen("tcp", conf.AddrGrpc)
    if err != nil {
        util.Info("", "transport", "GRPC", "during", "Listen", "err", err)
        os.Exit(1)
    }
    g.Add(func() error {
        util.Info("", "transport", "gRPC", "addr", conf.AddrGrpc)

        baseServer :=  grpc.NewServer()
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
            return fmt.Errorf("收到信号【%s】", sig)
        case <-cancelInterrupt:
            return nil
        }
    }, func(error) {
        close(cancelInterrupt)
    })
}
