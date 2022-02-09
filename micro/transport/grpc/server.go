package grpc

import (
    "context"
    "github.com/jasonfeng1980/pg/micro/endpoint"
    "github.com/jasonfeng1980/pg/micro/transport/grpc/pb"
    "github.com/jasonfeng1980/pg/util"
    "github.com/opentracing/opentracing-go"
    "google.golang.org/grpc/metadata"
)

type ErrorFunc func(ctx context.Context, err error)

type grpcServer struct {
    call GrpcHandler
}

type GrpcHandler interface {
    ServeGRPC(ctx context.Context, request interface{}) (context.Context, interface{}, error)
}

func (g *grpcServer) Call(ctx context.Context, req *pb.CallRequest) (*pb.CallReply, error) {
    _, rep, err := g.call.ServeGRPC(ctx, req)
    if err != nil {
        return nil, err
    }
    return rep.(*pb.CallReply), nil
}

// 获得服务句柄
func NewServer(endpoints endpoint.MicroEndpoint, tracer opentracing.Tracer, operationName string) pb.ServiceServer {
    m := &grpcServer{
        call: Server{
            ept: endpoints,
            operationName: operationName,
            tracer: tracer,
        },
    }
    return m
}

type Server struct {
    ept           endpoint.MicroEndpoint
    operationName   string
    tracer        opentracing.Tracer
}


// GRPC
func (s Server) ServeGRPC(ctx context.Context, req interface{}) (context.Context, interface{}, error) {
    var (
        request  interface{}
        response interface{}
        grpcResp interface{}
        err      error
    )

    // Retrieve gRPC metadata.
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        md = metadata.MD{}
    }

    // 处理trace
    ctx = TraceServer(s.tracer, "Call")(ctx, md)

    // 开启session
    ctx = util.SessionNew(ctx)

    // 解密request
    ctx, request, err = s.DecodeRequest(ctx, req)
    if err != nil {
        return ctx, nil, err
    }
    // 请求
    response, err = s.ept.Endpoint(ctx, request)
    if err != nil {
        return ctx, nil, err
    }

    // 加密response
    grpcResp, err = s.EncodeResponse(ctx, response)
    if err != nil {
        return ctx, nil, err
    }

    return ctx, grpcResp, nil
}

// 解密请求参数
func (s Server)DecodeRequest(ctx context.Context, r interface{}) (context.Context, interface{}, error) {
    var d map[string]interface{}
    err := util.JsonDecode([]byte(r.(*pb.CallRequest).Params), &d)
    if err != nil {
        return ctx, nil, err
    }

    return ctx,
    endpoint.CallRequest{
        Dns: r.(*pb.CallRequest).Dns,
        Params: d,
    }, nil
}

// 加密响应数据
func (s Server)EncodeResponse(_ context.Context, r interface{}) (interface{}, error) {
    resp := r.(endpoint.CallResponse)
    d, err := util.JsonEncode(resp.Data)
    return &pb.CallReply{
        Data: string(d[:]),
        Code: resp.Code,
        Msg:  resp.Msg,
    }, err
}



