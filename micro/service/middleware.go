package service

import (
    "context"
    "github.com/jasonfeng1980/pg/util"
    "net/http"
    "time"

    "github.com/go-kit/kit/log"
    "github.com/go-kit/kit/metrics"
)

type Middleware func(Service) Service

// 日志记录
func LoggingMiddleware() Middleware {
    logger := util.Log
    return func(next Service) Service {
        return &loggingMiddleware{logger, next}
    }
}
type loggingMiddleware struct {
    logger log.Logger
    next   Service
}
func (mw loggingMiddleware) Call(ctx context.Context, dns string, jsonParams map[string]interface{}) (data interface{}, code int64, msg string){
    var logArgs  []interface{}
    if dns[0:4] == "http" {
        logArgs = append(logArgs, "ip",  jsonParams[RequestHandle].(*http.Request).RemoteAddr)
    }
    defer func(begin time.Time) {
        logArgs = append(logArgs, "dns", dns, "use", time.Since(begin), "params", jsonParams)
        if util.Log.ShowDebug { // debug模式，显示response
            logArgs = append(logArgs,"response", util.M{
                    "data": data,
                    "msg": msg,
                    "code": code,
                })
        }
        mw.logger.Log(logArgs...)
    }(time.Now())
    data, code, msg = mw.next.Call(ctx, dns, jsonParams)
    return
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