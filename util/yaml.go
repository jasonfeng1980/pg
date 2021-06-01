package util

import "gopkg.in/yaml.v2"

// 读取YAML文件
func YamlRead(filePath string, m interface{}) error{
    filePath = FileRealPath(filePath)
    data, e := FileRead(filePath)
    if e != nil {
        return e
    }
    return yaml.Unmarshal(data, m)
}