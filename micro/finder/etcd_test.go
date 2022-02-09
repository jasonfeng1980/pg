package finder

import (
	"context"
	"fmt"
	"github.com/jasonfeng1980/pg/conf"
	"github.com/jasonfeng1980/pg/util"
	"testing"
	"time"
)

var (
	findEtcd *etcdClient
	err error
)

func TestMain(m *testing.M) {

	findEtcd,  err= NewEtcd(context.Background(), conf.Conf.ETCD)
	if err != nil {
		util.Panic(err.Error())
	}

	m.Run()
}

func TestEtcd_Register(t *testing.T) {
	findEtcd.Register("/PG/grpc/11", "127.0.0.1:9081")
	findEtcd.Register("/PG/grpc/22", "127.0.0.1:8081")
	findEtcd.Register("/PG/http/11", "127.0.0.1:80")
	findEtcd.Register("/PG/http/22", "127.0.0.1:8080")
}


func TestEntries(t *testing.T) {
	fmt.Println(findEtcd.GetEntries("/PG/grpc"))
	fmt.Println(findEtcd.GetEntries("/PG/http"))
}

func TestWatch(t *testing.T) {
	prefix := "grpc://pg/test"

	go func() {
		for {
			fmt.Println(findEtcd.Get(prefix))
			time.Sleep(time.Second * 2)
		}

	}()

	ch := make(chan int)
	select {
	case <-ch:

	}
}

func TestEdit(t *testing.T) {
	prefix := "grpc://pg/test"
	// 增加
	findEtcd.Register(prefix + "4", "44--" + util.TimeNowString())
	// 修改
	findEtcd.Register(prefix + "1", "111===" + util.TimeNowString())
	findEtcd.Register(prefix + "2", "2222===" + util.TimeNowString())

	time.Sleep(time.Second * 20)

}

