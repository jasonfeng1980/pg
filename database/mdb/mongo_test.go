package mdb

import (
    "fmt"
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/conf"
    "go.mongodb.org/mongo-driver/bson"
    "testing"
)

var mongoConf = make(map[string]conf.MongoConf)
var (
	user *Query
	pingErr error
)
func TestMain(m *testing.M) {
    mongoConf["USER"] = conf.MongoConf{
        Dns: "mongodb://admin:root@localhost:27017",
        Timeout: 3,
        Database: "user",
    }

    // 链接MYSQL连接池
    MONGO.Conn(mongoConf)
    defer MONGO.Close()

    user, pingErr = MONGO.Get("USER")
    if pingErr != nil {
        fmt.Println("无法链接mongo")
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
    fmt.Println(num, err)

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
        fmt.Println(k, "token:", v["token"], "sex:", info["sex"], "name:", info["name"])
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
        fmt.Println(k, v)
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
    fmt.Println(ret, err)

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
    fmt.Println(ret, err)
}

func TestQuery_Delete(t *testing.T) {
    ret, err := user.Delete().
        From("user").
        Where("password", "aaaa").
        One(true).
        Query().
        RowsAffected()
    fmt.Println(ret, err)
}