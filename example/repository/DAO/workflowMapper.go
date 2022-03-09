package DAO

import (
    "context"
    "github.com/jasonfeng1980/pg/ddd"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
)

// 根据表名获得一个Dao对象， 可以通过DO初始化数据
func NewWorkflowDao(ctx context.Context, tableName string, DO ...map[string]interface{}) (*ddd.DAO, error){
    if _, ok := DatabaseWorkflow.TableMap[tableName]; !ok{
        return nil, ecode.DaoWrongTable.Error("workflow", tableName)
    }
    var d = make(map[string]interface{})
    if len(DO) == 1 {
        d = DO[0]
    }
    dao := &ddd.DAO{
        Option: &ddd.Option{
            Ctx:          ctx,
            DBHandleName: "workflow",
            DatabaseMap:  DatabaseWorkflow,
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
var DatabaseWorkflow = &ddd.DataMap{
    TableMap: map[string]*ddd.Table{
        "enter_warehouse": {"workflow", "enter_warehouse", "enter_warehouse_id", []string{"enter_warehouse_id","workflow_tpl_code","product_id","EW1000","EW2000","EW3000","EW4100","EW4200","EW6000","enter_warehouse_end","created_at","op_id","version","enter_warehouse_deleted"}, []string{"workflow_tpl_code"}},
        "enter_warehouse_log": {"workflow", "enter_warehouse_log", "enter_warehouse_log_id", []string{"enter_warehouse_log_id","workflow_tpl_code","enter_warehouse_id","product_id","docket_code","enter_warehouse_log_params","enter_warehouse_log_rollback","created_at","op_id","version","enter_warehouse_log_deleted"}, []string{"docket_code","enter_warehouse_log_params","enter_warehouse_log_rollback"}},
        "enter_warehouse_node": {"workflow", "enter_warehouse_node", "enter_warehouse_node_id", []string{"enter_warehouse_node_id","workflow_tpl_code","enter_warehouse_id","product_id","node_code","node_status","node_is_end","enter_warehouse_node_created_at","enter_warehouse_node_updated_at","op_id","version","enter_warehouse_node_deleted"}, []string{"node_code","node_status"}},
    },
    FieldMap: map[string]*ddd.Field{
        "created_at": {"enter_warehouse","created_at", "timestamp", false, false},
        "workflow_tpl_code": {"enter_warehouse","workflow_tpl_code", "varchar(255)", false, true},
        "enter_warehouse_end": {"enter_warehouse","enter_warehouse_end", "tinyint", false, false},
        "enter_warehouse_deleted": {"enter_warehouse","enter_warehouse_deleted", "tinyint", false, false},
        "enter_warehouse_id": {"enter_warehouse","enter_warehouse_id", "int", true, false},
        "product_id": {"enter_warehouse","product_id", "int", false, false},
        "EW4100": {"enter_warehouse","EW4100", "varchar(255)", false, false},
        "EW6000": {"enter_warehouse","EW6000", "varchar(255)", false, false},
        "version": {"enter_warehouse","version", "int", false, false},
        "EW1000": {"enter_warehouse","EW1000", "varchar(255)", false, false},
        "EW3000": {"enter_warehouse","EW3000", "varchar(255)", false, false},
        "op_id": {"enter_warehouse","op_id", "int", false, false},
        "EW2000": {"enter_warehouse","EW2000", "varchar(255)", false, false},
        "EW4200": {"enter_warehouse","EW4200", "varchar(255)", false, false},
        "docket_code": {"enter_warehouse_log","docket_code", "varchar(255)", false, true},
        "enter_warehouse_log_deleted": {"enter_warehouse_log","enter_warehouse_log_deleted", "tinyint", false, false},
        "enter_warehouse_log_id": {"enter_warehouse_log","enter_warehouse_log_id", "int", true, false},
        "enter_warehouse_log_params": {"enter_warehouse_log","enter_warehouse_log_params", "text", false, true},
        "enter_warehouse_log_rollback": {"enter_warehouse_log","enter_warehouse_log_rollback", "text", false, true},
        "node_is_end": {"enter_warehouse_node","node_is_end", "tinyint", false, false},
        "enter_warehouse_node_updated_at": {"enter_warehouse_node","enter_warehouse_node_updated_at", "timestamp", false, false},
        "enter_warehouse_node_deleted": {"enter_warehouse_node","enter_warehouse_node_deleted", "tinyint", false, false},
        "enter_warehouse_node_id": {"enter_warehouse_node","enter_warehouse_node_id", "int", true, false},
        "node_code": {"enter_warehouse_node","node_code", "varchar(255)", false, true},
        "node_status": {"enter_warehouse_node","node_status", "varchar(255)", false, true},
        "enter_warehouse_node_created_at": {"enter_warehouse_node","enter_warehouse_node_created_at", "timestamp", false, false},
    },
    RelationMap: map[string]map[string][]*ddd.RelationKV{
        "enter_warehouse" : {"enter_warehouse_log": {{"enter_warehouse_id", "enter_warehouse_id"}},"enter_warehouse_node": {{"enter_warehouse_id", "enter_warehouse_id"}}},
        "enter_warehouse_log" : {"enter_warehouse": {{"enter_warehouse_log_id", "enter_warehouse_log_id"}}},
        "enter_warehouse_node" : {"enter_warehouse": {{"enter_warehouse_node_id", "enter_warehouse_node_id"}}},
    },
}

