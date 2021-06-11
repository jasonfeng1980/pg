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
    "time"
)

type queryOptions struct {
    // select
    Selects bson.M
    From string  // 表名
    Where []interface{}
    Having []interface{}
    OrderBy bson.D
    Limit int64
    Skip  int64
    Args map[string][]interface{}
    Insert []interface{} // 插入数据
    Into []string     // 表名
    Update []string   // 表名
    Upsert  bool       // 更新,没有就写入
    OnlyOne bool      // 只处理一个
    Set bson.D // 设置值的mapping
    Replace []string  // 表名
    Delete bool   // 使用了delete
    GroupBy map[string]interface{}  // 分组
    ForUpdate bool    // FOR UPDATE
    useCache bool    // 使用缓存
    Pipe     []bson.D  // pipe 有序list
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
        Args: make(map[string][]interface{}),
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

    q.options.Selects = bson.M{}
    for _, s:= range fields{
        q.options.Selects[s] = 1
    }
    return q
}

// 哪个表
func (q *Query)From(table string) *Query{
    q.options.From = table
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
    q.readWhere(&q.options.Where, args...)
    return q
}

func (q *Query)GroupByMap(group map[string]interface{}) *Query{
    q.options.GroupBy = group
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

    q.options.GroupBy = newW
    return q
}

// having
func (q *Query)Having(args  ...interface{}) *Query{
    q.readWhere(&q.options.Having, args...)
    return q
}

// 排序order by
func (q *Query)OrderBy(orderBy string) *Query{
    orderByArr := strings.Split(orderBy, ",")
    //q.options.orderBy = append(q.options.orderBy, orderByArr...)

    q.options.OrderBy = bson.D{}
    for _,v := range orderByArr{
        isAsc := 1
        if strings.Index(v, " desc")>0 {
            isAsc = -1
        }
        sortKey := strings.Replace(v, " desc", "", 1)
        sortKey = strings.Replace(sortKey, " asc", "", 1)
        sortKey = strings.Trim(sortKey, " ")
        q.options.OrderBy = append(q.options.OrderBy, bson.E{sortKey, isAsc})
    }

    return q
}

// limit 数量限制 skip跳过几条
func (q *Query)Limit(skip int, limit int) *Query{
    limit64 := int64(limit)
    skip64 := int64(skip)
    q.options.Limit = limit64
    q.options.Skip = skip64
    return q
}

func (q *Query)Insert(data interface{}) *Query{
    q.Clear()
    q.options.Insert = q.toList(data)
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
    q.options.Into = append(q.options.Into, table)
    return q
}

// 更新哪个表
func (q *Query)Update(table string) *Query{
    q.Clear()
    q.options.Update = append(q.options.Update, table)
    return q
}
// 更新，没有就插入
func (q *Query)Upsert(b bool) *Query{
    q.options.Upsert = b
    return q
}
// 只处理一次  DeleteOne  UpdateOne
func (q *Query)One(b bool) *Query{
    q.options.OnlyOne = b
    return q
}
func (q *Query)Set(set bson.D) *Query{
    q.options.Set = set
    return q
}
// replace哪个表
func (q *Query)Replace(table string) *Query{
    q.Clear()
    q.options.Replace = append(q.options.Replace, table)
    return q
}

// 删除哪个表
func (q *Query)Delete() *Query{
    q.Clear()
    q.options.Delete = true
    return q
}

// 缓存SQL
func (q *Query)CacheSql(ctx context.Context, c conf.MongoConf, op *queryOptions) *rdb.String{
    argByte, _ :=util.JsonEncode(op)
    cacheKey := fmt.Sprintf("%s/%s->%s", c.Dns, c.Database, string(argByte[:]))
    return &rdb.String{
        Key: rdb.Key{
            CTX: ctx,
            Name: "MDB_CACHE:" + util.StrMd5(cacheKey),
            Client: MONGO.CacheRedisClient,
            Expr: MONGO.CacheExpr,
        },
    }
}


