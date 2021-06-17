package main

import (
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/database/db"
)

// 生成ORM
func main(){
    root := "../"
    pg.YamlRead(root).
        Server("example/conf/demo/pg_11_dev.yaml").
        Mysql("example/conf/mysql.yaml").
        Set()

    mysqlConf := conf.Get().MySQLConf

    // 链接MYSQL连接池
    db.MYSQL.Conn(mysqlConf)
    defer db.MYSQL.Close()

    // 用MYSQL的DEMO别名配置， 在orm目录 生成 包名为 company的各个表的ORM文件
    db.OrmInit("DEMO", "company", "orm")

}