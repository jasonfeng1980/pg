package conf

import (
    "github.com/sony/gobreaker"
    "golang.org/x/time/rate"
    "runtime"
    "time"
)


var DefaultConf = Config{
    ServerName: "PG",
    ServerNo:   "01",

    // 日志
    LogDir:      "",        // 日志文件夹
    LogShowDebug: true,     // 是否记录测试日志

    // 网站配置
    WebMaxBodySizeM:     32<<20,   // 最大允许上传的大小
    WebReadTimeout:     time.Second*10,  // 读取超时时间
    WebWriteTimeout:    time.Second*30, // 写入超时时间

    // 微服务配置
    DebugAddr:  ":8080",
    HttpAddr:   ":80",
    //HttpsInfo:  [3]string{"","",""},   // http服务监听地址 [Addr, cert, key]

    GrpcAddr:   ":8082",

    // etcd
    EtcdAddr:   "127.0.0.1:2379",
    EtcdTimeout: time.Second * 3,
    EtcdKeepAlive: time.Second * 3,
    EtcdRetryTimes: 3,
    EtcdRetryTimeout: time.Second * 30,

    // 链路跟踪配置
    ZipkinUrl:  "http://localhost:9411/api/v2/spans",

    // 限流
    LimitServer:    rate.NewLimiter(rate.Limit(200*runtime.NumCPU()), 200*runtime.NumCPU()), // 限流，超过直接拒绝
    LimitClient:    rate.NewLimiter(rate.Limit(200*runtime.NumCPU()), 200*runtime.NumCPU()), // 限流，拒绝延时等待

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
            //if to == gobreaker.StateOpen {
            //   log.With(log.NewNopLogger(), "type", "warnning", "from", name, "to", to)
            //} else {
            //   log.With(log.NewNopLogger(),"from", name, "to", to)
            //}
        },
    },
    BreakerClient: gobreaker.Settings{
        Name:    "Gobreaker-client",
        Timeout: 30 * time.Second,
    },
}