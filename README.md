# pg
方便PHPer 快速使用GO 搭建微服务平台<br>
只需简单配置，就可以实现路由、日志、熔断、限流、链路追踪、连接池<br>
简化了MYSQL,MONGO,RabbitMQ,REDIS的API, 会SQL语法，就可以使用这些搭建系统

### 推荐目录结构
```text
┌── README.md
├── bin                                         # 启动目录
│   └── demo_server_01.go
├── conf                                        # 配置目录目录
│   ├── pg_demo.01.dev.json             # 配置文件
│   └── ssl                             # https证书目录
│       ├── local.com.crt
│       └── local.com.key
├── application                                 # 应用服务目录
│   └── demo
│       ├── companyController.go
│       ├── companyController_v2.go
│       └── companyController_v3.go
├── domain                                      # 领域服务目录
│   └── companyService.go
├── aggregate                                   # 多实体聚合根目录
│   └── companyRoot.go
├── entity                                      # 单个实体目录    自动生成的
│   └── demoEntity
│       ├── companyEntity.go
│       └── companyMemberEntity.go
├── repository                                  # 资源仓库目录    自动生成的
│   └── DAO
│       └── demoMapper.go
├── ecode                                       # 错误code目录
│   └── ecode.go
├── go.mod
├── go.sum
└── vendor                                      # vendor目录
    └── modules.txt
```
### 快速启用服务
```go
srv := pg.Server(context.Background())
srv.Run()
```
### 调用服务API
```go
svc, _ := pg.Client()
defer svc.Close()
// dns   [grpc|http]://服务名称/module/version/action
dns :=  "http://PG/auth/v1/login"
data, code, msg := svc.Call(context.Background(), dns, pg.M{
    "user_mobile": "186",
    "user_password": 11111,
})
pg.D(data, code, msg)
```

### 注册API
```go
api := pg.MicroApi()
api.Register("GET", "company", "info", "v1", CompanyInfo)
// 获得公司的成员信息
func CompanyInfo(ctx context.Context, params *util.Param)(data interface{}, code int64, msg string) {
  // 整理参数
  companyId := params.GetInt64("company_id")
  // 1. 获取关联实体-公司成员的数据
  return domain.CompanyDomain.CompanyInfo(ctx, companyId)
}
```

### 加载API项目  
```go
// 启动服务时
import (
    "github.com/jasonfeng1980/pg"
    _ "github.com/jasonfeng1980/pg/example/application/demo"
)
```
_ "github.com/jasonfeng1980/pg/example/application/demo"  是 api所在的包

### 全局配置
```json
{
  "Build": {
    "Package": "github.com/jasonfeng1980/pg/example" # 当前项目的包
  },
  "Server": {
    "Name": "pg_demo",        # 服务别名
    "Num": "01",              # 服务序号
    "Root": "..",             # 根目录  如果是相对路径，相对当前执行的文件
    "Env": "dev",             # 当前环境 dev debug  test product

    "LogDir": "",             # 日志目录
    "LogLevel": "debug",      # 日志级别

    "AddrDebug": ":8081",     # 测试服务地址
    "AddrHttp":  ":80",       # http服务地址
    "AddrHttps": [":443", "conf/ssl/local.com.crt", "conf/ssl/local.com.key"], # https服务地址
    "AddrGrpc":  ":8082",     # grpc服务地址

    "ETCD":   "etcd://:@tcp(127.0.0.1:2379)/?DialTimeout=3&KeepAlive=3&RetryTimes=3&RetryTimeout=30", # etcd地址
    "ZipkinUrl": "http://localhost:9411/api/v2/spans",    # zipkinUrl链路跟踪地址

    "CacheRedis": "demo",     # 缓存sql的redis别名
    "CacheSec":    60         # 缓存时间
  },
  "MySQL": {
    "demo": [                 # 别名 DNS格式为 driver://[user]:[password]@network(host:port)/[dbname][?param1=value1&paramN=valueN]
      "mysql://root:@tcp(localhost:3306)/demo?charset=utf8mb4&maxOpen=200&maxIdle=100&maxLifetime=30",  # 写
      "mysql://root:@tcp(localhost:3306)/demo?charset=utf8mb4&maxOpen=200&maxIdle=100&maxLifetime=30"   # 读
    ],
    "test": [
      "mysql://root:@tcp(localhost:3306)/demo?charset=utf8mb4&maxOpen=200&maxIdle=100&maxLifetime=30"   # 读+写
    ]
  },
  "Redis": {
    "demo": [                 # 别名   DNS格式为 driver://[user]:[password]@network(host:port)/[dbname][?param1=value1&paramN=valueN]
      "redis://:@tcp(localhost:6379)/0"
    ]
  },
  "Mongo": {
    "demo": [                 # 别名   DNS格式为 driver://[user]:[password]@network(host:port)/[dbname][?param1=value1&paramN=valueN]
      "mongodb://admin:root@tpc(localhost:27017)/demo?Timeout=3&AllowDiskUse=0",
      "mongodb://admin:root@tpc(localhost:27017)/demo?Timeout=3&AllowDiskUse=0"
    ]
  },
  "KAFKA": {
    "product": {              # 别名
      "Server": ["127.0.0.1:9092"],
      "Topic": ["test"],
      "GroupId": "product"
    }
  }
}
```


