package conf

import (
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "regexp"
    "time"
)

type YamlConf struct {
    Root   string
    ServerConf *Config
    tmpConf map[string][]string
    Err    error
}
func (y *YamlConf)Set(){
    Set(*y.ServerConf)
}

// 读取YAML文件
func (y *YamlConf)YamlRead(filePath string, m interface{}) error{
    if filePath[:1]!="/" {
        filePath = y.Root + filePath
    }
    return util.YamlRead(filePath, m)
}

////////////////////////////////
//
//  通过YAML 获取server的配置
////////////////////////////////
func (y *YamlConf)Server(filePath string) *YamlConf{
    var serverConf *Config
    var m map[string]interface{}
    e := y.YamlRead(filePath, &m)
    if e != nil {
        panic(e)
        return y
    }
    serverConf, e = y.serverMapConf(m)
    if e != nil {
        y.Err = e
        return y
    }

    // 添加默认的限流 和 熔断
    serverConf.ServerRoot = util.FileRealPath(y.Root)
    serverConf.LimitServer = DefaultConf.LimitServer
    serverConf.LimitClient = DefaultConf.LimitClient
    serverConf.BreakerServer = DefaultConf.BreakerServer
    serverConf.BreakerClient = DefaultConf.BreakerClient

    // 添加用到的mysql，mongo，redis， rabbitmq
    y.tmpConf = make(map[string][]string)
    if v, ok := m["MySQL"]; ok {
        y.tmpConf["MySQL"] = util.ListInterfaceToStr(v.([]interface{}))
    }
    if v, ok := m["Mongo"]; ok {
        y.tmpConf["Mongo"] = util.ListInterfaceToStr(v.([]interface{}))
    }
    if v, ok := m["Redis"]; ok {
        y.tmpConf["Redis"] = util.ListInterfaceToStr(v.([]interface{}))
    }
    if v, ok := m["RabbitMQ"]; ok {
        y.tmpConf["RabbitMQ"] = util.ListInterfaceToStr(v.([]interface{}))
    }

    y.ServerConf = serverConf

    // 修改默认日志
    if y.ServerConf.LogDir == "<nil>" {
        y.ServerConf.LogDir = ""
    }
    if y.ServerConf.LogDir != ""  {
        if y.ServerConf.LogDir[:1] != "/" {
            y.ServerConf.LogDir = y.ServerConf.ServerRoot + "/" + y.ServerConf.LogDir
        }
    }
    util.LogInit(y.ServerConf.LogDir, y.ServerConf.LogShowDebug, y.ServerConf.ServerName)

    return y
}
func (y *YamlConf)serverMapConf(m map[string]interface{}) (*Config, error){
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

    var server = &Config{
        ServerName: util.StrParse(m["ServerName"]),
        ServerNo:   util.StrParse(m["ServerNo"]),

        // 日志
        LogDir:      util.StrParse(m["LogDir"]),        // 日志文件夹 成功 access.年月日.小时.log 失败 error.201012.13.log
        LogShowDebug: util.StrParse(m["LogDebug"]) == "true",     // 是否记录测试日志

        // 网站配置
        WebMaxBodySizeM:     util.Int64HideErr(errList, m["WebMaxBodySizeM"])<<20,   // 最大允许上传的大小
        WebReadTimeout:     time.Second*time.Duration(util.Int64HideErr(errList, m["WebReadTimeout"])),  // 读取超时时间
        WebWriteTimeout:    time.Second*time.Duration(util.Int64HideErr(errList, m["WebWriteTimeout"])), // 写入超时时间

        // 微服务配置
        DebugAddr:  util.StrParse(m["DebugAddr"]),
        HttpAddr:   util.StrParse(m["HttpAddr"]),
        HttpsInfo:  HttpsInfo,

        GrpcAddr:   util.StrParse(m["GrpcAddr"]),

        // etcd
        EtcdAddr:   util.StrParse(m["EtcdAddr"]),
        EtcdTimeout: time.Second * time.Duration(util.Int64HideErr(errList, m["EtcdTimeout"])),
        EtcdKeepAlive: time.Second * time.Duration(util.Int64HideErr(errList, m["EtcdKeepAlive"])),
        EtcdRetryTimes: util.IntHideErr(errList, m["EtcdRetryTimes"]),
        EtcdRetryTimeout: time.Second * 30,

        // 链路跟踪配置
        ZipkinUrl:  util.StrParse(m["ZipkinUrl"]),

        // 缓存redis 别名
        CacheRedis: util.StrParse(m["CacheRedis"]),
        CacheSec: util.Int64HideErr(errList, m["CacheSec"]),
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
func (y *YamlConf)Redis(filePath string) *YamlConf{
    if y.Err != nil {
        return y
    }
    if _, ok := y.tmpConf["Redis"]; !ok { // 没有server配置 就返回
        return y
    }
    y.ServerConf.RedisConf = make(map[string]RedisConf)
    var m map[string]map[string]string
    err := y.YamlRead(filePath, &m)
    if err != nil {
        y.Err = err
        return y
    }
    for n, line := range m {
        if !util.ListHave(y.tmpConf["Redis"], n) {
            continue
        }
        conf, err := y.redisSetConf(line)
        if err != nil {
            y.Err = err
            return y
        }
        y.ServerConf.RedisConf[n] = *conf
    }
    return y
}

func (y *YamlConf)redisSetConf(confList map[string]string) (*RedisConf, error){
    fields := []string{"RedisType", "Network", "Addr", "Password", "DB", "MasterName", "PoolSize", "MinIdleConns",
        "DialTimeout", "ReadTimeout", "WriteTimeout", "IdleTimeout"}
    e := ecode.ConfNotComplete.Error("Redis")

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
    return &RedisConf{
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
//  通过YAML 获取全局mongo配置
////////////////////////////////
func (y *YamlConf)Mongo(filePath string) *YamlConf {
    if y.Err != nil {
        return y
    }
    if _, ok := y.tmpConf["Mongo"]; !ok { // 没有server配置 就返回
        return y
    }
    y.ServerConf.MongoConf = make(map[string]MongoConf)
    var m map[string]map[string]string
    err := y.YamlRead(filePath, &m)
    if err != nil {
        y.Err = err
        return y
    }
    for n, line := range m {
        if !util.ListHave(y.tmpConf["Mongo"], n) {
            continue
        }
        conf, err := y.mongoSetConf(line)
        if err != nil {
            y.Err = err
            return y
        }
        y.ServerConf.MongoConf[n] = *conf
    }
    return y
}
func (y *YamlConf)mongoSetConf(confList map[string]string) (*MongoConf, error){
    fields := []string{"Dns", "Timeout", "Database", "AllowDiskUse"}
    e := ecode.ConfNotComplete.Error("Mongo")

    for _, v:=range fields {
        if _, ok:= confList[v]; !ok{
            return nil, e
        }
    }
    Dns, _ := confList["Dns"]
    Timeout, _ := util.Int64Parse(confList["Timeout"])
    Database, _ := confList["Database"]
    return &MongoConf{
        Dns: Dns,
        Timeout: time.Duration(Timeout) * time.Second,
        Database: Database,
        AllowDiskUse: confList["AllowDiskUse"] == "1",
    }, nil
}
////////////////////////////////
//
//  通过YAML 获取全局rabbitmq配置
////////////////////////////////
func (y *YamlConf) Rabbitmq(filePath string) *YamlConf {
    if y.Err != nil {
        util.Log.Error(y.Err)
    }
    if _, ok := y.tmpConf["RabbitMQ"]; !ok { // 没有server配置 就返回
        return y
    }
    m := make(map[string]RabbitMQConf)
    err := y.YamlRead(filePath, &m)
    if err != nil {
        y.Err = err
        return y
    }
    y.ServerConf.RabbitMQConf = make(map[string]RabbitMQConf)
    for n, c := range m {
        if !util.ListHave(y.tmpConf["RabbitMQ"], n) {
            continue
        }
        y.ServerConf.RabbitMQConf[n] = c
    }
    return y
}
////////////////////////////////
//
//  通过YAML 获取全局mysql配置
////////////////////////////////
func (y *YamlConf)Mysql(filePath string) *YamlConf {
    if y.Err!=nil {
        return y
    }
    if _, ok := y.tmpConf["MySQL"]; !ok { // 没有server配置 就返回
        return y
    }
    y.ServerConf.MySQLConf = make(map[string]MysqlConfigs)
    var m map[string]map[string]string
    err := y.YamlRead(filePath, &m)
    if err != nil {
        y.Err = err
        return y
    }
    // 循环赋值
    for n, line := range m {
        if !util.ListHave(y.tmpConf["MySQL"], n) {
            continue
        }
        conf, err := y.mysqlMapConf(line)
        if err != nil {
            panic(err)
        }
        y.ServerConf.MySQLConf[n] = *conf
    }
    return y
}
func (y *YamlConf)mysqlMapConf(confList map[string]string) (*MysqlConfigs, error){
    w, ok := confList["W"]
    if  !ok {
        return nil, ecode.ConfMissWrite.Error("mysql")
    }
    confW, err := y.mysqlSetConf(w)
    if err != nil {
        return nil, err
    }
    r, ok := confList["R"]
    if !ok {
        r = w
    }
    confR, err := y.mysqlSetConf(r)
    if err != nil {
        return nil, err
    }
    return &MysqlConfigs{
        W: *confW,
        R: *confR,
    }, nil

}
func (y *YamlConf)mysqlSetConf(s string) (*MysqlConf, error){
    reg := regexp.MustCompile(`^(\w+):\/\/(.+):(.*)@(\w+)\((.+):(\d+)\)\/(.+)\?charset=(\w+)&maxOpen=(\d+)&maxIdle=(\d+)&maxLifetime=(\d+)$`)
    m := reg.FindAllStringSubmatch(s, 1)

    if len(m)!=1 || len(m[0])!=12{
        return nil, ecode.ConfWrong.Error("mysql")
    }
    match := m[0]

    port, _ := util.Int64Parse(match[6])
    MaxOpenConns, _ := util.IntParse(match[9])
    MaxIdleConns, _ := util.IntParse(match[10])
    ConnMaxLifetime, _ := util.Int64Parse(match[11])

    return &MysqlConf{
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
