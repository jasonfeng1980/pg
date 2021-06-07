# pg
方便PHPer 快速使用GO 搭建微服务平台
只需要简单配置，就可以实现路由、日志、熔断、限流、链路追踪、连接池
简化了MYSQL,MONGO,RabbitMQ,REDIS的API, 只需熟悉SQL语句，就可以使用这些搭建系统

### 推荐目录结构
```text
┌── README.md
├── apps                            # 服务目录
│   └── demo                          # 示例：一个服务文件夹
│       ├── usesr.go
│       └── auth.go
├── cmd                             # 命令目录
│   ├── batch                         # 批处理
│   │   └── client.go
│   └── micro                         # 微服务启动命令
│       └── sso.go
├── conf                            # 配置目录
│   ├── demo                          # 某一个服务的配置
│   │   └── pg_11_dev.yaml
│   ├── mysql.yaml                    # 全局mysql配置
│   └── redis.yaml                    # 全局redis配置
├── ecode                           # 错误code
│   └── ecode.go
├── go.mod
├── go.sum
├── log                             # 日志目录
│   ├── PG11.login.210601             # 服务名  服务编号 . 用户指定名称|默认日志级别 . 年月日
│   └── PG11.access.210601
├── orm                             # 生成的ORM目录 全局
│   └── demo
│       ├── company.go
│       └── company_member.go
├── ssl                             # SSL证书
│   ├── local.com.crt
│   └── local.com.key
├── upload                          # 上传目录
│   └── 2021-01-13
│       └── test.png
└── vendor                          # vendor目录
```
### 快速启用服务
```go
svc := pg.Server("../"")
svc.Run()
```
### 调用服务API
svc := pg.Client()  
data, code, msg := svc.Call(ctx, dns, pg.H{})  
dns  [grpc|http]://服务名称/module/version/action
```go
svc := pg.Client()
defer svc.Close()
# dns  grpc://服务名称/module/version/action
dns :=  "http://PG/auth/v1/login"
data, code, msg := svc.Call(context.Background(), dns, pg.M{
    "user_mobile": "186",
    "user_password": 11111,
})
util.Log.Infoln(data, code, msg)
```

### 请求的API
```go
api := pg.MicroApi()
api.Register("GET","orm", "v1", "page", ORMPage)
api.Register("GET","orm", "v1", "flow", ORMFlow)
func ORMPage(ctx context.Context, params map[string]interface{})(interface{}, int64, string) {
	ret := "成功"
	// do something .... 
	return pg.Success(ret)
}
```

### 加载API项目  
```go
// 启动服务时
import (
    "github.com/jasonfeng1980/pg"
    _ "github.com/jasonfeng1980/pg/example/api"
)
```
_ "github.com/jasonfeng1980/pg/example/api"  是 api所在的包

### 全局MYSQL配置
```yaml
# 别名
DEMO:
  # 写库 DNS
  W: mysql://root:@tcp(localhost:3306)/demo?charset=utf8mb4&maxOpen=200&maxIdle=100&maxLifetime=30
  # 读库 DNS
  R: mysql://root:@tcp(localhost:3306)/demo?charset=utf8mb4&maxOpen=200&maxIdle=100&maxLifetime=30
TEST:
  W: mysql://root:@tcp(localhost:3306)/demo?charset=utf8mb4&maxOpen=200&maxIdle=100&maxLifetime=30
  # 没有读库配置，就取写库的DNS
```

### 全局MONGO配置
```yaml
# 别名
USER:
  # DNS
  Dns: "mongodb://admin:root@localhost:27017"
  # 超时时间
  Timeout: 3
  # 数据库名
  Database: user
  # 是否允许使用Disk
  AllowDiskUse: 0

```

### 全局REDIS配置
```yaml
# 别名
DEMO:
  RedisType: CLIENT     # 类别 CLIENT 普通客户端；SENTINEL 哨兵； CLUSTER 集群
  Network:  tcp         # 链接方式 tcp | unix
  Addr:     docker.for.mac.host.internal:6379 # 服务地址，主机名:端口，默认localhost:6379
  Password:             # 密码  就是auth的部分
  DB:       0           # 数据库
  MasterName:           # 哨兵模式的 MasterName
  PoolSize: 40          # 连接池容量
  MinIdleConns: 10      # 闲置连接数量
  IdleTimeout: 300      # 空闲持续时间 默认300秒
  DialTimeout: 2        # 连接超时时间 单位秒
  ReadTimeout: 2        # 读超时时间 单位秒
  WriteTimeout: 2       # 写超时时间 单位秒
```

