package main

import (
    "context"
    "flag"
    "fmt"
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/ddd"
    "os"
)

var (
    configFile string
    dbHandleName string
)

func main(){
    // 根据参数加载配置文件
    flagSet := flag.NewFlagSet("build", flag.ExitOnError)
    flagSet.StringVar(&configFile, "c", "", "配置文件地址")
    flagSet.StringVar(&dbHandleName, "db", "", "配置文件地址")

    switch os.Args[1] {
    case "build":
        flagSet.Parse(os.Args[2:])
    default:
        fmt.Println("请输入命令  e.g.  pg build -c dev.json -db demo")
        return
    }

    if configFile == "" {
        fmt.Println("请出入配置文件的路径  e.g.  pg build -c dev.json -db demo")
    }
    if dbHandleName == "" {
        fmt.Println("请出入配置文件的路径  e.g.  pg build -c dev.json -db demo")
    }

    if err :=pg.Load(configFile);err!= nil {
        fmt.Println("加载配置错误", err)
        os.Exit(1)
    }

    srv := pg.Server(context.Background())
    _ = srv.Script()
    build()
}

func build() error{
    ddd.BuildEntity(dbHandleName, pg.Conf.GetString("Build.Package"), "repository")

    return nil
}
