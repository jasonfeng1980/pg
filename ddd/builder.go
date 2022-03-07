package ddd

import (
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/util"
    "strings"
)


var DaoTpl = `package DAO

import (
    "context"
    "github.com/jasonfeng1980/pg/ddd"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
)

// 根据表名获得一个Dao对象， 可以通过DO初始化数据
func New{DBHandleNameFirstUpper}Dao(ctx context.Context, tableName string, DO ...map[string]interface{}) (*ddd.DAO, error){
    if _, ok := Database{DBHandleNameFirstUpper}.TableMap[tableName]; !ok{
        return nil, ecode.DaoWrongTable.Error("{DBHandleName}", tableName)
    }
    var d = make(map[string]interface{})
    if len(DO) == 1 {
        d = DO[0]
    }
    dao := &ddd.DAO{
        Option: &ddd.Option{
            Ctx:          ctx,
            DBHandleName: "{DBHandleName}",
            DatabaseMap:  Database{DBHandleNameFirstUpper},
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
var Database{DBHandleNameFirstUpper} = &ddd.DataMap{
    TableMap: map[string]*ddd.Table{{TableMapFieldList}
    },
    FieldMap: map[string]*ddd.Field{{FieldMapFieldList}
    },
    RelationMap: map[string]map[string][]*ddd.RelationKV{{RealtionMapList}
    },
}

`
var RelationMap = make(map[string][]string)       // 反向表关联 b变名：a表名
var TplDAOTableMapFieldList = `
        "{TableName}": {"{DBHandleName}", "{TableName}", "{PK}", []string{{FieldList}}, []string{{NeedFieldList}}},`
var TplDaoFieldMapField = `
        "{FieldName}": {"{TableName}","{FieldName}", "{MySQLType}", {Unsigned}, {Need}},`
var TplDaoRelationMapOneTab = `"{RelationTableName}": {{"{PK}", "{PK}"}}`
var TplDaoRelationMap = `
        "{TableName}" : {{TplDaoRelationMapOneTab}},`
var EntityTpl = `package {DBHandleName}Entity

import (
    "context"
    "{AppPackage}/{EntityPath}/DAO"
    {RelationPageImport}
    "github.com/jasonfeng1980/pg/ddd"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
)

// 获取实体对象
func New{TableUFirst}Entity(ctx context.Context, pk ...interface{}) *{TableUFirst}Entity {
    dao, err := DAO.New{DBHandleNameFirstUpper}Dao(ctx, "{TableName}")
    if err != nil {
        util.Panic(ecode.DaoWrongTable.Error("{DBHandleName}", "{TableName}"))
    }
    o := &{TableUFirst}Entity{
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

type {TableUFirst}Entity struct {
    *ddd.Entity
}

//////////////////////////////////////////
//  关联实体
//////////////////////////////////////////
{RelationEntity}

//////////////////////////////////////////
//  Getter Setter
//////////////////////////////////////////
{FieldGetSet}

`
var TplRelationEntity = `
func (o *{TableUFirst}Entity) Relation{RelTableUFirst}() *ddd.Query {
    q := o.Relation("{RelTableName}")
    q.ResultFunc = func(query *db.Query) (interface{},  error){
        doList, err := query.Query().Array()
        if err != nil {
            return nil, err
        }
        var ret []*{RelTableUFirst}Entity
        relationPk := o.DatabaseMap.TableMap["{RelTableName}"].Pk
        for _, DO := range doList {
            relationEntity := New{RelTableUFirst}Entity(o.Ctx, DO[relationPk] )
            relationEntity.DO.Box = DO
            ret = append(ret, relationEntity)
        }
        return ret, err
    }
    return q
}
`

var TplFieldGetSet = `
// 获取 {FieldComment}
func (o *{TableUFirst}Entity) Get{FieldNameUFirst}(def ...{MySQLTypeToGoType}) {MySQLTypeToGoType}{
    return o.AutoLoad("{FieldName}").DO.{MySQLTypeToParam}("{FieldName}", def...)
}
// 设置 {FieldComment}
func (o *{TableUFirst}Entity) Set{FieldNameUFirst}(value interface{}) (err error){
    return o.Set("{FieldName}", value)
}
`

///////// 生成Entity
type EntityField struct {
    Name      string
    DataType  string
    MysqlType string
    Comment   string
    Unsigned  bool
    IsNeed    bool
}

type EntityTable struct {
    Name      string
    Pk        string
    FieldArr    []string
    Field    map[string]*EntityField
}

