# pg
方便PHPer 快速使用GO  
只需要简单配置，就可以实现路由、日志、熔断、限流、链路追踪、连接池  
同时支持WEB服务和内部微服务（HTTP,HTTPS,GRPC)  
### 推荐目录结构
```text
├── README.md
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
│   ├── PG11.DEBUG.210126
│   ├── PG11.ERROR.210126
│   ├── PG11.ETCD.210126
│   └── PG11.INFO.210126
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
pg.SetRoot("../../")
svc := pg.Server()
svc.Run()
```
### 调用服务API
svc := pg.Client()  
data, code, msg := svc.Call(dns, pg.H{})  
dns  服务类型://服务名称/module/version/action
```go
func main(){
    pg.SetRoot("../../")
    util.CmdWait("请输入请求方式", waitFunc)
}

func waitFunc(cmdString string) (string, bool){
    cmdString = strings.ToLower(cmdString)
    svc := pg.Client()
    defer svc.Close()
    switch cmdString {
    case "http", "grpc":
        // dns  服务类型://服务名称/module/version/action
        dns := cmdString + "://PG/request/v1/post"
        data, code, msg := svc.Call(dns, pg.H{
            "aa": 1,
            "bb": cmdString,
        })
        fmt.Println(data, code, msg)
        return "请输入请求方式", true
    case "exit":
        return "", false
    default:
        return "请输入正确的参数：http | grpc | exit", true
    }
}
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
    _ "test/apps/demo"
)
```
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

### 全局REDIS配置
```yamlll
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

### 微服务配置
```yaml
ServerName: PG                   # 服务名称
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

# 限流，熔断无需配置，取默认的

# MYSQL 取全局库的别名
MySQL:
  - DEMO
  - TEST

# REDIS 取全局库的别名
Redis:
  - DEMO
```
### 使用YAML配置，启动服务
```go
package main

import (
    "fmt"
    "github.com/jasonfeng1980/pg"
    "os"
    _ "test/apps/demo"
)

func main(){
    root := "../../"
    mysqlFile := root + "conf/mysql.yaml"
    redisFile := root + "conf/redis.yaml"
    serverFile := root + "conf/demo/pg_11_dev.yaml"

    err := pg.SetConfYaml(mysqlFile, redisFile, serverFile, root)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    svc := pg.Server()
    svc.Run()
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
// 直接使用sql - 不推荐
ret, err := db.Query("select * from company limit 2").Array()
// 推荐方式
ret, err := db.Select("*").
    From("company").
    Where("company_id <?", 200).
    Where("company_money>=? or company_money<?", 100, 500).
    GroupBy("company_money").
    Having("company_id >?", 1).
    OrderBy("company_money desc").
    Limit(3, 0).
    Cache(true).
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
### REDIS-直接操作
```go
rdb := pg.Redis.Client("demo")
rdb.TTL(ctx, nameKey)
rdb.Get(ctx, name1).Val()
rdb.RPush(ctx, nameList, "a", "b"）
rdb.SAdd(ctx, nameSet, "a", "b"）
rdb.SUnion(ctx, name1, name2)
rdb.ZAdd(ctx, name1,
        &redis.Z{1, "a"},
        &redis.Z{2, "b"})

```








