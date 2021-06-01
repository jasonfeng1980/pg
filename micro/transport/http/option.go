package http

import (
    "context"
    "github.com/go-kit/kit/tracing/opentracing"
    "github.com/jasonfeng1980/pg/ecode"
    callendpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    "github.com/jasonfeng1980/pg/util"
    "net/http"

    "github.com/go-kit/kit/transport"
    kitHttp "github.com/go-kit/kit/transport/http"
    stdopentracing "github.com/opentracing/opentracing-go"
)


func DefaultClientOptions(otTracer stdopentracing.Tracer) []kitHttp.ClientOption {
    logger := util.Log
    opts := []kitHttp.ClientOption{
        kitHttp.ClientBefore(opentracing.ContextToHTTP(otTracer,  logger)),
    }
    return opts
}

func DefaultServerOptions(otTracer stdopentracing.Tracer) []kitHttp.ServerOption {
    logger := util.Log
    opts := []kitHttp.ServerOption{
        kitHttp.ServerErrorEncoder(errorEncoder),
        kitHttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
    }
    opts = append(opts, kitHttp.ServerBefore(opentracing.HTTPToContext(otTracer, "Call", logger)))
    return opts
}

// 错误处理
func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(http.StatusInternalServerError)
    code, msg := ecode.ReadError(err)

    util.Json.NewEncoder(w).Encode(callendpoint.CallResponse{
        Code: code,
        Msg:  msg,
    })
}
