package rdb

import (
    "context"
    "github.com/go-redis/redis/v8"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "strings"
    "time"
)

type RedisConn struct {
   W *redis.Client
   R *redis.Client
}
type RedisConf struct{
    RedisType string
    Network string
    Addr    string
    Password  string
    DB      int

    DialTimeout time.Duration  // 连接超时时间
    ReadTimeout time.Duration   // 读超时时间
    WriteTimeout  time.Duration // 写超时时间
    ///// 连接池配置
    PoolSize int       // 连接池容量
    MinIdleConns int   // 闲置连接数量
    IdleTimeout time.Duration // 空闲持续时间 默认5分钟
}

type redisClass struct {
    log *util.Logger
    *redisPool
}

type redisPool struct {
    Client         map[string]*RedisConn
}

var CTX = context.Background()

//////////////////////////////////////////////
//
//  Redis 的连接 和连接池
//
//////////////////////////////////////////////
var Redis = &redisClass{
    log: util.Log(),
    redisPool: &redisPool{
        Client:         make(map[string]*RedisConn),
    },
}

func (r *redisClass)Conn(redisConfDnsMap map[string][]string){
    for  name, redisConfDns := range redisConfDnsMap{
        name = strings.ToUpper(name)
        redisConn := &RedisConn{}
        for k, dns := range redisConfDns {
            redisConf, err := r.UnmarshalDns(name, dns)
            if err != nil {
                r.log.Panic(err.Error())
            }
            conn, err := r.doConn(name, redisConf)
            if err != nil {
                r.log.Panic(err.Error())
            }
            if k == 0 {
                redisConn.W = conn
            } else {
                redisConn.R = conn
            }
        }
        if len(redisConfDns) == 1 {
            redisConn.R = redisConn.W
        }
        r.redisPool.Client[name] = redisConn
    }
}
func (r *redisClass)UnmarshalDns(name string, dns string) (ret *RedisConf, err error) {
        m, err := util.MapFromDns(dns)
        if  err !=nil {
            return nil, err
        }
        t := &RedisConf{
            RedisType: "CLIENT",
            Network:    m.GetStr("network", "tcp"),
            Addr:       m.GetStr("host", "localhost") + ":" + m.GetStr("port", "6379"),
            Password:   m.GetStr("password"),
            DB:         m.GetInt("db", 0),

            DialTimeout:m.GetTimeDuration("params.DialTimeout", 2) * time.Second,  // 连接超时时间
            ReadTimeout:m.GetTimeDuration("params.ReadTimeout", 2) * time.Second,  // 读超时时间
            WriteTimeout:m.GetTimeDuration("params.WriteTimeout", 2) * time.Second, // 写超时时间
            ///// 连接池配置
            PoolSize:   m.GetInt("params.PoolSize", 40),                     // 连接池容量
            MinIdleConns:m.GetInt("params.MinIdleConns", 10),                // 闲置连接数量
            IdleTimeout:m.GetTimeDuration("params.IdleTimeout", time.Second * 2),   // 空闲持续时间 默认5分钟
        }
        return t, nil
}

// 单机模式
func (r *redisClass)Client(name string)  (*RedisConn, error){
    name = strings.ToUpper(name)
    ret, ok := r.redisPool.Client[name]
    if !ok {
        return nil, ecode.DbWrongConfName.Error("redis-Client", name)
    }
    return ret, nil
}
// 关闭链接
func (r *redisClass)Close() error{
    for k, v := range r.redisPool.Client {
        if v.R != v.W {
            _ = v.R.Close()
        }
        _ = v.W.Close()
        r.log.S.Debugf("关闭REDIS-Client - 【别名 %s】 的链接", k)
    }
    return nil
}

func (r *redisClass)doConn(name string, conf *RedisConf) (*redis.Client, error){
    client := redis.NewClient(&redis.Options{
        //连接信息
        Network:  conf.Network,                  //网络类型，tcp or unix，默认tcp
        Addr:     conf.Addr, //主机名+冒号+端口，默认localhost:6379
        Password: conf.Password,                     //密码
        DB:       conf.DB,                      // redis数据库index
        //连接池容量及闲置连接数量
        PoolSize:     conf.PoolSize, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
        MinIdleConns: conf.MinIdleConns, //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。

        //钩子函数
        OnConnect: func(CTX context.Context, conn *redis.Conn) error { //仅当客户端执行命令时需要从连接池获取连接时，如果连接池需要新建连接时则会调用此钩子函数
            r.log.S.Debugf("连接Redis  %s(%v )   ----  成功", name, conn)
            return nil
        },
    })
    if _, err := client.Ping(CTX).Result(); err == nil { // 连接成功
        return client, nil
    } else { // 连接失败
        return client, err
    }
}