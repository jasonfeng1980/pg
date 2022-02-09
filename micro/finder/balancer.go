package finder

import (
    "github.com/jasonfeng1980/pg/micro/endpoint"
    "math/rand"
    "time"
)

type BalancerFunc func(srvList []string) (oneSrv string)

type Balancer struct {
    F       BalancerFunc
    ept     endpoint.Endpoint
}

// 随机分配一个地址
func Random(srvList []string) string{
    seed := time.Now().UnixNano()
    rand.New(rand.NewSource(seed))
    return srvList[rand.Intn(len(srvList))]
}