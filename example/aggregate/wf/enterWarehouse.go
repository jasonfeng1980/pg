package wf

import (
    "codeup.aliyun.com/5f119a766207a1a8b17f58ba/aplum_mis/workflow/ecode"
    "codeup.aliyun.com/5f119a766207a1a8b17f58ba/aplum_mis/workflow/entity/workflowEntity"
    "context"
    "encoding/json"
    "fmt"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/database/db"
    "github.com/jasonfeng1980/pg/ddd"
    "github.com/jasonfeng1980/pg/util"
    "github.com/jasonfeng1980/pg/util/workflow"
)


// 通过PK获取聚合根
func NewEnterWarehouseFromPK(ctx context.Context, pk int64) (wf *EnterWarehouseRoot) {
    var (
        prefix = "WF-Enter"
        cf workflow.Conf
    )

    // 获取配置
    if confJson, err := util.FileRead(conf.ConfBox.GetString("Workflow.enter_warehouse")); err !=nil {
        return nil
    } else if err =json.Unmarshal(confJson, &cf); err !=nil {
        return
    }

    // 获取实例，并绑定方法
    ew := &EnterWarehouseRoot{
        Ctx:                    ctx,
        EnterWarehouseEntity:   workflowEntity.NewEnterWarehouseEntity(ctx, pk),
        AggregateRoot:          ddd.NewAggregateRoot(prefix, pk),
    }
    ew.Workflow = workflow.New(&cf, ew)
    return ew
}

// 入库工作流
type EnterWarehouseRoot struct {
    Ctx             context.Context
    *workflow.Workflow
    *workflowEntity.EnterWarehouseEntity
    *ddd.AggregateRoot
}


// 更新各个节点的状态和版本
func (w *EnterWarehouseRoot)UpdateNodeStatus(operatorId int64, workflowId int64, nodeStatus *workflow.ConfNodeStatus, params map[string]interface{}) (rollbackAct workflow.RollbackParams, err error){
    var (
        ewNode  *workflow.EnterWarehouseNode
        ok          bool
        rows    int64
    )
    // 更新或插入节点数据
    if ewNode, ok = w.NodeVersion[nodeStatus.NodeCode]; !ok { // 如果不存在就创建
        nodeEntity := workflowEntity.NewEnterWarehouseNodeEntity(w.Ctx)
        // 赋值
        nodeEntity.SetWorkflowTplCode(w.Workflow.Conf.Info.Code)
        nodeEntity.SetEnterWarehouseId(workflowId)
        nodeEntity.SetProductId(params["product_id"])
        nodeEntity.SetNodeCode(nodeStatus.NodeCode)
        nodeEntity.SetNodeStatus(nodeStatus.Code)
        nodeEntity.SetNodeIsEnd(nodeStatus.End)
        nodeEntity.SetVersion(1)
        nodeEntity.SetEnterWarehouseNodeDeleted(0)
        nodeEntity.SetOpId(operatorId)
        // 创建
        if err = nodeEntity.Create(); err!=nil{
            return
        }
        // 整理回滚数据
        rollbackAct = workflow.RollbackParams{
            Name: "updateNode",
            Params: util.M{"delete": fmt.Sprintf("delete from %s where %s = %d",
                nodeEntity.TableName(), nodeEntity.PkName(), nodeEntity.Pk())},
        }
        // 更新工作流
        if _, rollbackAct2, e :=w.ChangeWorkflowStatus(operatorId, workflowId, nodeStatus, params); e!=nil {
            err = e
            return
        } else {
            rollbackAct.Params["updateWorkflow"] = rollbackAct2.Params["updateWorkflow"]
        }
        // 写入节点版本
        w.NodeVersion[nodeStatus.NodeCode] = &workflow.EnterWarehouseNode{
            Pk: nodeEntity.Pk(),
            Status: nodeStatus.Code,
            Version: 1,
            IsEnd: nodeStatus.End,
        }
    } else { // 如果存在就更新
        nodeEntity := workflowEntity.NewEnterWarehouseNodeEntity(w.Ctx, ewNode.Pk)
        rows, err = nodeEntity.Edit(util.M{
            "node_is_end": nodeStatus.End,
            "node_status": nodeStatus.Code,
            "version": ewNode.Version+1,
        }).Where(util.M{
            "enter_warehouse_node_id": ewNode.Pk,
            "version": ewNode.Version,
        }).Run().RowsAffected();
        if err != nil {
            return
        }
        if rows == 0 { // 有其他人修改节点状态，就返回错误
            err = ecode.NodeVersionChange.Error(nodeStatus.NodeCode, ewNode.Version, nodeStatus.Code)
            return
        }
        // 整理回滚数据
        rollbackAct = workflow.RollbackParams{
            Name: "updateNode",
            Params: util.M{"updateNode": fmt.Sprintf(
                "update %s set node_status = '%s', node_is_end='%d', version=version+1 where %s = %d ",
                nodeEntity.TableName(), ewNode.Status, ewNode.IsEnd, nodeEntity.PkName(), nodeEntity.Pk())},
        }

        // 更新工作流
        if _, rollbackActWorkflow, e :=w.ChangeWorkflowStatus(operatorId, workflowId, nodeStatus, params); e!=nil {
            err = e
            return
        } else {
            rollbackAct.Params["updateWorkflow"] = rollbackActWorkflow.Params["updateWorkflow"]
        }


        // 更新节点版本
        ewNode.Status = nodeStatus.Code
        ewNode.IsEnd = nodeStatus.End
        ewNode.Version++
    }
    // 执行观察者事件
    if rollbackActOb, e := w.NotifyOb(operatorId, workflowId, nodeStatus, params); e!=nil{
        err = e
        return
    } else {
        rollbackAct.Params["OB"] = rollbackActOb
    }
    return
}

