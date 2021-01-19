package db

import (
    "fmt"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "strings"
    "time"
)

type Orm struct {
    *Query        // 数据操作
    Name string           // 表名
    Pk string             // 主键
    Fields []string       // 其他字段
    Filter map[string]*FilterConf // 所有字段的过滤条件
    BeforeInsert map[string]ChangeFunc   // 插入前修改
    BeforeUpdate map[string]ChangeFunc   // 更新前修改
}

type ChangeFunc  func(m map[string]interface{}, name string)

func ChangeNow(m map[string]interface{}, name string){
    if _, ok := m[name]; !ok {
        m[name] = time.Now().Format("2006-01-02 15:04:05")
    }
}

// 插入之前
func (c *Orm)beforeInsert(m map[string]interface{}){
    if c.BeforeInsert != nil {
        for name, f := range c.BeforeInsert {
            f(m, name)
        }
    }
}

// 更新之前
func (c *Orm)beforeUpdate(m map[string]interface{}){
    if c.BeforeUpdate != nil {
        for name, f := range c.BeforeUpdate {
            f(m, name)
        }
    }
}

// 分页-分页布局
func (c *Orm)Page(pageSize int, pageNum int) *Query{
    // 整理数据
    if pageSize <=1 {
        pageSize = 1
    }
    if pageNum <=0 {
        pageNum = 1
    }
    skip := pageSize * (pageNum -1)
    limit := pageSize
    return c.Select("*").From(c.Name).Limit(skip, limit)
}

// 分页-流式布局
func (c *Orm)Flow(pageSize int, flowWhere string, FlowArgs ...interface{}) *Query{
    // 整理数据
    if pageSize <=1 {
        pageSize = 1
    }
    skip := 0
    limit := pageSize
    return c.Select("*").From(c.Name).Where(flowWhere, FlowArgs...).Limit(skip, limit)
}

// 创建
func (c *Orm) Create(args interface{}) (int64, error){
    argMap := args.(map[string]interface{})
    c.beforeInsert(argMap)
    // 筛检数据->检测格式->必填
    dataJson, err := c.Check(argMap, true)
    if err !=nil {
        return 0, err
    }
    // 判断格式 必填项
    // 插入数据
    insertId, _ := c.Insert(dataJson).
        Into(c.Name).
        Query().
        LastInsertId()
    return insertId, nil
}

// 筛检数据->检测格式->必填
func (c *Orm) Check(argMap map[string]interface{}, checkNeed bool) (ret []map[string]interface{}, err error) {
    dataArr, err := util.ListMapField(argMap, c.Fields)
    if err != nil {
        return nil, ecode.OrmWrongArgType.Error()
    }
    // 格式检测
    var errArr []string
    for _, info:= range dataArr{ // 数组
        for k, v := range info{ // 具体的一个line
            if c.Filter[k].Check(v) == false {
                errArr = append(errArr, k)
            }
        }
    }
    // 必填检测
    var errNeed []string
    if checkNeed { // 如果检测必填
        for k, v := range c.Filter{
            if v.Need { // 如果需要必填
                if _, ok :=dataArr[0][k]; !ok{
                    errNeed = append(errNeed, k)
                }
            }
        }
    }
    if len(errArr)>0 {
        return nil, ecode.OrmWrongColumnsType.Error(strings.Join(errArr, ","))
    }
    if len(errNeed)>0 {
        return nil, ecode.OrmMissColumnsNeed.Error(fmt.Sprint(errNeed))
    }

    return dataArr, nil
}

//////////////////////////////////////////////////////////
//  对一行记录操作
//////////////////////////////////////////////////////////
type ormLine struct {
    *Orm
    useCache bool
    pkId  int64
}
// 创建新的对象
func (c *Orm) Line(pkId int64)  *ormLine{
    return  &ormLine{
        Orm: c,
        pkId: pkId,
    }
}

// 删除
func (obj *ormLine) Remove() int64{
    lines, _ := obj.Delete().
        From(obj.Name).
        Where(obj.Pk + "=?", obj.pkId).
        Query().
        RowsAffected()
    return lines
}

// 编辑
func (obj *ormLine) Edit(argMap map[string]interface{}) (int64, error){
    obj.beforeUpdate(argMap)

    // 筛检数据->检测格式->必填
    dataJson, err := obj.Check(argMap, false)
    if err !=nil {
        return 0, err
    }

    ret, _ := obj.Update(obj.Name).
        Set(dataJson[0]).
        Where(obj.Pk +"=?", obj.pkId).
        Query().
        RowsAffected()
    return ret, nil
}

func (obj *ormLine) Cache(c bool) *ormLine{
    obj.useCache = c
    return obj
}

func (obj *ormLine) Info() map[string]interface{}{
    ret, _ := obj.Select("*").
        From(obj.Name).
        Where(obj.Pk + "=?", obj.pkId).
        Limit(0,1).
        Cache(obj.useCache).
        Query().
        Array()
    if len(ret) == 0 {
        return nil
    }
    return ret[0]
}


///////// 生成ORM
type OrmField struct {
    Name      string
    MysqlType string
    Unsigned  bool
    IsNeed    bool
}

type ormTable struct {
    Name      string
    Pk        string
    FieldArr    []string
    Field    map[string]*OrmField
}

