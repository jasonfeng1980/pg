package pg

import (
    "context"
    "fmt"
    "github.com/go-kit/kit/endpoint"
    callConf "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/database/rdb"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/micro"
    callendpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    "github.com/jasonfeng1980/pg/micro/service"
    "github.com/jasonfeng1980/pg/util"
    "github.com/sony/gobreaker"
    "time"
)

// 从YAML配置系统
func SetConfYaml(mysqlFile string, redisFile string, serverFile string, root string) error{
    globalMysql, err := yamlToMysql(mysqlFile)
    if err != nil {
        return err
    }
    globalRedis, err := yamlToRedis(redisFile)
    if err != nil {
        return err
    }

    serverConf, err := yamlToServer(serverFile, globalMysql, globalRedis)
    if err != nil {
        return err
    }

    root = util.FileRealPath(root)
    fmt.Println("系统根目录：", root)
    serverConf.ServerRoot = root
    callConf.Set(*serverConf)


    util.LogInit(serverConf)

    return nil
}

// 设置系统配置
func SetConf(c callConf.Config, root string){
    root = util.FileRealPath(root)
    fmt.Println("系统根目录：", root)
    c.ServerRoot = root
    callConf.Set(c)
    util.LogInit(&c)
}
func SetRoot(root string){
    c := callConf.Get()
    root = util.FileRealPath(root)
    fmt.Println("系统根目录：", root)
    c.ServerRoot = root
    callConf.Set(c)
}

///////// server  /////////
// 服务
func Server() *micro.Server {
    return &micro.Server{
        Conf:   callConf.Get(),
    }
}

///////// client  /////////
// 服务
func Client() *micro.Client {
    config := callConf.Get()
    c := &micro.Client{
        Conf: config,
        Ctx: context.Background(),
    }
    c.InitTraceClient()
    c.Middleware = []endpoint.Middleware{
        callendpoint.TraceClient("TraceClient", c.Tracer),
        callendpoint.ZipkinTrace(c.ZipkinTracer),
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
    MicroApi = service.Api
    MySQL = db.MYSQL
    Redis = rdb.Redis
    Filter = &db.Filter{}

    LogInfo = util.LogHandle("info")
    LogErr = util.LogHandle("error")
    LogDebug = util.LogHandle("debug")
    LogInit = util.LogInit

    Ecode = ecode.Err
)

type H map[string]interface{}
func Success(data interface{}) (interface{}, int64, string){
    return data, 200, ""
}
func Error(e error) (interface{}, int64, string) {
    code, msg := ecode.ReadError(e)
    return nil, code, msg
}
func ErrCode(code int64, msg string)(interface{}, int64, string) {
    return nil, code, msg
}
func D(kvs ...interface{}){
    LogDebug.Log(kvs...)
}

