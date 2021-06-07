package mdb

import (
    "context"
    "fmt"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/database/rdb"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "reflect"
    "strings"
)

type queryOptions struct {
    // select
    selects bson.M
    from string  // 表名
    where []interface{}
    having []interface{}
    orderBy bson.D
    limit int64
    skip  int64
    args map[string][]interface{}
    insert []interface{} // 插入数据
    into []string     // 表名
    update []string   // 表名
    upsert  bool       // 更新,没有就写入
    onlyOne bool      // 只处理一个
    set bson.D // 设置值的mapping
    replace []string  // 表名
    delete bool   // 使用了delete
    groupBy map[string]interface{}  // 分组
    forUpdate bool    // FOR UPDATE
    useCache bool    // 使用缓存
    pipe     []bson.D  // pipe 有序list
}

// SQL query
type Query struct {
    Conf    conf.MongoConf
    Conn    *mongo.Database
    options *queryOptions
}

// 清除options
func (q *Query)Clear(){
    q.options = &queryOptions{
        args: make(map[string][]interface{}),
    }
}

// 查询
func (q *Query)Select(fields ...string)  *Query{
    // 初始化option
    q.Clear()
    fieldLen := len(fields)
    if fieldLen==0 || fields[0] == "" || fields[0] == "*"{
        return q
    } else if fieldLen==1 {
        // 去除fields之间的空格，变成数组
        fields = strings.Split(fields[0], ",")
    }
    fields = util.ListTrim(fields, " ")

    q.options.selects = bson.M{}
    for _, s:= range fields{
        q.options.selects[s] = 1
    }
    return q
}

// 哪个表
func (q *Query)From(table string) *Query{
    q.options.from = table
    return q
}


func (q *Query)buildWhereSql(where interface{}) (str string) {
    switch where.(type) {
    case string:
        str = where.(string)
    case map[string]interface{}:

    default:
        panic(ecode.DbWrongWhere.Error())
    }
    return
}

func (q *Query)readWhere(obj *[]interface{}, args ...interface{}) {
    var where interface{}
    argLen := len(args)
    if argLen == 0 {
        return
    } else if argLen == 1 {
        where = args[0]
    } else {
        m := make(map[string]interface{})
        for i:=0; i<argLen; i=i+2{
            m[args[i].(string)] = args[i+1]
        }
        where = m
    }

    switch where.(type) {
    case string:
        var newW bson.M
        if err := util.JsonDecode(where, &newW); err!=nil{
            panic(err)
        }
        *obj = append(*obj, newW)
    case bson.A, bson.D, bson.E, bson.M, map[string]interface{}:
        *obj = append(*obj, where)
    default:
        mList := q.toList(where)
        for _, v := range mList {
            *obj = append(*obj, v)
        }
    }
}

// where条件
func (q *Query)Where(args  ...interface{}) *Query{
    q.readWhere(&q.options.where, args...)
    return q
}

func (q *Query)GroupByMap(group map[string]interface{}) *Query{
    q.options.groupBy = group
    return q
}

func (q *Query)GroupBy(group string) *Query{
    var newW bson.M
    if strings.HasPrefix(group, "{") { // 传的是JSON
        if err := util.JsonDecode(group, &newW); err!=nil{
            panic(err)
        }
    } else { // 只是单个字段
        newW = bson.M{
            "_id": "$" + group,
            "count": bson.M{"$sum": 1},
        }
    }

    q.options.groupBy = newW
    return q
}

// having
func (q *Query)Having(args  ...interface{}) *Query{
    q.readWhere(&q.options.having, args...)
    return q
}

// 排序order by
func (q *Query)OrderBy(orderBy string) *Query{
    orderByArr := strings.Split(orderBy, ",")
    //q.options.orderBy = append(q.options.orderBy, orderByArr...)

    q.options.orderBy = bson.D{}
    for _,v := range orderByArr{
        isAsc := 1
        if strings.Index(v, " desc")>0 {
            isAsc = -1
        }
        sortKey := strings.Replace(v, " desc", "", 1)
        sortKey = strings.Replace(sortKey, " asc", "", 1)
        sortKey = strings.Trim(sortKey, " ")
        q.options.orderBy = append(q.options.orderBy, bson.E{sortKey, isAsc})
    }

    return q
}

// limit 数量限制 skip跳过几条
func (q *Query)Limit(skip int, limit int) *Query{
    limit64 := int64(limit)
    skip64 := int64(skip)
    q.options.limit = limit64
    q.options.skip = skip64
    return q
}

