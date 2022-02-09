package rdb

import (
    "context"
    "fmt"
    "testing"
    "time"
)

var (
    ctx = context.Background()
    r *RedisConn
)

func TestMain(m *testing.M) {
    redisConfMap := map[string][]string{
        "demo": []string{
            "redis://:@tcp(localhost:6379)/0",
        },
    }
    Redis.Conn(redisConfMap)
    r, _ = Redis.Client("demo")
    m.Run()
}

func TestQuery(t *testing.T) {

    fmt.Println(r.W.Set(ctx, "aaa", "111", 0))
    fmt.Println(r.W.Get(ctx, "aaa"))
}

func TestKey(t *testing.T) {
    k := userCache(ctx, "userName")
    k.Set("hello world")
    fmt.Println(k.Get())
}

func userCache(ctx context.Context, name string) String {
    return String{
        Key: Key{
            CTX: ctx,
            Name: "USER_CACHE:" + name,
            Client: r,
            Expr: time.Second * 5,
        },
    }
}
