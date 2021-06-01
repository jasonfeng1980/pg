package db

import (
    "database/sql"
    "fmt"
    "github.com/go-redis/redis/v8"
    _ "github.com/go-sql-driver/mysql"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/util"
    "strings"
    "time"
)


var MYSQL = &mySQL{
    log: util.Log.Logger,
}

type mySQL struct {
    pool map[string]*Conn  // 连接池
    log  *util.Logger
    openConn map[string]*sql.DB // 已经打开的MYSQL链接
    CacheRedisClient *redis.Client // 缓存redis
    CacheExpr time.Duration // 缓存时间
}

type expr struct {
    S string
}
// update时 不需要转义的 字符串
func Expr(s string) expr{
    return expr{
        S: s,
    }
}

// 获取mysql连接池的集合
func (m *mySQL)Conn(dbConf map[string]conf.MysqlConfigs){
    // 创建MYSQL 连接池
    m.InitPool(dbConf)
}

func (m *mySQL)SetCacheRedis(client *redis.Client, expr time.Duration){
    m.CacheRedisClient = client
    m.CacheExpr = expr
}

func (m *mySQL)GetPool(confName string) *Conn{
    return m.pool[confName]
}

// mysql 链接的类， 包含执行方法
type Conn struct {
    Conf  *conf.MysqlConfigs // 配置文件
    Reader *sql.DB  // 读句柄
    Writer *sql.DB  // 写句柄
    CacheRedis *redis.Client    // 缓存REDIS
}
// 开启一个新的查询
func (m *Conn)new() *Query{
    query := &Query{
        db: m,
    }
    return query
}



//////////////////////////////////////////////
//
//  MySQL 的方法
//
//////////////////////////////////////////////
// 根据配置，链接数据库，初始化连接池
func (m *mySQL)InitPool(dbConf map[string]conf.MysqlConfigs){
    m.pool = make(map[string]*Conn)
    m.openConn = make(map[string]*sql.DB)
    // 循环配置，建立连接池
    for name, conf := range dbConf {
        m.pool[name] = &Conn{
            Conf:   &conf,
            Writer: m.conn(conf.W),
            Reader: m.conn(conf.R),
        }
    }
}

// 获取mysql conn指针
func (m *mySQL)getConn(name string) (conn *Conn, ok bool){
    conn, ok = m.pool[name]
    return
}

// 获取新的执行QUERY
func (m *mySQL)Get(name string)(*Query, bool){
    name = strings.ToUpper(name)
    conn, ok := m.getConn(name)
    if !ok {
        return nil, false
    } else {
        return conn.new(), true
    }
}

// 关闭读写的链接
func (m *mySQL)Close(){
    for k, v := range m.openConn {
        m.log.Infof("关闭mysql - %s 的链接", k)
        v.Close()
    }
}

// 根据配置，连接数据库
func (m *mySQL)conn(conf conf.MysqlConf) (db *sql.DB) {
    dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
        conf.User, conf.Pwd, conf.Host, conf.Port, conf.Database, conf.Charset)
    if cache, ok := m.openConn[dbDSN]; ok { // 缓存，防止重复链接
        return cache
    }

    // 打开连接失败
    db, dbErr := sql.Open("mysql", dbDSN)
    if nil != dbErr {
        panic("MYSQL创建连接失败: " + dbErr.Error())
    }
    // 最大连接数
    db.SetMaxOpenConns(conf.MaxOpenConns)
    // 闲置连接数
    db.SetMaxIdleConns(conf.MaxIdleConns)
    // 最大连接周期
    db.SetConnMaxLifetime(conf.ConnMaxLifetime)
    // 缓存链接
    m.openConn[dbDSN] = db
    m.log.Infof("连接MYSQL   (%s : %d )/%s ----  成功", conf.Host, conf.Port, conf.Database)
    return
}