func (q *Query)Insert(data interface{}) *Query{
    q.Clear()
    q.options.insert = q.toList(data)
    return q
}

func (q *Query)toList(h interface{}) (insertData []interface{}){
    t := reflect.ValueOf(h)
    switch t.Kind() {
    case reflect.Map:
        insertData = append(insertData, q.toMap(t))
    case reflect.Slice:
        len := t.Len()
        for i:=0; i<len; i++ {
            td := t.Index(i)
            for td.Kind() == reflect.Ptr { // 指针
                td = td.Elem()
            }
            if td.Kind() == reflect.Map {
                insertData = append(insertData, q.toMap(td))
            } else {
                panic("数据格式不对")
            }
        }
    default:
        panic("数据格式不对")
    }
    return
}

func (q *Query)toMap(t reflect.Value) map[string]interface{}{
    ret := make(map[string]interface{})
    for _,k := range t.MapKeys(){
        ret[k.String()] = t.MapIndex(k).Interface()
    }
    return ret
}

// 插入到什么表
func (q *Query)Into(table string) *Query{
    q.options.into = append(q.options.into, table)
    return q
}

// 更新哪个表
func (q *Query)Update(table string) *Query{
    q.Clear()
    q.options.update = append(q.options.update, table)
    return q
}
// 更新，没有就插入
func (q *Query)Upsert(b bool) *Query{
    q.options.upsert = b
    return q
}
// 只处理一次  DeleteOne  UpdateOne
func (q *Query)One(b bool) *Query{
    q.options.onlyOne = b
    return q
}
func (q *Query)Set(set bson.D) *Query{
    q.options.set = set
    return q
}
// replace哪个表
func (q *Query)Replace(table string) *Query{
    q.Clear()
    q.options.replace = append(q.options.replace, table)
    return q
}

// 删除哪个表
func (q *Query)Delete() *Query{
    q.Clear()
    q.options.delete = true
    return q
}

// 将mongo命令整理成sql语句
func (q *Query)buildSql() (retSql string, args []interface{}){// @todo
    //for k, v := range q.options.where {
    //    fmt.Println(k, v)
    //}
    return
}
// 统计配置
func (q *Query)countOptions() *options.CountOptions{
    countOptions := &options.CountOptions{
        MaxTime: &q.Conf.Timeout,
    }
    if q.options.skip >0 { // skip
        countOptions.SetSkip(q.options.skip)
    }
    if q.options.limit >0 { // limit
        countOptions.SetLimit(q.options.limit)
    }
    return countOptions
}
// select查询配置
func (q *Query)findOptions() *options.FindOptions {
    findOptions := options.Find()
    if len(q.options.selects) !=0 { // select
        findOptions.SetProjection(q.options.selects)
    }
    if len(q.options.orderBy) != 0 { // order by
        findOptions.SetSort(q.options.orderBy)
    }
    if q.options.skip >0 { // skip
        findOptions.SetLimit(q.options.skip)
    }
    if q.options.limit >0 { // limit
        findOptions.SetLimit(q.options.limit)
    }
    if q.Conf.AllowDiskUse {
        findOptions.SetAllowDiskUse(q.Conf.AllowDiskUse)
    }
    findOptions.SetMaxTime(q.Conf.Timeout)
    return findOptions
}
// pipe查询配置
func (q *Query)pipeline() mongo.Pipeline {
    pipeline := mongo.Pipeline{}
    if len(q.options.where) !=0 { // where
        pipeline = append(pipeline, bson.D{{"$match",bson.D{{"$and", q.options.where}}}})
    }
    if q.options.groupBy != nil { // group
        pipeline = append(pipeline, bson.D{{"$group", q.options.groupBy}})
    }
    if len(q.options.orderBy) != 0 { // order by
        pipeline = append(pipeline, bson.D{{"$sort", q.options.orderBy}})
    }
    if len(q.options.having) != 0 { // having
        pipeline = append(pipeline, bson.D{{"$match",bson.D{{"$and", q.options.having}}}})
    }
    if q.options.skip >0 { // skip
        pipeline = append(pipeline, bson.D{{"$skip",q.options.skip}})
    }
    if q.options.limit >0 { // limit
        pipeline = append(pipeline, bson.D{{"$limit",q.options.limit}})
    }
    if len(q.options.selects) !=0 { // select
        pipeline = append(pipeline, bson.D{{"$project",q.options.selects}})
    }
    return pipeline
}

