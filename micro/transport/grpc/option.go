package grpc

import (
    "github.com/go-kit/kit/tracing/opentracing"
    "github.com/go-kit/kit/transport"
    kitGrpc "github.com/go-kit/kit/transport/grpc"
    "github.com/jasonfeng1980/pg/util"

    stdopentracing "github.com/opentracing/opentracing-go"
)

func DefaultClientOptions(otTracer stdopentracing.Tracer) []kitGrpc.ClientOption {
    logger := util.Log
    var opts  []kitGrpc.ClientOption
    opts = append(opts, kitGrpc.ClientBefore(opentracing.ContextToGRPC(otTracer,  logger)))
    return opts
}

func DefaultServerOptions(otTracer stdopentracing.Tracer) []kitGrpc.ServerOption {
    logger := util.Log
    opts := []kitGrpc.ServerOption{
        kitGrpc.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
    }
    opts = append(opts, kitGrpc.ServerBefore(opentracing.GRPCToContext(otTracer, "Call", logger)))
    return opts
}