// 获取当前待处理的节点状态
func (w *EnterWarehouseRoot)GetWorkflowToDoStatus(workflowId int64) (workflowStatusList []string, err error){
    rs, err := w.RelationEnterWarehouseNode().Where("enter_warehouse_node_deleted", 0).Result()
    if err != nil {
        return nil, err
    }
    nodeList := rs.([]*workflowEntity.EnterWarehouseNodeEntity)
    // 生成节点状态，并记录待处理
    for _, v := range nodeList {
        w.NodeVersion[v.GetNodeCode()] = &workflow.EnterWarehouseNode{
            Pk: v.Pk(),
            Status: v.GetNodeStatus(),
            Version: v.GetVersion(),
            IsEnd: v.GetNodeIsEnd(),
        }
        if v.GetNodeIsEnd() == 0 {
            workflowStatusList = append(workflowStatusList, v.GetNodeStatus())
        }
    }
    return
}
// 获取指定节点状态
func (w *EnterWarehouseRoot)GetNodeStatus(workflowId int64, nodeCode string) (nodeStatus string, err error){
    rs, err := w.RelationEnterWarehouseNode().Where(util.M{
        "enter_warehouse_id":   workflowId,
        "node_code":            nodeCode,
    }).Limit(0,1).Result()
    if err != nil {
        return "", err
    }
    nodeList := rs.([]*workflowEntity.EnterWarehouseNodeEntity)
    if len(nodeList)==0 {
        return "", nil
    } else {
        return nodeList[0].GetNodeStatus(), nil
    }
}

