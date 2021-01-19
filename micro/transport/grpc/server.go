package grpc

import (
    "context"
    callEndpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    callPb "github.com/jasonfeng1980/pg/micro/transport/grpc/pb"

    kitGrpc "github.com/go-kit/kit/transport/grpc"

    jsoniter "github.com/json-iterator/go"
)
var json = jsoniter.ConfigCompatibleWithStandardLibrary

type grpcServer struct {
    call kitGrpc.Handler
}

func (g *grpcServer) Call(ctx context.Context, req *callPb.CallRequest) (*callPb.CallReply, error) {
    _, rep, err := g.call.ServeGRPC(ctx, req)
    if err != nil {
        return nil, err
    }
    return rep.(*callPb.CallReply), nil
}


// 获得服务句柄
func NewServer(endpoints callEndpoint.Set, options []kitGrpc.ServerOption) callPb.ServiceServer {
    m := &grpcServer{
        call: makeCallHandle(endpoints, options),
    }
    return m
}

func makeCallHandle(endpoints callEndpoint.Set, options []kitGrpc.ServerOption) kitGrpc.Handler {
    return kitGrpc.NewServer(
        endpoints.CallEndpoint,
        decodeRequest,
        encodeResponse,
        options...
    )
}

// 解密请求参数
func decodeRequest(_ context.Context, r interface{}) (interface{}, error) {
    var d map[string]interface{}
    err := json.Unmarshal([]byte(r.(*callPb.CallRequest).Params), &d)
    if err != nil {
        return nil, err
    }

    return callEndpoint.CallRequest{
        Dns: r.(*callPb.CallRequest).Dns,
        Params: d,
    }, nil
}

// 加密响应数据
func encodeResponse(_ context.Context, r interface{}) (interface{}, error) {
    resp := r.(callEndpoint.CallResponse)
    d, err := json.Marshal(resp.Data)
    return &callPb.CallReply{
        Data: string(d[:]),
        Code: resp.Code,
        Msg:  resp.Msg,
    }, err
}



