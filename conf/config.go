package conf

import (
    "github.com/jasonfeng1980/pg/util"
    "github.com/sony/gobreaker"
    "github.com/spf13/viper"
    "golang.org/x/time/rate"
)

type Config struct {
    // 整体配置
    Server

    // MYSQL
    MySQL  map[string][]string	`json:"MySQL"`
    // Mongo
    Mongo  map[string][]string	`json:"Mongo"`
    // REDIS
    Redis  map[string][]string	`json:"Redis"`
    // RabbitMQ
    RabbitMQ map[string]string	`json:"RabbitMQ"`
}

type Server struct{
    // 服务
    Name  string  		// 服务名称
    Num   string  		// 服务序号
    Root  string  		// 项目根目录
    Env   string  		// 系统环境 dev test pre prod
    // 日志
    LogDir      string  		// 日志文件夹
    LogLevel	string			// 日志记录级别
    // 服务配置
    AddrDebug   string  		// 调试监听地址 metrics查看端口
    AddrHttp    string  		// http服务监听地址
    AddrHttps   [3]string  	    // https服务监听地址 [Addr, cert, key]
    AddrGrpc    string  		// grpc服务监听地址

    ETCD        string          // 服务发现配置
    ZipkinUrl   string          // 链路跟踪配置

    CacheRedis string			// 缓存redis配置名
    CacheSec   int64			// 缓存时间-秒

    WebMaxBodySizeM  int64      // 最大允许上传的大小
    WebReadTimeout  int64  		// 读取超时时间
    WebWriteTimeout  int64 		// 写入超时时间

    // 限流
    LimitServer *rate.Limiter       // 限流，超过直接拒绝
    LimitClient *rate.Limiter       // 限流，拒绝延时等待

    // 熔断
    BreakerServer gobreaker.Settings      // 熔断配置
    BreakerClient gobreaker.Settings      // 熔断配置
}

var Conf = &Config{Server:DefaultServer};

type Loader struct {
    *viper.Viper
}
// 独立的VIPER实例，防止冲突
var ConfBox = &Loader{viper.New()}

// 加载配置文件
func Load(configFile string) (error){
    ConfBox.SetConfigFile(configFile)
    if err := ConfBox.ReadInConfig(); err != nil {
        return err
    }
    // 给内部配置赋值
    if err := ConfBox.Unmarshal(&Conf); err !=nil {
        return err
    }

    // 修改root
    if Conf.Root == "" {
        Conf.Root = util.FileRealPath(".")
    } else if Conf.Root[:1] == "." {
        Conf.Root = util.FileRealPath(Conf.Root)
    }
    // 初始化日志
    if Conf.LogDir != "" && Conf.LogDir[:1] != "/" {
        Conf.LogDir = util.FileRootPath(Conf.LogDir, Conf.Root)
    }
    util.LogInit(Conf.LogDir, Conf.Name + Conf.Num, Conf.LogLevel)
    return nil
}