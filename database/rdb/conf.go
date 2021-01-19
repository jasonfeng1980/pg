package rdb

import (
    "context"
    "github.com/go-redis/redis/v8"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/util"
    "strings"
)

type redisClass struct {
    log util.Logger
    *redisPool
}

type redisPool struct {
    Client         map[string]*redis.Client
    SentinelClient map[string]*redis.Client
    ClusterClient  map[string]*redis.ClusterClient
}

var CTX = context.Background()

//////////////////////////////////////////////
//
//  Redis 的连接 和连接池
//
//////////////////////////////////////////////
var Redis = &redisClass{
    log: util.LogHandle(""),
    redisPool: &redisPool{
        Client:         make(map[string]*redis.Client),
        SentinelClient: make(map[string]*redis.Client),
        ClusterClient:  make(map[string]*redis.ClusterClient),
    },
}

//



func (r *redisClass)Conn(redisConf map[string]conf.RedisConf){
    r.initPool(redisConf)
}
// 单机模式
func (r *redisClass)Client(name string)  *redis.Client{
    name = strings.ToUpper(name)
    return r.redisPool.Client[name]
}
// 哨兵模式
func (r *redisClass)SentinelClient(name string) *redis.Client{
    name = strings.ToUpper(name)
    return r.redisPool.SentinelClient[name]
}
// 集群模式
func (r *redisClass)ClusterClient(name string) *redis.ClusterClient {
    name = strings.ToUpper(name)
    return r.redisPool.ClusterClient[name]
}
// 关闭链接
func (r *redisClass)Close(){
    for k, v := range r.redisPool.Client {
        _ = v.Close()
        r.log.Logf("关闭REDIS-Client - %s 的链接", k)
    }
    for k, v := range r.redisPool.SentinelClient {
        _ = v.Close()
        r.log.Logf("关闭REDIS-SentinelClient - %s 的链接", k)
    }
    for k, v := range r.redisPool.ClusterClient {
        _ = v.Close()
        r.log.Logf("关闭REDIS-ClusterClient - %s 的链接", k)
    }

}

func (r *redisClass)initPool(confArr map[string]conf.RedisConf){
    for name, conf := range confArr {
        newName := strings.ToUpper(name)
        switch strings.ToUpper(conf.RedisType) {
        case "CLIENT":
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
                    r.log.Logf("连接Redis  %s(%v )   ----  成功", name, conn)
                    return nil
                },
            })
            if _, err := client.Ping(CTX).Result(); err == nil {
                r.redisPool.Client[newName] = client
            } else { // 连接失败
                panic("链接redis-client <" + newName +">失败:  "  + err.Error())
            }
        case "SENTINEL":
            sentinelClient := redis.NewFailoverClient(&redis.FailoverOptions{
                //连接信息
                MasterName: conf.MasterName,
                SentinelAddrs: strings.Split(conf.Addr, ","),
                SentinelPassword: conf.Password,
                //Network:  conf.Network,                  //网络类型，tcp or unix，默认tcp
                //Addr:     conf.Addr, //主机名+冒号+端口，默认localhost:6379

                Password: conf.Password,                     //密码
                DB:       conf.DB,                      // redis数据库index
                //连接池容量及闲置连接数量
                PoolSize:     conf.PoolSize, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
                MinIdleConns: conf.MinIdleConns, //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。

                //钩子函数
                OnConnect: func(CTX context.Context, conn *redis.Conn) error { //仅当客户端执行命令时需要从连接池获取连接时，如果连接池需要新建连接时则会调用此钩子函数
                    r.log.Logf("开启新的链接conn=%v", conn)
                    return nil
                },
            })
            if _, err := sentinelClient.Ping(CTX).Result(); err == nil {
                r.redisPool.SentinelClient[newName] = sentinelClient
            } else { // 连接失败
                panic("链接redis-sentinelClient <" + newName +">失败:  "  + err.Error())
            }
        case "CLUSTER":
            clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
                //连接信息
                Addrs:    strings.Split(conf.Addr, ","), //主机名+冒号+端口，默认localhost:6379

                Password: conf.Password,                     //密码
                //连接池容量及闲置连接数量
                PoolSize:     conf.PoolSize, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
                MinIdleConns: conf.MinIdleConns, //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。

                //钩子函数
                OnConnect: func(CTX context.Context, conn *redis.Conn) error { //仅当客户端执行命令时需要从连接池获取连接时，如果连接池需要新建连接时则会调用此钩子函数
                    r.log.Logf("开启新的链接conn=%v", conn)
                    return nil
                },
            })
            if _, err := clusterClient.Ping(CTX).Result(); err == nil {
                r.redisPool.ClusterClient[newName] = clusterClient
            } else { // 连接失败
                panic("链接redis-clusterClient <" + newName +">失败:  "  + err.Error())
            }
        default:

        }
    }
}