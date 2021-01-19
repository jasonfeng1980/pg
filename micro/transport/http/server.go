package http

import (
    "context"
    "github.com/jasonfeng1980/pg/conf"
    callEndpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    "github.com/jasonfeng1980/pg/util"
    "net/http"

    kitHttp "github.com/go-kit/kit/transport/http"
)


// 获得服务句柄
func NewServer(endpoints callEndpoint.Set, conf conf.Config, options []kitHttp.ServerOption) http.Server {
    return http.Server{
        Handler: kitHttp.NewServer(
            endpoints.CallEndpoint,
            decodeRequest(conf),
            encodeResponse,
            options...
        ),
    }
}

func decodeRequest(conf conf.Config) func (_ context.Context, r *http.Request) (interface{}, error) {
    return func (_ context.Context, r *http.Request) (interface{}, error) {
        dns, params, err := getRequestParams(conf, r)
        if err != nil {
        return nil, err
        }

        req := callEndpoint.CallRequest{
        Dns: dns,
        Params: params,
        }
        return req, nil
    }
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    //resp := response.(callEndpoint.CallResponse)
    return util.Json.NewEncoder(w).Encode(response.(callEndpoint.CallResponse))
}
