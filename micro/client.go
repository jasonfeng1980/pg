package micro

import (
    "context"
    "fmt"
    "github.com/go-kit/kit/endpoint"
    "github.com/go-kit/kit/sd"
    "github.com/go-kit/kit/sd/etcdv3"
    "github.com/go-kit/kit/sd/lb"
    callConf "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/ecode"
    callendpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    callGrpc "github.com/jasonfeng1980/pg/micro/transport/grpc"
    callHttp "github.com/jasonfeng1980/pg/micro/transport/http"
    "github.com/jasonfeng1980/pg/util"
    stdopentracing "github.com/opentracing/opentracing-go"
    zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
    zipkin "github.com/openzipkin/zipkin-go"
    "github.com/openzipkin/zipkin-go/reporter"
    zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
    "google.golang.org/grpc"
    "io"
    "net/url"
    "time"
)

type Client struct {
    Conf callConf.Config
    //ZipkinTracer *zipkin.Tracer
    Tracer stdopentracing.Tracer
    Report  reporter.Reporter
    Middleware []endpoint.Middleware
    Ctx  context.Context
}

type H map[string]interface{}
// 客户端
// e.g. http://serverName/module/version/action
// e.g. grpc://serverName/module/version/action
func (c *Client)Call(ctx context.Context, dns string, params map[string]interface{}) (data interface{}, code int64, msg string){
    m, e:= url.Parse(dns)
    if e != nil {
        return ecode.HttpDnsParseWrong.Parse()
    }
    // 服务发现
    svc := c.getSvcFromEtcd(m.Host, m.Scheme)
    // 添加中间件
    svc = svc.AddMiddleware(c.Middleware)

    return svc.Call(ctx, dns, params)
}

func (c *Client)Close() {
    c.Report.Close()
}

// 链路跟踪
func (c *Client)InitTraceClient() {
    //var err error
    if c.Conf.ZipkinUrl == "" {
        c.Tracer =  stdopentracing.GlobalTracer()
        return
    }

    //创建zipkin上报管理器
    c.Report    = zipkinhttp.NewReporter(c.Conf.ZipkinUrl)

    //创建trace跟踪器
    zEP, _ := zipkin.NewEndpoint(c.Conf.ServerName + ":" + c.Conf.ServerNo, "")
    zipkinTracer, err := zipkin.NewTracer(c.Report, zipkin.WithLocalEndpoint(zEP))

    c.Tracer = zipkinot.Wrap(zipkinTracer)
    if err != nil {
        util.Log.Error(err)
        panic("zipkintracer err:" + err.Error())
    }
}

// 初始化etcd
func (c *Client)getSvcFromEtcd(serverName string, scheme string) callendpoint.Set {
    var factoryFunc sd.Factory
    var prefix string
    if scheme == "grpc" {
        prefix  = fmt.Sprintf("/%s/grpc/", serverName)
        factoryFunc = func(instanceAddr string) (endpoint.Endpoint, io.Closer, error) {
            conn, err := grpc.DialContext(context.Background(), instanceAddr, grpc.WithInsecure())
            if err != nil {
                util.Log.Error(err.Error())
                panic(fmt.Sprintf("连接GRPC失败：%s", instanceAddr))
            }
            eps, err := callGrpc.NewClient(conn,
                callGrpc.DefaultClientOptions(c.Tracer),
            )
            return eps.CallEndpoint, conn, nil
        }
    } else {
        prefix  = fmt.Sprintf("/%s/http/", serverName)
        factoryFunc = func(instanceAddr string) (endpoint.Endpoint, io.Closer, error) {
            eps, err := callHttp.NewClient(instanceAddr,
                callHttp.DefaultClientOptions(c.Tracer),
            )
            return eps.CallEndpoint, nil, err
        }
    }

    //创建etcd连接
    logEtcd := util.Log
    client, err := etcdv3.NewClient(c.Ctx,
        []string{c.Conf.EtcdAddr},
        etcdv3.ClientOptions{
            DialTimeout: c.Conf.EtcdTimeout,
            DialKeepAlive: c.Conf.EtcdKeepAlive,
        })
    if err != nil {
        logEtcd.With( "server", prefix).Error(err.Error())
        panic(ecode.EtcdDisconnect.Error(c.Conf.EtcdAddr, err.Error()))
    }

    instancer, e := etcdv3.NewInstancer(client, prefix, logEtcd)
    if e != nil {
        logEtcd.With( "server", prefix).Error(err.Error())
        panic(ecode.NoMicroServer.Error(prefix, err.Error()))
    }
    endpointer := sd.NewEndpointer(instancer, factoryFunc, logEtcd)

    // 随机请求
    balancer := lb.NewRandom(endpointer, time.Now().UnixNano())
    reqEndPoint := lb.Retry(c.Conf.EtcdRetryTimes, c.Conf.EtcdTimeout, balancer)
    return callendpoint.Set{
        CallEndpoint: reqEndPoint,
    }
}
