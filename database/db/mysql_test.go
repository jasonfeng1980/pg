package db

import (
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/util"
    "testing"
    "time"
)

var demo *Query

type H map[string]interface{}


func TestMain(m *testing.M) {
    root := "../../"
    conf.ConfInit(root).
        Server("example/conf/demo/pg_11_dev.yaml").
        Mysql("example/conf/mysql.yaml").
        Set()
    c := conf.Get()
    MYSQL.Conn(c.MySQLConf)

    demo, _ = MYSQL.Get("DEMO")
    m.Run()
}

func TestSelect(t *testing.T) {
    rs, err := demo.Select("TABLE_SCHEMA,TABLE_NAME,TABLE_ROWS").
        From("information_schema.TABLES").
        //Where("TABLE_ROWS > ?", 10).
        Where("TABLE_ROWS", util.M{"$gte": 10}).
        Where("AVG_ROW_LENGTH", util.M{"$in": []int{0,1,2}}).
        //Where(util.M{"AVG_ROW_LENGTH": 1}).
        OrderBy("TABLE_ROWS  desc").
        Limit(1, 1).
        //Cache(true).
        Query().
        Array()
    if err != nil {
        t.Error(err)
    }
    util.Log.Log("ret", rs)
}

func _TestInsert(t *testing.T) {
    data := H{
        "company_name": "ff",
        "company_money": 888,
        "create_at":  util.TimeFormat(time.Now()),
        "update_at":  util.TimeFormat(time.Now()),
    }
    rs, err := demo.Insert(data).Into("company").Query().LastInsertId()
    if err != nil {
        t.Error(err)
    }
    util.Log.Debugln(rs)
}

func _TestUpload(t *testing.T) {
    where := H{
        "company_id": 210,
    }
    data := H{
        "company_name": "aaa",
        "company_money": Expr("company_money+1"),
        "update_at":  util.TimeFormat(time.Now()),
    }
    rs, err := demo.Update("company").Set(data).Where(where).Query().RowsAffected()
    if err != nil {
        t.Error(err)
    }
    util.Log.Debugln(rs)
}
