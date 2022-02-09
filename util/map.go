package util

import (
    "github.com/jasonfeng1980/pg/ecode"
    "regexp"
    "strings"
)

type M map[string]interface{}

// 获取一个map的 指定字段
func MapField(data map[string]interface{}, fieldArr []string, must bool) (ret map[string]interface{}, err error) {
    ret = make(map[string]interface{})
    // 循环赋值
    for _, name := range fieldArr {
        if v, ok := data[name]; ok {
            ret[name] = v
        } else if must{
            err = ecode.UtilMissNeedField.Error(name)
        }
    }
    return
}

// 增量合并
func MapMerge(base map[string]interface{}, others ...map[string]interface{}) (ret map[string]interface{}){
    if base == nil {
        ret = make(map[string]interface{})
    } else {
        ret = base
    }
    for _, other := range others {
        if other == nil {
            continue
        }
        for k, v := range other {
            if _, ok := base[k]; !ok {
                ret[k] = v
            }
        }
    }
    return
}

// 获取map的所有key
func MapKeys(m interface{}) (keys []string) {
    v1, ok1 := m.(map[string]interface{})
    v2, ok2 := m.(map[string][]interface{})
    i := 0
    if ok1 {
        keys = make([]string, len(v1))
        for k, _ := range v1 {
            keys[i] = k
            i++
        }
    }
    if ok2 {
        keys = make([]string, len(v2))
        for k, _ := range v2 {
            keys[i] = k
            i++
        }
    }
    return
}

func MapStringToString(m map[string]string) string{
    ret := ""
    for _, v := range m {
        ret += v
    }
    return ret
}

// map interface变字符串
// map[string]interface => map[string]string
func MapInterfaceToMapString(m map[string]interface{})  map[string]string{
    ret := make(map[string]string, len(m))
    for k, v := range m {
        ret[k] = Str(v)
    }
    return ret
}

// 将url的query 转换成 map[string]string
func MapFromUrlQuery(q string) map[interface{}]interface{}{
    ret := make(map[interface{}]interface{})
    for _, v := range strings.Split(q, "&") {
        param := strings.SplitN(v, "=", 2)
        if len(param) != 2 {
            continue
        }
        ret[param[0]] = param[1]
    }
    return ret
}

// 将DNS转换成  map[string]interface{}
// DNS格式为 driver://[user]:[password]@network(host:port)/[dbname][?param1=value1&paramN=valueN]
func MapFromDns(dns string) (*Param, error) {
    reg := regexp.MustCompile(`^(\w+):\/\/(.*):(.*)@(\w+)\((.+):(\d+)\)\/([^?]*)\??(.*)$`)
    m := reg.FindAllStringSubmatch(dns, 1)
    if len(m)== 0 {
        return nil, ecode.UtilWrongDns.Error()
    }
    params := MapFromUrlQuery(m[0][8])
    ret := map[string]interface{}{
        "driver": m[0][1],
        "user": m[0][2],
        "password": m[0][3],
        "network": m[0][4],
        "host": m[0][5],
        "port": m[0][6],
        "dbname": m[0][7],
        "hostList": strings.Split(m[0][5] + ":" + m[0][6], ","),
        "params": params,
    }
    return &Param{
        Box: ret,
    }, nil
}