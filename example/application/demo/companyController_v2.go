package demo

import (
	"context"
	"github.com/jasonfeng1980/pg"
	"github.com/jasonfeng1980/pg/example/domain"
	"github.com/jasonfeng1980/pg/util"
)

func init(){
	version := "v2"
	api := pg.MicroApi()
	api.Register("POST", "company", "trans", version, CompanyTransV2)
}

// A公司转账到B公司
func CompanyTransV2(ctx context.Context, params *util.Param)(data interface{}, code int64, msg string) {
	companyA := params.GetInt64("company_id_expand", 0)
	companyB := params.GetInt64("company_id_income", 0)
	money := params.GetInt("money", 0)

	// 开启事务
	pg.StarTransaction(ctx)

	// 1. 请求转账服务 A 给 B 转钱
	if data, code, msg = domain.CompanyDomain.Trans(ctx, companyA, companyB, money); code != 200 {
		pg.Rollback(ctx)
		return
	}

	// 提交事务
	pg.Commit(ctx)
	// 返回
	return pg.Suc("转账成功")
}