// 根据各个数据库的句柄，创建ORM池
func OrmInit(confName, app string, ormPath string){
    // 获取表结构
    conn := MYSQL.GetPool(confName)
    if conn == nil {
        panic("数据库配置名错误: configName=" + confName)
    }
    var (
        dbName = conn.Conf.W.Database
        tableArr   []string
        db, _ = MYSQL.Get(confName)
        //result = make(map[string]*OrmTable)
    )

    // 获取所有的表
    rs, _ := db.Query("SHOW FULL TABLES WHERE table_type = 'BASE TABLE'").Array()
    for _, line:= range rs {
        tableArr = append(tableArr, line["Tables_in_" + dbName].(string))
    }
    fmt.Println("获取表结构", tableArr)

    // 循环表获取各个字段
    for _, table := range tableArr {
        // 获取有默认值的fields
        rs, _ = db.Query("show COLUMNS FROM `" + table + "` where not ISNULL(`Default`)").Array()
        var defaultField = make(map[string]bool)

        for _, v := range rs {
            defaultField[v["Field"].(string)] = true
        }
        rs, _ = db.Query("show COLUMNS FROM `" + table + "`").Array()
        tableColumns := make(map[string]*OrmField)
        var (
            tablePk string
            tableField  []string
            isNeed bool
            unsigned bool
        )

        for _, line := range rs{
            isNeed = true
            if line["Key"] == "PRI" {
                tablePk = line["Field"].(string)
            } else {
                tableField = append(tableField, line["Field"].(string))
            }
            if line["Extra"].(string) == "auto_increment" { // 有自增ID，就不必填
                isNeed = false
            }

            if line["Null"]== "YES"  {
                isNeed = false
            } else if _, ok :=defaultField[line["Field"].(string)]; ok {
                isNeed = false
            }
            splitType := strings.Split(line["Type"].(string), " ")
            if len(splitType) ==2 && splitType[1]=="unsigned" {
                unsigned = true
            } else {
                unsigned = false
            }
            tableColumns[line["Field"].(string)] =  &OrmField{
                MysqlType: splitType[0],
                Unsigned: unsigned,
                IsNeed: isNeed,
            }
        }

        // 创建 一个ORM 表的GO文件
        createOrmTable(ormTable{
            Name: table,
            Pk: tablePk,
            FieldArr: tableField,
            Field: tableColumns,
        }, app, ormPath)
    }

}

// 创建ORM文件  /orm/$app+_tmp/ [$table1.go, $table2.go, ...]
func createOrmTable(table ormTable, app string, ormPath string){
    var ret = `package {package}

import (
  "github.com/jasonfeng1980/pg"
  "github.com/jasonfeng1980/pg/database/db"
)

type {Table.Name} struct {
  db.Orm
}

func {Table}() *{Table.Name}{
  ret := &{Table.Name}{}
  q, ok := pg.MySQL.Get("{package}")
  if !ok {
    panic("数据库配置名称错误 ：{package}" )
  }
  ret.Query = q
  ret.Name = "{Table.Name}"
  ret.Pk = "{Table.Pk}"
  ret.Fields = []string{{Table.Fields}}
  // filter 设置
  ret.Filter = make(map[string]*db.FilterConf)
  f := pg.Filter.MySQL
{Table.Filter}
  return ret
}
`
    var ormTplFilter = "  ret.Filter[\"{Table.Field}\"] = f(\"{Filter.mysqlType}\", {Filter.unsigned}, {Filter.isNeed})\n"

    TableFilter := "  ret.BeforeInsert = make(map[string]db.ChangeFunc)\n  ret.BeforeUpdate = make(map[string]db.ChangeFunc)\n\n"

    for field, filterObj := range table.Field{
        filterTpl := ormTplFilter
        unsigned, _ := util.Str(filterObj.Unsigned)
        isNeed, _ := util.Str(filterObj.IsNeed)
        filterTpl =  strings.Replace(filterTpl, "{Table.Field}", field, -1)
        filterTpl =  strings.Replace(filterTpl, "{Filter.mysqlType}", filterObj.MysqlType, -1)
        filterTpl =  strings.Replace(filterTpl, "{Filter.unsigned}", unsigned, -1)
        filterTpl =  strings.Replace(filterTpl, "{Filter.isNeed}", isNeed, -1)
        filterTpl =  strings.Replace(filterTpl, "{Table.Name}", table.Name, -1)
        TableFilter += filterTpl

        // 增加入口前的 默认数据修改
        if field == "create_at" {
            TableFilter += "  ret.BeforeInsert[\"create_at\"] = db.ChangeNow\n\n"
        }
        if field == "update_at" {
            TableFilter += "  ret.BeforeInsert[\"update_at\"] = db.ChangeNow\n"
            TableFilter += "  ret.BeforeUpdate[\"update_at\"] = db.ChangeNow\n\n"
        }
    }

    ret =  strings.Replace(ret, "{package}", app, -1)
    ret =  strings.Replace(ret, "{Table}", util.StrUFirstForSplit(table.Name, "_"), -1)
    ret =  strings.Replace(ret, "{Table.Name}", table.Name, -1)
    ret =  strings.Replace(ret, "{Table.Pk}", table.Pk, -1)
    ret =  strings.Replace(ret, "{Table.Fields}", "\"" + strings.Join(table.FieldArr, "\",\"") + "\"", -1)
    ret =  strings.Replace(ret, "{Table.Filter}", TableFilter, -1)

    ormPath = util.FileRealPath(ormPath)
    dir := ormPath + "/" + app + "_tmp/"
    fileName := table.Name + ".go"

    if err := util.FileWrite(dir, fileName, ret); err!=nil{
        fmt.Println("创建ORM文件：", dir, fileName , "  失败! -- Error")
        fmt.Println(err.Error())
    } else {
        fmt.Println("创建ORM文件：", dir, fileName , "  成功! -- OK")
    }
}


