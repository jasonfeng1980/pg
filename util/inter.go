package util

import (
    "github.com/jasonfeng1980/pg/ecode"
    "strconv"
)


func IntHideErr(errList []error, arg interface{}) int{
    i, e := IntParse(arg)
    if e!= nil {
        errList = append(errList, e)
    }
    return i
}

func Int64HideErr(errList []error, arg interface{}) int64{
    i, e := Int64Parse(arg)
    if e!= nil {
        errList = append(errList, e)
    }
    return i
}

// 转为int型  --> INT
func IntParse(arg interface{}) (ret int, err error){
    if arg == nil {
        return 0, nil
    }
    switch arg.(type) {
    case int:
        ret = arg.(int)
    case int32:
        ret = int(arg.(int32))
    case int64:
        ret = int(arg.(int64))
    case string:
        ret, err = strconv.Atoi(arg.(string))
    case float32:
        ret = int(arg.(float32))
    case float64:
        ret = int(arg.(float64))
    default:
        err = ecode.UtilCanNotBeInt.Error()
    }
    return
}

// 转为INT64型  -->INT64
func Int64Parse(arg interface{}) (ret int64, err error){
    if arg == nil {
        return 0, nil
    }
    switch arg.(type) {
    case int:
        ret = int64(arg.(int))
    case int32:
        ret = int64(arg.(int32))
    case int64:
        ret = arg.(int64)
    case string:
        ret, err = strconv.ParseInt(arg.(string), 10, 64)
    case float32:
        ret = int64(arg.(float32))
    case float64:
        ret = int64(arg.(float64))
    default:
        err = ecode.UtilCanNotBeInt64.Error()
    }
    return
}
