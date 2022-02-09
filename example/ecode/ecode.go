package ecode

import "github.com/jasonfeng1980/pg"

var (
	// company 域的错误 [10000, 11000)
	CompanyWrongMoney = pg.ECode(10001, "公司金额必须为大于0:【传递的是%d】")
	CompanyMoneyNotEnough = pg.ECode(10002, "公司余额不足:【小于%d】")

)

