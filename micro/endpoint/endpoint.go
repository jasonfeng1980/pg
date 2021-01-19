package endpoint

import (
    "context"
    "github.com/go-kit/kit/endpoint"
    "github.com/jasonfeng1980/pg/micro/service"
)


func New(s service.Service) Set {
    ept := func(ctx context.Context, request interface{}) (interface{}, error) {
        req := request.(CallRequest)
        data, code, msg := s.Call(ctx, req.Dns, req.Params)
        return CallResponse{
            Data: data,
            Code: code,
            Msg: msg,
        }, nil
    }
    return Set{CallEndpoint: ept}
}

type Set struct {
    CallEndpoint endpoint.Endpoint
}

// 实现 service.Service 接口
func (s Set) Call(ctx context.Context, dns string, params map[string]interface{}) (data interface{}, code int64, msg string) {
    request := CallRequest{
        Dns: dns,
        Params: params,
    }
    resp, err := s.CallEndpoint(ctx, request)
    if err != nil {
        return
    }
    //return resp.(map[string]interface{})
    r := resp.(CallResponse)
    return r.Data, r.Code, r.Msg


    //return map[string]interface{}{
    //   "data": r.Data,
    //   "msg": r.Msg,
    //   "code": r.Code,
    //}
}

func (s Set) AddMiddleware(mdw []endpoint.Middleware) Set {
    // 循环添加中间件
    for _, m := range mdw {
        if m == nil {
            continue
        }
        s.CallEndpoint = m(s.CallEndpoint)
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