func (q *Query)Count() (int64, error){
    if len(q.options.update)>0 || len(q.options.insert)>0 || q.options.delete ||  q.options.groupBy!=nil {
        return 0, ecode.MdbCountErr.Error()
    }
    ctx     := context.Background()
    database := q.Conn
    collection := database.Collection(q.options.from)
    countOptions := q.countOptions()
    where   := bson.D{}
    if len(q.options.where) > 0 {
        where = bson.D{{"$and", q.options.where}}
    }
    return collection.CountDocuments(ctx, where, countOptions)
}


func (q *Query)Query(args ...interface{}) (ret *Result){
    // 链接
    var (
        ctx     = context.Background()
    	collection *mongo.Collection
        where   = bson.D{}
    )
    ret     = &Result{ctx: ctx}

    // from
    if q.options.from != "" {
        collection = q.Conn.Collection(q.options.from)
    } else if len(q.options.update) > 0 {
        collection = q.Conn.Collection(q.options.update[0])
    } else if len(q.options.into) > 0 {
        collection = q.Conn.Collection(q.options.into[0])
    } else { // 没有指定 Collection
        ret.Err = ecode.MdbCollectionIsNil.Error()
        return
    }

    // where
    if len(q.options.where) > 0 {
        where = bson.D{{"$and", q.options.where}}
    }

    if q.options.groupBy == nil { // 不用Aggregate
        if q.options.delete { // 删除
            var deleteResult *mongo.DeleteResult
            if q.options.onlyOne {
                deleteResult, ret.Err = collection.DeleteOne(ctx, where)
            } else {
                deleteResult, ret.Err = collection.DeleteMany(ctx, where)
            }
            if ret.Err == nil {
                ret.RowsForAffected = deleteResult.DeletedCount
            }
        } else if len(q.options.update)>0 { // 更新
            updateOption := options.Update()
            var updateResult *mongo.UpdateResult
            if q.options.upsert {
                updateOption.SetUpsert(true)
            }
            if q.options.onlyOne {
                updateResult, ret.Err = collection.UpdateOne(ctx, where, q.options.set, updateOption);
            } else {
                updateResult, ret.Err = collection.UpdateMany(ctx, where, q.options.set, updateOption);
            }
            if ret.Err == nil {
                ret.RowsForAffected = updateResult.ModifiedCount + updateResult.UpsertedCount
            }
        } else if len(q.options.insert) >0 { // 插入
            ret.InsertManyResult, ret.Err =collection.InsertMany(ctx, q.options.insert)
        } else {
            findOptions := q.findOptions()
            ret.Cursor, ret.Err = collection.Find(ctx, where, findOptions)
        }
    } else {
        pipeline := q.pipeline()
        opts := options.Aggregate()
        if q.Conf.AllowDiskUse {
            opts.SetAllowDiskUse(q.Conf.AllowDiskUse)
        }
        opts.SetMaxTime(q.Conf.Timeout)
        fmt.Println(pipeline)
        ret.Cursor, ret.Err = collection.Aggregate(ctx, pipeline, opts)
    }


    return
}

/////////////////////////////////////////////////////
//
// result 处理
/////////////////////////////////////////////////////
// SQL query
type Result struct {
    ctx      context.Context
    useCache bool
    query string
    args  []interface{}
    cacheKey *rdb.String
    cacheDate string
    structFieldMap map[string]int   // 结构体 field对应关系 ， 使用tag - json
    Cursor  *mongo.Cursor
    InsertManyResult  *mongo.InsertManyResult
    RowsForAffected int64
    Err  error
}


// 最后插入的ID
func (r *Result) LastInsertId() ([]interface{}, error){
    if r.Err != nil {
        return nil, r.Err
    }
    if r.InsertManyResult == nil {
        return nil, ecode.MdbNotExecData.Error()
    }
    return r.InsertManyResult.InsertedIDs, nil
}
// 影响的行数
func (r *Result) RowsAffected() (int64, error){
    if r.Err != nil {
        return 0, r.Err
    }
    //if r.RowsForAffected == 0 {
    //    return 0, ecode.MdbNotExecData.Error()
    //}
    return r.RowsForAffected, nil
}
// 获取查询数据 - []map[string]string
func (r *Result) Array() ([]map[string]interface{}, error){
    var result = make([]map[string]interface{}, 0)
    err := r.Bind(&result)
    return result, err
}

func (r *Result) Bind(dest interface{}) error{
    if r.Err != nil {
        return r.Err
    }
    if r.Cursor == nil {
        return ecode.MdbNotExecData.Error()
    }
    defer r.Cursor.Close(r.ctx)

    return r.Cursor.All(r.ctx, dest)
}