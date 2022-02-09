package domain
/**
	领域服务  - 可以理解为  行为
	整合1个或者多个聚合根的能力，包装成不同的服务
	e.g.   电视机的不同行为：    调整声音大小   切换频道   画中画 录屏
 */

import (
	"context"
	"github.com/jasonfeng1980/pg"
	"github.com/jasonfeng1980/pg/example/aggregate"
	"github.com/jasonfeng1980/pg/ddd"
)

var CompanyDomain =  &companyDomain{}

type companyDomain struct {
	*ddd.Domain
}

// 获得公司信息
func (o *companyDomain) CompanyInfo(ctx context.Context, companyId int64) (interface{}, int64, string) {
	// 1. 获得company聚合根
	companyRoot := aggregate.NewCompanyRootFromPK(ctx, companyId)

	// 2. 返回实体的DO
	return pg.Suc(companyRoot.ParseToDO())
}

// 获得公司成员信息
func (o *companyDomain) MemberInfo(ctx context.Context, companyId int64, pageSize int, page int) (interface{}, int64, string){
	// 获得company聚合根
	companyRoot := aggregate.NewCompanyRootFromPK(ctx, companyId)

	/*  例子
	// 通过关联实体companyMemberEntity 再通过 ParseToDo 获取数据
	rs, err := companyEntity.RelationCompanyMember().Page(pageSize, page).Result()
	if err != nil {
		return pg.Err(err)
	}
	companyMemberEntityList := rs.([]*demoEntity.CompanyMemberEntity)
	 */

	// 2. 直接获得关联实体companyMember的数据
	companyMemberEntityList, err := companyRoot.RelationCompanyMember().Page(pageSize, page).Cache(true).Run().Array()
	if err != nil {
		return pg.Err(err)
	}
	// 返回
	return pg.Suc(pg.M{"company_member": companyMemberEntityList})
}

// 修改公司名称
func (o *companyDomain) ChangeName(ctx context.Context, companyId int64, newName string) (interface{}, int64, string){
	// 获得company聚合根
	companyRoot := aggregate.NewCompanyRootFromPK(ctx, companyId)

	// 1. 修改公司名称
	companyRoot.SetCompanyName(newName)
	rows, err := companyRoot.Edit().Result()
	if  err != nil {
		return pg.Err(err)
	}

	// 返回
	return pg.Suc(pg.M{"NewName": newName, "row": rows})
}

// 收入，进账
func (o *companyDomain) Income(ctx context.Context, companyId int64, money int) (interface{}, int64, string){
	// 获得公司聚合根
	companyRoot := aggregate.NewCompanyRootFromPK(ctx, companyId)

	// 1. 给公司入账
	return companyRoot.Income(ctx, companyId, money)
}

// 支出，出账
func (o *companyDomain) Expend(ctx context.Context, companyId int64, money int) (interface{}, int64, string){
	// 获得公司聚合根
	companyRoot := aggregate.NewCompanyRootFromPK(ctx, companyId)

	// 1. 给公司出账
	return companyRoot.Expend(ctx, companyId, money)
}

// 转账
func (o *companyDomain) Trans(ctx context.Context, companyA int64, companyB int64, money int) (data interface{}, code int64, msg string){
	/** 可以 复用领域服务
	// 1. 给A减钱
	if data, code, msg = o.Expend(ctx, companyA, money); code != 200 {
		return
	}
	// 2. 给B加钱
	return o.Income(ctx, companyB, money)
	 */
	// 也可以 复用聚合根能力

	// 1. 给A减钱
	if data, code, msg = aggregate.NewCompanyRootFromPK(ctx, companyA).Expend(ctx, companyA, money); code != 200 {
		return
	}
	// 2. 给B加钱
	return aggregate.NewCompanyRootFromPK(ctx, companyB).Income(ctx, companyB, money)
}