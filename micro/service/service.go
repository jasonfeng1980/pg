package service

import (
    "context"
    "fmt"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "net/url"
    "strings"
)

type Service interface {
    Call(ctx context.Context, dns string, params map[string]interface{})(data interface{}, code int64, msg string)
}

func New(mw []Middleware) Service {
    var svc = NewCallService()
    for _, m := range mw {
        svc = m(svc)
    }
    return svc
}

// client端的网关
type callService struct {}
type H map[string]interface{}

const UploadFile = "PG_UPLOAD_FILE__"

func (c callService) Call(ctx context.Context, dns string, params map[string]interface{}) (data interface{}, code int64, msg string){
    defer func() {
        if err:=recover(); err!=nil{
            msg = fmt.Sprint(err)
            util.LogHandle("error").Log("panic", msg)
            data, code, msg = ecode.CallServerPanic.Parse(dns, msg)
        }
    }()

    // 分解dns
    callApi := Api()
    u, e := url.Parse(dns)
    if e != nil {
        return ecode.HttpDnsParseWrong.Parse()
    }
    l := strings.Split(u.Path, "/")
    if len(l)<4 || l[1] == "" || l[2] == "" || l[3] == "" {
        return ecode.HttpUrlMissMVA.Parse()
    }

    someFunc, ok := callApi.Get(l[1], l[2], l[3])
    if !ok {
        return ecode.HttpCannotMatchDns.Parse(l[1], l[2], l[3])
    } else {
        // 如果有文件上传 params里有 _FILE_
        if v, o := params[UploadFile]; o {
            ctx = context.WithValue(ctx, UploadFile, v)
            delete(params, UploadFile)
        }

        return someFunc(ctx, params)
    }
}
func NewCallService() Service { // 将 callService 变成 接口 Service
    return &callService{}
}
func doCallFunc(someFunc Action, ctx context.Context, params map[string]interface{}) (data interface{}, code int64, msg string){
    defer func(ctx context.Context) {
        callErr := ""
        if err:=recover(); err!=nil{
            callErr = fmt.Sprint(err)
            util.LogHandle("error").Log("panic", )
        }
        ctx = context.WithValue(ctx, "CallErr", callErr)
    }(ctx)

    return someFunc(ctx, params)
}