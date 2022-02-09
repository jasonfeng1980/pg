package http

import (
    "context"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/micro/endpoint"
    "github.com/jasonfeng1980/pg/util"
    "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/ext"
    "net/http"
)

// 链路追踪-客户端
func TraceClient(tracer opentracing.Tracer) func(ctx context.Context, r *http.Request) context.Context {
    return func(ctx context.Context, r *http.Request) context.Context {
        if tracer == nil {
            return ctx
        }
        if span := opentracing.SpanFromContext(ctx); span != nil {
            if err := tracer.Inject(
                span.Context(),
                opentracing.HTTPHeaders,
                opentracing.HTTPHeadersCarrier(r.Header),
            ); err != nil {
                util.Error("", "err", err)
            }
        }
        return ctx
    }
}
// 链路追踪-服务端
func TraceServer(tracer opentracing.Tracer) func(ctx context.Context, r *http.Request) context.Context {
    return func(ctx context.Context, r *http.Request) context.Context{
        if tracer == nil {
            return ctx
        }
        var span opentracing.Span
        wireContext, err := tracer.Extract(
            opentracing.HTTPHeaders,
            opentracing.HTTPHeadersCarrier(r.Header),
        )
        if err != nil && err != opentracing.ErrSpanContextNotFound {
            util.Error("", "err", err)
        }
        span = tracer.StartSpan("Call", ext.RPCServerOption(wireContext))
        return opentracing.ContextWithSpan(ctx, span)
    }
}

// 错误处理
func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(http.StatusInternalServerError)
    code, msg := ecode.ReadError(err)

    util.JsonEncoder(w).Encode(endpoint.CallResponse{
        Code: code,
        Msg:  msg,
    })
}
