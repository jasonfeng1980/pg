package util

import (
    "gopkg.in/yaml.v2"
)

// 获取一个数组map的 指定字段
func ListMapField(dataArg interface{}, fieldArr []string) (ret []map[string]interface{}, err error) {
    data, err := InterfaceToArr(dataArg)
    if err!= nil {
        return nil, err
    }
    // 循环赋值
    for _, v := range data {
        d, _ := MapField(v, fieldArr, false)
        ret = append(ret, d)
    }
    return
}

// 读取YAML文件
func YamlRead(filePath string, m interface{}) error{
    filePath = FileRealPath(filePath)
    data, e := FileRead(filePath)
    if e != nil {
        return e
    }
    return yaml.Unmarshal(data, m)
}
