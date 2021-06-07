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
        //Where("TABLE_ROWS > ?", 10).
        Where("TABLE_ROWS", pg.M{"$gte": 10}).
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
    myLog := pg.Log.Get("login")
    myLog.With("userId", 8888, "userName", "张三丰").Infoln("登录成功")
    return pg.Suc(ret)
}
func dbMongo(ctx context.Context, params map[string]interface{})(interface{}, int64, string) {
    mdb, err := pg.Mongo.Get("USER")
    if err != nil {
        return pg.ErrCode(15003, "提供的mongoDB配置不存在")
    }
    ret, err := mdb.Select("*").
        From("user").
        Where(pg.M{"info.name": pg.M{"$lte":"王二"}}).
        //Where("info.name", pg.M{"$lte":"王二"}). // 效果同上
        //GroupByMap(pg.M{
        //    "_id":"$token",
        //    "sum": bson.D{{"$sum", "$info.sex"}},
        //    "count": bson.D{{"$sum", 1}},
        //}).
        GroupBy("token").
        //Having(pg.M{"count": pg.M{"$gte": 16}}).
        Having("count", pg.M{"$gte": 16}). // 效果同上
        OrderBy("create_at desc").
        Limit(0, 20).
        Query().
        Array()
    if err != nil {
        return pg.Err(err)
    }
    pg.D(ret)
    return pg.Suc(ret)
}
func dbRedis(_ context.Context, params map[string]interface{})(interface{}, int64, string) {
    rClient, err := pg.Redis.Client("DEMO")
    if err != nil {
        pg.Err(err)
    }
    ctx := context.Background()
    UserName := func() rdb.String{
        return rdb.String{
            Key: rdb.Key{
                CTX: ctx,
                Name: "userName",
                Client: rClient,
            },
        }
    }
    UserInfo := func(userId int) rdb.Hash{
        return rdb.Hash{
            Key: rdb.Key{
                CTX:    ctx,
                Name:   "userInfo",
                Client: rClient,
            },
            Field:    util.StrParse(userId),
            JoinMode: []string{"age", "desc"},
        }
    }
    // 用这些KEY来操作, 可以大幅减少redis内存空间
    u := UserName()
    ui := UserInfo(888)
    u.Set("张三丰")
        // 只取JoinMode里的key对应的值，不存储KEY
    info, _ := ui.Encode(pg.M{"age":18, "desc":"备注", "xxx":"无关信息"})

    ui.HSet(info)
    retName, err := u.Get()
    tInfo, _ := ui.HGet()       // 18〡备注
    retInfo, _ := ui.Decode(tInfo) // map[string]string{"age": "18", "desc": "备注"}
    return pg.Suc(pg.M{
        "userName": retName,
        "userInfo":   retInfo,
    })
}