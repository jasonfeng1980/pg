package util

import (
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "github.com/jasonfeng1980/pg/ecode"
    "strconv"
    "strings"
    "unicode/utf8"
)


// 转换成字符串
func Str(arg interface{}) string {
    return fmt.Sprint(arg)
}
// 数字|bool|字符串 =>字符串 其他错误
func StrParse(arg interface{})(ret string, err error){
    if arg == nil {
        return "", nil
    }
    switch arg.(type) {
    case string:
        ret = arg.(string)
    case int:
        ret = strconv.Itoa(arg.(int))
    case int64:
        ret = strconv.FormatInt(arg.(int64), 10)
    case float32:
        ret = strconv.FormatFloat(float64(arg.(float32)), 'G', -1, 32)
    case float64:
        ret = strconv.FormatFloat(arg.(float64), 'G', -1, 64)
    case bool:
        if arg.(bool) == true {
            ret = "true"
        } else {
            ret = "false"
        }
    default:
        err = ecode.UtilCanNotBeString.Error()
    }
    return
}

// []int64 => []string
func StrListInt(list []int64) []string {
    box := make([]string, len(list))
    for k, v := range list {
        box[k] = Str(v)
    }
    return box
}

// 首字母大写
func StrUFirst(str string) string {
    if len(str) < 1 {
        return ""
    }
    strArr := []rune(str)
    if strArr[0] >= 97 && strArr[0] <= 122 {
        strArr[0] -= 32
    }
    return string(strArr)
}

// 将以指定字符分割的字符串，变成全单词首字母大写
// e.g.  hello_world   ==>  HelloWorld
func StrUFirstForSplit(str string, spe string) string{
    arr := strings.Split(str, "_")
    ret := ""
    for _, v := range arr{
        ret += StrUFirst(v)
    }
    return ret
}
// 骆驼命名法
// e.g.  hello_world   ==>  helloWorld
func StrSecFirstForSplit(str string, spe string) string{
    str = StrUFirstForSplit(str, spe)
    return strings.ToLower(str[0:1]) + str[1:]
}

func StrSubStr(s string,begin int , len int) string{
    r := []rune(s)
    return string(r[begin:begin+len])
}

// 获取字符串格式，中文算一个
// "中国2" => 3
func StrLen(str string) int{
    return utf8.RuneCountInString(str)
}

// 字符串 MD5加密
func StrMd5(str string) string  {
    h := md5.New()
    h.Write([]byte(str))
    return hex.EncodeToString(h.Sum(nil))
}

// []int64 => string
// e.g. []int64{1,2,3,4} => "1,2,3,4"
func StrJoinFromInt64(iList []int64, sep string) string{
    strBox := make([]string, len(iList))
    for k, v := range iList {
        strBox[k] = Str(v)
    }
    return strings.Join(strBox, sep)
}

// 逗号隔开字符串 变成 数字数组
// "1,2,3,4" => []int64{1, 2, 3, 4}
func StrSplitToInt(str string) ([]int64, error) {
    if str == "" {
        return nil, nil
    }
    arr := strings.Split(str, ",")
    res := make([]int64, 0, len(arr))
    for _, v := range arr {
        if i, err := strconv.ParseInt(v, 10, 64); err != nil {
            return nil, err
        } else {
            res = append(res, i)
        }
    }
    return res, nil
}
