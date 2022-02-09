package ecode

import (
    "errors"
    "fmt"
    "strconv"
    "strings"
)

type ECode struct {
    Code int64
    Msg  string
}

func(e *ECode) Error(args ...interface{}) error{
    return errors.New(fmt.Sprintf(strconv.FormatInt(e.Code, 10)+ ":" + e.Msg, args...))
}
func(e *ECode) Parse(args ...interface{}) (interface{}, int64, string){
    return nil, e.Code, fmt.Sprintf(e.Msg, args...)
}
func NewError(code int64, msg string) error{
    return errors.New(fmt.Sprintf("%d,%s", code, msg))
}
func Err(code int64, msg string) *ECode{
    return &ECode{code, msg}
}
func ReadError(e error) (int64, string) {
    s := e.Error()
    l := strings.SplitN(s, ":", 2)

    if len(l) ==2 {
        if i64, err := strconv.ParseInt(l[0], 10, 64); err == nil {
            return i64, l[1]
        }
    }
    return 200, s
}

var (
    // 系统        [1000,1100)
    UtilCanNotBeInt = Err(1001, "无法转换成INT")
    UtilCanNotBeInt64 = Err(1002, "无法转换成INT64")
    UtilWrongDataType = Err(1003, "错误的数据格式")
    UtilMissNeedField = Err(1004, "缺少必需的字段：%s")
    UtilCanNotBeString = Err(1005, "无法转换成string")
    UtilNoUploadFile = Err(1006, "没有需要上传的文件")
    UtilWrongDir = Err(1007, "错误的文件目录:%s")
    UtilErrDecodeJson = Err(1008, "无法解密json，只接受[]byte和json格式")
    CallServerPanic = Err(1009, "执行dns【%s】时，出现Panic错误: %s")
    NoMicroServer = Err(1010, "没有名称为【%s】的微服务，错误为%s")
    EtcdDisconnect = Err(1011, "无法连接链接为【%s】的ETCD服务器.错误为%s")
    EtcdWrongAddr = Err(1012, "错误的etcd地址")
    EtcdMissServerK = Err(1013, "etcd缺失服务的Key")
    EtcdMissServerV = Err(1014, "etcd缺失服务的Value")
    EtcdEmptySrv    = Err(1015, "etcd没有可用的服务，前缀%s")
    EtcdFindErr = Err(1016, "etcd服务发现错误:%s")
    DnsWrongScheme    = Err(1018, "etcd-DNS只支持http和grpc")
    UtilSessionNotNew = Err(1020, "session没有初始化")

    UtilWrongDns = Err(1025, "错误的DNS格式")

    LimitErroring = Err(1051, "服务方限流，拒绝访问")

    // 浏览器请求   [1100,1200)
    HttpMissDns = Err(1100, "缺少DNS")
    HttpCannotMatchDns = Err(1101, "没有无法匹配的DNS:%s/%s/%s")
    HttpUrlMissMVC = Err(1102, "dns缺少module,version,action")
    HttpDataNotMap = Err(1103, "传递的参数不是map[string]interface{}格式")
    HttpDnsParseWrong = Err(1104, "dns解析失败")

    // 数据库      [1200,1300)
    DbNotExecData = Err(1201, "没有result-exec数据")
    DbNotQueryData = Err(1202, "没有result-query数据")
    DbWrongType = Err(1203, "dest类型不对")
    DbColumnsNotMatch = Err(1204, "columns与地址不匹配")
    DbWrongMap = Err(1205, "错误的MAP格式，必须是 map[string]interface")
    DbWrongWhere = Err(1206, "错误的where格式，必须是 string 或者 map[string]interface")
    DbWrongConfName = Err(1207, "%s 没有 别名%s 的配置")

    MdbWrongData= Err(1210, "Mongo创建连接失败: %s")
    MdbCloseConnErr = Err(1211, "Mongo关闭连接失败: %s")
    MdbPingErr = Err(1212, "Mongo ping失败: %s")
    MdbNotExecData = Err(1213, "Mongo没有result-exec数据")
    MdbNotQueryData = Err(1214, "Mongo没有result-query数据")
    MdbCollectionIsNil = Err(1215, "Mongo没有指定Collection")
    MdbCountErr = Err(1216, "Mongo只有select并且不用group时，可以用Count()")


    RdbCannotToString = Err(1220, "redis-hash key:%s field:%s无法转换成字符串")
    RdbWrongData= Err(1221, "redis-hash数据格式不正确 名称%s")
    RdbWrongDecodeJoin= Err(1222, "redis decodeJoin 数据格式不匹配")

    OrmWrongArgType= Err(1230, "ORM 参数无法转换成[]map[string]interface{}")
    OrmWrongColumnsType = Err(1231, "ORM %s字段格式不对")
    OrmMissColumnsNeed = Err(1232, "ORM 缺少必填项：%s")
    DaoWrongTable = Err(1233, "数据库%s不存在该数据表%s")
    DaoWrongField = Err(1234, "数据表%s不存在该字段%s")
    DaoEmptyResult = Err(1235, "没有满足条件【%s】的数据")
    DaoMissNeedField = Err(1236, "%s 缺少必填字段【%s】")

    EntityMissPK = Err(1250, "Entity【%s】缺失PK")
    EntityEmptyCreateParams = Err(1251, "Entity创建【%s】记录，没有数据")
    EntityEmptySearchCondition = Err(1252, "Entity搜索【%s】，没有查询条件")
    EntityWrongRelTableName = Err(1253, "Entity[%s]没有与【%s】的关联")
    EntityMissRelValue = Err(1254, "Entity[%s]缺失与【%s】的关联的字段值")
    EntityMissResultQueryFunc = Err(1255, "entity没有设置Result对应的方法")


    AggregateWrongGuid = Err(1270, "aggregate[%s]的guid=[%s]错误")

    // CONF  [1300, 1400)
    ConfYamlWrongFormat = Err(1301, "%s格式不正确")
    ConfYamlWrongMysql = Err(1302, "Mysql-yaml配置格式不正确")
    ConfYamlWrongMongo = Err(1303, "Mongo-yaml配置格式不正确")
    ConfYamlWrongRedis = Err(1304, "Redis-yaml配置格式不正确")
    ConfYamlWrongRabbitMQ = Err(1305, "RabbitMQ-yaml配置格式不正确")
    ConfNotComplete = Err(1306, "【%s】的配置文件不完整")
    ConfMissWrite = Err(1307, "【%s】缺少写库配置")
    ConfWrong = Err(1308, "【%s】配置不正确")

    // RabbitMQ [1400, 1420)
    RabbitMQDnsConnErr = Err(1401, "rabbit DNS%s链接失败,%s")
    RabbitMQNotDnsConf = Err(1402, "rabbit 没有dns:%s 的配置")
    RabbitMQNotExchangeConf = Err(1403, "rabbit 没有dns:%s exchange:%s 的配置")
    RabbitMQNotRoutingConf = Err(1404, "rabbit 没有dns:%s exchange:%s routing:%s的配置")

    // KafkaMQ [1420, 1440)
    KafkaWrongConf = Err(1421, "kafka 错误的配置：%s")
    KafkaWrongConfName = Err(1422, "kafka 没有 别名%s 的配置")



    // workflow [1500, 1550)
    WorkFlowNodeNotSupportBill = Err(1501, "工作流节点%s不支持单据%s")
)
