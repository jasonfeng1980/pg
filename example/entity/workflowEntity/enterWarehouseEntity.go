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
func NewEnterWarehouseEntity(ctx context.Context, pk ...interface{}) *EnterWarehouseEntity {
	dao, err := DAO.NewWorkflowDao(ctx, "enter_warehouse")
	if err != nil {
		util.Panic(ecode.DaoWrongTable.Error("workflow", "enter_warehouse"))
	}
	o := &EnterWarehouseEntity{
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

type EnterWarehouseEntity struct {
	*ddd.Entity
}

//////////////////////////////////////////
//  关联实体
//////////////////////////////////////////

func (o *EnterWarehouseEntity) RelationEnterWarehouseLog() *ddd.Query {
	q := o.Relation("enter_warehouse_log")
	q.ResultFunc = func(query *db.Query) (interface{}, error) {
		doList, err := query.Query().Array()
		if err != nil {
			return nil, err
		}
		var ret []*EnterWarehouseLogEntity
		relationPk := o.DatabaseMap.TableMap["enter_warehouse_log"].Pk
		for _, DO := range doList {
			relationEntity := NewEnterWarehouseLogEntity(o.Ctx, DO[relationPk])
			relationEntity.DO.Box = DO
			ret = append(ret, relationEntity)
		}
		return ret, err
	}
	return q
}

func (o *EnterWarehouseEntity) RelationEnterWarehouseNode() *ddd.Query {
	q := o.Relation("enter_warehouse_node")
	q.ResultFunc = func(query *db.Query) (interface{}, error) {
		doList, err := query.Query().Array()
		if err != nil {
			return nil, err
		}
		var ret []*EnterWarehouseNodeEntity
		relationPk := o.DatabaseMap.TableMap["enter_warehouse_node"].Pk
		for _, DO := range doList {
			relationEntity := NewEnterWarehouseNodeEntity(o.Ctx, DO[relationPk])
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

// 获取 操作时间
func (o *EnterWarehouseEntity) GetCreatedAt(def ...int) int {
	return o.AutoLoad("created_at").DO.GetInt("created_at", def...)
}

// 设置 操作时间
func (o *EnterWarehouseEntity) SetCreatedAt(value interface{}) (err error) {
	return o.Set("created_at", value)
}

// 获取 工作流模板code
func (o *EnterWarehouseEntity) GetWorkflowTplCode(def ...string) string {
	return o.AutoLoad("workflow_tpl_code").DO.GetStr("workflow_tpl_code", def...)
}

// 设置 工作流模板code
func (o *EnterWarehouseEntity) SetWorkflowTplCode(value interface{}) (err error) {
	return o.Set("workflow_tpl_code", value)
}

// 获取 工作流是否完结: 0进行中|1完结
func (o *EnterWarehouseEntity) GetEnterWarehouseEnd(def ...int) int {
	return o.AutoLoad("enter_warehouse_end").DO.GetInt("enter_warehouse_end", def...)
}

// 设置 工作流是否完结: 0进行中|1完结
func (o *EnterWarehouseEntity) SetEnterWarehouseEnd(value interface{}) (err error) {
	return o.Set("enter_warehouse_end", value)
}

// 获取 是否是删除: 0不是|1是
func (o *EnterWarehouseEntity) GetEnterWarehouseDeleted(def ...int) int {
	return o.AutoLoad("enter_warehouse_deleted").DO.GetInt("enter_warehouse_deleted", def...)
}

// 设置 是否是删除: 0不是|1是
func (o *EnterWarehouseEntity) SetEnterWarehouseDeleted(value interface{}) (err error) {
	return o.Set("enter_warehouse_deleted", value)
}

// 获取 自增ID
func (o *EnterWarehouseEntity) GetEnterWarehouseId(def ...int) int {
	return o.AutoLoad("enter_warehouse_id").DO.GetInt("enter_warehouse_id", def...)
}

// 设置 自增ID
func (o *EnterWarehouseEntity) SetEnterWarehouseId(value interface{}) (err error) {
	return o.Set("enter_warehouse_id", value)
}

// 获取 商品ID
func (o *EnterWarehouseEntity) GetProductId(def ...int) int {
	return o.AutoLoad("product_id").DO.GetInt("product_id", def...)
}

// 设置 商品ID
func (o *EnterWarehouseEntity) SetProductId(value interface{}) (err error) {
	return o.Set("product_id", value)
}

// 获取 拍摄-节点状态
func (o *EnterWarehouseEntity) GetEW4100(def ...string) string {
	return o.AutoLoad("EW4100").DO.GetStr("EW4100", def...)
}

// 设置 拍摄-节点状态
func (o *EnterWarehouseEntity) SetEW4100(value interface{}) (err error) {
	return o.Set("EW4100", value)
}

// 获取 网关拍摄完成且卖家完成定价-网关状态
func (o *EnterWarehouseEntity) GetEW6000(def ...string) string {
	return o.AutoLoad("EW6000").DO.GetStr("EW6000", def...)
}

// 设置 网关拍摄完成且卖家完成定价-网关状态
func (o *EnterWarehouseEntity) SetEW6000(value interface{}) (err error) {
	return o.Set("EW6000", value)
}

// 获取 版本ID
func (o *EnterWarehouseEntity) GetVersion(def ...int) int {
	return o.AutoLoad("version").DO.GetInt("version", def...)
}

// 设置 版本ID
func (o *EnterWarehouseEntity) SetVersion(value interface{}) (err error) {
	return o.Set("version", value)
}

// 获取 收货打签节-点状态
func (o *EnterWarehouseEntity) GetEW1000(def ...string) string {
	return o.AutoLoad("EW1000").DO.GetStr("EW1000", def...)
}

// 设置 收货打签节-点状态
func (o *EnterWarehouseEntity) SetEW1000(value interface{}) (err error) {
	return o.Set("EW1000", value)
}

// 获取 编辑-节点状态
func (o *EnterWarehouseEntity) GetEW3000(def ...string) string {
	return o.AutoLoad("EW3000").DO.GetStr("EW3000", def...)
}

// 设置 编辑-节点状态
func (o *EnterWarehouseEntity) SetEW3000(value interface{}) (err error) {
	return o.Set("EW3000", value)
}

// 获取 最后操作人
func (o *EnterWarehouseEntity) GetOpId(def ...int) int {
	return o.AutoLoad("op_id").DO.GetInt("op_id", def...)
}

// 设置 最后操作人
func (o *EnterWarehouseEntity) SetOpId(value interface{}) (err error) {
	return o.Set("op_id", value)
}

// 获取 鉴定-节点状态
func (o *EnterWarehouseEntity) GetEW2000(def ...string) string {
	return o.AutoLoad("EW2000").DO.GetStr("EW2000", def...)
}

// 设置 鉴定-节点状态
func (o *EnterWarehouseEntity) SetEW2000(value interface{}) (err error) {
	return o.Set("EW2000", value)
}

// 获取 卖家定价-节点状态
func (o *EnterWarehouseEntity) GetEW4200(def ...string) string {
	return o.AutoLoad("EW4200").DO.GetStr("EW4200", def...)
}

// 设置 卖家定价-节点状态
func (o *EnterWarehouseEntity) SetEW4200(value interface{}) (err error) {
	return o.Set("EW4200", value)
}
