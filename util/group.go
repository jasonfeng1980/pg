package util

import "sync"

// 并行执行方法
func Group() *group{
	return &group{
		wg: sync.WaitGroup{},
		box: make(map[string]*groupResult),
	}
}

type group struct {
	wg sync.WaitGroup
	box map[string]*groupResult
}
type groupResult struct {
	Data interface{}
	Code int64
	Msg  string
}

type Fn func()(data interface{}, code int64, msg string)
// 增加并行方法
func (g *group)Add(name string, fn Fn)() {
	g.box[name] = &groupResult{}
	g.wg.Add(1)
	go func() {
		g.box[name].Data, g.box[name].Code,g.box[name].Msg =  fn()
		g.wg.Done()
	}()
	return
}
// 开始等待
func (g *group)Wait(){
	g.wg.Wait()
}
// 获取执行的值
func (g *group)Get(name string) *groupResult{
	if v, ok := g.box[name]; ok {
		return v
	} else {
		return nil
	}
}
