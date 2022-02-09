package micro

import (
    "context"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/ecode"
    callendpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    "github.com/jasonfeng1980/pg/micro/finder"
    callGrpc "github.com/jasonfeng1980/pg/micro/transport/grpc"
    callHttp "github.com/jasonfeng1980/pg/micro/transport/http"
    "github.com/opentracing/opentracing-go"
    zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
    "github.com/openzipkin/zipkin-go"
    "github.com/openzipkin/zipkin-go/reporter"
    zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
    "github.com/sony/gobreaker"
    "net/url"
    "time"
)

type Client struct {
    ctx  context.Context
    c    *conf.Config
    Tracer opentracing.Tracer
    Report  reporter.Reporter
    Middleware []callendpoint.Middleware
}

var clientInstance *Client
func NewClient() (*Client, error){
    if clientInstance != nil { // 有缓存就直接返回
        return clientInstance, nil
    }
    tracer, report, err := InitTraceClient(conf.Conf)
    if err != nil {
        return nil, err
    }
    client := &Client{
        ctx: context.Background(),
        c:   conf.Conf,
        Tracer: tracer,
        Report: report,
    }

    if client.Tracer!= nil {
        client.Middleware = []callendpoint.Middleware{
            callendpoint.TraceClient("TraceClient", client.Tracer),
        }
    }
    client.Middleware = append(client.Middleware,
        callendpoint.LimitDelaying(conf.Conf.LimitClient),
        callendpoint.Gobreaking(gobreaker.Settings{
            Name:    "Gobreaking-" + conf.Conf.Name,
            Timeout: 30 * time.Second,
        }))
    clientInstance = client
    return clientInstance, nil
}

// 链路跟踪
func InitTraceClient(c *conf.Config) (tracer opentracing.Tracer, report reporter.Reporter, err error){
    if c.ZipkinUrl == "" {
        tracer =  opentracing.GlobalTracer()
        return
    }

    //创建zipkin上报管理器
    report    = zipkinhttp.NewReporter(c.ZipkinUrl)

    //创建trace跟踪器
    ept, _ := zipkin.NewEndpoint(c.Name + ":" + c.Num, "")
    zipkinTracer, err := zipkin.NewTracer(report, zipkin.WithLocalEndpoint(ept))

    tracer = zipkinot.Wrap(zipkinTracer)
    return
}

// 客户端
// e.g. http://serverName/module/version/action
// e.g. grpc://serverName/module/version/action
func (c *Client)Call(ctx context.Context, dns string, params map[string]interface{}) (data interface{}, code int64, msg string){
    m, err:= url.Parse(dns)
    if err != nil {
        return ecode.HttpDnsParseWrong.Parse()
    }
    // 服务发现
    svc, err := c.getSvcFromEtcd(m.Host, m.Scheme)
    if err != nil {
        return ecode.EtcdFindErr.Parse(err.Error())
    }
    // 添加中间件
    svc = svc.AddMiddleware(c.Middleware)

    return svc.Call(ctx, dns, params)
}

func (c *Client)Close() {
    // 为了复用，单个调用不重复创建Report
    //c.Report.Close()
}

// 初始化client
func (c *Client)getSvcFromEtcd(serverName string, scheme string) (callendpoint.MicroEndpoint, error) {
    etcdClient, err := finder.NewEtcd(c.ctx, conf.Conf.ETCD)
    if err != nil {
        return callendpoint.MicroEndpoint{}, err
    }

    ept, err := etcdClient.Endpoint(serverName, scheme, c.etcdFunc)

    return callendpoint.MicroEndpoint{Endpoint: ept}, err
}

func (c *Client)etcdFunc(scheme string, instanceAddr string) (callendpoint.Endpoint,  error) {
    switch scheme {
    case "grpc":
        return callGrpc.NewClient(instanceAddr, c.Tracer)
    case "http":
        return callHttp.NewClient(instanceAddr, c.Tracer)
    default:
        return nil, ecode.DnsWrongScheme.Error()
    }
}