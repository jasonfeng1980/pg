package grpc

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/jasonfeng1980/pg/util"
    "github.com/opentracing/opentracing-go"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    "reflect"

    endpoint "github.com/jasonfeng1980/pg/micro/endpoint"
    pb "github.com/jasonfeng1980/pg/micro/transport/grpc/pb"
)

func NewClient(instanceAddr string, tracer opentracing.Tracer) (endpoint.Endpoint, error) {
    conn, err := grpc.DialContext(context.Background(), instanceAddr, grpc.WithInsecure())
    if err != nil {
        util.Error(fmt.Sprintf("连接GRPC失败：%s", instanceAddr))
        return nil, err
    }
    CallEndpoint := Client{
        client: conn,
        method: "/pb.Service/Call",
        grpcReply: reflect.TypeOf(
            reflect.Indirect(
                reflect.ValueOf(pb.CallReply{}),
            ).Interface(),
        ),
        tracer: tracer,
    }.Endpoint()
    return CallEndpoint, nil
}

func encodeRequest(_ context.Context, r interface{}) (interface{}, error) {
    d, err := json.Marshal(r.(endpoint.CallRequest).Params)
    if err != nil {
        return nil ,err
    }
    return &pb.CallRequest{
        Dns: r.(endpoint.CallRequest).Dns,
        Params: string(d[:]),
    }, nil
}

func decodeResponse(_ context.Context, r interface{}) (interface{}, error) {
    var d interface{}
    resp := r.(*pb.CallReply)
    err := json.Unmarshal([]byte(resp.Data), &d)
    return endpoint.CallResponse{
        Data: d,
        Msg:  resp.Msg,
        Code: resp.Code,
    }, err
}

type Client struct {
    client      *grpc.ClientConn
    method      string
    grpcReply   reflect.Type
    tracer         opentracing.Tracer
}
func (c Client) Endpoint() endpoint.Endpoint {
    return func(ctx context.Context, request interface{}) (response interface{}, err error) {
        ctx, cancel := context.WithCancel(ctx)
        defer cancel()

        req, err := encodeRequest(ctx, request)
        if err != nil {
            return nil, err
        }

        md := &metadata.MD{}
        ctx = TraceClient(c.tracer)(ctx, md)
        ctx = metadata.NewOutgoingContext(ctx, *md)

        var header, trailer metadata.MD
        grpcReply := reflect.New(c.grpcReply).Interface()
        if err = c.client.Invoke(
            ctx, c.method, req, grpcReply, grpc.Header(&header),
            grpc.Trailer(&trailer),
        ); err != nil {
            return nil, err
        }

        response, err = decodeResponse(ctx, grpcReply)
        if err != nil {
            return nil, err
        }
        return response, nil
    }
}