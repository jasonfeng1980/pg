package service

import (
    "context"
    "fmt"
    "github.com/jasonfeng1980/pg/util"
    "strings"
    "sync"
)

// api 类
type api struct{
    mapping map[apiKey]apiValue
}
type apiKey struct {
    Module string   // 模块
    Version string  // 版本
    Action string   // 方法
}
type apiValue struct {
    Method string   // 允许的请求方式 GRPC GET POST PUT DELETE
    Func Action
}
type Action func(c context.Context, p *util.Param)(data interface{}, code int64, msg string)

var (
    apiOnce sync.Once // 单例
    apiInstance *api  // 实例
)

// 对外入口
func Api () *api {
    apiOnce.Do(func() {
        apiInstance = &api{
            mapping: make(map[apiKey]apiValue),
        }
    })
    return apiInstance
}

// 公开的方法
// 注册新的mapping
func (api *api) Register(httpMethod string, module string, action string,  version string, someFunc Action){
    httpMethod = strings.ToUpper(httpMethod)
    util.Debug(fmt.Sprintf("添加[%s]方法%s/%s/%s", httpMethod, module, action, version))
    key := api.makeApiKey(module, action, version)
    if _, ok := api.mapping[key]; ok {
        panic(fmt.Sprintf("方法重复： %s/%s/%s", module, action, version))
    }
    api.mapping[key] = apiValue{
        Method: httpMethod,
        Func: someFunc,
    }
}
// 获取方法
func (api *api)Get(module, action, version string) (someFunc Action, ok bool) {
    key := api.makeApiKey(module, action, version)

    value, ok := api.mapping[key]
    return value.Func, ok

}

// 以下 是内部方法
// 生成mapping的key
func (api *api)makeApiKey(module string, version string, action string) apiKey {
    return apiKey{module, version, action}
}