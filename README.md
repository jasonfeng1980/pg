# pg
方便PHPer 快速使用GO
无需考虑 熔断，限流，链路追踪，连接池
同时支持 WEB请求和微服务（HTTP,HTTPS,GRPC)

### 快速开始
```go
pg.SetRoot("../../")
svc := pg.Server()
svc.Run()
```

### 请求的API
```go
api := pg.MicroApi()
api.Register("GET","orm", "v1", "page", ORMPage)
func ORMPage(ctx context.Context, params map[string]interface{})(interface{}, int64, string) {
	ret := "成功"
	// do something .... 
	return pg.Success(ret)
}
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

# MYSQL 取全局库的别名 + 权限  e.g.
MySQL:
  - DEMO
  - TEST

# REDIS 取全局库的别名
Redis:
  - DEMO
```



