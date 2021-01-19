package endpoint

import (
    "context"
    "fmt"
    "github.com/go-kit/kit/tracing/opentracing"
    "github.com/jasonfeng1980/pg/util"
    stdopentracing "github.com/opentracing/opentracing-go"
    otext "github.com/opentracing/opentracing-go/ext"
    stdzipkin "github.com/openzipkin/zipkin-go"
    "github.com/openzipkin/zipkin-go/model"
    "time"

    "github.com/go-kit/kit/circuitbreaker"
    "github.com/sony/gobreaker"

    "github.com/go-kit/kit/endpoint"
    "github.com/go-kit/kit/metrics"
    "github.com/go-kit/kit/ratelimit"
    "golang.org/x/time/rate"
)

// 统计监控
// InstrumentingMiddleware returns an endpoint middleware that records
// the duration of each invocation to the passed histogram. The middleware adds
// a single field: "success", which is "true" if no error is returned, and
// "false" otherwise.
func InstrumentingMiddleware(duration metrics.Histogram) endpoint.Middleware {
    return func(next endpoint.Endpoint) endpoint.Endpoint {
        return func(ctx context.Context, request interface{}) (response interface{}, err error) {
            defer func(begin time.Time) {
                duration.With("success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
            }(time.Now())
            return next(ctx, request)

        }
    }
}

// LoggingMiddleware returns an endpoint middleware that logs the
// duration of each invocation, and the resulting error, if any.
func LoggingMiddleware() endpoint.Middleware {
    return func(next endpoint.Endpoint) endpoint.Endpoint {
        return func(ctx context.Context, request interface{}) (response interface{}, err error) {

            defer func(begin time.Time) {
                if err != nil {
                    util.LogHandle("error").Log("transport_error", err, "took", time.Since(begin))
                }
            }(time.Now())
            return next(ctx, request)
        }
    }
}

// 限流， 超过报错
func LimitErroring(limit *rate.Limiter) endpoint.Middleware {
    return ratelimit.NewErroringLimiter(limit)
}
// 限流， 超过延时等待
func LimitDelaying(limit *rate.Limiter) endpoint.Middleware {
    return ratelimit.NewDelayingLimiter(limit)
}
// 熔断
func Gobreaking(setting gobreaker.Settings)endpoint.Middleware {
    return circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(setting))
}
// trace server
func TraceServer(otTracer stdopentracing.Tracer) endpoint.Middleware {
    return func(next endpoint.Endpoint) endpoint.Endpoint {
        return func(ctx context.Context, request interface{}) (interface{}, error) {
            name := request.(CallRequest).Dns
            serverSpan := stdopentracing.SpanFromContext(ctx)
            if serverSpan == nil {
                // All we can do is create a new root span.
                serverSpan = otTracer.StartSpan(name)
            } else {
                serverSpan.SetOperationName(name)
            }
            defer serverSpan.Finish()
            otext.SpanKindRPCServer.Set(serverSpan)
            ctx = stdopentracing.ContextWithSpan(ctx, serverSpan)
            return next(ctx, request)
        }
    }
    //return opentracing.TraceServer(otTracer, "call")
}
// zipkinTracer
func ZipkinTrace(zipkinTracer *stdzipkin.Tracer) endpoint.Middleware {
    if zipkinTracer == nil {
        return nil
    }
    return func(next endpoint.Endpoint) endpoint.Endpoint {
        return func(ctx context.Context, request interface{}) (interface{}, error) {
            var sc model.SpanContext
            if parentSpan := stdzipkin.SpanFromContext(ctx); parentSpan != nil {
                sc = parentSpan.Context()
            }
            name := request.(CallRequest).Dns
            sp := zipkinTracer.StartSpan(name, stdzipkin.Parent(sc))
            defer sp.Finish()

            ctx = stdzipkin.NewContext(ctx, sp)
            return next(ctx, request)
        }
    }
    //return zipkin.TraceEndpoint(zipkinTracer, name)
}

// trace client
func TraceClient(name string, otTracer stdopentracing.Tracer) endpoint.Middleware {
    return opentracing.TraceClient(otTracer, name)
}

