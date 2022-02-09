package DAO

import (
    "context"
    "github.com/jasonfeng1980/pg/ddd"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
)

// 根据表名获得一个Dao对象， 可以通过DO初始化数据
func NewDemoDao(ctx context.Context, tableName string, DO ...map[string]interface{}) (*ddd.DAO, error){
    if _, ok := DatabaseDemo.TableMap[tableName]; !ok{
        return nil, ecode.DaoWrongTable.Error("demo", tableName)
    }
    var d = make(map[string]interface{})
    if len(DO) == 1 {
        d = DO[0]
    }
    dao := &ddd.DAO{
        Option: &ddd.Option{
            Ctx:          ctx,
            DBHandleName: "demo",
            DatabaseMap:  DatabaseDemo,
            TableName:    tableName,
        },
        DO: util.Param{
            d,
        },
        Params : util.Param{
            make(map[string]interface{}),
        },
    }
    dao.Conn()
    return dao, nil
}

// 数据mapper
var DatabaseDemo = &ddd.DataMap{
    TableMap: map[string]*ddd.Table{
        "company": {"demo", "company", "company_id", []string{"company_id","company_name","company_money","company_update_at","company_create_at"}},
        "company_member": {"demo", "company_member", "company_member_id", []string{"company_member_id","company_member_ctime","company_id","company_member_name","company_member_birthday","company_member_id_parent"}},
    },
    FieldMap: map[string]*ddd.Field{
        "company_name": {"company","company_name", "varchar(50)", false, false},
        "company_money": {"company","company_money", "int(11)", false, false},
        "company_update_at": {"company","company_update_at", "datetime", false, false},
        "company_create_at": {"company","company_create_at", "datetime", false, false},
        "company_id": {"company","company_id", "bigint(20)", true, false},
        "company_member_id": {"company_member","company_member_id", "bigint(20)", false, false},
        "company_member_ctime": {"company_member","company_member_ctime", "datetime", false, false},
        "company_member_name": {"company_member","company_member_name", "varchar(255)", false, true},
        "company_member_birthday": {"company_member","company_member_birthday", "date", false, false},
        "company_member_id_parent": {"company_member","company_member_id_parent", "int(11)", false, false},
    },
    RelationMap: map[string]map[string][]*ddd.RelationKV{
        "company" : {"company_member": {{"company_id", "company_id"}}},
    },
}

