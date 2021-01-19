package util

import (
    jsoniter "github.com/json-iterator/go"
)

var Json = jsoniter.ConfigCompatibleWithStandardLibrary
// map => byte
func JsonEncode(data interface{}) ([]byte, error){
    return Json.Marshal(data)
}
func JsonDecode(data []byte, v interface{}) error{
    return Json.Unmarshal(data, v)
}

// json => map
func JsonDecodeToMap(jsonStr string) (jsonMap map[string]interface{}, err error) {
    jsonMap = make(map[string]interface{})
    err = Json.Unmarshal([]byte(jsonStr), &jsonMap)
    return
}

// json => []map
func JsonDecodeToListMap(jsonStr string) (jsonMap []map[string]interface{}, err error) {
    jsonMap = make([]map[string]interface{}, 0)
    err = Json.Unmarshal([]byte(jsonStr), &jsonMap)
    return
}
