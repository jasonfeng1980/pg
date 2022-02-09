package util

import (
	"context"
	"fmt"
	"github.com/jasonfeng1980/pg/ecode"
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	ctx := context.Background()
	ctx = SessionNew(ctx)
	SessionSet(ctx, "a", 1111)
	s := SessionHandle(ctx)
	fmt.Println(s.Get("a"))
}

func TestHideErr(t *testing.T) {
	ctx1 := context.Background()
	c2 := SessionNew(ctx1)

	c3 := HideErrorWatch(c2)
	go func() {
		tt(c3)
	}()
	func() {
		select {
		case <-c3.Done():
			if err := HideErrorGet(c3); err!=nil {
				fmt.Println("-------",err)
				return
			}
		}
	}()

	fmt.Println("33333")
}

func tt(ctx context.Context){
	time.Sleep(time.Second*2)

	HideErrorCancel(ctx, ecode.ConfWrong.Error("xxxxx"))
	fmt.Println("111111")
	time.Sleep(time.Second*2)
	fmt.Println("2222")
}