package mdb

import (
    "context"
    "github.com/jasonfeng1980/pg/ecode"

    "github.com/go-redis/redis/v8"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/util"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "time"

)

var MONGO = &Mongo{
    log: util.Log.Logger,
}

type Mongo struct {
    ctx context.Context
    log  *util.Logger
    confList map[string]conf.MongoConf //
    Pool map[string]*mongo.Client
    CacheRedisClient *redis.Client // 缓存redis
    CacheExpr time.Duration // 缓存时间
}

func (m *Mongo)Conn(mdbConf map[string]conf.MongoConf){
    m.confList = mdbConf
    m.Pool = make(map[string]*mongo.Client, len(mdbConf))
    for _, conf := range mdbConf {
        if _, ok := m.Pool[conf.Dns]; ok {
            continue
        }
        m.ctx = context.Background()
        clientOptions := options.Client().ApplyURI(conf.Dns)
        client, mdbErr := mongo.Connect(m.ctx, clientOptions)
        if mdbErr != nil {
            panic("Mongo创建连接失败: " + mdbErr.Error())
        }
        m.Pool[conf.Dns] = client
        m.log.Debugf("连接MONGO   (dns:%s,database: %s ) ----  成功", conf.Dns, conf.Database)
    }
}

// 获取新的执行QUERY
func (m *Mongo)Get(name string)(*Query, error){
    if conf, ok := m.confList[name]; ok {
        client := m.Pool[conf.Dns]
        err := client.Ping(m.ctx, nil)
        if err != nil {
            return nil, ecode.MdbPingErr.Error(err.Error())
        }
        return &Query{
            Conf: conf,
            Conn: client.Database(conf.Database),
        }, nil
    } else {
        return nil, nil
    }
}

func (m *Mongo)Close(){
    for dns, client := range m.Pool {
        m.log.Debugf("关闭Mongo - 【%s】的链接", dns)
        client.Disconnect(m.ctx)
    }
}
