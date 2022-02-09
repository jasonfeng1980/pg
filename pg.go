package pg

import (
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/database/mdb"
    "github.com/jasonfeng1980/pg/database/rdb"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/micro"
    "github.com/jasonfeng1980/pg/micro/service"
    "github.com/jasonfeng1980/pg/mq/rabbitmq"
    "github.com/jasonfeng1980/pg/util"
)

var (
    Conf = conf.ConfBox      // 配置，viper的实例
    MicroApi = service.Api   // 提供API的快捷方式
    ECode = ecode.Err       // 根据错误码和错误生成ECode实例
    NewError   = ecode.NewError // 生成待错误码的错误

    Logger   = util.Log()

    Client = micro.NewClient
    Server = micro.NewServer

    MySQL  = db.MYSQL
    Redis  = rdb.Redis
    Mongo = mdb.MONGO
    RabbitMQ = rabbitmq.RabbitMq

    Filter = &db.Filter{}


    StarTransaction = util.StartTransaction
    Commit = util.Commit
    Rollback = util.Rollback
)

// 加载配置
func Load(configFile string)  error {
    return conf.Load(configFile)
}


/////////////////////////////////////////////////////////////////
// 简化操作方法
/////////////////////////////////////////////////////////////////

type M map[string]interface{}

// 执行成功
func Suc(data interface{}) (interface{}, int64, string){
    return data, 200, ""
}
// 执行失败
// error => nil, code, msg  （ api 输出格式）
func Err(e error) (interface{}, int64, string) {
    code, msg := ecode.ReadError(e)
    return nil, code, msg
}
// 执行失败，并指定错误码
// code, msg  => nil, code, msg  （ api 输出格式）
func ErrCode(code int64, msg string)(interface{}, int64, string) {
    return nil, code, msg
}
// 调试错误
func D(msgList ...interface{}){
    util.Logs(msgList...)
}