type mysqlTypeRelation struct {
    goType      string
    paramFunc   string
}
var mysqlTypeMap = map[string]mysqlTypeRelation{
    "enum": {"string", "GetStr"},
    "tinyint": {"int", "GetInt"},
    "smallint": {"int", "GetInt"},
    "int": {"int", "GetInt"},
    "mediumint": {"int", "GetInt"},
    "bigint": {"int64", "GetInt64"},
    "timestamp": {"int", "GetInt"},
    "date": {"string", "GetStr"},
    "datetime": {"string", "GetStr"},
    "char": {"string", "GetStr"},
    "varchar": {"string", "GetStr"},
    "text": {"string", "GetStr"},
    "json": {"string", "GetStr"},
    "logblob": {"string", "GetStr"},
}

// 根据各个数据库的句柄，创建Entity池
func BuildEntity(confName, appPackage string, path string){
    // 获取表结构
    conn := db.MYSQL.GetPool(confName)
    if conn == nil {
        panic("数据库配置名错误: configName=" + confName)
    }
    var (
        dbName = conn.Conf.W.Database
        tableArr   []string
        db, _ = db.MYSQL.Get(confName)
        //result = make(map[string]*EntityTable)
    )

    // 获取所有的表
    rs, _ := db.Query("SHOW FULL TABLES WHERE table_type = 'BASE TABLE'").Array()
    for _, line:= range rs {
        tableArr = append(tableArr, line["Tables_in_" + dbName].(string))
    }
    util.Info("获取表结构", "tableArr", tableArr)

    // 循环表获取各个字段
    DaoTableMapFieldList := ""
    DaoFieldMapFieldList := ""
    DaoRelationBoxList  := make(map[string][]string)
    DaoRelationBoxStr   := ""

    FieldBox := make(map[string]int)
    RelationTableMap := make(map[string][]string)
    // 获取所有的主键
    pkList, _ := db.Query("select TABLE_NAME, COLUMN_NAME from information_schema.columns where TABLE_SCHEMA=? and COLUMN_KEY = 'PRI'", dbName).Array()
    for _, v := range pkList {
        // 获取其他表管理自己主键的记录
        vTable := util.Str(v["TABLE_NAME"])
        vRs, _ := db.Query("select TABLE_NAME from information_schema.columns where TABLE_schema=? and TABLE_NAME!=? and COLUMN_NAME=?", dbName, vTable, util.Str(v["COLUMN_NAME"])).Array()
        for _, vv := range vRs {
            if !util.ListHave(RelationTableMap[vTable], util.Str(vv["TABLE_NAME"])) { // 存在就不加
                RelationTableMap[vTable] = append(RelationTableMap[vTable], util.Str(vv["TABLE_NAME"]))
            }
        }
        // 给关联的表加记录
        for _, rTableName := range RelationTableMap[vTable] {
            if !util.ListHave(RelationTableMap[rTableName], vTable) { // 存在就不加
                RelationTableMap[rTableName] = append(RelationTableMap[rTableName], vTable)
            }
        }
    }

    for _, table := range tableArr {
        EntityRelationList := ""
        rs, _ := db.Query("select *,if (ISNULL(COLUMN_DEFAULT), \"0\", \"1\") as hasDefault from information_schema.columns where TABLE_schema =? and TABLE_NAME=?", dbName, table).Array()

        tableColumns := make(map[string]*EntityField)
        var (
            tablePk string
            tableField  []string
            isNeed bool
            unsigned bool
        )

        for _, line := range rs{
            isNeed = true
            if line["COLUMN_KEY"] == "PRI" {
                tablePk = line["COLUMN_NAME"].(string)
            }
            tableField = append(tableField, line["COLUMN_NAME"].(string))

            if line["EXTRA"].(string) == "auto_increment" { // 有自增ID，就不必填
                isNeed = false
            }

            if line["IS_NULLABLE"]== "YES" || line["hasDefault"] == "1" {
                isNeed = false
            }
            splitType := strings.Split(line["COLUMN_TYPE"].(string), " ")
            if len(splitType) ==2 && splitType[1]=="unsigned" {
                unsigned = true
            } else {
                unsigned = false
            }
            tableColumns[line["COLUMN_NAME"].(string)] =  &EntityField{
                DataType:  line["DATA_TYPE"].(string),
                MysqlType: splitType[0],
                Comment:   line["COLUMN_COMMENT"].(string),
                Unsigned: unsigned,
                IsNeed: isNeed,
            }
        }

        // 整理关联表
        if len(RelationTableMap[table]) >0 {

            for _, v := range RelationTableMap[table]{
                relationTableName := v
                //tplRelation := TplDaoRelationMap
                tplRelation := TplDaoRelationMapOneTab
                tplRelation = strings.Replace(tplRelation, "{TableName}", table, -1)
                tplRelation = strings.Replace(tplRelation, "{RelationTableName}", relationTableName, -1)
                tplRelation = strings.Replace(tplRelation, "{PK}", tablePk, -1)
                if _, ok := DaoRelationBoxList[table]; !ok { // 还没创建，就创建
                    DaoRelationBoxList[table] = []string{}
                }
                DaoRelationBoxList[table] = append(DaoRelationBoxList[table], tplRelation)

                tplEntityRelation := TplRelationEntity
                tplEntityRelation = strings.Replace(tplEntityRelation, "{TableUFirst}", util.StrUFirstForSplit(table, "_"), -1)
                tplEntityRelation = strings.Replace(tplEntityRelation, "{RelTableUFirst}", util.StrUFirstForSplit(relationTableName, "_"), -1)
                tplEntityRelation = strings.Replace(tplEntityRelation, "{RelTableName}", relationTableName, -1)
                EntityRelationList += tplEntityRelation

                RelationMap[relationTableName] = append(RelationMap[relationTableName], table)
            }
        }

        // 创建获取多个和field相关的替换内容
        daoTableStr, daoFieldStr, fieldGetSetStr := getTableField(EntityTable{
            Name: table,
            Pk: tablePk,
            FieldArr: tableField,
            Field: tableColumns,
        }, confName, appPackage, path, FieldBox)
        DaoTableMapFieldList += daoTableStr
        DaoFieldMapFieldList += daoFieldStr

        // 生成Entity文件
        entityStr := EntityTpl
        entityStr =  strings.Replace(entityStr, "{DBHandleName}", confName, -1)
        entityStr =  strings.Replace(entityStr, "{AppPackage}", appPackage, -1)
        entityStr =  strings.Replace(entityStr, "{EntityPath}", path, -1)
        entityStr =  strings.Replace(entityStr, "{TableUFirst}", util.StrUFirstForSplit(table, "_"), -1)
        entityStr =  strings.Replace(entityStr, "{TableName}", table, -1)
        entityStr =  strings.Replace(entityStr, "{pkName}", tablePk, -1)
        entityStr =  strings.Replace(entityStr, "{FieldGetSet}", fieldGetSetStr, -1)
        entityStr =  strings.Replace(entityStr, "{RelationEntity}", EntityRelationList, -1)
        entityStr =  strings.Replace(entityStr, "{DBHandleNameFirstUpper}", util.StrUFirst(confName), -1)
        RelationPageImport := ""
        if EntityRelationList != "" {
            RelationPageImport = `"github.com/jasonfeng1980/pg/database/db"`
        }
        entityStr =  strings.Replace(entityStr, "{RelationPageImport}", RelationPageImport, -1)
        entityPath := util.FileRealPath("tmp")
        dir := entityPath + "/entity/" + confName +"Entity/"
        fileName :=  util.StrSecFirstForSplit(table, "_") + "Entity.go"
        if err := util.FileWrite(dir, fileName, entityStr); err!=nil{
            util.Logs("创建Entity文件：", dir, fileName , "  失败! -- Error")
            util.Logs(err.Error())
        } else {
            util.Logs("创建Entity文件：", dir, fileName , "  成功! -- OK")
        }

    }

    // 整理 DaoRelationBoxList
    for realTable, v := range DaoRelationBoxList {
        tplRelationStr := TplDaoRelationMap
        tplRelationStr = strings.Replace(tplRelationStr, "{TableName}", realTable, -1)
        tplRelationStr = strings.Replace(tplRelationStr, "{TplDaoRelationMapOneTab}", util.ListStringJoin(v, ","), -1)

        DaoRelationBoxStr += tplRelationStr
    }

    // 生成dao文件
    daoStr := DaoTpl
    daoStr =  strings.Replace(daoStr, "{DBHandleName}", confName, -1)
    daoStr =  strings.Replace(daoStr, "{DBHandleNameFirstUpper}", util.StrUFirst(confName), -1)
    daoStr =  strings.Replace(daoStr, "{TableMapFieldList}", DaoTableMapFieldList, -1)
    daoStr =  strings.Replace(daoStr, "{FieldMapFieldList}", DaoFieldMapFieldList, -1)
    daoStr =  strings.Replace(daoStr, "{RealtionMapList}", DaoRelationBoxStr, -1)
    daoPath := util.FileRealPath(path)
    dir := daoPath + "/DAO/"
    fileName :=  confName + "Mapper.go"
    if err := util.FileWrite(dir, fileName, daoStr); err!=nil{
        util.Logs("创建DAO文件：", dir, fileName , "  失败! -- Error")
        util.Logs(err.Error())
    } else {
        util.Logs("创建DAO文件：", dir, fileName , "  成功! -- OK")
    }




}

