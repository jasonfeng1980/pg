package util

import (
    "encoding/json"
    "github.com/jasonfeng1980/pg/ecode"
    jsoniter "github.com/json-iterator/go"
)

var Json = jsoniter.ConfigCompatibleWithStandardLibrary

// map => json_byte
func JsonEncode(data interface{}) ([]byte, error){
    return Json.Marshal(data)
}
// json_byte or string => map
func JsonDecode(data interface{}, v interface{}) error{
    var b []byte
    switch data.(type) {
    case []byte:
        b = data.([]byte)
    case string:
        b = []byte(data.(string))
    default:
        return ecode.UtilErrDecodeJson.Error()
    }
    return Json.Unmarshal(b, v)
}
func JsonIndent(v interface{}) (string, error){
    r, err := json.MarshalIndent(v, "", "\t")
    if err != nil {
        return "", err
    }
    return string(r[:]), nil
}

// json_string => map
func JsonDecodeToMap(jsonStr string) (jsonMap map[string]interface{}, err error) {
    jsonMap = make(map[string]interface{})
    err = Json.Unmarshal([]byte(jsonStr), &jsonMap)
    return
}

// json_string => []map
func JsonDecodeToListMap(jsonStr string) (jsonMap []map[string]interface{}, err error) {
    jsonMap = make([]map[string]interface{}, 0)
    err = Json.Unmarshal([]byte(jsonStr), &jsonMap)
    return
}
