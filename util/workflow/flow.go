package workflow

import (
    "github.com/jasonfeng1980/pg/ddd"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
)


/////////////////////////////////////////////////////////
// 工作流  主记录，包含信息+状态
/////////////////////////////////////////////////////////
type Flow struct {
    BillMap    map[string]*Bill // 拥有的各个节点对应的单据 map[Node.Name]*Bill

    *ddd.DAO         // 工作流对应的数据库表的DAO
}

// 工作流状态修改方法 (当前的Node，驱动的单据) 错误
type Action func(node *Node, bill *Bill) error

// 工作流通知方法   (当前的Node，驱动的单据) 错误
type Notify func(node *Node, bill *Bill) error

/////////////////////////////////////////////////////////
// 工作流-单据，工作流主驱动
/////////////////////////////////////////////////////////
type Bill struct {
    Name   string

    ToNode string   // 要去什么节点

    Flow   *Flow    // 单据所在的工作流

    ActionFn Action     // 在各个节点需要执行的方法
    NotifyFn Notify     // 单据对应的通知事件

    *ddd.DAO         // 单据对应的数据库表的DAO
}

// 执行单据，外部加事务, 返回日志主键
func (b *Bill)Save(node *Node) (pk int64, err error){
    // 1. 验证当今节点是否支持bill
    if err = node.isAllow(b); err !=nil {
        return
    }

    // 2. 创建单据记录
        // 验证必填
    if err = b.CheckParams(); err != nil {
        return
    }
    if pk, err = b.Create(); err != nil {
        return
    }

    // 3. 执行action改变节点
    if err = node.Change(b); err != nil {
        return
    }

    // 4. 通知执行成功事件 -- 不在事务内，出错写error日志
    if err = b.NotifyFn(node, b); err != nil {
        util.Error(b.Name + "-推送通知失败", "err", err)
    }

    return
}

/////////////////////////////////////////////////////////
// 工作流-节点
/////////////////////////////////////////////////////////
type Node struct {
    Name        string
    Code        string            // 对应的标识
    Action      map[string]Action // 支持的单据，map[bill名称]Action
}

// 当前节点是否支持传来的单据
func (n *Node)isAllow(bill *Bill) error{
    _, ok := n.Action[bill.Name]
    if !ok {
        return ecode.WorkFlowNodeNotSupportBill.Error(n.Name, bill.Name)
    }
    return nil
}

func (n *Node)Change(bill *Bill) error{
    // 1. 获取当今节点，单据对应的行为
    actFn, ok := n.Action[bill.Name]
    if !ok {
        return ecode.WorkFlowNodeNotSupportBill.Error(n.Name, bill.Name)
    }
    // 2. 执行行为
    return actFn(n, bill)
}

