package test

import (
    "context"
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/util"
    "sync"
)

// 通过初始化，注册ping关系
func init() {
    api := pg.MicroApi()
    api.Register("POST", "auth", "test", "v1", authTest)
    api.Register("POST", "auth", "login", "v1", authLogin)
    api.Register("GET", "auth", "logout", "v1", authLogout)
    api.Register("GET", "auth", "logout", "v2", authLogoutV2)
}

func authTest(ctx context.Context, params *util.Param)(interface{}, int64, string) {
    svc, _ := pg.Client()
    defer svc.Close()

    wg := sync.WaitGroup{}
    util.TraceTag(ctx, "DO", "请求v2logout 和 v1logout")
    wg.Add(2)
    go func() {
       svc.Call(ctx, "grpc://pg/auth/v2/logout", pg.M{})
       wg.Done()
    }()
    go func() {
       svc.Call(ctx, "http://pg/auth/v1/logout", pg.M{})
       wg.Done()
    }()
    wg.Wait()
    return svc.Call(ctx, "grpc://pg/auth/v1/login", pg.M{
        "user_mobile": params.Get("u"),
        "user_password": params.Get("p"),
    })
}

func authLogin(ctx context.Context, params *util.Param)(interface{}, int64, string) {
    util.TraceTag(ctx, "DO", "检查账号密码")
    mobile :=  params.GetStr("user_mobile")
    password := params.GetStr("user_password")
    if mobile == "186" && password == "1" {
        return pg.Suc("登录成功")
    } else {
        return pg.ErrCode(1023, "登录失败")
    }
}

func authLogoutV2(ctx context.Context, params *util.Param)(interface{}, int64, string) {
    return pg.Suc("退出成功")
}

func authLogout(ctx context.Context, params *util.Param)(interface{}, int64, string) {
    svc, _ := pg.Client()
    defer svc.Close()
    return svc.Call(ctx, "http://demo/auth/v2/logout", pg.M{})
}
