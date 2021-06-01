package conf

import (
    "github.com/jasonfeng1980/pg/util"
    "testing"
)

var log = util.Log
func TestMain(m *testing.M) {
    m.Run()
}

func TestOther(t *testing.T) {
    root := "../"
    serverFile := "example/conf/demo/pg_11_dev.yaml"
    ConfInit(root).
        Server(serverFile).
        Mysql("example/conf/mysql.yaml").
        Mongo("example/conf/mongo.yaml").
        Redis("example/conf/redis.yaml").
        Rabbitmq("example/conf/rabbitmq.yaml").
        Set()

    c := Get()
    log.With("mysql", c.MySQLConf,
        "mongo", c.MongoConf,
        "redis", c.RedisConf,
        "rabbitmq", c.RabbitMQConf).Info()
}

