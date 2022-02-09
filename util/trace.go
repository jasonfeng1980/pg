package util

import (
    "context"
    "github.com/opentracing/opentracing-go"
)

func TraceTag(ctx context.Context, k string, v interface{}) {
    span := opentracing.SpanFromContext(ctx)
    span.SetTag(k, v)
}
