package conf

import (
    "github.com/sony/gobreaker"
    "golang.org/x/time/rate"
    "time"
)

type Config struct {
    // 整体配置
    ServerName  string  // 服务名称
    ServerNo    string  // 服务序号
    ServerRoot  string  // 项目根目录

    // 日志
    LogDir      string  // 日志文件夹
    LogShowDebug bool   // 日志是否记录debug


    // 网站配置
    WebMaxBodySizeM      int64
    WebReadTimeout  time.Duration  // 读取超时时间
    WebWriteTimeout  time.Duration // 写入超时时间

    // 服务配置
    DebugAddr   string  // 调试监听地址 metrics查看端口
    HttpAddr    string  // http服务监听地址

    HttpsInfo    [3]string  // http服务监听地址 [Addr, cert, key]

    GrpcAddr    string  // grpc服务监听地址

    // 服务发现配置
    EtcdAddr    string
    EtcdTimeout time.Duration
    EtcdKeepAlive time.Duration
    EtcdRetryTimes int
    EtcdRetryTimeout time.Duration

    // 链路跟踪配置
    ZipkinUrl   string  // zipkin URL

    // 限流
    LimitServer *rate.Limiter       // 限流，超过直接拒绝
    LimitClient *rate.Limiter       // 限流，拒绝延时等待

    // 熔断
    BreakerServer gobreaker.Settings      // 熔断配置
    BreakerClient gobreaker.Settings      // 熔断配置

    // MYSQL
    MySQLConf  map[string]MysqlConfigs
    // REDIS
    RedisConf  map[string]RedisConf

}

var systemConf = DefaultConf
func Set(c Config){
    systemConf = c
}
func Get() Config {
    return systemConf
}


type RedisConf struct{
    RedisType string
    Network string
    Addr    string
    Password  string
    DB      int
    MasterName    string

    DialTimeout time.Duration  // 连接超时时间
    ReadTimeout time.Duration   // 读超时时间
    WriteTimeout  time.Duration // 写超时时间
    ///// 连接池配置
    PoolSize int       // 连接池容量
    MinIdleConns int   // 闲置连接数量
    IdleTimeout time.Duration // 空闲持续时间 默认5分钟
}

type MysqlConf struct {
    Driver  string
    User    string
    Pwd     string
    Host    string
    Port    int64
    Database  string
    Charset string
    ///// 连接池配置
    MaxOpenConns int
    MaxIdleConns int
    ConnMaxLifetime time.Duration
}

type MysqlConfigs struct {
    W MysqlConf
    R MysqlConf
}