package pg

import (
    "errors"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "regexp"
    "time"
)

////////////////////////////////
//
//  通过YAML 获取某个server的配置
////////////////////////////////
func yamlToServer(filePath string, mysqlList map[string]conf.MysqlConfigs, redisList map[string]conf.RedisConf) (*conf.Config, error){
    var serverConf *conf.Config
    var m map[string]interface{}
    e := util.YamlRead(filePath, &m)
    if e != nil {
        return nil, e
    }
    serverConf, e = serverMapConf(m)
    if e != nil {
        return nil, e
    }

    // 选择mysql
    mysqlConf , err := serverSetMysql(m["MySQL"], mysqlList)
    if err != nil {
        return nil, err
    }
    serverConf.MySQLConf = mysqlConf

    // 选择redis
    redisConf , err := serverSetRedis(m["Redis"], redisList)
    if err != nil  {
        return nil, err
    }
    serverConf.RedisConf = redisConf

    // 添加默认的限流 和 熔断
    serverConf.LimitServer = conf.DefaultConf.LimitServer
    serverConf.LimitClient = conf.DefaultConf.LimitClient
    serverConf.BreakerServer = conf.DefaultConf.BreakerServer
    serverConf.BreakerClient = conf.DefaultConf.BreakerClient

    return serverConf, nil
}

func serverSetMysql(m interface{}, mysqlList map[string]conf.MysqlConfigs) (map[string]conf.MysqlConfigs, error){
    if m == nil {
        return nil, nil
    }
    var mysqlConf = make(map[string]conf.MysqlConfigs)

    if mL, ok := m.([]interface{}); ok{
        for _, v := range mL {
            if mS, err := util.Str(v); err == nil {
                if mysqlV, mysqlOk:= mysqlList[mS]; mysqlOk{
                    mysqlConf[mS] = mysqlV
                } else {
                    return nil, ecode.ConfYamlWrongMysql.Error()
                }
            }
        }

    }
    return mysqlConf, nil
}

func serverSetRedis(m interface{}, redisList map[string]conf.RedisConf) (map[string]conf.RedisConf, error){
    if m == nil {
        return nil, nil
    }
    var redisConf = make(map[string]conf.RedisConf)
    if mL, ok := m.([]interface{}); ok{
        for _, v := range mL {
            if mS, err := util.Str(v); err == nil {
                if redisV, redisOk:= redisList[mS]; redisOk{
                    redisConf[mS] = redisV
                } else {
                    return nil, ecode.ConfYamlWrongRedis.Error()
                }
            }
        }

    }
    return redisConf, nil
}

func serverMapConf(m map[string]interface{}) (*conf.Config, error){
    var (
        errList []error
        HttpsInfo [3]string
    )

    if m["HttpsInfo"] != nil {
        v, ok := m["HttpsInfo"].([]interface{})
        if !ok || len(v) != 3{
            return nil, ecode.ConfYamlWrongFormat.Error("HttpsInfo")
        }
        HttpsInfo = [3]string{
            v[0].(string), v[1].(string), v[2].(string),
        }
    }

    var server = &conf.Config{
        ServerName: util.StrHideErr(errList, m["ServerName"]),
        ServerNo:   util.StrHideErr(errList, m["ServerNo"]),

        // 日志
        LogDir:      util.StrHideErr(errList, m["LogDir"]),        // 日志文件夹 成功 access.年月日.小时.log 失败 error.201012.13.log
        LogShowDebug: util.StrHideErr(errList, m["LogDir"]) == "true",     // 是否记录测试日志

        // 网站配置
        WebMaxBodySizeM:     util.Int64HideErr(errList, m["WebMaxBodySizeM"])<<20,   // 最大允许上传的大小
        WebReadTimeout:     time.Second*time.Duration(util.Int64HideErr(errList, m["WebReadTimeout"])),  // 读取超时时间
        WebWriteTimeout:    time.Second*time.Duration(util.Int64HideErr(errList, m["WebWriteTimeout"])), // 写入超时时间

        // 微服务配置
        DebugAddr:  util.StrHideErr(errList, m["DebugAddr"]),
        HttpAddr:   util.StrHideErr(errList, m["HttpAddr"]),
        HttpsInfo:  HttpsInfo,

        GrpcAddr:   util.StrHideErr(errList, m["GrpcAddr"]),

        // etcd
        EtcdAddr:   util.StrHideErr(errList, m["EtcdAddr"]),
        EtcdTimeout: time.Second * time.Duration(util.Int64HideErr(errList, m["EtcdTimeout"])),
        EtcdKeepAlive: time.Second * time.Duration(util.Int64HideErr(errList, m["EtcdKeepAlive"])),
        EtcdRetryTimes: util.IntHideErr(errList, m["EtcdRetryTimes"]),
        EtcdRetryTimeout: time.Second * 30,

        // 链路跟踪配置
        ZipkinUrl:  util.StrHideErr(errList, m["ZipkinUrl"]),
    }
    if len(errList) > 0 {
        return nil, ecode.ConfYamlWrongFormat.Error("")
    }
    return server, nil
}

