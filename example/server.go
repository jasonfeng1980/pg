package main

import (
    "github.com/jasonfeng1980/pg"
    _ "github.com/jasonfeng1980/pg/example/api"
    "os"
)

func main(){
    confPath := "pg_11_dev.yaml"
    if len(os.Args) == 2 {
        confPath = "client_01_dev.yaml"
    }

    root := "../"
    pg.YamlRead(root).
        Server("example/conf/demo/" + confPath).
        Mysql("example/conf/mysql.yaml").
        Mongo("example/conf/mongo.yaml").
        Redis("example/conf/redis.yaml").
        Rabbitmq("example/conf/rabbitmq.yaml").
        Set()

    svc := pg.Server()
    svc.Run()
}