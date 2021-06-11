package company

import (
  "github.com/jasonfeng1980/pg"
  "github.com/jasonfeng1980/pg/database/db"
)

type company struct {
  db.Orm
}

func Company() *company{
  ret := &company{}
  q, err := pg.MySQL.Get("DEMO")
  if err != nil {
    panic("数据库配置名称错误 ：DEMO" )
  }
  ret.Query = q
  ret.Name = "company"
  ret.Pk = "company_id"
  ret.Fields = []string{"company_name","company_money","create_at","update_at"}
  // filter 设置
  ret.Filter = make(map[string]*db.FilterConf)
  f := pg.Filter.MySQL
  ret.BeforeInsert = make(map[string]db.ChangeFunc)
  ret.BeforeUpdate = make(map[string]db.ChangeFunc)

  ret.Filter["company_id"] = f("bigint", true, false)
  ret.Filter["company_name"] = f("varchar(255)", false, false)
  ret.Filter["company_money"] = f("int", false, true)
  ret.Filter["create_at"] = f("datetime", false, false)
  ret.BeforeInsert["create_at"] = db.ChangeNow

  ret.Filter["update_at"] = f("datetime", false, false)
  ret.BeforeInsert["update_at"] = db.ChangeNow
  ret.BeforeUpdate["update_at"] = db.ChangeNow


  return ret
}
