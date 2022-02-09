package conf

import (
    "github.com/jasonfeng1980/pg/util"
    "github.com/sony/gobreaker"
    "golang.org/x/time/rate"
    "runtime"
    "time"
)

var DefaultServer =  Server{
    // 服务
    Name:  "pg",  		// 服务名称
    Num:   "01",  		// 服务序号
    Root:  "",  		// 项目根目录
    Env:   "dev",  		// 系统环境 dev test pre prod
    // 日志
    LogDir:      "log",  		// 日志文件夹
    LogLevel:	"info",			// 日志记录级别
    // 服务配置
    AddrDebug:   ":8081",  		// 调试监听地址 metrics查看端口
    AddrHttp:    ":80",  		// http服务监听地址
    //AddrHttps:   [3]string,  	    // https服务监听地址 [Addr, cert, key]
    AddrGrpc:    ":8082",  		// grpc服务监听地址

    ETCD:        "etcd://:@tcp(127.0.0.1:2379,127.0.0.1:2379)/?DialTimeout=3&KeepAlive=3&RetryTimes=3&RetryTimeout=30",          // 服务发现配置  e.g. etcd://:@tcp(127.0.0.1:2379,127.0.0.1:2379)/?DialTimeout=3&KeepAlive=3&RetryTimes=3&RetryTimeout=30
    ZipkinUrl:   "",          // 链路跟踪配置

    CacheRedis: "",			// 缓存redis配置名
    CacheSec:   60,			// 缓存时间-秒

    WebMaxBodySizeM:  32<<20,      // 最大允许上传的大小 32M
    WebReadTimeout:   10,  		// 读取超时时间
    WebWriteTimeout:  30, 		// 写入超时时间

    // 限流
    LimitServer:    rate.NewLimiter(rate.Limit(500*runtime.NumCPU()), 500*runtime.NumCPU()), // 限流，超过直接拒绝
    LimitClient:    rate.NewLimiter(rate.Limit(500*runtime.NumCPU()), 500*runtime.NumCPU()), // 限流，拒绝延时等待

    // 熔断
    BreakerServer: gobreaker.Settings{
        Name:    "Gobreaker-server",
        Timeout: time.Second * 10,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            if counts.TotalFailures > 100 || counts.ConsecutiveFailures > 10 {
                return true
            }
            return false
        },
        OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
            if to == gobreaker.StateOpen {
                util.Warn(name, "from", from, "to", to)
            } else {
                util.Warn(name, "from", from, "to", to)
            }
        },
    },
    BreakerClient: gobreaker.Settings{
        Name:    "Gobreaker-client",
        Timeout: 30 * time.Second,
    },
}


