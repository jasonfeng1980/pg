package api

import (
    "context"
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/util"
    "sync"
)

// 通过初始化，注册ping关系
func init() {
    api := pg.MicroApi()
    api.Register("POST", "auth", "v1", "test", authTest)
    api.Register("POST", "auth", "v1", "login", authLogin)
    api.Register("GET", "auth", "v1", "logout", authLogout)
}

func authTest(ctx context.Context, params map[string]interface{})(interface{}, int64, string) {
    log := pg.Log
    svc := pg.Client()
    wg := sync.WaitGroup{}

    wg.Add(2)
    go func() {
       data, code, msg := svc.Call(ctx, "grpc://PG/auth/v1/logout", nil)
        log.Infoln(data, code, msg)
       wg.Done()
    }()
    go func() {
       data, code, msg := svc.Call(ctx, "http://PG/auth/v1/logout", nil)
        log.Infoln(data, code, msg)
       wg.Done()
    }()
    wg.Wait()
    return svc.Call(ctx, "grpc://PG/auth/v1/login", pg.M{
        "user_mobile": params["u"],
        "user_password": params["p"],
    })
}

func authLogin(ctx context.Context, params map[string]interface{})(interface{}, int64, string) {
    mobile :=  util.StrParse(params["user_mobile"])
    password := util.StrParse(params["user_password"])
    if mobile == "186" && password == "1" {
        return pg.Suc("登录成功")
    } else {
        return pg.ErrCode(1023, "登录失败")
    }
}

func authLogout(ctx context.Context, params map[string]interface{})(interface{}, int64, string) {
    return pg.Suc("退出成功")


}
