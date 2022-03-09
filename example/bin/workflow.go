package main

import (
    "codeup.aliyun.com/5f119a766207a1a8b17f58ba/aplum_mis/workflow/aggregate/wf"
    "context"
    "fmt"
    "github.com/jasonfeng1980/pg"
    "github.com/jasonfeng1980/pg/util"
)

var (
    params = util.M{"product_id": 888}
    enterWarehouse *wf.EnterWarehouseRoot
    workflowId int64
    err error
    ctx context.Context
    tipIndex  = 0
    operatorId int64
)

func main(){
    if err :=pg.Load("../conf/workflow.01.dev.json");err!= nil {
        util.Panic("加载配置错误", "error", err)
        return
    }
    srv := pg.Server(context.Background())
    defer srv.Close()
    ctx = srv.Script()


    util.ConsoleTip("请输入输入操作人ID(整数)", docketSubmit)

}

func docketSubmit(cmdString string) (string, bool){
    switch true {
    case cmdString == "quit" || cmdString == "exit":
        return "bye", false
    case tipIndex == 0:
        operatorId, err = util.Int64Parse(cmdString)
        if err != nil || operatorId<=0{
            return "请正确的输入操作人ID(整数)", true
        }
        tipIndex++
        return "请输入工作流ID(整数)，开启新的工作流输入0", true
    case tipIndex == 1:
        workflowId, err = util.Int64Parse(cmdString)
        if err != nil {
            return "请输入正确工作流ID(整数)，开启新的工作流输入0", true
        }
        enterWarehouse = wf.NewEnterWarehouseFromPK(ctx, workflowId)
        tipIndex++
        return "请输入提交的单据编码", true
    default:
        workflowId, _, err = enterWarehouse.Submit(operatorId, workflowId, cmdString, params)
        if err != nil{
            fmt.Println("出现错误:" , err.Error())
        }
        printWaitNode(enterWarehouse)
        return "请输入提交的单据编码", true
    }
}

// 测试输出当前待处理的工作流
func printWaitNode(w *wf.EnterWarehouseRoot){
    var waitNode []string

    for _, v := range w.NodeVersion{
        if v.IsEnd ==0 {
            waitNode = append(waitNode, w.MapNodeStatus[v.Status].Name)
        }
    }
    if len(waitNode)==0 {
        fmt.Println("执行完毕: 【工作流处理完毕】\n")
    } else {
        fmt.Printf("执行完毕: 当前需要待处理节点为【%s】\n\n", util.Str(waitNode))
    }
}