////////////////////////////////
//
//  通过YAML 获取全局redis配置
////////////////////////////////
func yamlToRedis(filePath string) (map[string]conf.RedisConf, error){
    redisConf := make(map[string]conf.RedisConf)
    var m map[string]map[string]string
    e := util.YamlRead(filePath, &m)
    if e != nil {
        return nil, e
    }
    for n, line := range m {
        conf, err := RedisSetConf(line)
        if err != nil {
            return nil, err
        }
        redisConf[n] = *conf
    }
    return redisConf, nil
}

func RedisSetConf(confList map[string]string) (*conf.RedisConf, error){
    fields := []string{"RedisType", "Network", "Addr", "Password", "DB", "MasterName", "PoolSize", "MinIdleConns",
        "DialTimeout", "ReadTimeout", "WriteTimeout", "IdleTimeout"}
    e := errors.New("配置文件不完整")

    for _, v:=range fields {
        if _, ok:= confList[v]; !ok{
            return nil, e
        }
    }
    DB, _ := util.IntParse(confList["DB"])
    PoolSize, _ := util.IntParse(confList["PoolSize"])
    MinIdleConns, _ := util.IntParse(confList["MinIdleConns"])
    IdleTimeout, _ := util.Int64Parse(confList["IdleTimeout"])
    DialTimeout, _ := util.Int64Parse(confList["DialTimeout"])
    ReadTimeout, _ := util.Int64Parse(confList["ReadTimeout"])
    WriteTimeout, _ := util.Int64Parse(confList["WriteTimeout"])
    return &conf.RedisConf{
        RedisType: confList["RedisType"],
        Network: confList["Network"],
        Addr: confList["Addr"],
        Password: confList["Password"],
        DB: DB,
        MasterName: confList["MasterName"],
        DialTimeout: time.Duration(DialTimeout) * time.Second,  // 连接超时时间
        ReadTimeout: time.Duration(ReadTimeout) * time.Second,   // 读超时时间
        WriteTimeout: time.Duration(WriteTimeout) * time.Second, // 写超时时间
        ///// 连接池配置
        PoolSize: PoolSize,       // 连接池容量
        MinIdleConns: MinIdleConns,  // 闲置连接数量
        IdleTimeout: time.Duration(IdleTimeout) * time.Second, // 空闲持续时间 默认5分钟
    }, nil
}

////////////////////////////////
//
//  通过YAML 获取全局mysql配置
////////////////////////////////
func yamlToMysql(filePath string) (map[string]conf.MysqlConfigs, error){
    mysqlConf := make(map[string]conf.MysqlConfigs)
    var m map[string]map[string]string
    err := util.YamlRead(filePath, &m)
    if err != nil {
        return nil, err
    }
    // 循环赋值
    for n, line := range m {
        conf, err := mysqlMapConf(line)
        if err != nil {
            panic(err)
        }
        mysqlConf[n] = *conf
    }

    return mysqlConf, nil
}

func mysqlMapConf(confList map[string]string) (*conf.MysqlConfigs, error){
    w, ok := confList["W"]
    if  !ok {
        return nil, errors.New("缺少写库配置")
    }
    confW, err := mysqlSetConf(w)
    if err != nil {
        return nil, err
    }
    r, ok := confList["R"]
    if !ok {
        r = w
    }
    confR, err := mysqlSetConf(r)
    if err != nil {
        return nil, err
    }
    return &conf.MysqlConfigs{
        W: *confW,
        R: *confR,
    }, nil

}
func mysqlSetConf(s string) (*conf.MysqlConf, error){
    reg := regexp.MustCompile(`^(\w+):\/\/(\w+):(.*)@(\w+)\((\w+):(\d+)\)\/(\w+)\?charset=(\w+)&maxOpen=(\d+)&maxIdle=(\d+)&maxLifetime=(\d+)$`)

    m := reg.FindAllStringSubmatch(s, 1)
    if len(m)!=1 || len(m[0])!=12{
        return nil, errors.New("mysql配置不正确")
    }
    match := m[0]

    port, _ := util.Int64Parse(match[6])
    MaxOpenConns, _ := util.IntParse(match[9])
    MaxIdleConns, _ := util.IntParse(match[10])
    ConnMaxLifetime, _ := util.Int64Parse(match[11])

    return &conf.MysqlConf{
        Driver: match[1],
        User: match[2],
        Pwd: match[3],
        Host: match[5],
        Port: port,
        Database: match[7],
        Charset: match[8],
        MaxOpenConns: MaxOpenConns,
        MaxIdleConns: MaxIdleConns,
        ConnMaxLifetime: time.Duration(ConnMaxLifetime) * time.Second,
    }, nil
}
