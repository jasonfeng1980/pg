package http

import (
    "bytes"
    "context"
    "errors"
    "github.com/jasonfeng1980/pg/util"
    "github.com/opentracing/opentracing-go"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"

    "github.com/jasonfeng1980/pg/micro/endpoint"
)

func NewClient(instance string, tracer opentracing.Tracer) (endpoint.Endpoint, error) {
    if !strings.HasPrefix(instance, "http") {
        instance = "http://" + instance
    }
    u, err := url.Parse(instance)
    if err != nil {
        return nil, err
    }

    CallEndpoint := Client{
        client: http.DefaultClient,
        method: "POST",
        tgt:    copyURL(u, "/api"),
        tracer: tracer,
    }.Endpoint()

    return CallEndpoint, nil
}

type Client struct {
    client         *http.Client
    method         string
    tgt            *url.URL
    tracer         opentracing.Tracer
}

func (c Client) Endpoint() endpoint.Endpoint {
    return func(ctx context.Context, request interface{}) (interface{}, error) {
        // 获得超时取消方法
        ctx, cancel := context.WithCancel(ctx)

        var (
            resp *http.Response
            err  error
        )

        req, err := http.NewRequest(c.method, c.tgt.String(), nil)
        if err != nil {
            cancel()
            return nil, err
        }
        // 加密request
        if err = c.EncodeRequest(ctx, req, request); err != nil {
            cancel()
            return nil, err
        }
        // 链路追踪
        ctx = TraceClient(c.tracer)(ctx, req)
        // 执行
        resp, err = c.client.Do(req.WithContext(ctx))
        if err != nil {
            cancel()
            return nil, err
        }
        defer resp.Body.Close()
        defer cancel()
        // 解密Response
        response, err := c.DecodeResponse(ctx, resp)
        if err != nil {
            return nil, err
        }

        return response, nil
    }
}


func (c Client) EncodeRequest(_ context.Context, r *http.Request, request interface{}) error {
    var buf bytes.Buffer
    if err := util.JsonEncoder(&buf).Encode(request); err != nil {
        return err
    }
    r.Body = ioutil.NopCloser(&buf)
    return nil
}

func (c Client) DecodeResponse(_ context.Context, r *http.Response) (interface{}, error) {
    if r.StatusCode != http.StatusOK { // 相应状态不对
      return nil, errors.New(r.Status)
    }

    var resp endpoint.CallResponse
    err := util.JsonDecoder(r.Body).Decode(&resp)
    return resp, err
}

func copyURL(base *url.URL, path string) (next *url.URL) {
    n := *base
    n.Path = path
    next = &n
    return
}