package finder

import (
    "context"
    "github.com/jasonfeng1980/pg/micro/endpoint"
    "github.com/jasonfeng1980/pg/util"
    "time"
)

type Retry struct {
    ctx context.Context
    ept endpoint.Endpoint
    Timeout time.Duration
    RetryTimes int
}

func (r Retry)Endpoint(eptF DoEptFunc, balancer Balancer, srvList []string, scheme string) endpoint.Endpoint {

    return func(ctx context.Context, request interface{}) (response interface{}, err error) {
        var (
            // 超时取消函数
            newctx, cancel = context.WithTimeout(ctx, r.Timeout)
            responses      = make(chan interface{}, 1)
            errs           = make(chan error, 1)
            instanceAddr  string
        )
        defer cancel()
        // 重试
        for i := 1; ; i++ {
            // 获取MicroEndpoint
            instanceAddr = balancer.F(srvList)
            go func() {
                // 获取新的Endpoint
                ept, err := eptF(scheme, instanceAddr)
                if err != nil {
                    errs <- err
                    return
                }
                response, err := ept(newctx, request)
                if err != nil {
                    errs <- err
                    return
                }
                responses <- response
            }()

            select {
            case <-newctx.Done():  	// ctx结束了
                return nil, newctx.Err()

            case response := <-responses:	// 执行完成
                return response, nil

            case err := <-errs:		// 出现访问错误，不是返回错误
                util.Error("访问【" + instanceAddr + "】出现错误", "err", err)
                if i > r.RetryTimes {
                    return nil, err
                }
                continue
            }
        }

    }
}
