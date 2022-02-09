package demoEntity

import (
    "context"
    "github.com/jasonfeng1980/pg/example/repository/DAO"
    
    "github.com/jasonfeng1980/pg/ddd"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
)

// 获取实体对象
func NewCompanyMemberEntity(ctx context.Context, pk interface{}) *CompanyMemberEntity {
    dao, err := DAO.NewDemoDao(ctx, "company_member")
    if err != nil {
        util.Panic(ecode.DaoWrongTable.Error("demo", "company_member"))
    }
    o := &CompanyMemberEntity{
        &ddd.Entity{DAO: dao},
    }
    if v, _ := util.Int64Parse(pk); err == nil {
        o.SetPk(v)
    }
    return o
}

type CompanyMemberEntity struct {
    *ddd.Entity
}

//////////////////////////////////////////
//  关联实体
//////////////////////////////////////////


//////////////////////////////////////////
//  Getter Setter
//////////////////////////////////////////

// 获取 创建时间
func (o *CompanyMemberEntity) GetCompanyMemberCtime(def ...string) string{
    return o.AutoLoad("company_member_ctime").DO.GetStr("company_member_ctime", def[0])
}
// 设置 创建时间
func (o *CompanyMemberEntity) SetCompanyMemberCtime(value interface{}) (err error){
    return o.Set("company_member_ctime", value)
}


// 获取 公司id
func (o *CompanyMemberEntity) GetCompanyId(def ...int) int{
    return o.AutoLoad("company_id").DO.GetInt("company_id", def[0])
}
// 设置 公司id
func (o *CompanyMemberEntity) SetCompanyId(value interface{}) (err error){
    return o.Set("company_id", value)
}


// 获取 成员名称
func (o *CompanyMemberEntity) GetCompanyMemberName(def ...string) string{
    return o.AutoLoad("company_member_name").DO.GetStr("company_member_name", def[0])
}
// 设置 成员名称
func (o *CompanyMemberEntity) SetCompanyMemberName(value interface{}) (err error){
    return o.Set("company_member_name", value)
}


// 获取 成员生日
func (o *CompanyMemberEntity) GetCompanyMemberBirthday(def ...string) string{
    return o.AutoLoad("company_member_birthday").DO.GetStr("company_member_birthday", def[0])
}
// 设置 成员生日
func (o *CompanyMemberEntity) SetCompanyMemberBirthday(value interface{}) (err error){
    return o.Set("company_member_birthday", value)
}


// 获取 成员上级ID
func (o *CompanyMemberEntity) GetCompanyMemberIdParent(def ...int) int{
    return o.AutoLoad("company_member_id_parent").DO.GetInt("company_member_id_parent", def[0])
}
// 设置 成员上级ID
func (o *CompanyMemberEntity) SetCompanyMemberIdParent(value interface{}) (err error){
    return o.Set("company_member_id_parent", value)
}


// 获取 自增ID
func (o *CompanyMemberEntity) GetCompanyMemberId(def ...int64) int64{
    return o.AutoLoad("company_member_id").DO.GetInt64("company_member_id", def[0])
}
// 设置 自增ID
func (o *CompanyMemberEntity) SetCompanyMemberId(value interface{}) (err error){
    return o.Set("company_member_id", value)
}


