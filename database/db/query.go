package db

import (
    "context"
    "database/sql"
    "fmt"
    "github.com/jasonfeng1980/pg/database/rdb"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "reflect"
    "sort"
    "strings"
    "time"
)

//////////////////////////////////////////////
//
//  QUERY  执行的query
//
//////////////////////////////////////////////
type queryJoin struct {
    joinType string
    table string
    on    string
}

type queryOptions struct {
    // select
    selects []string
    from []string  // 表名
    join []*queryJoin
    where []string
    having []string
    orderBy []string
    limit int
    skip  int
    args map[string][]interface{}
    insert []map[string]interface{} // 插入数据
    into []string     // 表名
    update []string   // 表名
    set map[string]interface{} // 设置值的mapping
    replace []string  // 表名
    delete bool   // 使用了delete
    groupBy []string  // 分组
    forUpdate bool    // FOR UPDATE
    tx     *sql.Tx   // 事务句柄
    useCache bool    // 使用缓存
}

// SQL query
type Query struct {
    options *queryOptions
    db *Conn
    tx *sql.Tx
}


// 清除options
func (q *Query)Clear(){
    q.options = &queryOptions{
        args: make(map[string][]interface{}),
    }
}

// 查询
func (q *Query)Select(fields string)  *Query{
    // 初始化option
    q.Clear()
    var selectOptions []string
    if fields == "" {
        fields = "*"
    }
    // 去除fields之间的空格，变成数组
    var arr = strings.Split(fields, ",")
    for _, v := range arr {
        selectOptions = append(selectOptions, strings.Trim(v, " "))
    }
    q.options.selects = selectOptions
    return q
}

// 哪个表
func (q *Query)From(table string) *Query{
    q.options.from = append(q.options.from, table)
    return q
}

// join联合其他表
func (q *Query)Join(table string, on string) *Query{
    q.options.join = append(q.options.join, &queryJoin{
        joinType: "JOIN",
        table: table,
        on: on,
    })
    return q
}

// 左联其他表
func (q *Query)LeftJoin(table string, on string) *Query{
    q.options.join = append(q.options.join, &queryJoin{
        joinType: "LEFT JOIN",
        table: table,
        on: on,
    })
    return q
}

