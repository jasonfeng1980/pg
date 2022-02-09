package ddd
/**
  聚合根   -  可以理解为 能力
  尽量设计最小的功能
  高度内聚
  全局唯一的标识
  操作聚合的一组实体
  封装真正的不变性
  聚合中的实体和值对象具有相同的生命周期，并应该属于一个业务场景

  e.g.    转账 只设计 进账 和 出账，  不在聚合根里提供转账的组合行为
  e.g.    电视机的不同能力     声音播放  图像显示   频道播放  开机动画  关机提示
*/
import (
    "fmt"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
)

// 通过guid活动本地主键
func GetPkFromGuid(prefix string, guid string)(int64, error){
    l := len(prefix)
    if guid[0:l] == prefix {
        return util.Int64Parse(guid[l-1:])
    }
    return 0, ecode.AggregateWrongGuid.Error(prefix, guid)
}

// 获取一个聚合根基类
func NewAggregateRoot(prefix string, pk int64) *AggregateRoot {
    return &AggregateRoot{
        prefix: prefix,
        pk:     pk,
    }
}

type AggregateRoot struct {
    pk     int64     // 本地唯一标识
    prefix string    // 标识前缀
}
// 获取全局唯一标识
func (a *AggregateRoot) GetGuid() string{
    return fmt.Sprintf("%s%d", a.prefix, a.pk)
}

