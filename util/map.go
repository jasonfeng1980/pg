package util

import "github.com/jasonfeng1980/pg/ecode"

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

// map interface变字符串
// map[string]interface => map[string]string
func MapInterfaceToString(m map[string]interface{})  map[string]string{
    ret := make(map[string]string, len(m))
    for k, v := range m {
        ret[k] = StrParse(v)
    }
    return ret
}