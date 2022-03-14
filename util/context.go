package util

import (
	"context"
	"database/sql"
	"time"
)


//////////////////////////////////////////////////
//  session相关操作
//////////////////////////////////////////////////
const (
	SessionKey = "PG_SESSION"
	SessionTransaction = "PG_SESSION_TRANSACTION"
)

type session struct {
	*Param
}
// session删除某个key
func SessionDelete(ctx context.Context, k string) {
	m := ctx.Value(SessionKey)
	m.(*session).Delete(k)
}
// session添加一个K-V
func SessionSet(ctx context.Context, k string, v interface{}) {
	m := ctx.Value(SessionKey)
	m.(*session).Set(M{k:v})
}
// 获取session句柄
func SessionHandle(ctx context.Context) *session{
	m := ctx.Value(SessionKey)
	if m == nil {
		//Panic(ecode.UtilSessionNotNew.Error())
		return &session{&Param{}}
	}
	return m.(*session)
}
// 开启一个session
func SessionNew(ctx context.Context) context.Context{
	m := ctx.Value(SessionKey)
	if m == nil {
		ctx = context.WithValue(ctx, SessionKey, &session{&Param{}})
	}
	return ctx
}
//////////////////////////////////////////////////
//  隐藏error相关操作
//  必须先开启session
//////////////////////////////////////////////////
const (
	hideErrorCancelKey = "HIDE_ERROR_CANCEL"
	hideErrorKey = "HIDE_ERROR"
)
func HideErrorWatch(ctx context.Context) context.Context{
	ctxNew, cancel := context.WithCancel(SessionNew(ctx))
	ctxNew = context.WithValue(ctxNew, hideErrorCancelKey, cancel)
	return ctxNew
}
func HideErrorCancel(ctx context.Context, err error){
	c := ctx.Value(hideErrorCancelKey)
	if c != nil {
		cancel := c.(context.CancelFunc)
		SessionSet(ctx, hideErrorKey, err)
		cancel()
		time.Sleep(time.Millisecond*10)
		return
	}
	// 没有cancel ，就直接panic，让 micro/service/service.go 处理
	panic(err)
}
func HideErrorGet(ctx context.Context) error{
	s := SessionHandle(ctx)
	if err := s.Get(hideErrorKey); err !=nil {
		return err.(error)
	}
	return nil
}


//////////////////////////////////////////////////
//  事务相关操作
//////////////////////////////////////////////////
// 开启事务
func StartTransaction(ctx context.Context) {
	SessionSet(ctx, SessionTransaction, make(map[string]*sql.Tx))
}
// 是否开启了事务
func IsTransaction(ctx context.Context) bool{
	session := SessionHandle(ctx)
	return  session.Get(SessionTransaction) == nil

}
func Commit(ctx context.Context) {
	session := SessionHandle(ctx)
	tx := session.Get(SessionTransaction)
	if tx == nil {
		return
	}
	for _, v := range tx.(map[string]*sql.Tx){
		v.Commit()
	}
	SessionDelete(ctx, SessionTransaction)
}
func Rollback(ctx context.Context) {
	session := SessionHandle(ctx)
	tx := session.Get(SessionTransaction)
	if tx == nil {
		return
	}
	for _, v := range tx.(map[string]*sql.Tx){
		v.Rollback()
	}
	SessionDelete(ctx, SessionTransaction)
}