package service

import (
    "context"
    "errors"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "net/url"
    "strings"
)

const (
    UploadFile = "__HTTP_UPLOAD_FILE"
    RequestHandle = "__HTTP_REQUEST_HANDLE"
)

// 定义服务接口
type Service interface {
    Call(ctx context.Context, dns string, params map[string]interface{})(data interface{}, code int64, msg string)
}

// 构造服务，并添加中间件
func New(mw []Middleware) Service {
    var svc = gateway()
    for _, m := range mw {
        svc = m(svc)
    }
    return svc
}

// 入口网关
func gateway() Service { // 将 callService 变成 接口 Service
    return &callService{}
}
type callService struct {}
func (c callService) Call(ctx context.Context, dns string, params map[string]interface{}) (data interface{}, code int64, msg string) {
    defer func() {
        if err:=recover(); err!=nil{ // 出现错误处理
            msg = util.Str(err)
            code, msg = ecode.ReadError(errors.New(msg))
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
        return ecode.HttpUrlMissMVC.Parse()
    }

    // 调用方法
    someFunc, ok := callApi.Get(l[1], l[2], l[3])
    if !ok { // 如果不存在对应的application
        return ecode.HttpCannotMatchDns.Parse(l[1], l[2], l[3])
    }

    return  someFunc(ctx, &util.Param{Box: params})

}

