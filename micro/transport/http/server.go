package http

import (
    "context"
    "github.com/jasonfeng1980/pg/micro/endpoint"
    "github.com/jasonfeng1980/pg/util"
    "github.com/opentracing/opentracing-go"
    "net/http"
)


// 获得服务句柄
func NewServer(endpoints endpoint.MicroEndpoint, tracer opentracing.Tracer, operationName string) http.Server {
    return http.Server{
        Handler: Server{
            ept:        endpoints,
            operationName:   operationName,
            tracer: tracer,
        },
    }
}

type ErrorFunc func(ctx context.Context, err error)

type Server struct {
    ept           endpoint.MicroEndpoint
    tracer        opentracing.Tracer
    operationName   string
}

// ServeHTTP implements http.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    // 链路追踪 - 还原
    ctx = TraceServer(s.tracer)(ctx, r)
    // 开启session
    ctx = util.SessionNew(ctx)

    // 解密request
    ctx, request, err := s.DecodeRequest(ctx, r)
    if err != nil {
        errorEncoder(ctx, err, w)
        return
    }
    // 执行
    response, err := s.ept.Endpoint(ctx, request)
    if err != nil {
        errorEncoder(ctx, err, w)
        return
    }
    // 加密response
    if err := s.EncodeResponse(ctx, w, response); err != nil {
        errorEncoder(ctx, err, w)
        return
    }
}
// 解密方法
func (s Server) DecodeRequest(ctx context.Context, r *http.Request) (context.Context, interface{}, error) {
    ctx, dns, params, err := getRequestParams(ctx, r)
    if err != nil {
        return ctx, nil, err
    }

    req := endpoint.CallRequest{
        Dns: dns,
        Params: params,
    }
    return ctx, req, nil
}
// 加密方法
func (s Server) EncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    return util.JsonEncoder(w).Encode(response.(endpoint.CallResponse))
}