// 创建工作流,记录日志 - 统一返回格式
func (w *EnterWarehouseRoot)CreateWorkflow(operatorId int64, params map[string]interface{}) (nowWorkflowId int64, rollbackAct []workflow.RollbackParams, err error){
    // 添加默认值
    params["workflow_tpl_code"] = w.Workflow.Conf.Info.Code
    params["op_id"] = operatorId
    // 赋值
    errList := w.EnterWarehouseEntity.SetMany(params)
    if len(errList)>0 {
        return 0, nil, errList[0]
    }

    // 创建 - - product_id 必填
    if err = w.EnterWarehouseEntity.Create("product_id"); err!=nil{
        return
    }
    // 整理回滚数据
    fmt.Println(" - - - - -创建工作流: pk =", w.EnterWarehouseEntity.Pk())
    nowWorkflowId = w.EnterWarehouseEntity.Pk()
    rollbackAct = []workflow.RollbackParams{{
        Name: "createWorkflow",
        Params: util.M{"delete": fmt.Sprintf("delete from %s where %s = %d",
            w.EnterWarehouseEntity.TableName(), w.EnterWarehouseEntity.PkName(), nowWorkflowId)},
    }}

    return
}
// 更新节点状态,记录日志 - 统一返回格式
func (w *EnterWarehouseRoot)ChangeNodeStatus(operatorId int64, workflowId int64, nodeStatus *workflow.ConfNodeStatus, params map[string]interface{}) (nowWorkflowId int64, rollbackAct workflow.RollbackParams, err error){
    fmt.Printf(" - - - - -更新节点状态到 %s(%s)\n", nodeStatus.Name , nodeStatus.Code)
    rollbackAct, err = w.UpdateNodeStatus(operatorId, workflowId, nodeStatus, params)
    return workflowId, rollbackAct, err
}
// 更新工作流状态,记录日志 - 统一返回格式
func (w *EnterWarehouseRoot)ChangeWorkflowStatus(operatorId int64, workflowId int64, nodeStatus *workflow.ConfNodeStatus, params map[string]interface{}) (nowWorkflowId int64, rollbackAct workflow.RollbackParams, err error){
    ew := w.EnterWarehouseEntity
    ew.SetMany(util.M{
        nodeStatus.NodeCode: nodeStatus.Code,
    })
    ew.SetVersion(db.Expr("version+1"))
    ew.SetEnterWarehouseEnd(nodeStatus.Finish)
    ew.SetOpId(operatorId)
    // 更新
    if _, e := w.EnterWarehouseEntity.Edit().Run().RowsAffected(); e!=nil{
        return workflowId, rollbackAct, e
    }
    // 整理回滚数据
    var oldStatus=""
    if ewNode, ok := w.NodeVersion[nodeStatus.NodeCode]; ok {
        oldStatus = ewNode.Status
    }
    rollbackAct = workflow.RollbackParams{
        Name: "updateNode",
        Params: util.M{"updateWorkflow": fmt.Sprintf(
            "update %s set %s = '%s', enter_warehouse_end='%d' where %s = %d ",
            ew.TableName(), nodeStatus.NodeCode, oldStatus, 0, ew.PkName(), ew.Pk())},
    }

   return
}
// 写单据提交日志和回滚数据
func (w *EnterWarehouseRoot)SaveSubmitLog(operatorId int64, nowWorkflowId int64, docketCode string, params map[string]interface{}, e error, rollbackAct []workflow.RollbackParams) (submitLogId string, err error){
    fmt.Println(" - - - - -写单据提交日志和回滚数据")
    var (
        rollbackByte []byte
        paramsStr []byte
    )

    submitLog := workflowEntity.NewEnterWarehouseLogEntity(w.Ctx)

    // 默认值
    if rollbackByte, err = util.JsonEncode(rollbackAct); err != nil{
        return
    }
    if paramsStr, err = util.JsonEncode(params); err != nil{
        return
    }

    data := util.M{
        "workflow_tpl_code": w.Workflow.Conf.Info.Code,
        "enter_warehouse_id": nowWorkflowId,
        "product_id":       params["product_id"],
        "docket_code":      docketCode,
        "enter_warehouse_log_params": string(paramsStr[:]),
        "enter_warehouse_log_rollback": string(rollbackByte[:]),
        "version": 1,
    }
    // 赋值
    errList := submitLog.SetMany(data)
    if len(errList)>0 {
        return "", errList[0]
    }
    // 写日志
    if err = submitLog.Create(); err!=nil{
        return
    }
    return util.Str(submitLog.Pk()), nil
}
// 执行回滚
func (w *EnterWarehouseRoot)RunRollbackAct(operatorId int64, workflowId int64, rollbackAct []workflow.RollbackParams) {
    fmt.Println(" - - - - -执行回滚")
    for _, l:= range rollbackAct{
        for _, sql := range l.Params{
            w.Query.Query(sql)
        }
    }
    return
}
// 自定义行为
func (w *EnterWarehouseRoot)CustomAct(operatorId int64, workflowId int64, docketCode string, act *workflow.ConfAct, params map[string]interface{})(nowWorkflowId int64, rollbackAct []workflow.RollbackParams, err error){
    nowWorkflowId = workflowId
    switch act.Fn {
    case "identifyAssign":
        var nodeStatus *workflow.ConfNodeStatus
        if params["to"] == "待复审" {
            fmt.Println(" - - - - -去复审")
            nodeStatus = w.MapNodeStatus["EW2020"]
        } else {
            fmt.Println(" - - - - -去初审")
            nodeStatus = w.MapNodeStatus["EW2010"]
        }
        return w.Workflow.DefaultChangeStatus(operatorId, nowWorkflowId, docketCode, nodeStatus, params)
    case "beginReturnProcess":
        fmt.Println(" - - - - -开启退货流程")
        return
        break
    case "force_to": // 强制扭转
        break;
    }
    return
}
// 观察者行为
func (w *EnterWarehouseRoot)ObAct(operatorId int64, workflowId int64, nodeStatus *workflow.ConfNodeStatus, ob *workflow.ConfOB, params map[string]interface{}) (rollbackAct []workflow.RollbackParams, err error){
    fmt.Println(" = O B =  ", ob.Name, "\t执行 ", ob.Fn)
    //switch ob.Code {
    //case "OB100":
    //    fmt.Println(" = O B =  卖家视角-状态更新", "\t执行 ", ob.Fn)
    //case "OB110":
    //    fmt.Println(" = O B =  买家视角-状态更新", "\t执行 ", ob.Fn)
    //case "OB200":
    //    fmt.Println(" = O B =  财务视角-流水更新", "\t执行 ", ob.Fn)
    //}
    return
}
