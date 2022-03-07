package main

import (
    "context"
    "fmt"
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/util"
    "os"
    "strings"
)

func main(){
    if err :=pg.Load("conf/pg.01.dev.json");err!= nil {
        fmt.Println("加载配置错误", err)
        os.Exit(1)
    }
    srv := pg.Server(context.Background())
    srv.Script()
    clientTest()
}

func clientTest() error {
    util.ConsoleTip("请输入请求方式", waitFunc)
    return nil
}

func waitFunc(cmdString string) (string, bool){
    cmdString = strings.ToLower(cmdString)
    svc, _ := pg.Client()
    defer svc.Close()
    switch cmdString {
    case "test":
        dns := "http://demo/auth/v1/test"
        data, code, msg := svc.Call(context.Background(), dns, pg.M{
            "u": "186",
            "p": 1,
        })
        pg.D(data, code, msg)
        return "请输入请求参数", true
    case "http", "grpc":
        // dns  服务类型://服务名称/module/version/action
        dns := cmdString + "://pg/auth/v1/login"
        data, code, msg := svc.Call(context.Background(), dns, pg.M{
            "user_mobile": "186",
            "user_password": 1,
        })
        pg.D(data, code, msg)
        return "请输入请求参数", true
    case "mysql", "mongo", "redis", "orm":
        dns := "grpc://pg/db/v1/" + cmdString
        data, code, msg := svc.Call(context.Background(), dns, pg.M{})
        pg.D(data, code, msg)
        return "请输入请求参数", true
    case "publish", "consume":
        dns := "grpc://pg/mq/v1/" + cmdString
        data, code, msg := svc.Call(context.Background(), dns, pg.M{"date": util.TimeNowString()})
        pg.D(data, code, msg)
        return "请输入请求参数", true
    case "exit":
        return "", false
    default:
        return "请输入正确的参数：http | grpc | exit", true
    }
}