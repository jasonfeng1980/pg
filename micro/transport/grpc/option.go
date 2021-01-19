package grpc

import (
    "github.com/go-kit/kit/tracing/opentracing"
    "github.com/go-kit/kit/tracing/zipkin"
    "github.com/go-kit/kit/transport"
    kitGrpc "github.com/go-kit/kit/transport/grpc"
    "github.com/jasonfeng1980/pg/util"

    stdopentracing "github.com/opentracing/opentracing-go"
    stdzipkin "github.com/openzipkin/zipkin-go"
)

func DefaultClientOptions(otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) []kitGrpc.ClientOption {
    logger := util.LogHandle("info")
    var opts  []kitGrpc.ClientOption
    if zipkinTracer != nil {
        opts = append(opts, zipkin.GRPCClientTrace(zipkinTracer))
    }
    opts = append(opts, kitGrpc.ClientBefore(opentracing.ContextToGRPC(otTracer,  logger)))
    return opts
}

func DefaultServerOptions(otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) []kitGrpc.ServerOption {
    logger := util.LogHandle("error")
    opts := []kitGrpc.ServerOption{
        kitGrpc.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
    }
    if zipkinTracer != nil {
        opts = append(opts, zipkin.GRPCServerTrace(zipkinTracer))
    }
    opts = append(opts, kitGrpc.ServerBefore(opentracing.GRPCToContext(otTracer, "Call", logger)))
    return opts
}