### 启动服务
```go
package main

import (
  "context"
  "github.com/jasonfeng1980/pg"
  "github.com/jasonfeng1980/pg/util"

  _ "github.com/jasonfeng1980/pg/example/application/demo"
)

func main(){
  if err :=pg.Load("../conf/pg_demo.01.dev.json");err!= nil {
    util.Panic("加载配置错误", "error", err)
    return
  }
  srv := pg.Server(context.Background())
  srv.Run()
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
    svc.Script()、
    test()
}

func test()  {
    // ...
}
```

### 带错误码的error
```go
// 在ecode文件里添加一个错误， %s可以替换为不同的字符串，同fmt.Sprintf
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
    Cache(true).  // 用redis缓存结果，读配置的CacheRedis和CacheSec
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

### MONGO操作
```go
mdb, err := pg.Mongo.Get("USER")
if err != nil {
    return pg.ErrCode(15003, "提供的mongoDB配置不存在")
}
ret, err := mdb.Select("*").
  From("user").
  Where(pg.M{"info.name": pg.M{"$regex":"王"}}).
  //Where("info.name", pg.M{"$regex":"王"}). // 效果同上条
  //GroupByMap(pg.M{
  //    "_id":"$token",
  //    "sum": bson.D{{"$sum", "$info.sex"}},
  //    "count": bson.D{{"$sum", 1}},
  //}).
  GroupBy("token").     // 等同上条的 _id, count的效果
  //Having(pg.M{"count": pg.M{"$gte": 16}}).
  Having("count", pg.M{"$gte": 16}). // 效果同上条
  OrderBy("create_at desc").
  Limit(0, 1).
  Cache(true).
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
  // 在配置文件里设计redis的KEY，可以放在单独文件里，统一管理
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
info, _ := ui.Encode(pg.M{"age":18, "desc":"备注", "xxx":"无关信息不存储"})
ui.HSet(info)
retName, err := u.Get()
tInfo, _ := ui.HGet()       // 18〡备注
retInfo, _ := ui.Decode(tInfo) // map[string]string{"age": "18", "desc": "备注"}
```

### LOG-日志
- 日志级别默认是Info
- 如果配置LogLevel = debug, 显示文件行号和请求的response
```go
// 记录DEBUG日志 日志级别DEBUG 会显示文件，行号，美化json
pg.D("a", "b", "c")  
// 记录DEBUG日志并退出
pg.DD("a")  

util.Info("msg", "key", "value", "k2", "v2")
util.Error("msg")

```

### 自动生成资源仓库和实体文件
复制下面的代码到XXX.go
```go
package main

import "github.com/jasonfeng1980/pg/build"