// 右联其他表
func (q *Query)RightJoin(table string, on string) *Query{
    q.options.join = append(q.options.join, &queryJoin{
        joinType: "RIGHT JOIN",
        table: table,
        on: on,
    })
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

// where条件
func (q *Query)Where(where interface{}, args  ...interface{}) *Query{
    var whereSql string
    switch where.(type) {
    case string:
        whereSql = where.(string)
    case map[string]interface{}:
        for k, v := range where.(map[string]interface{}) {
            if whereSql == "" {
                whereSql = fmt.Sprintf("%s=? ", k)
            } else {
                whereSql = fmt.Sprintf("%s and %s=? ", whereSql, k)
            }
            args = append(args, v)
        }
    default:
        mList := q.toList(where)
        return q.Where(mList[0])
        //panic(ecode.DbWrongWhere.Error())
    }
    if strings.Index(strings.ToLower(whereSql), " or ") > -1 { // 包含or
        whereSql =  "( " + whereSql + " )"
    }
    q.options.where = append(q.options.where, whereSql)
    for _ ,v := range args{
        q.options.args["where"] = append(q.options.args["where"], v)
    }
    return q
}

func (q *Query)GroupBy(fields string) *Query{
    q.options.groupBy = append(q.options.groupBy, fields)
    return q
}

// having
func (q *Query)Having(havingSql string, args  ...interface{}) *Query{
    q.options.having = append(q.options.having, havingSql)
    for _ ,v := range args{
        q.options.args["having"] = append(q.options.args["having"], v)
    }
    return q
}

// 排序order by
func (q *Query)OrderBy(orderSql string) *Query{
    q.options.orderBy = append(q.options.orderBy, orderSql)
    return q
}

// limit 数量限制 skip跳过几条
func (q *Query)Limit(skip int, limit int) *Query{
    q.options.limit = limit
    q.options.skip = skip
    return q
}

func (q *Query)Insert(data interface{}) *Query{
    q.Clear()
    q.options.insert = q.toList(data)
    return q
}

func (q *Query)toList(h interface{}) (insertData []map[string]interface{}){
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

func (q *Query)buildInsertSql() (sql string, args []interface{}){
    data := q.options.insert
    if len(data) ==0 {
        panic("没有插入数据")
        return
    }
    var keys []string
    var valueArr []string

    // 获取name
    for keyName, _ := range data[0] {
        keys = append(keys, keyName)
        valueArr = append(valueArr, "?")
    }
    sort.Strings(keys[:])

    // 生成头SQL
    sql += fmt.Sprintf(" (%s) VALUES ", strings.Join(keys, ",") )
    // 生成 values
    for k, v := range data {
        for _, name := range keys {
            args = append(args, v[name])
        }
        format := " (%s) "
        if k>0 {
            format = ", (%s) "
        }
        sql += fmt.Sprintf(format,  strings.Join(valueArr, ","))

    }
    return
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

func (q *Query)Set(set map[string]interface{}) *Query{
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

// 开始事务
func (q *Query)StartTransaction() *Query{
    if tx, err :=q.db.Writer.Begin(); err==nil {
        return &Query{
            options: q.options,
            db: q.db,
            tx: tx,
        }
    } else {
        panic("开启事务失败")
    }
}

// 提交事务
func (q *Query)Commit() error{
    err :=q.tx.Commit()
    q.tx = nil
    return err
}

// 回滚
func (q *Query)Rollback() error{
    err := q.tx.Rollback()
    q.tx = nil
    return err
}

// 锁
func (q *Query)ForUpdate() *Query{
    q.options.forUpdate = true
    return q
}

func (q *Query)Cache(useCache bool) *Query{
    if MYSQL.CacheRedisClient != nil { // 如果配置了 缓存redis
        q.options.useCache = useCache
    } else {
        panic("没有配置缓存redis，无法使用MYSQL CACHE")
    }
    return q
}

// 查询
func (q *Query)Query(args ...interface{}) *Result{
    var query string
    //根据参数做不同的判断 获取 sql  和 args
    argLen := len(args)
    switch argLen{
    case 0:
        query, args = q.buildSql()
    case 1:
        query = args[0].(string)
        args = nil
    default:
        query = args[0].(string)
        args = args[1:]
    }

    t := time.Now()
    cacheGet := false
    defer func() {
        driver := "MYSQL"
        if cacheGet {
            driver = "REDIS缓存"
        }
        util.Log.Logger.With("driver", driver, "query", query, "args", fmt.Sprint(args),
            "useTime", time.Since(t) ).Debug()
    }()

    if argLen == 1{
        return q.execForArray(query)
    } else if len(q.options.selects)>0 {
        if q.options.useCache { // 使用缓存
            cacheKey := q.CacheSql(context.Background(), q.db.Conf.W.Database, query, args)
            if rs, err := cacheKey.Get(); err == nil {
                cacheGet = true
                return &Result{
                    useCache:  true,
                    query:     query,
                    args:      args,
                    cacheKey: cacheKey,
                    cacheDate: rs,
                    Err:       err,
                }
            } else { // 没有取到缓存
                ret := q.execForArray(query, args...)
                ret.useCache = q.options.useCache
                ret.cacheKey = cacheKey
                return ret
            }
        }
        // 不需要缓存
        return q.execForArray(query, args...)
    } else {
        return q.execForLineNum(query, args...)
    }
}

// 缓存SQL
func (q *Query)CacheSql(ctx context.Context, dbConfName, sql string, args []interface{}) *rdb.String{
    argStr := fmt.Sprint(args)
    return &rdb.String{
        Key: rdb.Key{
            CTX: ctx,
            //Name: "SQL_CACHE:" + util.StrMd5(dbConfName + ":" + sql + "[" + strings.Join(argStr, ",") + "]"),
            Name: "SQL_CACHE:" + util.StrMd5(dbConfName + ":" + sql + " - "+ argStr),
            Client: MYSQL.CacheRedisClient,
            Expr: MYSQL.CacheExpr,
        },
    }
}


func (q *Query)buildSql() (retSql string, args []interface{}){
    var o = q.options

    // 初始命令
    if len(o.selects) >0 {
        retSql = "SELECT " + strings.Join(o.selects, ", ")
    } else if len(o.update) >0 && len(o.set)>0 {
        retSql = "UPDATE " + strings.Join(o.update, ", ")
    } else if len(o.insert) >0 && len(o.into) >0 {
        retSql = "INSERT INTO "
    } else if o.delete {
        retSql = "DELETE "
    } else if len(o.replace) >0 {
        retSql = "REPLACE INTO " + strings.Join(o.replace, ", ")
    }

    // 整理FROM
    if len(o.from) >0 {
        retSql = retSql + " FROM " + strings.Join(o.from, ", ")
    }

    // 整理INTO
    if len(o.into) >0 {
        retSql +=  strings.Join(o.into, ", ")
        insertInfo, insertArgs := q.buildInsertSql()
        retSql +=  insertInfo
        args = q.appendArgs(args, insertArgs)
    }

    // 整理SET
    if len(o.set) >0 {
        var setArr []string
        for k, v := range o.set {
            switch v.(type) {
            case expr:
                setArr = append(setArr, k + "=" + v.(expr).S)
            default:
                setArr = append(setArr, k + "=?")
                args = append(args, v)
            }
        }
        retSql = retSql + " SET " + strings.Join(setArr, ", ")
    }

    // 整理JOIN
    retSql = retSql + q.buildJoin()

    // 整理where
    if len(o.where) >0 {
        retSql = retSql + " WHERE " + strings.Join(o.where, " AND ")
        args = q.appendArgs(args, o.args["where"])
    }

    // 整理group by
    if len(o.groupBy) >0 {
        retSql = retSql + " GROUP BY " + strings.Join(o.groupBy, ", ")
    }

    // 整理having
    if len(o.having) >0 {
        retSql = retSql + " HAVING " + strings.Join(o.having, " AND` ")
        args = q.appendArgs(args, o.args["having"])
    }

    // 整理order by
    if len(o.orderBy) >0 {
        retSql = retSql + " ORDER BY " + strings.Join(o.orderBy, ", ")
    }

    // 整理LIMIT
    if o.limit >0 {
        retSql = retSql + fmt.Sprintf(" LIMIT %d, %d", o.skip, o.limit)
    }

    // 整理for update
    if o.forUpdate == true {
        retSql = retSql + " FOR UPDATE"
    }
    return
}

// 整理JOIN 的SQL
func (q *Query)buildJoin() (joinSql string){
    for _, v := range q.options.join{
        joinSql = joinSql + fmt.Sprintf(" %s %s ON %s", v.joinType, v.table, v.on)
    }
    return
}

func (q *Query)appendArgs(args []interface{}, addArgs []interface{}) []interface{} {
    //for _, v := range  addArgs {
    args = append(args, addArgs...)
    //}
    return args
}

// 获取影响条数
func (q *Query) execForLineNum(query string, args ...interface{}) *Result {
    db := q.db.Writer

    if dbErr := db.Ping(); nil != dbErr {
        panic("连接数据库失败: " + dbErr.Error())
    }

    var (
        handle *sql.Stmt
        err error
        res  sql.Result
    )
    if q.tx == nil { // SQL
        handle, err = db.Prepare(query)
    } else { // 事务
        handle, err = q.tx.Prepare(query)
    }

    useCache := false
    if q.options != nil && q.options.useCache{
        useCache = q.options.useCache
    }

    if err ==nil {
        res, err = handle.Exec(args...)
    }
    defer handle.Close()

    return &Result{
        useCache: useCache,
        query: query,
        args: args,

        ExecResult: res,
        Err: err,
    }
}

// 获取数组
func (q *Query) execForArray(query string, args ...interface{}) *Result{
    db := q.db.Reader
    if q.options !=nil && q.options.forUpdate{
        db = q.db.Writer
    }

    if dbErr := db.Ping(); nil != dbErr {
        panic("链接数据库失败: " + dbErr.Error())
    }

    rows, err := db.Query(query, args...)

    useCache := false
    if q.options != nil && q.options.useCache{
        useCache = q.options.useCache
    }

    return &Result{
        useCache: useCache,
        query: query,
        args: args,

        QueryResult: rows,
        Err: err,
    }

    //return result
}


/////////////////////////////////////////////////////
//
// result 处理
/////////////////////////////////////////////////////
// SQL query
type Result struct {
    useCache bool
    query string
    args  []interface{}
    cacheKey *rdb.String
    cacheDate string
    structFieldMap map[string]int   // 结构体 field对应关系 ， 使用tag - json
    QueryResult  *sql.Rows
    ExecResult  sql.Result
    Err  error
}


// 最后插入的ID
func (r *Result) LastInsertId() (int64, error){
    if r.Err != nil {
        return 0, r.Err
    }
    if r.ExecResult == nil {
        return 0, ecode.DbNotExecData.Error()
    }
    return r.ExecResult.LastInsertId()
}
// 影响的行数
func (r *Result) RowsAffected() (int64, error){
    if r.Err != nil {
        return 0, r.Err
    }
    if r.ExecResult == nil {
        return 0, ecode.DbNotExecData.Error()
    }
    return r.ExecResult.RowsAffected()
}
// 获取查询数据 - []map[string]string
func (r *Result) Array() ([]map[string]interface{}, error){
    var result = make([]map[string]interface{}, 0)
    err := r.Bind(&result)
    return result, err
}


// 获取查询数据 - 生成object
func (r *Result) Bind(dest interface{}) error{
    // 判断缓存 如果有直接返回
    if r.useCache && r.cacheDate!="" { // 如果允许缓存，并且有缓存
        if err :=util.JsonDecode([]byte(r.cacheDate), &dest); err ==nil {
            return err
        } else { // 缓存decode出错
            panic("decode缓存数据出错")
        }
    }

    // 走MYSQL处理
    if r.Err != nil {
        return r.Err
    }
    if r.QueryResult == nil {
        return ecode.DbNotQueryData.Error()
    }
    // 关闭row链接
    defer r.QueryResult.Close()


    t := reflect.TypeOf(dest)
    v := reflect.ValueOf(dest)

    typeErr := ecode.DbWrongType.Error()
    if t.Kind() != reflect.Ptr {
        return typeErr
    }
    //如果是用 var userPtr *User 方式声明的变量，则不可取址
    if !v.Elem().CanAddr() {
        return typeErr
    }

    v = v.Elem()
    t = t.Elem()

    column, err:= r.QueryResult.Columns()
    if err != nil {
        return err
    }
    switch t.Kind() {
    case reflect.Map:
        for r.QueryResult.Next() {
            m, err := r.setMap(r.QueryResult, t, column)
            if err != nil {
                return err
            }
            v.Set(m)
        }
    case reflect.Ptr, reflect.Struct:
        for t.Kind() == reflect.Ptr {
            t = t.Elem()
        }
        for r.QueryResult.Next() {
            destination, err := r.setStruct(r.QueryResult, t, column)
            if err != nil {
                return err
            }

            switch v.Kind() {
            case reflect.Ptr, reflect.Map:
                v.Set(destination)
            default:
                v.Set(destination.Elem())
            }
        }
    case reflect.Slice:
        dt := t.Elem()
        for dt.Kind() == reflect.Ptr {
            dt = dt.Elem()
        }
        sl := reflect.MakeSlice(t, 0, 0)

        for r.QueryResult.Next() {
            var destination reflect.Value
            if dt.Kind() == reflect.Map {
                destination, err = r.setMap(r.QueryResult, dt, column)
            } else {
                destination, err = r.setStruct(r.QueryResult, dt, column)
            }
            if err != nil {
                return err
            }
            //区分切片元素是否指针
            switch t.Elem().Kind() {
            case reflect.Ptr, reflect.Map:
                sl = reflect.Append(sl, destination)
            default:
                sl = reflect.Append(sl, destination.Elem())
            }
        }
        v.Set(sl)
    }
    // 如果要缓存，记录缓存数据
    if r.useCache && r.cacheKey != nil {
        if rs, err := util.JsonEncode(dest); err==nil {
            r.cacheKey.Set(rs)
        }
    }
    return nil
}

func (r *Result)getStructFieldMap(dest reflect.Value, columns []string) map[string]int{
    if r.structFieldMap == nil {
        r.structFieldMap = make(map[string]int)
        dest = dest.Elem()
        t := dest.Type()
        for n, l := 0, t.NumField(); n < l; n++ {
            tf := t.Field(n)
            if tf.Anonymous { // 嵌入字段
                continue
            }
            column := strings.Split(tf.Tag.Get("json"), ",")[0]
            if column == "" {
                continue
            }

            //只取选定的字段的地址
            for _, col := range columns {
                if col == column {
                    r.structFieldMap[col] = n
                    break
                }
            }
        }
    }
    return r.structFieldMap
}

func (r *Result)address(dest reflect.Value, columns []string) []interface{} {
    //dest = dest.Elem()
    //t := dest.Type()
    addrList := make([]interface{}, 0)
    switch dest.Elem().Type().Kind() {
    case reflect.Struct:
        structFieldMap := r.getStructFieldMap(dest, columns)
        for _, col := range columns {
            if n, ok := structFieldMap[col]; ok {
                addrList = append(addrList, dest.Elem().Field(n).Addr().Interface())
            }
        }
    default:
        addrList = append(addrList, dest.Elem().Addr().Interface())
    }
    return addrList
}


//适用于基类型和struct
func (r *Result)setStruct(rows *sql.Rows, t reflect.Type, columns []string) (reflect.Value, error) {
    dest := reflect.New(t)
    addrList := r.address(dest, columns)
    if len(columns) != len(addrList) {
        return reflect.ValueOf(nil), ecode.DbColumnsNotMatch.Error()
    }
    if err := rows.Scan(addrList...); err != nil {
        return reflect.ValueOf(nil), err
    }
    return dest, nil
}

// map 都是 interface
func (r *Result)setMap(rows *sql.Rows, t reflect.Type, columns []string) (reflect.Value, error) {
    if t.Elem().Kind() != reflect.Interface {
        return reflect.ValueOf(nil), ecode.DbWrongMap.Error()
    }
    m := reflect.MakeMap(t)
    addrList := make([]interface{}, len(columns))
    for idx := range columns {
        addrList[idx] = new(interface{})
    }
    if err := rows.Scan(addrList...); err != nil {
        return reflect.ValueOf(nil), err
    }
    for idx, column := range columns {
        //从指针剥出interface{}，再剥出实际值
        switch reflect.ValueOf(addrList[idx]).Elem().Elem().Kind() {
        case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // 整数
            m.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(reflect.ValueOf(addrList[idx]).Elem().Elem().Int()))
        case reflect.Float64, reflect.Float32: // 浮点数
            m.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(reflect.ValueOf(addrList[idx]).Elem().Elem().Float()))
        case reflect.String, reflect.Invalid: // 字符串
            m.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(reflect.ValueOf(addrList[idx]).Elem().Elem().String()))
        case reflect.Slice:
            m.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(string(reflect.ValueOf(addrList[idx]).Elem().Elem().Bytes())))
        default: // 默认
            m.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(reflect.ValueOf(addrList[idx]).Elem().Elem().Bytes()))
        }
        //m.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(addrList[idx]).Elem().Elem())
    }
    return m, nil
}