// 创建DAO文件  entityPath/dao/dbHandleName.go
func getTableField(table EntityTable, dbHandleName string, appPackage string, entityPath string, FieldBox map[string]int) (string, string, string){
    var (
        tableMapFieldList  string
        FieldMapFieldList string
        FieldMapNeedList []string

    	FieldGetSetList  []string
    )

    fieldStrs := "\"" + strings.Join(table.FieldArr, "\",\"") + "\""

    tableMapField := TplDAOTableMapFieldList
    tableMapField =  strings.Replace(tableMapField, "{TableName}", table.Name, -1)
    tableMapField =  strings.Replace(tableMapField, "{DBHandleName}", dbHandleName, -1)
    tableMapField =  strings.Replace(tableMapField, "{PK}", table.Pk, -1)
    tableMapField =  strings.Replace(tableMapField, "{FieldList}", fieldStrs, -1)
    //tableMapFieldList += tableMapField

    for field, filterObj := range table.Field{
        // Entity的GetSet
        fieldGetSetTmp := TplFieldGetSet
        fieldGetSetTmp =  strings.Replace(fieldGetSetTmp, "{FieldComment}", filterObj.Comment, -1)
        fieldGetSetTmp =  strings.Replace(fieldGetSetTmp, "{TableUFirst}", util.StrUFirstForSplit(table.Name, "_"), -1)
        fieldGetSetTmp =  strings.Replace(fieldGetSetTmp, "{FieldNameUFirst}", util.StrUFirstForSplit(field, "_"), -1)
        fieldGetSetTmp =  strings.Replace(fieldGetSetTmp, "{MySQLTypeToGoType}", mysqlTypeMap[filterObj.DataType].goType, -1)
        fieldGetSetTmp =  strings.Replace(fieldGetSetTmp, "{MySQLTypeToParam}", mysqlTypeMap[filterObj.DataType].paramFunc, -1)
        fieldGetSetTmp =  strings.Replace(fieldGetSetTmp, "{FieldName}", field, -1)
        FieldGetSetList = append(FieldGetSetList, fieldGetSetTmp)

        // 配置DAO
        if _, ok:= FieldBox[field]; ok {
            continue
        }
        FieldBox[field] = 1

        fieldMapField := TplDaoFieldMapField
        unsigned, _ := util.StrParse(filterObj.Unsigned)
        isNeed, _ := util.StrParse(filterObj.IsNeed)
        fieldMapField =  strings.Replace(fieldMapField, "{FieldName}", field, -1)
        fieldMapField =  strings.Replace(fieldMapField, "{MySQLType}", filterObj.MysqlType, -1)
        fieldMapField =  strings.Replace(fieldMapField, "{Unsigned}", unsigned, -1)
        fieldMapField =  strings.Replace(fieldMapField, "{Need}", isNeed, -1)
        fieldMapField =  strings.Replace(fieldMapField, "{TableName}", table.Name, -1)
        //fieldMapList = append(fieldMapList, tableMapField)
        FieldMapFieldList += fieldMapField

        if filterObj.IsNeed {
            FieldMapNeedList = append(FieldMapNeedList, field)
        }

        // 增加入口前的 默认数据修改
        //if field == "create_at" {
        //    TableFilter += "  ret.BeforeInsert[\"create_at\"] = db.ChangeNow\n\n"
        //}
        //if field == "update_at" {
        //    TableFilter += "  ret.BeforeInsert[\"update_at\"] = db.ChangeNow\n"
        //    TableFilter += "  ret.BeforeUpdate[\"update_at\"] = db.ChangeNow\n\n"
        //}
    }
    needFieldStrs := "\"" + strings.Join(FieldMapNeedList, "\",\"") + "\""
    tableMapField =  strings.Replace(tableMapField, "{NeedFieldList}", needFieldStrs, -1)
    tableMapFieldList += tableMapField
    return tableMapFieldList, FieldMapFieldList, strings.Join(FieldGetSetList, "\n")
}

