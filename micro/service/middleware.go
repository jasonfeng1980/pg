package service

import (
    "context"
    "github.com/jasonfeng1980/pg/util"
    "time"

    "github.com/go-kit/kit/log"
    "github.com/go-kit/kit/metrics"
)

type Middleware func(Service) Service

// 日志记录
func LoggingMiddleware() Middleware {
    logger := util.LogHandle("info")
    return func(next Service) Service {
        return &loggingMiddleware{logger, next}
    }
}
type loggingMiddleware struct {
    logger log.Logger
    next   Service
}
func (mw loggingMiddleware) Call(ctx context.Context, dns string, jsonParams map[string]interface{}) (data interface{}, code int64, msg string){
    defer func(begin time.Time) {
        mw.logger.Log("dns", dns, "time", time.Since(begin))
    }(time.Now())
    return mw.next.Call(ctx, dns, jsonParams)
}

// 服务监控 每请求一次+1
func InstrumentingMiddleware(calls metrics.Counter) Middleware {
    return func(next Service) Service {
        return instrumentingMiddleware{
            calls:  calls,
            next:  next,
        }
    }
}
type instrumentingMiddleware struct {
    calls metrics.Counter
    next  Service
}
func (mw instrumentingMiddleware) Call(ctx context.Context, dns string, jsonParams map[string]interface{}) (data interface{}, code int64, msg string){
    data, code, msg =mw.next.Call(ctx, dns, jsonParams)
    mw.calls.Add(1)
    return
}