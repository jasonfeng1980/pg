package grpc

import (
    "context"
    "encoding/base64"
    "github.com/jasonfeng1980/pg/util"
    "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/ext"
    "google.golang.org/grpc/metadata"
    "strings"
)
// 链路追踪-客户端
func TraceClient(tracer opentracing.Tracer) func(context.Context, *metadata.MD) context.Context {
    return func(ctx context.Context, md *metadata.MD) context.Context {
        if tracer == nil {
            return ctx
        }
        if span := opentracing.SpanFromContext(ctx); span != nil {
            if err := tracer.Inject(span.Context(), opentracing.TextMap, metadataReaderWriter{md}); err != nil {
                util.Error("", "err", err)
            }
        }
        return ctx
    }
}
// 链路追踪-服务端
func TraceServer(tracer opentracing.Tracer, operationName string)  func(context.Context, metadata.MD) context.Context{
    return func(ctx context.Context, md metadata.MD) context.Context {
        if tracer == nil {
            return ctx
        }
        var span opentracing.Span
        wireContext, err := tracer.Extract(opentracing.TextMap, metadataReaderWriter{&md})
        if err != nil && err != opentracing.ErrSpanContextNotFound {
            util.Error("", "err", err)
        }
        span = tracer.StartSpan(operationName, ext.RPCServerOption(wireContext))
        return opentracing.ContextWithSpan(ctx, span)
    }
}


type metadataReaderWriter struct {
    *metadata.MD
}
func (w metadataReaderWriter) Set(key, val string) {
    key = strings.ToLower(key)
    if strings.HasSuffix(key, "-bin") {
        val = base64.StdEncoding.EncodeToString([]byte(val))
    }
    (*w.MD)[key] = append((*w.MD)[key], val)
}
func (w metadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
    for k, vals := range *w.MD {
        for _, v := range vals {
            if err := handler(k, v); err != nil {
                return err
            }
        }
    }
    return nil
}