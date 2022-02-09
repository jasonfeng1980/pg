package demo

import (
	"context"
	"github.com/jasonfeng1980/pg"
	"github.com/jasonfeng1980/pg/util"
)

func init(){
	version := "v3"
	api := pg.MicroApi()
	api.Register("POST", "company", "trans", version, CompanyTransV3)
}


// 调用转账，并返回双方的信息  -- 例子
func CompanyTransV3(ctx context.Context, params *util.Param)(data interface{}, code int64, msg string) {
	// 获得调用客户端实例
	client, _ := pg.Client()
	defer client.Close()

	// 1. 转账
	data, code, msg = client.Call(ctx, "grpc://pg_demo/company/trans/v2",params.ParseToDTO())
	if code != 200 { // 不成功直接返回
		return
	}

	// 获得group实例
	g   := util.Group()

	// 2. 并行 - 获取消费公司信息
	g.Add("company_expand", func() (data interface{}, code int64, msg string) {
		return client.Call(ctx,
			"http://pg_demo/company/info/v1",
			pg.M{"company_id":params.GetInt64("company_id_expand")})
	})
	// 3. 并行 - 获取消费公司信息
	g.Add("company_income", func() (data interface{}, code int64, msg string) {
		return client.Call(ctx,
			"grpc://pg_demo/company/info/v1",
			pg.M{"company_id":params.GetInt64("company_id_income")})
	})
	// 等待并行协程执行完毕
	g.Wait()

	// 返回
	return pg.Suc(pg.M{
		"trans": data,
		"company_expand": g.Get("company_expand").Data,
		"company_income": g.Get("company_income").Data,
	})
}
