package db

import (
    "fmt"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/database/rdb"
    "github.com/jasonfeng1980/pg/util"
    "go.mongodb.org/mongo-driver/bson"
    "math/rand"
    "os"
    "testing"
    "time"
)

var testDB *Query

type H map[string]interface{}

func TestMain(m *testing.M) {
    mysqlConfMap := map[string][]string{
        "demo": []string{
            "mysql://root:@tcp(localhost:3306)/demo?charset=utf8mb4&maxOpen=200&maxIdle=100&maxLifetime=30",
        },
    }
    MYSQL.Conn(mysqlConfMap)

    //MYSQL.SetCacheRedis()
    conf.Conf.LogLevel = "debug"

    testDB, _ = MYSQL.Get("demo")

    redisConfMap := map[string][]string{
        "demo": []string{
            "redis://:@tcp(localhost:6379)/0",
        },
    }
    rdb.Redis.Conn(redisConfMap)
    client, _ := rdb.Redis.Client("demo")
    MYSQL.SetCacheRedis(client, 10)
    a, _ := MYSQL.Get("demo")
    tx := a.StartTransaction()
    rs, err := tx.Update("company").Set(H{"company_money":1}).Where("company_id", 1).Query().RowsAffected()
    fmt.Println(rs, err)
    tx.Rollback()
    //tx.Commit()
    //tx.StartTransaction()
    tx.Update("company").Set(H{"company_money":2}).Where("company_id", 1).Query().RowsAffected()
    //tx.Rollback()
    //m.Run()
}
func TestManyQuery(t *testing.T) {
    a, _ := MYSQL.Get("demo")
    b, _ := MYSQL.Get("demo")
    tmpA := a.Select("*").From("company").Where("company_id", 1)
    tmpB := b.Select("*").From("company").Where("company_id", 2)
    retA, _ := tmpA.Query().Array()
    retB, _ := tmpB.Query().Array()
    fmt.Println(retA)
    fmt.Println(retB)
    os.Exit(1)
}

func TestXA(t *testing.T) {
    xid := testDB.Xid()
    util.Debug(xid)
    xa := testDB.XA(xid)
    r, err := xa.XaRecover()
    fmt.Println( r, err)

    err = xa.XaStart()
    fmt.Println(1, err)
    line, err := xa.Update("company").
        Set(util.M{"company_money": "11211"}).
        Where("company_id =? ", 1).
        Query().
        RowsAffected()
    fmt.Println(2, line, err)
    err = xa.XaEnd()
    fmt.Println(3, err)
    err = xa.XaPrepare()
    fmt.Println(4, err)
    xa.XaCommit()
}

func TestSelect(t *testing.T) {
    rs, err := testDB.Select("TABLE_SCHEMA,TABLE_NAME,TABLE_ROWS").
        From("information_schema.TABLES").
        Where("TABLE_ROWS", 1).
        Where("TABLE_ROWS", 1,2,3,4).
        Where("TABLE_ROWS", []int{1,2,3}).
        Where("TABLE_ROWS > ? and 1 < ?", 10, 3).
        Where("TABLE_ROWS", bson.M{"$gte": 9}).
        Where("AVG_ROW_LENGTH", util.M{"$in": []int{0,1,2}}).
        Where(util.M{"AVG_ROW_LENGTH": 111}).
        OrderBy("TABLE_ROWS  desc").
        Limit(2, 1).
        Cache(true).
        Query().
        Array()
    if err != nil {
        t.Error(err)
    }
    util.Debug("ret", rs)
}

func TestInsert(t *testing.T) {
    rand.Seed(time.Now().UnixNano())
    data := H{
        "company_name": "tt",
        "company_money": rand.Intn(9999),
        "company_create_at":  util.TimeFormat(time.Now()),
        "company_update_at":  util.TimeFormat(time.Now()),
    }
    rs, err := testDB.Insert(data).Into("company").Query().LastInsertId()
    if err != nil {
        t.Error(err)
    }
    util.Debug(rs)
}

func TestUpload(t *testing.T) {
    where := H{
        "company_id": 1,
    }
    data := H{
        "company_name": "aaa",
        "company_money": Expr("company_money+1"),
        "company_update_at":  util.TimeFormat(time.Now()),
    }
    rs, err := testDB.Update("company").Set(data).Where(where).Query().RowsAffected()
    if err != nil {
        t.Error(err)
    }
    util.Debug(rs)
}