func main()  {
    build.Build()
}
```
- 详见  https://github.com/jasonfeng1980/pg/blob/master/build/build.go
```bash
go run xxx.go build -c dev.json -db demo     # pg build -c 配置文件   -db 数据库别名
```

### 基于单据的工作流引擎
实现8个接口:
```go
// 1. 获取当前待处理的节点状态
GetWorkflowToDoStatus(workflowId int64) (workflowStatusList []string, err error)
// 2. 获取指定节点状态
GetNodeStatus(workflowId int64, nodeCode string) (nodeStatus string, err error)
// 3. 创建工作流,记录日志 - 统一返回格式
CreateWorkflow(operatorId int64, params map[string]interface{}) (nowWorkflowId int64, rollbackAct []RollbackParams, err error)
// 4. 更新节点状态,记录日志 - 统一返回格式
ChangeNodeStatus(operatorId int64, workflowId int64, nodeStatus *ConfNodeStatus, params map[string]interface{}) (nowWorkflowId int64, rollbackAct RollbackParams, err error)
// 5. 写单据提交日志和回滚数据
SaveSubmitLog(operatorId int64, nowWorkflowId int64, docketCode string, params map[string]interface{}, e error, rollbackAct []RollbackParams) (submitLogId string, err error)
// 6. 执行回滚
RunRollbackAct(operatorId int64, workflowId int64, rollbackAct []RollbackParams)
// 7. 自定义行为
CustomAct(operatorId int64, workflowId int64, docketCode string, act *ConfAct, params map[string]interface{})(nowWorkflowId int64, rollbackAct []RollbackParams, err error)
// 8. 观察者行为
ObAct(operatorId int64, workflowId int64, nodeStatus *ConfNodeStatus, obAct *ConfOB, params map[string]interface{}) (rollbackAct []RollbackParams, err error)
```
配置工作流json:
```json
{
// 工作流的基本信息,
  "info": {"code": "enter_warehouse", "name": "入仓工作流", "desc": "货物从收货打签到待上架的流程"},  
// 节点列表,
  "node": [   
// 某个节点基本信息：  type node节点|gateway网关
    {"code": "EW1000", "name": "收货打签", "type": "node", "status": [ 
// 节点状态 end节点流程结算|finish工作流结束|to到达后自动跳转到节点状态，可以是多个|arrive是到达后执行的多个行为
// 行为  如果有check是需要同时满足的节点状态，满足后执行后面的行为， 可以是默认 status状态改变|act自定义行为code
    ]},
    {"code": "EW2000", "name": "拍摄", "type": "node", "status": [
      {"code": "EW6001", "name": "等待拍摄完成且卖家定价完成", "arrive": [{"check": ["EW4111", "EW4210"] ,"status": ["EW6010"]}]},
    ]}
  ],
  "act": [
// 自定义行为  fn 是自定义行为的方法标记| status 默认状态改变方法
    {"code": "122", ",name": "鉴定-分配", "fn": "identifyAssign"},
    {"code": "131", ",name": "强制去待编辑", "status": ["EW3001"]},
  ],
  "ob": [
// 观察者  同步执行  allow允许的工作流待处理的节点状态|观察者的标识
    {"code": "OB100", "name": "卖家视角-更新状态", "allow": ["EW1001","EW2021","EW2022","EW4201","EW4210","EW9001"], "fn": "sellerChangeStatus"}
  ],
  "docket": [
// 定义单据  allow允许的工作流待处理的节点状态，[]表示不限制|do要执行的行为，status默认状态改变，act用上面配置的自定义行为code
// {"act": "WF_CreateWorkflow"} 默认创建工作流的方法
    {"code": "101", "name": "快递-签收单", "allow": [], "do": [{"act": "WF_CreateWorkflow"}, {"status": ["EW1001"]}]},
    {"code": "102", "name": "收货打签-完成单", "allow": ["EW1001"], "do": [{"status": ["EW1090"]}]}
  ]
}
```
启用
```go
import "github.com/jasonfeng1980/pg/util/workflow"
// 入库工作流
type EnterWarehouseRoot struct {
    Ctx             context.Context
    *workflow.Workflow
    *workflowEntity.EnterWarehouseEntity
    *ddd.AggregateRoot
}

// 获取配置
if confJson, err := util.FileRead(conf.ConfBox.GetString("Workflow.enter_warehouse")); err !=nil {
    return nil
} else if err =json.Unmarshal(confJson, &cf); err !=nil {
    return
}

// 获取实例，并绑定方法
ew := &EnterWarehouseRoot{
    Ctx:                    ctx,
    EnterWarehouseEntity:   workflowEntity.NewEnterWarehouseEntity(ctx, pk),
    AggregateRoot:          ddd.NewAggregateRoot(prefix, pk),
}
ew.Workflow = workflow.New(&cf, ew)
// 执行
workflowId, _, err = enterWarehouse.Submit(operatorId, workflowId, "101", params)
```
- 详见  https://github.com/jasonfeng1980/pg/blob/master/example/bin/workflow.go



