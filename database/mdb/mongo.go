package mdb

import (
    "context"
    "fmt"
    "github.com/jasonfeng1980/pg/database/rdb"
    "github.com/jasonfeng1980/pg/ecode"

    "github.com/jasonfeng1980/pg/util"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "time"
)

type MongoConf struct {
    Dns        string
    Timeout    time.Duration
    AllowDiskUse bool
    Database   string
}
type MongoConfigs struct {
    W MongoConf
    R MongoConf
}
type MongoClients struct {
    W *mongo.Client
    R *mongo.Client
}

var MONGO = &Mongo{
    log: util.Log(),
}

type Mongo struct {
    ctx context.Context
    log  *util.Logger
    confList map[string]*MongoConfigs //
    Pool map[string]*mongo.Client
    CacheRedisClient *rdb.RedisConn // 缓存redis
    CacheExpr time.Duration // 缓存时间
}

func (m *Mongo)Conn(mdbDnsList map[string][]string){
    l := len(mdbDnsList)
    m.confList = make(map[string]*MongoConfigs, l)
    // 循环配置，建立连接池
    for name, DNSList := range mdbDnsList {
        conf, err := m.UnmarshalDns(name, DNSList)
        if err !=nil {
            m.log.Panic(err.Error())
        }
        m.confList[name] = conf
    }
    //m.confList = mdbConf
    m.Pool = make(map[string]*mongo.Client)
    m.ctx = context.Background()
    for _, conf := range m.confList {
        m.conn(conf.W)
        m.conn(conf.R)
    }
}

func (m *Mongo) conn(conf MongoConf) {
    if _, ok := m.Pool[conf.Dns]; ok {
        return
    }
    clientOptions := options.Client().ApplyURI(conf.Dns)
    client, mdbErr := mongo.Connect(m.ctx, clientOptions)
    if mdbErr != nil {
        panic("Mongo创建连接失败: " + mdbErr.Error())
    }
    m.Pool[conf.Dns] = client
    m.log.S.Debugf("连接MONGO   (dns:%s,database: %s ) ----  成功", conf.Dns, conf.Database)
}

func (m *Mongo)UnmarshalDns(name string, l []string) (ret *MongoConfigs, err error) {
    ret = &MongoConfigs{}
    for k, dns := range l {
        m, err := util.MapFromDns(dns)
        if  err !=nil {
            return nil, err
        }
        t := MongoConf{
            //mongodb://admin:root@localhost:27017"
            Dns:       fmt.Sprintf("mongodb://%s:%s@%s:%s",
                m.GetStr("user", ""), m.GetStr("password", ""),
                m.GetStr("host", "127.0.0.1"), m.GetStr("port", "27017")),
            Timeout:     m.GetTimeDuration("params.Timeout", 3),
            AllowDiskUse:       m.GetInt("params.AllowDiskUse", 0) == 1,
            Database:   m.GetStr("dbname"),
        }
        if k== 0 {
            ret.W = t
        } else {
            ret.R = t
        }
    }
    if len(l) == 1 {
        ret.R = ret.W
    }
    return ret, nil
}

func (m *Mongo)SetCacheRedis(client *rdb.RedisConn, expr time.Duration){
    m.CacheRedisClient = client
    m.CacheExpr = expr
}

// 获取新的执行QUERY
func (m *Mongo)Get(name string)(*Query, error){
    if conf, ok := m.confList[name]; ok {
        clientW := m.Pool[conf.W.Dns]
        err := clientW.Ping(m.ctx, nil)
        if err != nil {
            return nil, ecode.MdbPingErr.Error(err.Error())
        }
        clientR := m.Pool[conf.W.Dns]
        err = clientR.Ping(m.ctx, nil)
        if err != nil {
            return nil, ecode.MdbPingErr.Error(err.Error())
        }
        return &Query{
            Conf: conf,
            ConnW: clientW.Database(conf.W.Database),
            ConnR: clientW.Database(conf.R.Database),
        }, nil
    } else {
        return nil, nil
    }
}

func (m *Mongo)Close() error{
    for dns, client := range m.Pool {
        m.log.S.Debugf("关闭Mongo - 【%s】的链接", dns)
        client.Disconnect(m.ctx)
    }
    return nil
}
