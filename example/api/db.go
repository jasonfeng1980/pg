package api

import (
    "context"
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/database/rdb"
    "github.com/jasonfeng1980/pg/util"
)

// 通过初始化，注册ping关系
func init() {
    api := pg.MicroApi()
    api.Register("POST", "db", "v1", "mysql", dbMysql)
    api.Register("POST", "db", "v1", "mongo", dbMongo)
    api.Register("GET", "db", "v1", "redis", dbRedis)
}

//select *
//from information_schema.`TABLES`
//where TABLE_ROWS > 1
//order by TABLE_ROWS desc
//limit 1, 2
func dbMysql(ctx context.Context, params map[string]interface{})(interface{}, int64, string) {
    db, err := pg.MySQL.Get("DEMO")
    if err != nil {
        return pg.Err(err)
    }
    ret, err := db.Select("*").
        From("information_schema.TABLES").
        Where("TABLE_ROWS > ?", 10).
        Where(pg.M{"AVG_ROW_LENGTH": 0}).
        OrderBy("TABLE_ROWS  desc").
        Limit(1, 1).
        Cache(true).
        Query().
        Array()
    if err != nil {
        return pg.Err(err)
    }
    pg.D(ret)
    return pg.Suc(ret)
}
func dbMongo(ctx context.Context, params map[string]interface{})(interface{}, int64, string) {
    mdb, err := pg.Mongo.Get("USER")
    if err != nil {
        return pg.ErrCode(15003, "提供的mongoDB配置不存在")
    }
    ret, err := mdb.Select("create_at, info").
        From("user").
        Where("info.name", pg.M{"$lte":"王二"}).
        OrderBy("create_at desc").
        Limit(1, 2).
        Query().
        Array()
    if err != nil {
        return pg.Err(err)
    }
    pg.D(ret)
    return pg.Suc(ret)
}
func dbRedis(ctx context.Context, params map[string]interface{})(interface{}, int64, string) {
    rClient, err := pg.Redis.Client("DEMO")
    if err != nil {
        pg.Err(err)
    }
    key := rdb.String{
        Key: rdb.Key{
            CTX: context.Background(),
            Name: "test",
            Client: rClient,
        },
    }
    ret, err := key.Get()
    if err != nil {
        return pg.Err(err)
    }
    num, _ := util.Int64Parse(ret)
    key.Set( num + 1)
    pg.D(ret)
    return pg.Suc(ret)
}