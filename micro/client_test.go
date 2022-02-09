package micro

import (
    "context"
    "github.com/jasonfeng1980/pg/util"
    "testing"
)

func TestNewClient(t *testing.T) {
    ctx := context.Background()
    srv, _ := NewClient()
    data, code, msg := srv.Call(ctx, "grpc://pg/auth/v1/test", map[string]interface{}{
        "fff": 11,
        "dd":  22,
    })
    util.Info("grpc", "data", data, "code", code, "msg", msg)

    //data, code, msg = srv.Call(ctx, "http://pg/auth/v1/test", map[string]interface{}{
    //    "u": 186,
    //    "p":  1,
    //})
    //util.Info("http", "data", data, "code", code, "msg", msg)
}