### 全局RabbitMQ配置
```yaml
USER:
  Dns:  "amqp://guest:guest@localhost:5672//test"
  Exchange:
    logs:
      Kind:   direct          # type fanout|direct|topic, durable
      Info:   [true, false, false, false] # durable, auto-deleted, internal, no-wait
      Args:     # x-expires, x-max-length, x-max-length-bytes, x-message-ttl, x-max-priority, x-queue-mode, x-queue-master-locator
        x-expires: 300
      Query:
        q_log:                    # name
          Routing: [q_log_q1, q_log_q2]        # 路由KEY
          Info:   [true, false, false, false]       # durable 持久化, autoDelete 自动删除, exclusive 排他, NoWait 不需要服务器的任何返回
          Delay:  [2, 5, 10, 100, 600, 1800, 3600, 86400] # 延时多少秒再次到队列里执行
          Qos:    [1, 0, 0]        # count, size, global

```

### 微服务配置
```yaml
ServerName: PG       # 服务名称
ServerNo:   11       # 服务序号

# 日志配置
LogDir:                # 日志文件夹 info.年月日.log | error.201012.log
LogShowDebug: true     # 是否记录测试日志

# 网站配置
WebMaxBodySizeM:     32        # 最大允许上传的大小，单位M
WebReadTimeout:     10 # 读取超时时间
WebWriteTimeout:    30 # 写入超时时间

# 微服务配置
DebugAddr:  :8180    # 测试服务地址
HttpAddr:   :8181    # HTTP服务地址
HttpsInfo:
  - :443
  - ssl/local.com.crt
  - ssl/local.com.key           # https 和相应证书
GrpcAddr:   :8182    # grpc服务地址

# etcd 服务发现
EtcdAddr:   127.0.0.1:2379   # etcd地址
EtcdTimeout: 3         # 超时时间 单位秒
EtcdKeepAlive: 3       # 保持时间 单位秒
EtcdRetryTimes: 3      # 重试次数
EtcdRetryTimeout: 30   # 重试超时时间 单位秒

# 链路跟踪配置
ZipkinUrl:  http://localhost:9411/api/v2/spans  # zipkin地址

# 缓存redis别名
CacheRedis:   DEMO,
CacheSec:     600,

# 限流，熔断无需配置，取默认的

# MYSQL 取全局库的别名
MySQL:
  - DEMO
  - TEST

# Mongo 取全局库的别名
Mongo:
  - USER

# REDIS 取全局库的别名
Redis:
  - DEMO

# Rabbitmq 取全局库的别名
RabbitMQ:
  - USER

```

### 使用YAML配置，启动服务
```go
package main

import (
    "github.com/jasonfeng1980/pg"
    _ "github.com/jasonfeng1980/pg/example/api"
)

func main(){
    root := "../"
    pg.YamlRead(root).
        Server("example/conf/demo/pg_11_dev.yaml").
        Mysql("example/conf/mysql.yaml").
        Mongo("example/conf/mongo.yaml").
        Redis("example/conf/redis.yaml").
        Rabbitmq("example/conf/rabbitmq.yaml").
        Set()

    svc := pg.Server()
    svc.Run()
}
```

### 启动脚本
```go
func main(){
    root := "../"
    pg.YamlRead(root).
        Server("example/conf/demo/pg_11_dev.yaml").
        Set()
    svc := pg.Server()
    svc.Script(test)       # 满足 func() error 就可以
}

func test() error {
    return nil
}
```

### 带错误码的error
```go
// 在ecode里添加一个错误
MYSQLNoHandle := pg.Ecode(200001, "无法获得配置名为【%s】的MYSQL句柄")

// 在API里可以直接返回 nil, 200001 "无法获得配置名为【DEMO】的MYSQL句柄"
return ecode.MYSQLNoHandle.Parse("DEMO")

// 获得对应的error  等于 errors.New("无法获得配置名为【DEMO】的MYSQL句柄")
err := ecode.MYSQLNoHandle.Error("DEMO")
// 可以将err解析成 code msg  返回 200001， "无法获得配置名为【%s】的MYSQL句柄"
code, msg := pg.ReadError(err)
```

### mysql操作
```go
db, err := pg.MySQL.Get("DEMO")
if err != nil {
    return pg.Err(err)
}
// 直接使用sql - 不推荐
ret, err := db.Query("select * from company limit 2").Array()
// 推荐方式
ret, err := db.Select("*").
    From("company").
    //Where("company_id <=?", 200).
    Where(pg.M{"company_id": pg.M{"$lte": 200}}).  // 和上面效果一样
    Where("company_money>=? or company_money<?", 100, 500).
    Where(pg.M{"company_money":222}).
    GroupBy("company_money").
    Having("company_id >?", 1).
    OrderBy("company_money desc").
    Limit(3, 0).
    Cache(true). # 用redis缓存结果，读配置的CacheRedis和CacheSec
    Query().   
    Array()
// 更新
updateLine, err := db.Update("company").
    Set(update).
    Where("company_id=?", companyId).
    Query().
    RowsAffected()
// 插入
insertId, err := db.Insert(dataJson).
    Into("company").
    Query().
    LastInsertId()
// 删除
deleteLine, err := db.Delete().
    From("company").
    Where("company_id=?", companyId).
    Query().
    RowsAffected()
// replace
replaceLine, err := db.Replace("company").
    Set(replace).
    Query().
    RowsAffected()
// 事务
tx := db.StartTransaction()     // 开启事务
ret, err := tx.Select("*").
    From("company").
    Where("company_id =?", params["company_id"]).
    ForUpdate().
    Query().
    Array()
tx.Commit()     // 提交
tx.Rollback()   // 回滚
```

