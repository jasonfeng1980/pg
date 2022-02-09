package main

import (
	"context"
	"github.com/jasonfeng1980/pg"
	"github.com/jasonfeng1980/pg/util"

	_ "github.com/jasonfeng1980/pg/example/application/demo"
)

func main(){
	if err :=pg.Load("../conf/pg_demo.01.dev.json");err!= nil {
		util.Panic("加载配置错误", "error", err)
		return
	}
	srv := pg.Server(context.Background())
	srv.Run()
}
