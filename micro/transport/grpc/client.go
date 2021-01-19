package grpc

import (
    "context"
    "google.golang.org/grpc"

    kitGrpc "github.com/go-kit/kit/transport/grpc"

    callEndpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    callpb "github.com/jasonfeng1980/pg/micro/transport/grpc/pb"
)

func NewClient(conn *grpc.ClientConn, options []kitGrpc.ClientOption) (callEndpoint.Set, error) {
    CallEndpoint := kitGrpc.NewClient(
        conn,
        "pb.Service",
        "Call",
        encodeRequest,
        decodeResponse,
        callpb.CallReply{},
        options...,
    ).Endpoint()

    return callEndpoint.Set{CallEndpoint: CallEndpoint}, nil
}

func encodeRequest(_ context.Context, r interface{}) (interface{}, error) {
    d, err := json.Marshal(r.(callEndpoint.CallRequest).Params)
    if err != nil {
        return nil ,err
    }
    return &callpb.CallRequest{
        Dns: r.(callEndpoint.CallRequest).Dns,
        Params: string(d[:]),
    }, nil
}

func decodeResponse(_ context.Context, r interface{}) (interface{}, error) {
    var d interface{}
    resp := r.(*callpb.CallReply)
    err := json.Unmarshal([]byte(resp.Data), &d)
    return callEndpoint.CallResponse{
        Data: d,
        Msg:  resp.Msg,
        Code: resp.Code,
    }, err
}

