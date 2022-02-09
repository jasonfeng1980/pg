package endpoint

import (
    "context"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/ext"
    "golang.org/x/time/rate"
    "time"

    "github.com/sony/gobreaker"
)


// endpoint日志错误
func LoggingMiddleware() Middleware {
    return func(next Endpoint) Endpoint {
        return func(ctx context.Context, request interface{}) (response interface{}, err error) {
            defer func() {
                if err != nil {
                    util.Error("endpoint出现错误", "error", err, "ts", time.Now())
                }
            }()
            return next(ctx, request)
        }
    }
}

// 限流， 超过报错
func LimitErroring(limit *rate.Limiter) Middleware {
    return func(next Endpoint) Endpoint {
        return func(ctx context.Context, request interface{}) (interface{}, error) {
            if !limit.Allow() {
                return nil, ecode.LimitErroring.Error()
            }
            return next(ctx, request)
        }
    }
}
// 限流， 超过延时等待
func LimitDelaying(limit *rate.Limiter) Middleware {
    return func(next Endpoint) Endpoint {
        return func(ctx context.Context, request interface{}) (interface{}, error) {
            if err := limit.Wait(ctx); err != nil {
                return nil, err
            }
            return next(ctx, request)
        }
    }
}

// 熔断
func Gobreaking(setting gobreaker.Settings)Middleware {
    circuitBreaker := gobreaker.NewCircuitBreaker(setting)
    return func(next Endpoint) Endpoint {
        return func(ctx context.Context, request interface{}) (interface{}, error) {
            return circuitBreaker.Execute(func() (interface{}, error) { return next(ctx, request) })
        }
    }
}
// trace server
func TraceServer(otTracer opentracing.Tracer) Middleware {
    return func(next Endpoint) Endpoint {
        return func(ctx context.Context, request interface{}) (interface{}, error) {
            name := request.(CallRequest).Dns
            var serverSpan opentracing.Span
            if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
                serverSpan = otTracer.StartSpan(
                    name,
                    opentracing.ChildOf(parentSpan.Context()),
                )
            } else {
                serverSpan = otTracer.StartSpan(name)
            }
            defer serverSpan.Finish()
            ext.SpanKindRPCServer.Set(serverSpan)
            ctx = opentracing.ContextWithSpan(ctx, serverSpan)
            return next(ctx, request)
        }
    }
}

// trace client
func TraceClient(name string, otTracer opentracing.Tracer) Middleware {
    return func(next Endpoint) Endpoint {
        return func(ctx context.Context, request interface{}) (interface{}, error) {
            var clientSpan opentracing.Span
            if clientSpan = opentracing.SpanFromContext(ctx); clientSpan == nil {
                clientSpan = otTracer.StartSpan(name)
            }
            clientSpan.LogKV("Call", request.(CallRequest).Dns)
            ext.SpanKindRPCClient.Set(clientSpan)
            ctx = opentracing.ContextWithSpan(ctx, clientSpan)
            return next(ctx, request)
        }
    }
}

