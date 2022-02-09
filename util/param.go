package util

import (
    "strings"
    "time"
)

type Param struct {
    Box map[string]interface{}
}
func (p *Param)ParseToDTO() map[string]interface{} {
    return p.Box
}
func (p *Param)Delete(key string) {
    if p.Box == nil {
        p.Box = make(map[string]interface{})
    } else {
        delete(p.Box, key)
    }

}

func (p *Param)Set(s map[string]interface{}) {
    if p.Box == nil {
        p.Box = make(map[string]interface{})
    }
    p.Box = s
}

func (p *Param)Get(name string) (ret interface{}){
    if p.Box == nil {
        p.Box = make(map[string]interface{})
    }
    nList := strings.Split(name, ".")
    l := len(nList)
    if l==1 {
        if v, ok := p.Box[nList[0]]; ok {
            return v
        }
        return nil
    }
    // 循环分层取数据
    val := p.Box
    for k, n := range nList{
        if v, ok := val[n]; ok {
            if k == l -1 {
                return v
            }
            if InterfaceType(v) == "map[string]interface {}" {
                val = v.(map[string]interface{})
                continue
            }
        }
        return nil
    }
    return nil
}

func (p *Param)GetInt(name string, defs ...int) int{
    def := 0
    if len(defs) == 1{
        def = defs[0]
    }
    if v := p.Get(name); v!=nil {
        if ret, err := IntParse(v); err == nil{
            return ret
        }
    }
    return  def
}
func (p *Param)GetInt64(name string, defs ...int64) int64{
    def := int64(0)
    if len(defs) == 1{
        def = defs[0]
    }
    if v := p.Get(name); v!=nil {
        if ret, err := Int64Parse(v); err == nil{
            return ret
        }
    }
    return  def
}
func (p *Param) GetStr(name string, defs ...string) string{
    def := ""
    if len(defs) == 1{
        def = defs[0]
    }
    if v := p.Get(name); v!=nil {
        if ret, err := StrParse(v); err == nil{
            return ret
        }
    }

    return  def
}
func (p *Param)GetTimeDuration(name string, defs ...time.Duration) time.Duration{

    var def time.Duration
    if len(defs) == 1{
        def = defs[0]
    }
    if v := p.Get(name); v!=nil {
        ret, err := Int64Parse(v)
        if err == nil{
            return time.Duration(ret)
        }
    }
    return  def
}
func (p *Param)GetListString(name string, defs ...[]string) []string{

    var def []string
    if len(defs) == 1{
        def = defs[0]
    }
    if v := p.Get(name); v!=nil {
        ret, ok := v.([]string)
        if ok {
            return ret
        }
    }
    return  def
}