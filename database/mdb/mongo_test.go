package mdb

import (
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/util"
    "go.mongodb.org/mongo-driver/bson"
    "testing"
)

var (
	mongoConf = map[string][]string{
        "USER": []string{
            "mongodb://admin:root@tpc(localhost:27017)/demo?Timeout=3&AllowDiskUse=0",
        },
    }
    user *Query
    pingErr error
)
func TestMain(m *testing.M) {

    // 链接MYSQL连接池
    MONGO.Conn(mongoConf)
    defer MONGO.Close()

    user, pingErr = MONGO.Get("USER")
    if pingErr != nil {
        util.Error("无法链接mongo")
    }
    m.Run()
}

func TestQuery_Select(t *testing.T) {
    num, err := user.Select().
        From("user").
        Where("token", "111111").
        Where("info.sex",pg.M{"$gt":4000}).
        //Where(`{"info.sex":{"$gt":2000}}`).
        //GroupBy("info.name").
        //Where(pg.M{"info.sex": pg.M{
        //"$gt": 3000,
        //}}).
        //OrderBy("info.sex, info.name asc").
        //Limit(1, 5).
        Count()
    util.Debug(num, "err", err)

    ret, err := user.Select().
        From("user").
        Where("token", "111111").
        Where("info.sex",pg.M{"$gt":4000}).
        //Where(`{"info.sex":{"$gt":2000}}`).
        //GroupBy("info.name").
        //Where(pg.M{"info.sex": pg.M{
        //"$gt": 3000,
        //}}).
        OrderBy("info.sex, info.name asc").
        Limit(1, 5).
        //Count().
        Query().
        Array()
    for k, v := range ret {
        info := v["info"].(map[string]interface{})
        util.Debug(k, "token:", v["token"], "sex:", info["sex"], "name:", info["name"])
    }


}

func TestQuery_GroupBy(t *testing.T) {
    ret, err := user.Select().
        From("user").
        Where(bson.M{"token":"111111", "password":"bbbb"}).
        Where("info.sex",bson.M{"$gte":0}).
        //Where(`{"info.sex":{"$gt":2000}}`).
        //GroupByMap(bson.M{"_id": "$info.sex", "count": bson.M{"$sum":1}}).
        GroupBy(`{"_id":"$password", "count":{"$sum":1}}`).
        OrderBy("_id").
        //Having("_id", bson.M{"$gte": "0000"}).
        Limit(0, 8).
        Query().
        Array()
    if err != nil {
        t.Error(err)
    }
    for k, v := range ret {
        util.Debug(k, "v", v)
    }
}

func TestQuery_Insert(t *testing.T) {
    data := []bson.M{
        bson.M{
            "account": "aaa",
            "password": "eeee",
            "token": "5555",
            "info": bson.M{
                "name" : "张三",
                "sex" : 1,
                "email" : "jasonfeng1@gmail.com",
                "mobile" : "222222222",
            },
        },bson.M{
            "account": "bbbb",
            "password": "反反复复",
            "token": "6666",
            "info": bson.M{
                "name" : "李四",
                "sex" : 2,
                "email" : "jasonfeng1@sina.com",
                "mobile" : "111",
            },
        },
    }
    ret, err := user.Insert(data). // 单个多个都可以
        Into("user").
        Query().
        LastInsertId()
    util.Debug(ret, "err", err)

}

func TestQuery_Update(t *testing.T) {
    update := bson.D{
        {"$push", bson.M{"list2": "a"}},
        {"$push", bson.M{"list1": bson.M{"$each": []interface{}{1,2,3,4,5,6}}}},
    }
    ret, err := user.Update("user").
        Set(update).
        Where("password", "aaaa", "token", "3333").
        //Upsert(true).
        Query().
        RowsAffected()
    util.Debug(ret, "err", err)
}

func TestQuery_Delete(t *testing.T) {
    ret, err := user.Delete().
        From("user").
        Where("password", "aaaa").
        One(true).
        Query().
        RowsAffected()
    util.Debug(ret, "err", err)
}