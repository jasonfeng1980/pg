package util

import (
    "github.com/jasonfeng1980/pg/ecode"
    jsoniter "github.com/json-iterator/go"
    "io"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary


func JsonEncoder(w io.Writer) *jsoniter.Encoder{
    return json.NewEncoder(w)
}

func JsonDecoder(r io.Reader) *jsoniter.Decoder{
    return json.NewDecoder(r)
}

// map => json_byte
func JsonEncode(data interface{}) ([]byte, error){
    return json.Marshal(data)
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
    return json.Unmarshal(b, v)
}
func JsonIndent(v interface{}) (string, error){
    r, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        return "", err
    }
    return string(r[:]), nil
}

// json_string => map
func JsonDecodeToMap(jsonStr string) (jsonMap map[string]interface{}, err error) {
    jsonMap = make(map[string]interface{})
    err = json.Unmarshal([]byte(jsonStr), &jsonMap)
    return
}

// json_string => []map
func JsonDecodeToListMap(jsonStr string) (jsonMap []map[string]interface{}, err error) {
    jsonMap = make([]map[string]interface{}, 0)
    err = json.Unmarshal([]byte(jsonStr), &jsonMap)
    return
}
