package aggregate
/**
    聚合根   -  可以理解为 能力
    尽量设计最小的功能
    高度内聚
    全局唯一的标识
    操作聚合的一组实体
    封装真正的不变性
    聚合中的实体和值对象具有相同的生命周期，并应该属于一个业务场景

    e.g.    转账 只设计 进账 和 出账，  不在聚合根里提供转账的组合行为
    e.g.    电视机的不同实体的能力     二级管发光  放大器调整信号 喇叭播放声音
 */

import (
    "context"
    "fmt"
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/example/ecode"
    "github.com/jasonfeng1980/pg/example/entity/demoEntity"
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/ddd"
    "github.com/jasonfeng1980/pg/util"
)

// 通过GUID 获取聚合根
func NewCompanyRootFromGuid(ctx context.Context, guid string) *CompanyRoot {
    prefix := "COMPANY"
    pk, err := ddd.GetPkFromGuid(prefix, guid)
    if err != nil {
        // 隐藏错误：退出，直接返回错误码
        util.HideErrorCancel(ctx, err)
        return nil
    }
    return &CompanyRoot{
        CompanyEntity:demoEntity.NewCompanyEntity(ctx, pk),
        AggregateRoot:ddd.NewAggregateRoot(prefix, pk),
    }
}

// 通过PK获取聚合根
func NewCompanyRootFromPK(ctx context.Context, pk int64) *CompanyRoot {
    prefix := "COMPANY"
    return &CompanyRoot{
        CompanyEntity:demoEntity.NewCompanyEntity(ctx, pk),
        AggregateRoot:ddd.NewAggregateRoot(prefix, pk),
    }
}
type CompanyRoot struct {
    *demoEntity.CompanyEntity

    *ddd.AggregateRoot
}

//////////////////////////////////////////
//  聚合行为
//////////////////////////////////////////
// 收入，进账
func (o *CompanyRoot) Income(ctx context.Context, companyId int64, money int) (interface{}, int64, string){
    // 1. 检查money
    if money <=0 {
        return ecode.CompanyWrongMoney.Parse(money)
    }
    // 2. 给公司入账
    err := o.SetCompanyMoney(db.Expr(fmt.Sprintf("company_money+%d", money)))
    if err != nil { // 出现赋值错误
        return pg.Err(err)
    }
    rows, err := o.Edit().Result()
    if err != nil { // 编辑失败
        return pg.Err(err)
    }

    // 返回
    return pg.Suc(pg.M{"rows": rows})
}

// 支出，出账
func (o *CompanyRoot) Expend(ctx context.Context, companyId int64, money int) (interface{}, int64, string){
    // 1. 检查参数
    if money <=0 {
        return ecode.CompanyWrongMoney.Parse(money)
    }

    // 3. 给公司出账
    err := o.SetCompanyMoney(db.Expr("company_money-%d", money))
    if err != nil { // 出现赋值错误
        return pg.Err(err)
    }
    rows, err := o.Edit().Where("company_money", pg.M{"$gte": money}).Result()
    if err != nil { // 编辑出现错误
        return pg.Err(err)
    }
    if rows.(int64) == 0 { // 钱不够
        return ecode.CompanyMoneyNotEnough.Parse(money)
    }
    // 返回
    return pg.Suc(pg.M{"rows": rows})
}