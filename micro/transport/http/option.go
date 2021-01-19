package http

import (
    "context"
    "github.com/jasonfeng1980/pg/ecode"
    callendpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    "github.com/jasonfeng1980/pg/util"
    "net/http"

    stdopentracing "github.com/opentracing/opentracing-go"
    stdzipkin "github.com/openzipkin/zipkin-go"

    "github.com/go-kit/kit/tracing/opentracing"
    "github.com/go-kit/kit/tracing/zipkin"
    "github.com/go-kit/kit/transport"
    kitHttp "github.com/go-kit/kit/transport/http"
)


func DefaultClientOptions(otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) []kitHttp.ClientOption {
    logger := util.LogHandle("info")
    opts := []kitHttp.ClientOption{
        kitHttp.ClientBefore(opentracing.ContextToHTTP(otTracer,  logger)),
    }
    if zipkinTracer != nil {
        opts = append(opts, zipkin.HTTPClientTrace(zipkinTracer))
    }
    return opts
}

func DefaultServerOptions(otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) []kitHttp.ServerOption {
    logger := util.LogHandle("error")
    opts := []kitHttp.ServerOption{
        kitHttp.ServerErrorEncoder(errorEncoder),
        kitHttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
    }
    if zipkinTracer != nil {
        opts = append(opts, zipkin.HTTPServerTrace(zipkinTracer))
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
