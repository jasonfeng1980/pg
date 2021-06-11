package company

import (
  "github.com/jasonfeng1980/pg"
  "github.com/jasonfeng1980/pg/database/db"
)

type company_member struct {
  db.Orm
}

func CompanyMember() *company_member{
  ret := &company_member{}
  q, err := pg.MySQL.Get("DEMO")
  if err != nil {
    panic("数据库配置名称错误 ：DEMO" )
  }
  ret.Query = q
  ret.Name = "company_member"
  ret.Pk = "company_member_id"
  ret.Fields = []string{"company_member_ctime","company_id","company_member_name","company_member_birthday","company_member_id_parent"}
  // filter 设置
  ret.Filter = make(map[string]*db.FilterConf)
  f := pg.Filter.MySQL
  ret.BeforeInsert = make(map[string]db.ChangeFunc)
  ret.BeforeUpdate = make(map[string]db.ChangeFunc)

  ret.Filter["company_member_birthday"] = f("date", false, false)
  ret.Filter["company_member_id_parent"] = f("int", false, false)
  ret.Filter["company_member_id"] = f("bigint", false, false)
  ret.Filter["company_member_ctime"] = f("datetime", false, false)
  ret.Filter["company_id"] = f("int", false, true)
  ret.Filter["company_member_name"] = f("varchar(255)", false, true)

  return ret
}
