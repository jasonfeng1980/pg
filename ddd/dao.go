package ddd

import (
    "context"
    "database/sql"
    "errors"
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
)

// 关联
type RelationKV struct {
    InternalKey	string
    ForeignKey	string
}

// Repository   -- 判断， 验证， 归纳到表
type DataMap struct {
    TableMap  map[string]*Table      // map[tableName]*Table
    FieldMap  map[string]*Field      // map[fieldName]*Field
    RelationMap map[string]map[string][]*RelationKV // map[tableName]*Relation
}

type Table struct {
    DBName          string
    TableName      string
    Pk        string
    FieldArr  []string
}

type Field struct {
    TableName   string
    FieldName   string
    MysqlType   string
    Unsigned    bool
    Need        bool
}
func (f *Field)Check(arg interface{}) error{
    ok :=pg.Filter.MySQL(f.MysqlType, f.Unsigned, f.Need).Check(arg)
    if  !ok{
        return errors.New("验证失败")
    }
    return nil
}

// DAO
type DAO struct {
    *db.Query

    DO       util.Param         // 加载后的数据
    Params   util.Param         // 传入的数据
    *Option
}
type Option struct {
    Ctx context.Context
    DBHandleName  string
    DatabaseMap  *DataMap
    TableName string
}
// 获取表名
func(o *DAO) GetTableName()string{
    return o.TableName
}
// 获取PK名
func(o *DAO) GetPKName()string{
    return o.DatabaseMap.TableMap[o.TableName].Pk
}
// 表中是否存在该字段
func (o *DAO) HadField(fieldName string) bool{
    return util.ListHave(o.DatabaseMap.TableMap[o.TableName].FieldArr, fieldName)
}
// 验证必填
func (o *DAO) CheckParams(needFieldNameList ...[]string) (err error) {
    var misList []string
    // 指定了必填字段
    if len(needFieldNameList) == 1{
        // 循环判断所有的need
        for _, fieldName := range  needFieldNameList[0] {
            if !o.HadField(fieldName){
                misList = append(misList, fieldName)
            }
        }
        if len(misList) >0 {
            return ecode.DaoMissNeedField.Error(o.TableName, util.ListStringJoin(misList, ","))
        }
    }
    // 用系统默认的必填字段验证
    for _,v := range o.DatabaseMap.FieldMap{
        if v.Need && !o.HadField(v.FieldName){ // 要求必填，但没有该字段
            misList = append(misList, v.FieldName)
        }
    }
    if len(misList) >0 {
        return ecode.DaoMissNeedField.Error(o.TableName, util.ListStringJoin(misList, ","))
    }

    return
}
// 为字段赋值
func (o *DAO) Set(fieldName string, value interface{}) (err error){
    // 如果存在字段，就赋值
    if !o.HadField(fieldName){
        return ecode.DaoWrongField.Error(o.TableName, fieldName)
    }
    if util.InterfaceType(value) == "db.expr" {
        o.Params.Box[fieldName] = value
        return nil
    }
    // 检验
    if err := o.DatabaseMap.FieldMap[fieldName].Check(value);err == nil {
        o.Params.Box[fieldName] = value
    }
    return err

}
// 为字段批量赋值
func (o *DAO) SetMany(params map[string]interface{}) (errList []error){
    // 循环参数
    for k, v := range params{
        // 如果存在字段，就赋值
        if util.ListHave(o.DatabaseMap.TableMap[o.TableName].FieldArr, k) {
            // 检验
            if err := o.DatabaseMap.FieldMap[k].Check(v);err != nil {
                o.Params.Box[k] = v
            } else {
                errList = append(errList, err)
            }
        }
    }
    return
}
// 新增
func (o *DAO) Create() (int64, error){
    return o.Conn().Insert(o.Params.Box).Into(o.TableName).Query().RowsAffected()
}

// 转换成DO
func (o *DAO) ParseToDO() map[string]interface{}{
    return o.DO.Box
}
// 转换成json
func (o *DAO) ParseToJson() ([]byte, error){
    return util.JsonEncode(o.DO.Box)
}
// 根据优先级获取指定字段  DO > Params
func (o *DAO) GetField(fieldName string) interface{}{
    if v := o.DO.Get(fieldName); v != nil {
        return v
    }
    if v := o.Params.Get(fieldName); v != nil {
        return v
    }
    return nil
}
// 获取关联数据
func (o *DAO)Relation(RelTableName string) (*db.Query, error){
    relCondition, ok := o.DatabaseMap.RelationMap[o.TableName]
    if !ok {
        return nil, ecode.EntityWrongRelTableName.Error(o.TableName, RelTableName)
    }
    conditionList, ok := relCondition[RelTableName]
    if !ok {
        return nil, ecode.EntityWrongRelTableName.Error(o.TableName, RelTableName)
    }
    q := o.Conn().Select("*").From(RelTableName)
    for _, v := range conditionList{
        myVal := o.GetField(v.InternalKey)
        if myVal == nil {
            return nil, ecode.EntityMissRelValue.Error(o.TableName, RelTableName)
        }
        q = q.Where(v.ForeignKey, myVal)
    }
    return q, nil
}


// 数据库相关操作
const TxKey = "PG_TX"
// 获取数据库链接句柄
func(o *DAO)Conn() *db.Query{
    var err error
    if o.Query == nil {
        if o.Query, err = db.MYSQL.Get(o.DBHandleName); err != nil {
            util.Error("链接MYSQL出错", "err", err)
        }
    }
    // 判断是否开启了事务
    session  := util.SessionHandle(o.Ctx)
    useTx := session.Get(TxKey)

    if useTx != nil { // 开启事务
        if o.Query.Tx != nil { // 自己已经开启了事务
            return o.Query
        }
        // 目前没有开启事务
        TxBox := useTx.(map[string]*sql.Tx)
        // 是否存在存在的sql.Tx
        if tx, ok := TxBox[o.DBHandleName]; ok { // 有存在的sql.Tx
            o.Query.UseTx(tx)
        } else { // 没有sql.Tx
            o.Query = o.Query.StartTransaction()
            TxBox[o.DBHandleName] = o.Query.Tx
            util.SessionSet(o.Ctx, TxKey, TxBox)
        }
    }

    return o.Query
}