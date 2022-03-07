module github.com/jasonfeng1980/pg

go 1.16

replace github.com/jasonfeng1980/pg => /Users/fengshengqi/Documents/code/golang/src/pg

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/Shopify/sarama v1.30.0
	github.com/coreos/etcd v3.3.27+incompatible // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/go-redis/redis/v8 v8.11.4
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0 // indirect
	github.com/json-iterator/go v1.1.12
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/oklog/oklog v0.3.2
	github.com/oklog/run v1.1.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go v0.3.0
	github.com/prometheus/client_golang v1.4.0
	github.com/sony/gobreaker v0.5.0
	github.com/spf13/viper v1.10.1
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	go.etcd.io/etcd v3.3.27+incompatible
	go.mongodb.org/mongo-driver v1.8.1
	go.uber.org/zap v1.19.1
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11
	google.golang.org/grpc v1.43.0
)
