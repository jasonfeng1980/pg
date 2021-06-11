package main

import (
    "context"
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/util"
    "strings"
)

func main(){
    util.ConsoleTip("请输入请求方式", waitFunc)
}

func waitFunc(cmdString string) (string, bool){
    cmdString = strings.ToLower(cmdString)
    svc := pg.Client()
    defer svc.Close()
    switch cmdString {
    case "http", "grpc":
        // dns  服务类型://服务名称/module/version/action
        dns := cmdString + "://PG/auth/v1/login"
        data, code, msg := svc.Call(context.Background(), dns, pg.M{
            "user_mobile": "186",
            "user_password": 1,
        })
        pg.D(data, code, msg)
        return "请输入请求参数", true
    case "mysql", "mongo", "redis", "orm":
        dns := "grpc://PG/db/v1/" + cmdString
        data, code, msg := svc.Call(context.Background(), dns, pg.M{})
        pg.D(data, code, msg)
        return "请输入请求参数", true
    case "publish", "consume":
        dns := "grpc://PG/mq/v1/" + cmdString
        data, code, msg := svc.Call(context.Background(), dns, pg.M{"date": util.TimeNowString()})
        pg.D(data, code, msg)
        return "请输入请求参数", true
    case "exit":
        return "", false
    default:
        return "请输入正确的参数：http | grpc | exit", true
    }
}