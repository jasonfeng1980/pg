package util

import (
    "context"
    "github.com/jasonfeng1980/pg/ecode"
    "io"
    "mime/multipart"
    "os"
)

const UploadFile = "__PG_UPLOAD"

type upload struct {
    FileHeader map[string][]*multipart.FileHeader
}

func Upload(ctx context.Context) (ret *upload, err error){
    f := ctx.Value(UploadFile)
    if v, ok := f.(map[string][]*multipart.FileHeader);ok {
        ret =  &upload{
            FileHeader: v,
        }
    } else {
        err = ecode.UtilNoUploadFile.Error()
    }
    return
}

func (u *upload)keys() []string{
    return MapKeys(u.FileHeader)
}

// 保存
func (u *upload)Save(key string, fileDir string, fileName string) error{
    list, ok := u.FileHeader[key]
    if !ok || len(list)==0{
        return ecode.UtilNoUploadFile.Error()
    }
    src, err := list[0].Open()
    if err != nil {
        return err
    }
    defer src.Close()
    FileMakeDir(fileDir)
    dst, err := os.OpenFile(fileDir + "/" + fileName, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        return err
    }
    defer dst.Close()
    io.Copy(dst, src)
    return nil
}

