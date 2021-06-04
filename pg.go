package pg

import (
    "context"
    "github.com/go-kit/kit/endpoint"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/database/mdb"
    "github.com/jasonfeng1980/pg/database/rdb"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/micro"
    callendpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    "github.com/jasonfeng1980/pg/micro/service"
    "github.com/jasonfeng1980/pg/mq/rabbitmq"
    "github.com/jasonfeng1980/pg/util"
    "github.com/sony/gobreaker"
    "net/http"
    "os"
    "time"
)

///////// server  /////////
// 服务
func Server(root ...string) *micro.Server {
    ret := &micro.Server{
        Conf:   conf.Get(),
    }
    if len(root) == 1 {
        ret.Conf.ServerRoot = util.FileRealPath(root[0])
    }
    return ret
}

///////// client  /////////
// 客户端
func Client() *micro.Client {
    config := conf.Get()
    c := &micro.Client{
        Conf: config,
        Ctx: context.Background(),
    }
    c.InitTraceClient()
    c.Middleware = []endpoint.Middleware{
        callendpoint.TraceClient("TraceClient", c.Tracer),
        callendpoint.LimitDelaying(config.LimitClient),
        callendpoint.Gobreaking(gobreaker.Settings{
            Name:    "Gobreaking-" + config.ServerName,
            Timeout: 30 * time.Second,
        }),
    }
    return c
}

///////// 快捷方法  /////////
var (
    Ecode = ecode.Err
    NewError   = ecode.NewError
    Root = conf.ConfInit

    MicroApi = service.Api
    MySQL = db.MYSQL
    Mongo = mdb.MONGO
    Expr = db.Expr
    Redis = rdb.Redis
    RabbitMQ = rabbitmq.RabbitMq
    Filter = &db.Filter{}
    YamlRead = conf.ConfInit
    Log = &util.Log

)
type M map[string]interface{}
func Suc(data interface{}) (interface{}, int64, string){
    return data, 200, ""
}
// error => nil, code, msg  （ api 输出格式）
func Err(e error) (interface{}, int64, string) {
    code, msg := ecode.ReadError(e)
    return nil, code, msg
}
// code, msg  => nil, code, msg  （ api 输出格式）
func ErrCode(code int64, msg string)(interface{}, int64, string) {
    return nil, code, msg
}
// 调试输出， 需开启debug模式
func D(l ...interface{}){
    for _, v:=range l{
        util.Log.LogPretty(v, 3)
    }
}
// 调试输出 并退出
func DD(l ...interface{}) {
    D(l...)
    os.Exit(1)
}
// 获取HTTP请求的的Request句柄
func Request(ctx context.Context) (r *http.Request, ok bool){
    ret := ctx.Value(service.RequestHandle)
    r, ok = ret.(*http.Request)
    return
}
