package http

import (
    "fmt"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/micro/service"
    "github.com/jasonfeng1980/pg/util"
    "mime/multipart"
    "net/http"
)


func getRequestParams(conf conf.Config, r *http.Request) (dns string, params map[string]interface{}, err error){
    dns = fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)

    var (
    	jsonRequest map[string]interface{}
        files  map[string][]*multipart.FileHeader
    )

    if r.URL.Path == "/api"{ // 服务调用
        util.Json.NewDecoder(r.Body).Decode(&jsonRequest)
        if v, ok:=jsonRequest["Dns"]; ok {// 传参
            dns = v.(string)
            if jsonRequest["Params"] != nil {
                if params, ok = jsonRequest["Params"].(map[string]interface{}); !ok{
                    err = ecode.HttpDataNotMap.Error()
                }
            }
        }
    } else {
        get := httpDecodeArgs(r.URL.Query())
        var post map[string]interface{}
        r.ParseMultipartForm(conf.WebMaxBodySizeM)
        if r.MultipartForm != nil {
            if r.MultipartForm.File != nil {
                files = r.MultipartForm.File
            }

            post = httpDecodeArgs(r.MultipartForm.Value)

        } else {
            util.Json.NewDecoder(r.Body).Decode(&jsonRequest)
            r.ParseForm()
            post = httpDecodeArgs(r.PostForm)
        }
        params = util.MapMerge(jsonRequest, post, get)
        if files != nil { // 文件上传
            params[service.UploadFile] = files
        }
    }

    params[service.RequestHandle] = r
    return
}



func httpDecodeArgs(args map[string][]string)  map[string]interface{} {
    handle := make(map[string]interface{})
    for k, arg := range args {
        for _, v := range arg {
            //var handle map[string]interface{}
            isArray := false
            // 将a[]  变成 a
            if lenK := len(k); lenK > 2 && k[lenK-2:] == "[]" {
                k = k[0 : lenK-2]
                isArray = true
            }

            if old, ok := handle[k]; ok { // 如果之前是否设置过， 就要变成数组
                switch old.(type) {
                case []string:
                    handle[k] = append(old.([]string), v)
                case string:
                    handle[k] = []string{old.(string), v}
                }
            } else {
                if isArray {
                    handle[k] = []string{v}
                } else {
                    handle[k] = v
                }
            }
        }
    }
    return handle
}