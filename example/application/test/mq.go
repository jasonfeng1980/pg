package test

import (
    "context"
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/util"
)

// 通过初始化，注册ping关系
func init() {
   api := pg.MicroApi()
   api.Register("POST", "mq", "publish", "v1", mqRabbitMQPublish)
   api.Register("POST", "mq", "consume", "v1", mqRabbitMQConsume)
}


func mqRabbitMQPublish(ctx context.Context, params *util.Param)(interface{}, int64, string) {
   ch, err := pg.RabbitMQ.Exchange("USER", "logs")
   if err != nil {
       return pg.Err(err)
   }
   defer ch.Close()
   routing := "q_log_q1"
   data, _ := util.JsonEncode(params)
   // 正常发布
   err = ch.Publish(routing, data)
   if err != nil {
       return pg.Err(err)
   }
   // 延迟发布
   err = ch.PublishDelay(routing, data, 10)
   if err != nil {
       return pg.Err(err)
   }
   return pg.Suc("发布成功")
}

func mqRabbitMQConsume(ctx context.Context, params *util.Param)(interface{}, int64, string) {
   ch, err := pg.RabbitMQ.Get("USER", "logs")
   if err != nil {
       return pg.Err(err)
   }
   defer ch.Close()
   msg, err := ch.Consume("q_log")
   if err!= nil {
       return pg.Err(err)
   }

   var ret  []string
   for d := range msg{
       util.Log().S.Debugf("<- %s  接收成功\n", d.Body)
       d.Ack(false)
       ret = append(ret, string(d.Body[:]))
       break
   }
   return pg.Suc(ret)
}