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
func NewEnterWarehouseNodeEntity(ctx context.Context, pk ...interface{}) *EnterWarehouseNodeEntity {
	dao, err := DAO.NewWorkflowDao(ctx, "enter_warehouse_node")
	if err != nil {
		util.Panic(ecode.DaoWrongTable.Error("workflow", "enter_warehouse_node"))
	}
	o := &EnterWarehouseNodeEntity{
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

type EnterWarehouseNodeEntity struct {
	*ddd.Entity
}

//////////////////////////////////////////
//  关联实体
//////////////////////////////////////////

func (o *EnterWarehouseNodeEntity) RelationEnterWarehouse() *ddd.Query {
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

// 获取 工作流模板code
func (o *EnterWarehouseNodeEntity) GetWorkflowTplCode(def ...string) string {
	return o.AutoLoad("workflow_tpl_code").DO.GetStr("workflow_tpl_code", def...)
}

// 设置 工作流模板code
func (o *EnterWarehouseNodeEntity) SetWorkflowTplCode(value interface{}) (err error) {
	return o.Set("workflow_tpl_code", value)
}

// 获取 是否是完结状态: 0不是|1是
func (o *EnterWarehouseNodeEntity) GetNodeIsEnd(def ...int) int {
	return o.AutoLoad("node_is_end").DO.GetInt("node_is_end", def...)
}

// 设置 是否是完结状态: 0不是|1是
func (o *EnterWarehouseNodeEntity) SetNodeIsEnd(value interface{}) (err error) {
	return o.Set("node_is_end", value)
}

// 获取 最后更新时间
func (o *EnterWarehouseNodeEntity) GetEnterWarehouseNodeUpdatedAt(def ...int) int {
	return o.AutoLoad("enter_warehouse_node_updated_at").DO.GetInt("enter_warehouse_node_updated_at", def...)
}

// 设置 最后更新时间
func (o *EnterWarehouseNodeEntity) SetEnterWarehouseNodeUpdatedAt(value interface{}) (err error) {
	return o.Set("enter_warehouse_node_updated_at", value)
}

// 获取 最后操作人
func (o *EnterWarehouseNodeEntity) GetOpId(def ...int) int {
	return o.AutoLoad("op_id").DO.GetInt("op_id", def...)
}

// 设置 最后操作人
func (o *EnterWarehouseNodeEntity) SetOpId(value interface{}) (err error) {
	return o.Set("op_id", value)
}

// 获取 是否是删除: 0不是|1是
func (o *EnterWarehouseNodeEntity) GetEnterWarehouseNodeDeleted(def ...int) int {
	return o.AutoLoad("enter_warehouse_node_deleted").DO.GetInt("enter_warehouse_node_deleted", def...)
}

// 设置 是否是删除: 0不是|1是
func (o *EnterWarehouseNodeEntity) SetEnterWarehouseNodeDeleted(value interface{}) (err error) {
	return o.Set("enter_warehouse_node_deleted", value)
}

// 获取 自增ID
func (o *EnterWarehouseNodeEntity) GetEnterWarehouseNodeId(def ...int) int {
	return o.AutoLoad("enter_warehouse_node_id").DO.GetInt("enter_warehouse_node_id", def...)
}

// 设置 自增ID
func (o *EnterWarehouseNodeEntity) SetEnterWarehouseNodeId(value interface{}) (err error) {
	return o.Set("enter_warehouse_node_id", value)
}

// 获取 工作流ID
func (o *EnterWarehouseNodeEntity) GetEnterWarehouseId(def ...int) int {
	return o.AutoLoad("enter_warehouse_id").DO.GetInt("enter_warehouse_id", def...)
}

// 设置 工作流ID
func (o *EnterWarehouseNodeEntity) SetEnterWarehouseId(value interface{}) (err error) {
	return o.Set("enter_warehouse_id", value)
}

// 获取 商品ID
func (o *EnterWarehouseNodeEntity) GetProductId(def ...int) int {
	return o.AutoLoad("product_id").DO.GetInt("product_id", def...)
}

// 设置 商品ID
func (o *EnterWarehouseNodeEntity) SetProductId(value interface{}) (err error) {
	return o.Set("product_id", value)
}

// 获取 节点code
func (o *EnterWarehouseNodeEntity) GetNodeCode(def ...string) string {
	return o.AutoLoad("node_code").DO.GetStr("node_code", def...)
}

// 设置 节点code
func (o *EnterWarehouseNodeEntity) SetNodeCode(value interface{}) (err error) {
	return o.Set("node_code", value)
}

// 获取 节点状态code
func (o *EnterWarehouseNodeEntity) GetNodeStatus(def ...string) string {
	return o.AutoLoad("node_status").DO.GetStr("node_status", def...)
}

// 设置 节点状态code
func (o *EnterWarehouseNodeEntity) SetNodeStatus(value interface{}) (err error) {
	return o.Set("node_status", value)
}

// 获取 创建时间
func (o *EnterWarehouseNodeEntity) GetEnterWarehouseNodeCreatedAt(def ...int) int {
	return o.AutoLoad("enter_warehouse_node_created_at").DO.GetInt("enter_warehouse_node_created_at", def...)
}

// 设置 创建时间
func (o *EnterWarehouseNodeEntity) SetEnterWarehouseNodeCreatedAt(value interface{}) (err error) {
	return o.Set("enter_warehouse_node_created_at", value)
}

// 获取 版本ID
func (o *EnterWarehouseNodeEntity) GetVersion(def ...int) int {
	return o.AutoLoad("version").DO.GetInt("version", def...)
}

// 设置 版本ID
func (o *EnterWarehouseNodeEntity) SetVersion(value interface{}) (err error) {
	return o.Set("version", value)
}