### ORM操作
```go
import orm_demo "test/orm/demo"

// 获取company 实例
orm := orm_demo.Company()
// ORM-创建
rs, err := orm.Create(params)
// 获取单行实例
line := orm.Line(id)
// ORM-单行编辑
rs, err := line.Edit(params)
// ORM-删除
ret := line.Remove()
// ORM-查看
ret := line.Info() 
ret := line.Cache(true).Info() // 带redis缓存
// ORM-分页-普通
ret, err := orm.Page(5, 1).
  Where("company_money=?",params["company_money"]).
  OrderBy("company_id asc").
  Cache(true).  // 使用缓存， 不缓存可以不使用Cache
  Query().
  Array()
// ORM-分页-流式
ret, _:= orm.Flow(5, "company_id>?", params["company_id"]).
  Where("company_money=?",params["company_money"]).
  OrderBy("company_id asc").
  Cache(true).   // 使用缓存， 不缓存可以不使用Cache
  Query().
  Array()
```

### MONGO操作
```go
mdb, err := pg.Mongo.Get("USER")
if err != nil {
    return pg.ErrCode(15003, "提供的mongoDB配置不存在")
}
ret, err := mdb.Select("*").
  From("user").
  Where(pg.M{"info.name": pg.M{"$lte":"王二"}}).
  //Where("info.name", pg.M{"$lte":"王二"}). // 效果同上条
  //GroupByMap(pg.M{
  //    "_id":"$token",
  //    "sum": bson.D{{"$sum", "$info.sex"}},
  //    "count": bson.D{{"$sum", 1}},
  //}).
  GroupBy("token").     // 等同上条的 _id, count的效果
  //Having(pg.M{"count": pg.M{"$gte": 16}}).
  Having("count", pg.M{"$gte": 16}). // 效果同上条
  OrderBy("create_at desc").
  Limit(0, 20).
  Query().
  Array()
```

### RabbitMQ操作
```go
ch, _ := pg.RabbitMQ.Get("USER", "logs")
defer ch.Close()
routing := "q_log_q1"
data, _ := util.JsonEncode(params)
// 正常发布
ch.Publish(routing, data)
// 延迟发布
ch.PublishDelay(routing, data, 10)
// 消费
msg, _ := ch.Consume("q_log")  # 队列别名
for d := range msg{
    d.Ack(false)
    util.Log.Infoln(string(d.Body[:]))
}
```

### REDIS-操作
pg.Redis.Client("DEMO")
  - CLIENT 普通客户端
  - SENTINEL 哨兵
  - CLUSTER 集群
```go
rClient, _ := pg.Redis.Client("DEMO")
// 直接操作
rdb.TTL(ctx, nameKey)
rdb.Get(ctx, name1).Val()
rdb.RPush(ctx, nameList, "a", "b"）
rdb.SAdd(ctx, nameSet, "a", "b"）
rdb.SUnion(ctx, name1, name2)
rdb.ZAdd(ctx, name1,
        &redis.Z{1, "a"},
        &redis.Z{2, "b"})
// 推荐操作
  // 在配置文件里设计redis的KEY，可以防止单独文件里，统一管理
ctx := context.Background()
UserName := func() rdb.String{
    return rdb.String{
        Key: rdb.Key{
            CTX: ctx,
            Name: "userName",
            Client: rClient,
        },
    }
}
UserInfo := func(userId int) rdb.Hash{
    return rdb.Hash{
        Key: rdb.Key{
            CTX:    ctx,
            Name:   "userInfo",
            Client: rClient,
        },
        Field:    util.StrParse(userId),
        JoinMode: []string{"age", "desc"},
    }
}
  // 用这些KEY来操作, 可以大幅减少redis内存空间
u := UserName()
ui := UserInfo(888)
u.Set("张三丰")
    // 只取JoinMode里的key对应的值，不存储KEY
info, _ := ui.Encode(pg.M{"age":18, "desc":"备注", "xxx":"无关信息"})

ui.HSet(info)
retName, err := u.Get()
tInfo, _ := ui.HGet()       // 18〡备注
retInfo, _ := ui.Decode(tInfo) // map[string]string{"age": "18", "desc": "备注"}
```

### LOG-日志
- 日志级别默认是Info
- 如果配置LogDebug:true,日志级别是Trace
```go
// 记录DEBUG日志 日志级别DEBUG 会显示文件，行号
pg.D("a", "b", "c")  
// 记录DEBUG日志并退出
pg.DD("a")  

// 获取句柄, 如果配置LogDir!= "", 会生成新的日志文件
myLog := pg.Log.Get("login")
    // With(kvs  ...interface{})
myLog.With("userId", 8888, "userName", "张三丰").
    // 支持Trace|Debug|Info|Warn|Error|Fatal|Panic()
      Info("登录成功") 
```







