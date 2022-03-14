package workflowEntity

import (
	"context"
	"github.com/jasonfeng1980/pg/database/db"
	"github.com/jasonfeng1980/pg/ddd"
	"github.com/jasonfeng1980/pg/ecode"
	"github.com/jasonfeng1980/pg/example/repository/DAO"
	"github.com/jasonfeng1980/pg/util"
)

// 获取实体对象
func NewEnterWarehouseLogEntity(ctx context.Context, pk ...interface{}) *EnterWarehouseLogEntity {
	dao, err := DAO.NewWorkflowDao(ctx, "enter_warehouse_log")
	if err != nil {
		util.Panic(ecode.DaoWrongTable.Error("workflow", "enter_warehouse_log"))
	}
	o := &EnterWarehouseLogEntity{
		&ddd.Entity{DAO: dao},
	}
	if len(pk) == 0 { // 如果是创建，就不用传PK
		return o
	}
	// 传了PK， 只取第一个，转换为int64
	if v, _ := util.Int64Parse(pk[0]); err == nil {
		o.SetPk(v)
	}
	return o
}

type EnterWarehouseLogEntity struct {
	*ddd.Entity
}

//////////////////////////////////////////
//  关联实体
//////////////////////////////////////////

func (o *EnterWarehouseLogEntity) RelationEnterWarehouse() *ddd.Query {
	q := o.Relation("enter_warehouse")
	q.ResultFunc = func(query *db.Query) (interface{}, error) {
		doList, err := query.Query().Array()
		if err != nil {
			return nil, err
		}
		var ret []*EnterWarehouseEntity
		relationPk := o.DatabaseMap.TableMap["enter_warehouse"].Pk
		for _, DO := range doList {
			relationEntity := NewEnterWarehouseEntity(o.Ctx, DO[relationPk])
			relationEntity.DO.Box = DO
			ret = append(ret, relationEntity)
		}
		return ret, err
	}
	return q
}

//////////////////////////////////////////
//  Getter Setter
//////////////////////////////////////////

// 获取 商品ID
func (o *EnterWarehouseLogEntity) GetProductId(def ...int) int {
	return o.AutoLoad("product_id").DO.GetInt("product_id", def...)
}

// 设置 商品ID
func (o *EnterWarehouseLogEntity) SetProductId(value interface{}) (err error) {
	return o.Set("product_id", value)
}

// 获取 单据code
func (o *EnterWarehouseLogEntity) GetDocketCode(def ...string) string {
	return o.AutoLoad("docket_code").DO.GetStr("docket_code", def...)
}

// 设置 单据code
func (o *EnterWarehouseLogEntity) SetDocketCode(value interface{}) (err error) {
	return o.Set("docket_code", value)
}

// 获取 是否是删除: 0不是|1是
func (o *EnterWarehouseLogEntity) GetEnterWarehouseLogDeleted(def ...int) int {
	return o.AutoLoad("enter_warehouse_log_deleted").DO.GetInt("enter_warehouse_log_deleted", def...)
}

// 设置 是否是删除: 0不是|1是
func (o *EnterWarehouseLogEntity) SetEnterWarehouseLogDeleted(value interface{}) (err error) {
	return o.Set("enter_warehouse_log_deleted", value)
}

// 获取 自增ID
func (o *EnterWarehouseLogEntity) GetEnterWarehouseLogId(def ...int) int {
	return o.AutoLoad("enter_warehouse_log_id").DO.GetInt("enter_warehouse_log_id", def...)
}

// 设置 自增ID
func (o *EnterWarehouseLogEntity) SetEnterWarehouseLogId(value interface{}) (err error) {
	return o.Set("enter_warehouse_log_id", value)
}

// 获取 工作流模板code
func (o *EnterWarehouseLogEntity) GetWorkflowTplCode(def ...string) string {
	return o.AutoLoad("workflow_tpl_code").DO.GetStr("workflow_tpl_code", def...)
}

// 设置 工作流模板code
func (o *EnterWarehouseLogEntity) SetWorkflowTplCode(value interface{}) (err error) {
	return o.Set("workflow_tpl_code", value)
}

// 获取 工作流ID
func (o *EnterWarehouseLogEntity) GetEnterWarehouseId(def ...int) int {
	return o.AutoLoad("enter_warehouse_id").DO.GetInt("enter_warehouse_id", def...)
}

// 设置 工作流ID
func (o *EnterWarehouseLogEntity) SetEnterWarehouseId(value interface{}) (err error) {
	return o.Set("enter_warehouse_id", value)
}

// 获取 操作人
func (o *EnterWarehouseLogEntity) GetOpId(def ...int) int {
	return o.AutoLoad("op_id").DO.GetInt("op_id", def...)
}

// 设置 操作人
func (o *EnterWarehouseLogEntity) SetOpId(value interface{}) (err error) {
	return o.Set("op_id", value)
}

// 获取 版本ID
func (o *EnterWarehouseLogEntity) GetVersion(def ...int) int {
	return o.AutoLoad("version").DO.GetInt("version", def...)
}

// 设置 版本ID
func (o *EnterWarehouseLogEntity) SetVersion(value interface{}) (err error) {
	return o.Set("version", value)
}

// 获取 回滚数据
func (o *EnterWarehouseLogEntity) GetEnterWarehouseLogParams(def ...string) string {
	return o.AutoLoad("enter_warehouse_log_params").DO.GetStr("enter_warehouse_log_params", def...)
}

// 设置 回滚数据
func (o *EnterWarehouseLogEntity) SetEnterWarehouseLogParams(value interface{}) (err error) {
	return o.Set("enter_warehouse_log_params", value)
}

// 获取 回滚数据
func (o *EnterWarehouseLogEntity) GetEnterWarehouseLogRollback(def ...string) string {
	return o.AutoLoad("enter_warehouse_log_rollback").DO.GetStr("enter_warehouse_log_rollback", def...)
}

// 设置 回滚数据
func (o *EnterWarehouseLogEntity) SetEnterWarehouseLogRollback(value interface{}) (err error) {
	return o.Set("enter_warehouse_log_rollback", value)
}

// 获取 操作时间
func (o *EnterWarehouseLogEntity) GetCreatedAt(def ...int) int {
	return o.AutoLoad("created_at").DO.GetInt("created_at", def...)
}

// 设置 操作时间
func (o *EnterWarehouseLogEntity) SetCreatedAt(value interface{}) (err error) {
	return o.Set("created_at", value)
}
