package service

import (
    "context"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/util"
    "github.com/prometheus/client_golang/prometheus"
    "net/http"
    "time"
)

// 定义中间件类型
type Middleware func(Service) Service

// 请求日志记录中间件
func LoggingMiddleware() Middleware {
    return func(next Service) Service {
        return &loggingMiddleware{next}
    }
}
type loggingMiddleware struct {
    next   Service
}
func (mw loggingMiddleware) Call(ctx context.Context, dns string, jsonParams map[string]interface{}) (data interface{}, code int64, msg string){
    var logArgs  []interface{}

    if dns[0:4] == "http" {
        session := util.SessionHandle(ctx)
       logArgs = append(logArgs, "ip",  session.Get(RequestHandle).(*http.Request).RemoteAddr)
    }
    defer func(begin time.Time) {
        logArgs = append(logArgs, "dns", dns, "params", jsonParams)
        if conf.Conf.LogLevel == "debug"  { // debug模式，显示response
            logArgs = append(logArgs,"response", map[string]interface{}{
                "data": data,
                "msg": msg,
                "code": code,
            })
        }
        logArgs = append(logArgs, "use", time.Since(begin))
        util.Info("", logArgs...)
    }(time.Now())
    data, code, msg = mw.next.Call(ctx, dns, jsonParams)
    return
}


// 服务监控 每请求一次+1
func InstrumentingMiddleware(calls *prometheus.CounterVec, duration *prometheus.SummaryVec) Middleware {
    return func(next Service) Service {
        return instrumentingMiddleware{
            calls:  calls,
            duration: duration,
            next:  next,
        }
    }
}
type instrumentingMiddleware struct {
    calls *prometheus.CounterVec
    duration *prometheus.SummaryVec
    next  Service
}
func (mw instrumentingMiddleware) Call(ctx context.Context, dns string, jsonParams map[string]interface{}) (data interface{}, code int64, msg string){
    defer func(begin time.Time) {
        mw.duration.WithLabelValues(
            util.Str(code == 200),
            util.Str(dns),
        ).Observe(time.Since(begin).Seconds())
    }(time.Now())

    data, code, msg =mw.next.Call(ctx, dns, jsonParams)
    mw.calls.WithLabelValues(
        util.Str(code),
        dns,
        ).Add(1)
    return
}