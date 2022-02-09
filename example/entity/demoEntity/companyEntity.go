package demoEntity
/**
    实体层<br>
    . 脱离数据库<br>
    . 基本的业务能力(实体内部的)<br>
    . 原则上不允许拥有跨实体的能力<br>
    . 继承DAO<br>
    . 常用的get set <br>
    . 系统生成在 tmp/repository里，检查合并到ORM里<
    e.g.   电视机的   二极管 喇叭 信号放大器
*/

import (
    "context"
    "github.com/jasonfeng1980/pg/example/repository/DAO"
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/ddd"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
)

// 获取实体对象
func NewCompanyEntity(ctx context.Context, pk ...interface{}) *CompanyEntity {
    dao, err := DAO.NewDemoDao(ctx, "company")
    if err != nil {
        util.Panic(ecode.DaoWrongTable.Error("demo", "company"))
    }
    o := &CompanyEntity{
        &ddd.Entity{DAO: dao},
    }
    if len(pk) == 0 {
        return o
    }
    if v, _ := util.Int64Parse(pk[0]); err == nil {
        o.SetPk(v)
    }
    return o
}

type CompanyEntity struct {
    *ddd.Entity
}

//////////////////////////////////////////
//  关联实体
//////////////////////////////////////////

func (o *CompanyEntity) RelationCompanyMember() *ddd.Query {
    q := o.Relation("company_member")
    q.ResultFunc = func(query *db.Query) (interface{},  error){
        doList, err := query.Query().Array()
        if err != nil {
            return nil, err
        }
        var ret []*CompanyMemberEntity
        relationPk := o.DatabaseMap.TableMap["company_member"].Pk
        for _, DO := range doList {
            relationEntity := NewCompanyMemberEntity(o.Ctx, DO[relationPk] )
            relationEntity.DO.Box = DO
            ret = append(ret, relationEntity)
        }
        return ret, err
    }
    return q
}


//////////////////////////////////////////
//  Getter Setter
//////////////////////////////////////////

// 获取 公司收益
func (o *CompanyEntity) GetCompanyMoney(def ...int) int{
    return o.AutoLoad("company_money").DO.GetInt("company_money", def[0])
}
// 设置 公司收益
func (o *CompanyEntity) SetCompanyMoney(value interface{}) (err error){
    return o.Set("company_money", value)
}


// 获取 更新时间
func (o *CompanyEntity) GetCompanyUpdateAt(def ...string) string{
    return o.AutoLoad("company_update_at").DO.GetStr("company_update_at", def[0])
}
// 设置 更新时间
func (o *CompanyEntity) SetCompanyUpdateAt(value interface{}) (err error){
    return o.Set("company_update_at", value)
}


// 获取 创建时间
func (o *CompanyEntity) GetCompanyCreateAt(def ...string) string{
    return o.AutoLoad("company_create_at").DO.GetStr("company_create_at", def[0])
}
// 设置 创建时间
func (o *CompanyEntity) SetCompanyCreateAt(value interface{}) (err error){
    return o.Set("company_create_at", value)
}


// 获取 公司ID
func (o *CompanyEntity) GetCompanyId(def ...int64) int64{
    return o.AutoLoad("company_id").DO.GetInt64("company_id", def[0])
}
// 设置 公司ID
func (o *CompanyEntity) SetCompanyId(value interface{}) (err error){
    return o.Set("company_id", value)
}


// 获取 公司名称
func (o *CompanyEntity) GetCompanyName(def ...string) string{
    return o.AutoLoad("company_name").DO.GetStr("company_name", def[0])
}
// 设置 公司名称
func (o *CompanyEntity) SetCompanyName(value interface{}) (err error){
    return o.Set("company_name", value)
}


