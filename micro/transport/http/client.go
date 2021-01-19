package http

import (
    "bytes"
    "context"
    "github.com/jasonfeng1980/pg/util"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"

    kitHttp "github.com/go-kit/kit/transport/http"
    callEndpoint "github.com/jasonfeng1980/pg/micro/endpoint"
)

func NewClient(instance string, options []kitHttp.ClientOption) (callEndpoint.Set, error) {
    if !strings.HasPrefix(instance, "http") {
        instance = "http://" + instance
    }
    u, err := url.Parse(instance)
    if err != nil {
        return callEndpoint.Set{}, err
    }

    CallEndpoint := kitHttp.NewClient("POST",
        copyURL(u, "/api"),
        encodeRequest,
        decodeResponse,
        options...
    ).Endpoint()

    return callEndpoint.Set{CallEndpoint: CallEndpoint}, nil
}

func encodeRequest(_ context.Context, r *http.Request, request interface{}) error {
    var buf bytes.Buffer
    if err := util.Json.NewEncoder(&buf).Encode(request); err != nil {
        return err
    }
    r.Body = ioutil.NopCloser(&buf)
    return nil
}

func decodeResponse(_ context.Context, r *http.Response) (interface{}, error) {
    //if r.StatusCode != http.StatusOK {
    //   return nil, errors.New(r.Status)
    //}

    var resp callEndpoint.CallResponse
    err := util.Json.NewDecoder(r.Body).Decode(&resp)
    return resp, err
}

func copyURL(base *url.URL, path string) (next *url.URL) {
    n := *base
    n.Path = path
    next = &n
    return
}