// 将mongo命令整理成sql语句
func (q *Query)buildSql() (retSql string){
    var o = q.options

    // 初始命令
    if len(o.Update) >0 && len(o.Set)>0 {
        retSql = "UPDATE " + strings.Join(o.Update, ", ")
    } else if len(o.Insert) >0 && len(o.Into) >0 {
        retSql = "INSERT INTO "
    } else if o.Delete {
        retSql = "DELETE "
    } else if len(o.Replace) >0 {
        retSql = "REPLACE INTO " + strings.Join(o.Replace, ", ")
    } else {
        retSql = "SELECT " + strings.Join(util.MapKeys(o.Selects), ", ")
    }

    // 整理FROM
    if len(o.From) >0 {
        retSql = retSql + " FROM " + o.From
    }


    // 整理SET
    if len(o.Set) >0 {
        var setArr []string
        for _, v := range o.Set {
            setArr = append(setArr, v.Key + ":" + util.StrParse(v.Value) + ",")
        }
        retSql = retSql + " SET {" + strings.Join(setArr, ", ") + "}"
    }

    // 整理where
    if len(o.Where) >0 {
        whereT, _ := util.JsonEncode(o.Where)
        retSql = retSql + " WHERE " + string(whereT[:])
    }

    // 整理group by
    if len(o.GroupBy) >0 {
        groupByT, _ := util.JsonEncode(o.GroupBy)
        retSql = retSql + " GROUP BY " + string(groupByT[:])
    }

    // 整理having
    if len(o.Having) >0 {
        havingT, _ := util.JsonEncode(o.Having)
        retSql = retSql + " HAVING " + string(havingT[:])
    }

    // 整理order by
    if len(o.OrderBy) >0 {
        var orderArr []string
        for _, v := range o.OrderBy {
            orderArr = append(orderArr, v.Key + ":" + util.StrParse(v.Value) + ",")
        }
        retSql = retSql + " ORDER BY {" + strings.Join(orderArr, ", ") + "}"
    }

    // 整理LIMIT
    if o.Limit >0 {
        retSql = retSql + fmt.Sprintf(" LIMIT %d, %d", o.Skip, o.Limit)
    }

    // 整理for update
    if o.ForUpdate == true {
        retSql = retSql + " FOR UPDATE"
    }
    return
}
// 统计配置
func (q *Query)countOptions() *options.CountOptions{
    countOptions := &options.CountOptions{
        MaxTime: &q.Conf.Timeout,
    }
    if q.options.Skip >0 { // skip
        countOptions.SetSkip(q.options.Skip)
    }
    if q.options.Limit >0 { // limit
        countOptions.SetLimit(q.options.Limit)
    }
    return countOptions
}
// select查询配置
func (q *Query)findOptions() *options.FindOptions {
    findOptions := options.Find()
    if len(q.options.Selects) !=0 { // select
        findOptions.SetProjection(q.options.Selects)
    }
    if len(q.options.OrderBy) != 0 { // order by
        findOptions.SetSort(q.options.OrderBy)
    }
    if q.options.Skip >0 { // skip
        findOptions.SetLimit(q.options.Skip)
    }
    if q.options.Limit >0 { // limit
        findOptions.SetLimit(q.options.Limit)
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
    if len(q.options.Where) !=0 { // where
        pipeline = append(pipeline, bson.D{{"$match",bson.D{{"$and", q.options.Where}}}})
    }
    if q.options.GroupBy != nil { // group
        pipeline = append(pipeline, bson.D{{"$group", q.options.GroupBy}})
    }
    if len(q.options.OrderBy) != 0 { // order by
        pipeline = append(pipeline, bson.D{{"$sort", q.options.OrderBy}})
    }
    if len(q.options.Having) != 0 { // having
        pipeline = append(pipeline, bson.D{{"$match",bson.D{{"$and", q.options.Having}}}})
    }
    if q.options.Skip >0 { // skip
        pipeline = append(pipeline, bson.D{{"$skip",q.options.Skip}})
    }
    if q.options.Limit >0 { // limit
        pipeline = append(pipeline, bson.D{{"$limit",q.options.Limit}})
    }
    if len(q.options.Selects) !=0 { // select
        pipeline = append(pipeline, bson.D{{"$project",q.options.Selects}})
    }
    return pipeline
}

func (q *Query)Count() (int64, error){
    if len(q.options.Update)>0 || len(q.options.Insert)>0 || q.options.Delete ||  q.options.GroupBy!=nil {
        return 0, ecode.MdbCountErr.Error()
    }
    ctx     := context.Background()
    database := q.Conn
    collection := database.Collection(q.options.From)
    countOptions := q.countOptions()
    where   := bson.D{}
    if len(q.options.Where) > 0 {
        where = bson.D{{"$and", q.options.Where}}
    }
    return collection.CountDocuments(ctx, where, countOptions)
}

func (q *Query)Cache(useCache bool) *Query{
    if MONGO.CacheRedisClient != nil { // 如果配置了 缓存redis
        q.options.useCache = useCache
    } else {
        util.Log.Infoln("没有配置缓存redis，无法使用MongoDB CACHE")
    }
    return q
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
    if q.options.From != "" {
        collection = q.Conn.Collection(q.options.From)
    } else if len(q.options.Update) > 0 {
        collection = q.Conn.Collection(q.options.Update[0])
    } else if len(q.options.Into) > 0 {
        collection = q.Conn.Collection(q.options.Into[0])
    } else { // 没有指定 Collection
        ret.Err = ecode.MdbCollectionIsNil.Error()
        return
    }

    // where
    if len(q.options.Where) > 0 {
        where = bson.D{{"$and", q.options.Where}}
    }

    t := time.Now()
    cacheGet := false
    cacheKeyName := ""
    defer func() {
        if !util.Log.ShowDebug {
            return
        }
        driver := "MONGO"
        if cacheGet {
            driver = "REDIS缓存-MONGO"
        }
        util.Log.Logger.With("driver", driver, "cacheName", cacheKeyName,
            "useTime", time.Since(t) ).Debug(q.buildSql())
    }()

    if q.options.useCache { // 使用缓存
        cacheKey := q.CacheSql(ctx, q.Conf, q.options)
        cacheKeyName = cacheKey.Name
        ret.useCache = true
        ret.cacheKey = cacheKey
        if rs, err := cacheKey.Get(); err == nil {
            cacheGet = true
            ret.cacheDate = rs
            return ret
        }
    }

    if q.options.GroupBy == nil { // 不用Aggregate
        if q.options.Delete { // 删除
            var deleteResult *mongo.DeleteResult
            if q.options.OnlyOne {
                deleteResult, ret.Err = collection.DeleteOne(ctx, where)
            } else {
                deleteResult, ret.Err = collection.DeleteMany(ctx, where)
            }
            if ret.Err == nil {
                ret.RowsForAffected = deleteResult.DeletedCount
            }
        } else if len(q.options.Update)>0 { // 更新
            updateOption := options.Update()
            var updateResult *mongo.UpdateResult
            if q.options.Upsert {
                updateOption.SetUpsert(true)
            }
            if q.options.OnlyOne {
                updateResult, ret.Err = collection.UpdateOne(ctx, where, q.options.Set, updateOption)
            } else {
                updateResult, ret.Err = collection.UpdateMany(ctx, where, q.options.Set, updateOption)
            }
            if ret.Err == nil {
                ret.RowsForAffected = updateResult.ModifiedCount + updateResult.UpsertedCount
            }
        } else if len(q.options.Insert) >0 { // 插入
            ret.InsertManyResult, ret.Err =collection.InsertMany(ctx, q.options.Insert)
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
    return r.RowsForAffected, nil
}
// 获取查询数据 - []map[string]string
func (r *Result) Array() ([]map[string]interface{}, error){
    var result = make([]map[string]interface{}, 0)
    err := r.Bind(&result)
    return result, err
}

func (r *Result) Bind(dest interface{}) error{
    // 判断缓存 如果有直接返回
    if r.useCache && r.cacheDate!="" { // 如果允许缓存，并且有缓存
        if err :=util.JsonDecode([]byte(r.cacheDate), &dest); err ==nil {
            return err
        } else { // 缓存decode出错
            panic("decode缓存数据出错")
        }
    }

    // mongo处理
    if r.Err != nil {
        return r.Err
    }
    if r.Cursor == nil {
        return ecode.MdbNotExecData.Error()
    }
    defer r.Cursor.Close(r.ctx)

    if err := r.Cursor.All(r.ctx, dest); err != nil {
        return err
    }
    // 如果要缓存，记录缓存数据
    if r.useCache && r.cacheKey != nil {
        if rs, err := util.JsonEncode(dest); err==nil {
            r.cacheKey.Set(rs)
        }
    }
    return nil
}