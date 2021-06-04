package util

import (
    "strings"
)

// 获取[]map的 指定字段的值
// e.g. 查询mysql，取结果集的指定字段
// []interface{} | []map[string]interface{} |map[string]interface{}
func ListMapField(dataArg interface{}, fieldArr []string) (ret []map[string]interface{}, err error) {
    data, err := InterfaceToListMap(dataArg)
    if err!= nil {
        return nil, err
    }
    // 循环赋值
    for _, v := range data {
        d, _ := MapField(v, fieldArr, false)
        ret = append(ret, d)
    }
    return
}

// 批量清除List里字符串的前后指定字符
func ListTrim(l []string, cutset string) (ret []string){
    for _, v := range l{
        ret = append(ret, strings.Trim(v, cutset))
    }
    return
}

// 判断值 是否在list里
func ListHave(l interface{}, need interface{}) bool {
    switch key := need.(type) {
    case int:
        for _, v := range l.([]int) {
            if v == key {
                return true
            }
        }
    case string:
        for _, v := range l.([]string) {
            if v == key {
                return true
            }
        }
    case int64:
        for _, v := range l.([]int64) {
            if v == key {
                return true
            }
        }
    case float64:
        for _, v := range l.([]float64) {
            if v == key {
                return true
            }
        }
    }
    return false
}

func ListInterfaceToStr(l []interface{}) (ret []string){
    for _, v:=range l {
        ret = append(ret, StrParse(v))
    }
    return
}