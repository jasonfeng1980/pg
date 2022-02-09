package kafka

import (
    "context"
    "github.com/Shopify/sarama"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "github.com/oklog/oklog/pkg/group"
)

var Kafka = &kafka{
    confList: make(map[string]*Conf),
}

// 配置
func Conn(confMap map[string]Conf) {
    config := sarama.NewConfig()
    config.Consumer.Return.Errors = false
    config.Version = sarama.V0_11_0_2
    config.Consumer.Offsets.Initial = sarama.OffsetOldest

    for name, c := range confMap {
        c.Config = config
        Kafka.confList[name] = &c
    }
}

type kafka struct {
    confList map[string]*Conf
}

// 获取新的client
func Get(name string) (*Client, error) {
    if c, ok := Kafka.confList[name]; ok {
        consumerGroup, err := sarama.NewConsumerGroup(c.Server, c.GroupId, c.Config)
        return &Client{
            client: consumerGroup,
            Conf:   c,
        }, err
    } else {
        return nil, ecode.KafkaWrongConfName.Error(name)
    }
}

type Conf struct {
    Server  []string
    Topic   []string
    GroupId string

    Config *sarama.Config
}

// 消费端
type Client struct {
    Conf   *Conf
    client sarama.ConsumerGroup
}

// 关闭
func (c *Client) Close() error{
    if c.client != nil {
        util.Info("退出kafka - client")
        c.client.Close()
    }
    return nil
}

// 监听消息
func (c *Client) ConsumerHandler(ctx context.Context, g *group.Group, handler sarama.ConsumerGroupHandler)  {
    c.consumer(ctx, g, c.client, handler)
    c.quit(ctx, g)
    c.trace(g)
}

// 简化监听消息
type HandlerFunc func(message *sarama.ConsumerMessage) (ack bool)

func (c *Client) Consumer(ctx context.Context, g *group.Group, f HandlerFunc) {
    handler := consumerGroupHandler{
        f: f,
    }
    c.ConsumerHandler(ctx, g, handler)
}

// 消费
func (c *Client) consumer(ctx context.Context, g *group.Group, consumerGroup sarama.ConsumerGroup, handler sarama.ConsumerGroupHandler) {
    g.Add(func() error {
        for {
            if err := consumerGroup.Consume(ctx, c.Conf.Topic, handler); err != nil {
                util.Error("kafka consumerGroup.Consume err:" + err.Error())
                return err
            }
            if ctx.Err() != nil {
                return ctx.Err()
            }
        }
    }, func(err error) {
    })
}

func (c *Client) trace(g *group.Group) {
    g.Add(func() error {
        for err := range c.client.Errors() {
            util.Debug("", "kafka", c.Conf.GroupId, " consume err", err.Error())
        }
        return nil
    }, func(err error) {
    })
}

func (c *Client) quit(ctx context.Context, g *group.Group) {
    g.Add(func() error {
        for {
            select {
            // 主进程退出，通知consumer关闭
            case <-ctx.Done():
                _ = c.client.Close()
                util.Info("退出kafka consumer ")
                return nil
            }
        }
    }, func(err error) {

    })
}

// consumerHandler 消费句柄
type consumerGroupHandler struct {
    f HandlerFunc
}

func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession,
    claim sarama.ConsumerGroupClaim) error {
    for message := range claim.Messages() {
        if ok := h.f(message); ok { // 返回true  确认消息
            sess.MarkMessage(message, "")
        }
    }
    return nil
}
