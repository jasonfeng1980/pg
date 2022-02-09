package endpoint

import (
    "context"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/micro/service"
)

type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)
type Middleware func(Endpoint) Endpoint

func New(s service.Service) MicroEndpoint {
    ept := func(ctx context.Context, request interface{}) (interface{}, error) {
        req := request.(CallRequest)
        data, code, msg := s.Call(ctx, req.Dns, req.Params)
        return CallResponse{
            Data: data,
            Code: code,
            Msg: msg,
        }, nil
    }
    return MicroEndpoint{Endpoint: ept}
}

type MicroEndpoint struct {
    Endpoint Endpoint
}

// 实现 service.Service 接口
func (s MicroEndpoint) Call(ctx context.Context, dns string, params map[string]interface{}) (data interface{}, code int64, msg string) {
    request := CallRequest{
        Dns: dns,
        Params: params,
    }
    resp, err := s.Endpoint(ctx, request)
    if err != nil {
        _, msg := ecode.ReadError(err)
        return ecode.CallServerPanic.Parse(dns, msg)
    }
    r := resp.(CallResponse)
    return r.Data, r.Code, r.Msg

}

func (s MicroEndpoint) AddMiddleware(mdw []Middleware) MicroEndpoint {
    // 循环添加中间件
    for _, m := range mdw {
        if m == nil {
            continue
        }
        s.Endpoint = m(s.Endpoint)
    }
    return s
}

// 定义请求格式
type CallRequest struct {
    Dns string
    Params map[string]interface{}
}

// 定义相应的格式
type CallResponse struct {
    Data interface{} `json:"data"`
    Msg  string      `json:"msg"`
    Code int64       `json:"code"`
}

