package ddd
/**
实体层
. 本地唯一的标识ID
. 脱离数据库
. 基本的业务能力(实体内部的)
. 原则上不允许拥有跨实体的能力
. 继承DAO
. 常用的get set
. e.g.  电视机里的     二极管    信号放大器   显示屏
*/
import (
    "fmt"
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
)

type Query struct {
    *db.Query
    ResultFunc func(query *db.Query) (interface{},  error)
}
func (r *Query) Where(where interface{}, args  ...interface{}) *Query{
    r.Query.Where(where, args...)
    return r
}
func (r *Query) Cache(useCache bool) *Query{
    r.Query.Cache(useCache)
    return r
}
func (r *Query) Limit(skip int, limit int) *Query {
    r.Query.Limit(skip, limit)
    return r
}
func (r *Query) Page(pageSize int, pageNum int) *Query {
    // 整理数据
    if pageSize <=1 {
        pageSize = 1
    }
    if pageNum <=0 {
        pageNum = 1
    }
    skip := pageSize * (pageNum -1)
    limit := pageSize
    return r.Limit(skip, limit)
}
func (r *Query) Result() (interface{}, error) {
    if r.ResultFunc != nil {
        return r.ResultFunc(r.Query)
    }
    return nil, ecode.EntityMissResultQueryFunc.Error()
}

// Entity
type Entity struct {
    *DAO

    pk        int64                     // 唯一标识
    isLoad    bool                      // 是否加载过
}
func (o *Entity)TableName() string{
    return o.DAO.GetTableName()
}
func (o *Entity)PkName() string{
    return o.DAO.GetPKName()
}
func (o *Entity)Pk() int64{
    return o.pk
    //return o.GetInt64(o.PkName(), 0)
}
// 创建新的记录
func (o *Entity)Create(needFieldNameList ...[]string) error {
    if len(o.Params.Box) == 0 {
        return ecode.EntityEmptyCreateParams.Error(o.TableName())
    }
    // 验证必填
    if err := o.DAO.CheckParams(needFieldNameList...); err != nil {
        return err
    }
    // 创建时，无需PK
    o.DAO.Params.Delete(o.PkName())
    rs, err := o.DAO.Create()
    if err == nil { // 没有错误,就更新PK
        o.pk = rs
    }
    return err
}
// 搜索
func (o *Entity)Search(condition map[string]interface{}) ([]map[string]interface{}, error) {
    search := make(map[string]interface{})
    for k, v := range condition {
        if o.HadField(k) {
            search[k] = v
        }
    }
    // 没有需要查询的数据
    if len(search) == 0 {
        return nil, ecode.EntityEmptySearchCondition.Error(o.DAO.GetTableName())
    }
    return o.Conn().Select("*").From(o.TableName()).Where(search).Query().Array()
}
// 设置PK ，不加载
func (o *Entity)SetPk(pk int64) {
    o.pk = pk
}
// 通过PK，加载实体
func (o *Entity)Find(pk int64) error{
    o.pk = pk
    return o.Load()
}
// 通过主键，加载数据
func (o *Entity)Load() error{
    if o.Pk() == 0 {
        return ecode.EntityMissPK.Error(o.DAO.GetTableName())
    }
    rs, err := o.Conn().Select("*").From(o.DAO.GetTableName()).Where(o.DAO.GetPKName(), o.Pk()).Limit(0, 1).Query().Array()
    if err != nil {// 出现错误
        o.pk = 0
        return err
    }
    if len(rs) == 0 { // 没有数据
        o.pk = 0
        return ecode.DaoEmptyResult.Error(fmt.Sprintf("%s=%d", o.DAO.GetPKName(), o.Pk()))
    }
    o.DO.Box = rs[0]
    o.isLoad = true
    return err
}
// 转换成DO
func (o *Entity) ParseToDO() map[string]interface{}{
    return o.AutoLoad(o.PkName()).DAO.ParseToDO()
}
// 编辑
func (o *Entity)Edit(condition ...map[string]interface{}) *Query{
    if o.Pk() == 0 {
        panic(ecode.EntityMissPK.Error(o.DAO.GetTableName()))
        return nil
    }
    // 不允许更新唯一标识
    data := o.Params.Box
    delete(data, o.PkName())
    q := o.Conn().Update(o.DAO.GetTableName()).Set(data).Where(o.DAO.GetPKName(), o.Pk())

    return &Query{
        Query:q,
        ResultFunc: func(query *db.Query) (interface{}, error) {
            return q.Query().RowsAffected()
        },
    }
}
// 删除
func (o *Entity)Remove() *Query{
    if o.Pk() == 0 {
        panic(ecode.EntityMissPK.Error(o.DAO.GetTableName()))
        return nil
    }
    q := o.Conn().Delete().From(o.TableName()).Where(o.DAO.GetPKName(), o.Pk())
    return &Query{
        Query:q,
        ResultFunc: func(query *db.Query) (interface{}, error) {
            return q.Query().RowsAffected()
        },
    }
}
// 判断属性值，没有就自动加载
func (o *Entity)AutoLoad(fieldName string) *Entity{
    if _, ok :=o.DO.Box[fieldName]; !ok {
        o.Load()
    }
    return o
}
// 获取关联数据
func (o *Entity)Relation(RelTableName string) *Query{
    if o.Pk() == 0 {
        util.Panic(ecode.EntityMissPK.Error(o.TableName()))
        return nil
    }

    q, err := o.DAO.Relation(RelTableName)
    if err == nil { // 没有错误, 就直接返回
        return  &Query{Query:q}
    }
    // 出错了
    if code, _ := ecode.ReadError(err); code != ecode.EntityMissRelValue.Code { // 如果错误，不是 "没有获得关联字段是值"
        util.Panic(ecode.EntityMissPK.Error(o.TableName()))
        return nil
    }
    // 非主键关联， 就load下数据，再relation下
    // 自己load一下
    if err = o.Load(); err !=nil {
        util.Panic(err)
        return nil
    }
    // 再请求一下
    q, err = o.DAO.Relation(RelTableName)
    if err != nil {
        util.Panic(err)
    }
    return  &Query{Query:q}
}