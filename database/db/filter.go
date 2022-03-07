package db

import (
    "github.com/jasonfeng1980/pg/util"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type Filter struct {
}

type FilterFunc  func(check interface{}, conf *FilterConf)bool

type FilterConf struct{
    Name FilterFunc
    Min  int64
    Max  int64
    LenMin int
    LenMax int
    MaxStr string
    Reg  string
    Msg  string
    Pass bool
    Need bool
    Enum []string
}
// 设置最长长度
func (f *FilterConf)SetLenMax(lenMax int) *FilterConf{
    f.LenMax = lenMax
    return f
}

// 设置必填
func (f *FilterConf)SetNeed(need bool) *FilterConf{
    f.Need = need
    return f
}

// 判断是否必填
func (f *FilterConf)IsNeed() bool{
    return f.Need
}

func (f *FilterConf)Check(arg interface{})bool{
    return f.Name(arg, f)
}

func (f *Filter)MySQL(mysqlType string, unsigned bool, isNeed bool)  *FilterConf{
    var (
    	check string
    	enum  []string
    )
    if unsigned {
        check = "u-" + mysqlType
    } else {
        check = mysqlType
    }
    // 获取 varchar(255) 里的 255
    reg, _ := regexp.Compile("(\\d+)")
    lenMax := reg.FindString(mysqlType)
    if len(lenMax) > 0 {
        check = strings.Replace(check, "("+ lenMax + ")", "", 1)
    }
    // 获取 enum('secondhand','designer','resell','plum','reclo','big_customer','vip','brandnew','zhongxin')
    if len(mysqlType)>=4 && mysqlType[0:4] == "enum" {
        check = "enum"
        enum = strings.Split(strings.ReplaceAll(mysqlType[5:len(mysqlType)-1], "'", ""), ",")
    }

    switch check  {
    case "enum":
        return &FilterConf{Name: FilterFunc(f.CheckEnum), Enum: enum}
    case "int" :
        return &FilterConf{Name:FilterFunc(f.CheckInt),  Min:-2147483648, Max:2147483647, Need:isNeed}
    case "u-int":
        return &FilterConf{Name:FilterFunc(f.CheckInt),  Min:0, Max:4294967295, Need:isNeed}

    case "bigint":
        return &FilterConf{Name:FilterFunc(f.CheckInt),  Min:-9223372036854775808, Max:9223372036854775807, Need:isNeed}
    case "u-bigint":
        return &FilterConf{Name:FilterFunc(f.CheckUBigint), Need:isNeed}

    case "mediumint":
        return &FilterConf{Name:FilterFunc(f.CheckInt),  Min:-8388608, Max:8388607, Need:isNeed}
    case "u-mediumint":
        return &FilterConf{Name:FilterFunc(f.CheckInt),  Min:0, Max:16777215, Need:isNeed}

    case "smallint":
        return &FilterConf{Name:FilterFunc(f.CheckInt),  Min:-32768, Max:32767, Need:isNeed}
    case "u-smallint":
        return &FilterConf{Name:FilterFunc(f.CheckInt),  Min:0, Max:65535, Need:isNeed}

    case "tinyint":
        return &FilterConf{Name:FilterFunc(f.CheckInt),  Min:-128, Max:127, Need:isNeed}
    case "u-tinyint":
        return &FilterConf{Name:FilterFunc(f.CheckInt),  Min:0, Max:255, Need:isNeed}

    case "date":
        return &FilterConf{Name:FilterFunc(f.CheckDate), Need:isNeed}
    case "datetime":
        return &FilterConf{Name:FilterFunc(f.CheckDateTime), Need:isNeed}

    case "timestamp":
        return &FilterConf{Name:FilterFunc(f.CheckInt), Min:0, Max:2147483647, Need:isNeed}

    case "char", "varchar":
        if lenMax == "" {
            return &FilterConf{Name:FilterFunc(f.CheckStr), LenMin: 0}  // lenMax 使用者配置
        } else if lenMaxInt, err := strconv.Atoi(lenMax); err==nil {
            return &FilterConf{Name:FilterFunc(f.CheckStr), LenMin: 0, LenMax: lenMaxInt, Need:isNeed}  // lenMax 使用者配置
        }

    case "text", "json":
        return &FilterConf{Name:FilterFunc(f.CheckStr), LenMin: 0, Pass: true, Need:isNeed}  // lenMax 使用者配置
    case "logblob":
        return &FilterConf{Name:FilterFunc(f.CheckPass), Pass: true, Need:isNeed}  // lenMax 使用者配置
    }

    return nil
}

// 检测枚举型
func (f *Filter)CheckEnum(check interface{}, conf *FilterConf) bool{
    if ret, err := util.StrParse(check); err == nil{
        if util.ListHave(conf.Enum, ret){
            return true
        }
    }
    return false
}
// 检测整型
func (f *Filter)CheckInt(check interface{}, conf *FilterConf) bool{
    if ret, err := util.Int64Parse(check); err == nil{
        if ret >=conf.Min && ret<=conf.Max {
            return true
        }
    }
    return false
}

// 检测字符串
func (f *Filter)CheckStr(check interface{}, conf *FilterConf) bool{
    if ret, err := util.StrParse(check); err == nil{
        if conf.Pass { // 如果直接通过，就不检测
            return true
        }
        // 获取字符串的长度， 字符数，中文算1个
        strLen := util.StrLen(ret)
        if strLen >=conf.LenMin && (conf.LenMax == 0 ||  strLen<=conf.LenMax) {
            return true
        }
    }
    return false
}

// 检测日期
func (f *Filter)CheckDate(check interface{}, conf *FilterConf) bool{
    if ret, err := util.StrParse(check); err == nil{
        if _, err := time.Parse("2006-01-02 15:04:05", ret + " 00:00:00"); err==nil {
            return true
        }
    }
    return false
}

// 检测时间
func (f *Filter)CheckDateTime(check interface{}, conf *FilterConf) bool{
    if ret, err := util.StrParse(check); err == nil{
        if _, err := time.Parse("2006-01-02 15:04:05", ret); err==nil {
            return true
        }
    }
    return false
}

// 检测Ubigint
func (f *Filter)CheckUBigint(check interface{}, conf *FilterConf) bool{
    regx, _:=regexp.Compile("\\d{0,20}")
    if ret, err := util.StrParse(check); err == nil{ // 字符串判断
        if regx.Match([]byte(ret)){ // 正则判断
            if len(ret) == 20 && ret > "18446744073709551615" { // 超过最大的uint64
                return false
            } else {
                return true
            }
        }
    }
    return false
}

// 检测正则
func (f *Filter)CheckRegx(check interface{}, conf *FilterConf) bool{
    if regx, err:=regexp.Compile(conf.Reg); err==nil{
        if ret, err := util.StrParse(check); err == nil{
            return regx.Match([]byte(ret))
        }
    }
    return false
}

// 检测-直接通过
func (f *Filter)CheckPass(check interface{}, conf *FilterConf) bool{
    return true
}

