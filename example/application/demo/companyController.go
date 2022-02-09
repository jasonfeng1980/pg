package demo
/**
	应用服务层    为了和领域服务区分，用Controller后缀
	梳理不同的业务逻辑，真正对外提供服务
	e.g.   电视机的不同的应用：   机身遥控按键     手持遥控器     微信遥控器
 */

import (
	"context"
	"github.com/jasonfeng1980/pg"
	"github.com/jasonfeng1980/pg/example/domain"
	"github.com/jasonfeng1980/pg/util"
)

// 注册api
func init() {
	version := "v1"
	api := pg.MicroApi()
	api.Register("GET", "company", "info", version, CompanyInfo)
	api.Register("GET", "company", "member", version, CompanyMemberInfo)
	api.Register("POST", "company", "rename", version, CompanyChangeName)
	api.Register("POST", "company", "trans", version, CompanyTrans)
}

// 获得公司的成员信息
func CompanyInfo(ctx context.Context, params *util.Param)(data interface{}, code int64, msg string) {
	// 整理参数
	companyId := params.GetInt64("company_id")
	// 1. 获取关联实体-公司成员的数据
	return domain.CompanyDomain.CompanyInfo(ctx, companyId)
}

// 获得公司的成员信息
func CompanyMemberInfo(ctx context.Context, params *util.Param)(data interface{}, code int64, msg string) {
	// 整理参数
	companyId := params.GetInt64("company_id")
	pageSize  := params.GetInt("pageSize", 20)
	page  := params.GetInt("page", 1)
	// 1. 获取关联实体-公司成员的数据
	return domain.CompanyDomain.MemberInfo(ctx, companyId, pageSize, page)
}
// 修改公司名称
func CompanyChangeName(ctx context.Context, params *util.Param)(data interface{}, code int64, msg string) {
	// 1. 改变名称
	return domain.CompanyDomain.ChangeName(ctx,
		params.GetInt64("company_id"),
		params.GetStr("company_name"))

}

// A公司转账到B公司
func CompanyTrans(ctx context.Context, params *util.Param)(data interface{}, code int64, msg string) {
	companyA := params.GetInt64("company_id_expand", 0)
	companyB := params.GetInt64("company_id_income", 0)
	money := params.GetInt("money", 0)

	// 开启事务
	pg.StarTransaction(ctx)
	// 1. 添加转账日志
		// ....
	// 2. 给A减钱
	if data, code, msg = domain.CompanyDomain.Expend(ctx, companyA, money); code != 200 {
		pg.Rollback(ctx)
		return
	}
	// 3. 给B加钱
	if data, code, msg = domain.CompanyDomain.Income(ctx, companyB, money); code != 200 {
		pg.Rollback(ctx)
		return
	}

	// 提交事务
	pg.Commit(ctx)

	// 返回
	return pg.Suc("转账成功")
}

