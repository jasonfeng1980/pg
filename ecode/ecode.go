package ecode

import (
    "errors"
    "fmt"
    "strconv"
    "strings"
)

type Ecode struct {
    Code int64
    Msg  string
}
func(e *Ecode) Error(args ...interface{}) error{
    return errors.New(fmt.Sprintf(strconv.FormatInt(e.Code, 10)+ ":" + e.Msg, args...))
}
func(e *Ecode) Parse(args ...interface{}) (interface{}, int64, string){
    return nil, e.Code, fmt.Sprintf(e.Msg, args...)
}

func Err(code int64, msg string) *Ecode{
    return &Ecode{code, msg}
}
func ReadError(e error) (int64, string) {
    s := e.Error()
    l := strings.SplitN(s, ":", 2)

    if len(l) ==2 {
        if i64, err := strconv.ParseInt(l[0], 10, 64); err == nil {
            return i64, l[1]
        }
    }
    return 200, s
}

var (
    // 系统        [1000,1100)
    UtilCanNotBeInt = Err(1001, "无法转换成INT")
    UtilCanNotBeInt64 = Err(1002, "无法转换成INT64")
    UtilWrongDataType = Err(1003, "错误的数据格式")
    UtilMissNeedField = Err(1004, "缺少必需的字段：%s")
    UtilCanNotBeString = Err(1005, "无法转换成string")
    UtilNoUploadFile = Err(1006, "没有需要上传的文件")
    UtilWrongDir = Err(1007, "错误的文件目录:%s")


    // 浏览器请求   [1100,1200)
    HttpMissDns = Err(11, "缺少DNS")
    HttpCannotMatchDns = Err(1101, "没有无法匹配的DNS:%s/%s/%s")
    HttpUrlMissMVA = Err(1102, "dns缺少module,version,action")
    HttpDataNotMap = Err(1103, "传递的参数为空或者不是map[string]interface{}格式")
    HttpDnsParseWrong = Err(1104, "dns解析失败")

    // 数据库      [1200,1300)
    DbNotExecData = Err(1201, "没有result-exec数据")
    DbNotQueryData = Err(1202, "没有result-query数据")
    DbWrongType = Err(1203, "dest类型不对")
    DbColumnsNotMatch = Err(1204, "columns与地址不匹配")
    DbWrongMap = Err(1205, "错误的MAP格式，必须是 map[string]interface")

    RdbCannotToString = Err(1210, "redis-hash key:%s field:%s无法转换成字符串")
    RdbWrongData= Err(1211, "redis-hash数据格式不正确 名称%s")
    RdbWrongDecodeJoin= Err(1212, "redis decodeJoin 数据格式不匹配")

    OrmWrongArgType= Err(1220, "ORM 参数无法转换成[]map[string]interface{}")
    OrmWrongColumnsType = Err(1221, "ORM %s字段格式不对")
    OrmMissColumnsNeed = Err(1222, "ORM 缺少必填项：%s")

    // CONF  [1300, 1400)
    ConfYamlWrongFormat = Err(1301, "%s格式不正确")
    ConfYamlWrongMysql = Err(1302, "Mysql-yaml配置格式不正确")
    ConfYamlWrongRedis = Err(1303, "Redis-yaml配置格式不正确")




)
