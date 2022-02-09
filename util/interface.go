package util

import (
    "github.com/jasonfeng1980/pg/ecode"
    "reflect"
)


// 获取interface{}的类型
func InterfaceType(arg interface{}) string{
    return reflect.Indirect(reflect.ValueOf(arg)).Type().String()
}
// interface{}是不是map
func InterfaceIsMap(arg interface{}) bool{
    return reflect.ValueOf(arg).Kind() == reflect.Map
}
// interface{}是不是切片
func InterfaceIsSlice(arg interface{}) bool{
    return reflect.ValueOf(arg).Kind() == reflect.Slice
}

// interface => []map[string]interface
// []interface{} | []map[string]interface{} |map[string]interface{}
func InterfaceToListMap(dataArg interface{}) (ret []map[string]interface{}, err error){
    err = ecode.UtilWrongDataType.Error()
    switch dataArg.(type) {
    case []interface{}:
        arr:= dataArg.([]interface{})
        for _, line := range arr {
            if mapInfo, isMap := line.(map[string]interface{}); isMap {
                ret = append(ret, mapInfo)
                err = nil
            }
        }
    case []map[string]interface{}:
        ret = dataArg.([]map[string]interface{})
        err = nil
    case map[string]interface{}:
        mapData := dataArg.(map[string]interface{})
        ret = append(ret, mapData)
        err = nil
    }
    return
}
