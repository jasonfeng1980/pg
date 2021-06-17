package util

import (
    "context"
    stdopentracing "github.com/opentracing/opentracing-go"
)

//var tracer stdopentracing.Tracer
//
//// 设置Trace
//func TraceSet(t stdopentracing.Tracer){
//    tracer = t
//}
//
//// 获取Trace
//func TraceGet() stdopentracing.Tracer{
//    return tracer
//}

//// 通过Ctx获取Span
//func TraceSpan(ctx context.Context, name string) (serverSpan stdopentracing.Span){
//    if parentSpan := stdopentracing.SpanFromContext(ctx); parentSpan != nil {
//        serverSpan = tracer.StartSpan(
//            name,
//            stdopentracing.ChildOf(parentSpan.Context()),
//        )
//    } else {
//        serverSpan = tracer.StartSpan(name)
//    }
//    return
//}
func TraceTag(ctx context.Context, k string, v interface{}) {
    span := stdopentracing.SpanFromContext(ctx)
    span.SetTag(k, v)
}