package main

import (
    "github.com/jasonfeng1980/pg"
)

func main(){
    root := "../"
    pg.YamlRead(root).
        Server("example/conf/demo/pg_11_dev.yaml").
        Mysql("example/conf/mysql.yaml").
        Mongo("example/conf/mongo.yaml").
        Redis("example/conf/redis.yaml").
        Rabbitmq("example/conf/rabbitmq.yaml").
        Set()
    svc := pg.Server()
    svc.Script(test)
}

func test() error {
    return nil
}