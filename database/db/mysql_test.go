package db

import (
    "fmt"
    "github.com/jasonfeng1980/pg/util"
    "testing"
    "time"
)

var demo *Query

type H map[string]interface{}


func TestMain(m *testing.M) {
    /*mysqlFile := "../../example/mysql.yaml"

    mysqlConf, err := util.YamlToMysql(mysqlFile)
    if err != nil {
        fmt.Println(err)
    }


    // 链接MYSQL连接池
    MYSQL.Conn(mysqlConf)
    defer MYSQL.Close()

    demo, _ = MYSQL.Get("demo")

    m.Run()*/
}

func TestSelect(t *testing.T) {
    rs, err := demo.Select("*").From("company").Limit(0,1).Query().Array()
    if err != nil {
        t.Error(err)
    }
    fmt.Println(rs)
}

func TestInsert(t *testing.T) {
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
    fmt.Println(rs)
}

func TestUpload(t *testing.T) {
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
    fmt.Println(rs)
}
