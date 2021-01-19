package util

import (
    "io/ioutil"
    "os"
    "path/filepath"
)


// 判断并创建目录，
func FileMakeDir(dir string) (err error){
    if !FilePathExists(dir) { // 没有就创建
        if err =os.MkdirAll(dir, os.ModePerm); err == nil{
            os.Chmod(dir, 0766)
        }
    }
    return
}
// 判断所给路径文件/文件夹是否存在
func FilePathExists(path string) bool {
    _, err := os.Stat(path)    //os.Stat获取文件信息
    if err != nil {
        if os.IsExist(err) {
            return true
        }
        return false
    }
    return true
}

// 判断所给路径是否为文件夹
func FilePathIsDir(path string) bool {
    s, err := os.Stat(path)
    if err != nil {
        return false
    }
    return s.IsDir()
}

// 判断所给路径是否为文件
func FilePathIsFile(path string) bool {
    return !FilePathIsDir(path)
}

// 一次性读整个文件
func FileRead(filename string) ([]byte, error){
    data, err := ioutil.ReadFile(filename)
    return data, err
}

// 写文件，没有文件夹就自动创建
func FileWrite(dir string, filename string, data string) error{
    //syscall.Umask(0000)
    if err :=FileMakeDir(dir); err !=nil{
        return err
    }
    return ioutil.WriteFile(dir + "/" + filename, []byte(data), 0755)
}

func FileRealPath(path string) string{
    if path == "" {
        path = "."
    }
    if path[:1]!="/" {
        path, _ = filepath.Abs(path)
    }
    return path
}