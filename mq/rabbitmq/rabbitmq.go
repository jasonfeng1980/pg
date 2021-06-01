package rabbitmq

import (
    "fmt"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "github.com/streadway/amqp"
    "sync"
    "time"
)

var RabbitMq = &rabbitMQ{
    Log: util.Log.Logger,
}

type rabbitMQ struct {
    Log  *util.Logger
    Conf  map[string]conf.RabbitMQConf
    Pool  sync.Map
}

func (r *rabbitMQ)Conn(conf map[string]conf.RabbitMQConf) {
    r.Conf = conf
    // 预链接
    for n, _:=range conf{
        r.conn(n)
    }
}
func (r *rabbitMQ)conn(dnsName string) (*amqp.Connection, error) {
    // 判断是否存在配置
    if l, ok := r.Conf[dnsName]; ok {
        // 链接
        conn, err := amqp.Dial(l.Dns)
        r.Log.Infof("连接RabbitMQ   %s: %s ----  成功", dnsName, l.Dns)
        if err != nil {
            return  nil, ecode.RabbitMQDnsConnErr.Error(l.Dns, err.Error())
        }
        r.Pool.Store(dnsName, conn)
        return conn, nil
    }
    return nil, ecode.RabbitMQNotDnsConf.Error(dnsName)
}
func (r *rabbitMQ)GetConn(dnsName string) (*amqp.Connection, error){
    var (
        err error
        conn *amqp.Connection
    )
    c, ok := r.Pool.Load(dnsName)
    if ok {
        conn = c.(*amqp.Connection)
        ok = !conn.IsClosed()
    }
    if !ok {
        conn, err  = r.conn(dnsName)
    }
    return conn, err
}
func (r *rabbitMQ)CloseConn(dnsName string) {
    // 获取链接
    var conn *amqp.Connection
    c, ok := r.Pool.Load(dnsName)
    if ok {
        conn = c.(*amqp.Connection)
        ok = !conn.IsClosed()
    }
    if conn !=nil && !conn.IsClosed() {
        conn.Close()
    }
}
// 获取交换机配置
func (r *rabbitMQ)ExchangeConf(dns string, exchangeName string) (*conf.RabbitMQExchange, error) {
    err := ecode.RabbitMQNotExchangeConf.Error(dns, exchangeName)
    if d, ok :=r.Conf[dns]; ok {
        if e, ok := d.Exchange[exchangeName]; ok {
            return &e, nil
        }
    }
    return nil, err
}
// 获取query配置
func (r *rabbitMQ)QueryConf(dns string, exchangeName string, queryName string) (ret *conf.RabbitMQQuery, err error) {
    exchangeConf, err := r.ExchangeConf(dns,exchangeName)
    if err == nil {
        if q, ok:= exchangeConf.Query[queryName]; ok {
            ret = &q
        } else {
            err = ecode.RabbitMQNotRoutingConf.Error(dns, exchangeName, queryName)
        }
    }
    return
}
func (r *rabbitMQ)Close() {
    //for k, conn := range r.Pool{
    //    r.Log.Logf("关闭RabbitMQ - %s 的链接", k)
    //    conn.Close()
    //}
    r.Pool.Range(func(k, v interface{}) bool {
        if conn, ok := v.(*amqp.Connection); ok {
            conn.Close()
        }
        return true
    })
}

func (r *rabbitMQ)Exchange(dnsName string, exchangeName string) (*Channel, error){
    // 判断是否存在该exchange的配置
    eL, err := r.ExchangeConf(dnsName, exchangeName)
    if err != nil {
        return nil, err
    }
    // 获取链接
    conn, err := r.GetConn(dnsName)
    if err !=nil {
        return nil, err
    }
    // channel
    ch, err := conn.Channel()
    if err != nil {
        return nil, err
    }

    // 交换机声明
    err = ch.ExchangeDeclare(exchangeName, eL.Kind, eL.Info[0], eL.Info[1], eL.Info[2], eL.Info[3], eL.Args)
    return &Channel{
        Channel: ch,
        Dns:     dnsName,
        Exchange: exchangeName,
    }, err
}


type Channel struct {
    Channel *amqp.Channel
    Dns      string
    Exchange string
}
func (c *Channel)Close(){
    c.Channel.Close()
    RabbitMq.Log.Infof("关闭RabbitMQ-Channel - %s 的链接", c.Dns)
}
func (c *Channel)ReConn() error{
    newCh, err := RabbitMq.Exchange(c.Dns, c.Exchange)
    if err == nil {
        c.Channel = newCh.Channel
    }
    return err
}
func (c *Channel)queryConf(queueName string) (*conf.RabbitMQQuery, error) {
    return RabbitMq.QueryConf(c.Dns, c.Exchange, queueName)
}
// 发布信息
func (c *Channel)Publish(routing string, data []byte) error{
    return c.PublishMore(routing, false, false, data)
}
/*
routing 路由KEY
mandatory routing至少可以找到一个队列接收，不然就调用basic.return返还生成者
immediate 队列必须有消费才投递；如果所有的队列都没有消费，就调用basic.return返还生成者
data   string
*/
func (c *Channel)PublishMore(routing string, mandatory bool, immediate bool, data []byte) (err error){
    var i int64
    for i=0;i<3;i++ {
        err =  c.Channel.Publish(c.Exchange, routing, mandatory, immediate, amqp.Publishing{
            DeliveryMode: amqp.Persistent,
            ContentType: "text/plain",
            Body: data,
        })
        if err == nil {
            break
        } else {
            RabbitMq.Log.Infof("错误 ----> %s", err.Error())
            time.Sleep(time.Microsecond * 500 * time.Duration(i))
            c.ReConn()
        }
    }
    return

}
// 发布到死信队列
func (c *Channel)PublishDelay(routing string, data []byte, delay int) error{
    // 声明延时队列
    args := amqp.Table{
        "x-dead-letter-exchange": c.Exchange,
        "x-dead-letter-routing-key": routing,
        "x-message-ttl": delay * 1000,
    }
    queueName := fmt.Sprintf("DELAY.%s.%d", routing, delay)
    routingName := queueName
    q, err := c.Channel.QueueDeclare(queueName, true, false, false, false, args)
    if err != nil {
        return  err
    }
    // 绑定路由
    err = c.Channel.QueueBind(q.Name, routingName, c.Exchange, false, nil)
    if err != nil {
        return err
    }
    // 发布死信
    return c.PublishMore(routingName, false, false, data)
}
// 消费
func (c *Channel)Consume(queueName string) (<-chan amqp.Delivery, error) {
    conf, err := c.queryConf(queueName)
    if err != nil {
        return nil, err
    }
    // 队列声明
    q, err := c.Channel.QueueDeclare(queueName, conf.Info[0], conf.Info[1], conf.Info[2], conf.Info[3], conf.Args)
    if err != nil {
        return nil, err
    }
    //Qos
    if len(conf.Qos) == 3 {
        glo := conf.Qos[2] == 1
        c.Channel.Qos(conf.Qos[0], conf.Qos[1], glo)
    }
    // 绑定路由
    for _, routing := range conf.Routing {
        err = c.Channel.QueueBind(q.Name, routing, c.Exchange, false, nil)
        if err != nil {
            return nil, err
        }
    }
    // 消费者
    return c.Channel.Consume(
        q.Name,
        "",
        false,
        conf.Info[2],
        false,
        conf.Info[3],
        nil,
    )
}
