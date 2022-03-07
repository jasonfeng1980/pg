package workflow

import (
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
)

// 获取工作流操作实例
func New(config *Conf, fn ConfFn) *Workflow{
    // 1. 根据参数生成工作流类
    wf := &Workflow{
        Conf: config,
        ConfFn: fn,

        MapDocket:      make(map[string]*ConfDocket),
        MapAct:         make(map[string]*ConfAct),
        MapNode:        make(map[string]*ConfNode),
        MapNodeStatus:  make(map[string]*ConfNodeStatus),

        NodeVersion:    make(map[string]*EnterWarehouseNode),
    }
    // 2. 整理数据
        // 2.1 单据
    for _, v := range wf.Conf.Docket {
        wf.MapDocket[v.Code] = v
    }
        // 2.2 行为
    for _, v := range wf.Conf.Act {
        wf.MapAct[v.Code] = v
    }
        // 2.3.1 节点
    for _, v := range wf.Conf.Node {
        wf.MapNode[v.Code] = v
        for _, vs := range v.Status {
            // 2.3.2 节点状态
            vs.NodeCode = v.Code
            wf.MapNodeStatus[vs.Code] = vs
        }
    }

    return wf
}

///////////////////////////////////////////////////
// 需要实现
///////////////////////////////////////////////////
type ConfFn interface {
    // 1. 获取当前待处理的节点状态
    GetWorkflowToDoStatus(workflowId int64) (workflowStatusList []string, err error)
    // 2. 获取指定节点状态
    GetNodeStatus(workflowId int64, nodeCode string) (nodeStatus string, err error)
    // 3. 创建工作流,记录日志 - 统一返回格式
    CreateWorkflow(operatorId int64, params map[string]interface{}) (nowWorkflowId int64, rollbackAct []RollbackParams, err error)
    // 4. 更新节点状态,记录日志 - 统一返回格式
    ChangeNodeStatus(operatorId int64, workflowId int64, nodeStatus *ConfNodeStatus, params map[string]interface{}) (nowWorkflowId int64, rollbackAct RollbackParams, err error)
    // 5. 写单据提交日志和回滚数据
    SaveSubmitLog(operatorId int64, nowWorkflowId int64, docketCode string, params map[string]interface{}, e error, rollbackAct []RollbackParams) (submitLogId string, err error)
    // 6. 执行回滚
    RunRollbackAct(operatorId int64, workflowId int64, rollbackAct []RollbackParams)
    // 7. 自定义行为
    CustomAct(operatorId int64, workflowId int64, docketCode string, act *ConfAct, params map[string]interface{})(nowWorkflowId int64, rollbackAct []RollbackParams, err error)
}

type RollbackParams struct {
    Name   string                           // 回滚标记名称
    Params map[string]interface{}           // 回滚参数
}

// 工作流
type Workflow struct {
    Conf *Conf                          // 配置文件

    ConfFn ConfFn                      // 基本操作的方法

    MapDocket   map[string]*ConfDocket   // 单据配置map
    MapAct      map[string]*ConfAct      // 行为配置Act
    MapNode     map[string]*ConfNode     // 节点配置Node
    MapNodeStatus     map[string]*ConfNodeStatus     // 节点配置Node


    NodeVersion map[string]*EnterWarehouseNode  // 工作流中各个节点的状态
}
// 节点状态
type EnterWarehouseNode struct {
    Pk          int64
    Status      string
    Version     int
    IsEnd       int
}

//// 1. 获取当前待处理的节点状态
//type GetWorkflowToDoStatus func (workflowId int64) (workflowStatusList []string, err error)
//// 2. 获取指定节点状态和版本
//type GetNodeStatus func(workflowId int64, nodeCode string) (nodeStatus string, err error)
//// 3. 创建工作流,记录日志 - 统一返回格式
//type CreateWorkflow func(operatorId int64, workflowConf *Conf, params map[string]interface{}) (nowWorkflowId int64, rollbackAct []RollbackParams, err error)
//// 4. 更新节点状态,记录日志 - 统一返回格式
//type ChangeNodeStatus func(workflowId int64, nodeStatus *ConfNodeStatus, params map[string]interface{}) (nowWorkflowId int64, rollbackAct RollbackParams, err error)
//// 5. 更新工作流状态,记录日志 - 统一返回格式
//type ChangeWorkflowStatus func(workflowId int64, nodeStatus *ConfNodeStatus, params map[string]interface{}) (nowWorkflowId int64, rollbackAct RollbackParams, err error)
//// 6. 写单据提交日志和回滚数据
//type SaveSubmitLog func(operatorId int64, nowWorkflowId int64, docketCode string, params map[string]interface{}, e error, rollbackAct []RollbackParams) (submitLogId string, err error)
//// 7. 执行回滚
//type RunRollbackAct func(operatorId int64, workflowId int64, rollbackAct []RollbackParams)


// 提交单据
func (w *Workflow)Submit(operatorId int64, workflowId int64, docketCode string, params map[string]interface{}) (nowWorkflowId int64, submitLogId string, err error){
    var (
        confDocket *ConfDocket
        workflowStatusList []string
        ok          bool
    )
    if params==nil {
        params = make(map[string]interface{})
    }
    nowWorkflowId = workflowId

    // 1. 判断当前工作流是否支持此单据
    if confDocket, ok = w.MapDocket[docketCode]; !ok {
        err = ecode.WorkflowNotSuupotDocket.Error(w.Conf.Info.Name, docketCode)
        return
    }

    if nowWorkflowId > 0 { // 如果已经创建了工作流
        // 2. 获取当前工作流待处理的状态
        workflowStatusList, err = w.ConfFn.GetWorkflowToDoStatus(nowWorkflowId)
    }

    // 3. 判断待处理的节点状态是否支持单据
    if err = w.isAllowNodeStatus(workflowStatusList, confDocket.Allow); err != nil {
        return
    }

    // 4. 顺序执行行为，并收集回滚信息
    var rollbackActList []RollbackParams
    nowWorkflowId, rollbackActList, err = w.RunAct(operatorId, nowWorkflowId, docketCode, confDocket.Do, params)

    // 5. 如果出现错误，执行回滚指令
    if err != nil {
        // 执行回滚
        w.ConfFn.RunRollbackAct(operatorId, nowWorkflowId, rollbackActList)
        // 返回原工作流ID
        return workflowId, "", err
    }

    // 6. 记录单据操作日志和回滚数据
    submitLogId, err = w.ConfFn.SaveSubmitLog(operatorId, nowWorkflowId, docketCode, params, err, rollbackActList)

    return
}

// 执行一组行为
func (w *Workflow)RunAct(operatorId int64, workflowId int64, docketCode string, doList []*ConfDo, params map[string]interface{}) (nowWorkflowId int64, rollbackActList []RollbackParams, err error){
    nowWorkflowId = workflowId
    var rollbackAct []RollbackParams
    for _, d := range doList{
        if nowWorkflowId, rollbackAct, err = w.runOneAct(operatorId, nowWorkflowId, docketCode, d, params); err != nil{
            return
        }
        rollbackActList = append(rollbackAct, rollbackActList...)
    }
    return
}

// 执行某一个行为
func (w *Workflow)runOneAct(operatorId int64, workflowId int64, docketCode string, d *ConfDo, params map[string]interface{}) (nowWorkflowId int64, rollbackAct []RollbackParams, err error){
    nowWorkflowId = workflowId
    var nowNodeStatus string
    // 如果有Check，等待状态
    if len(d.Check)>0 {
        for _, ns := range d.Check{
            nodeStatus, ok:= w.MapNodeStatus[ns];
            if !ok{
                err = ecode.WorkflowWrongNodeStatus.Error(ns)
                return
            }
            // 获取指定节点的状态
            nowNodeStatus, err = w.ConfFn.GetNodeStatus(nowWorkflowId, nodeStatus.NodeCode)
            // 如果出错 或者 还需要等待就返回
            if err != nil || nowNodeStatus != ns {
                return
            }
        }
    }

    // 如果有Delay，延时执行

    // 如果有Plan，计划执行

    // 如果有Point，指定时间点

    // 执行方法
    return w.runOneDo(operatorId, workflowId, docketCode, d, params)
}

// 执行动作-指令
func (w *Workflow)runOneDo(operatorId int64, workflowId int64, docketCode string, d *ConfDo, params map[string]interface{})(nowWorkflowId int64, rollbackAct []RollbackParams, err error){
    nowWorkflowId = workflowId
    var (
    	nodeStatus *ConfNodeStatus
    	ok   bool
    )

    if d.Act != "" { // 执行自定义动作指令
        if d.Act == "WF_CreateWorkflow" { // 创建工作流
            nowWorkflowId, rollbackAct, err  = w.ConfFn.CreateWorkflow(operatorId, params)
        } else if act, ok := w.MapAct[d.Act]; !ok { // 不存在指定的动作
            err = ecode.WorkflowWrongActCode.Error(d.Act)
        } else { // 都存在，就正常执行
            nowWorkflowId, rollbackAct, err  = w.ConfFn.CustomAct(operatorId, nowWorkflowId, docketCode, act,  params)
        }
    } else if len(d.Status)>0 { // 执行默认动作指令
        // 循环所有的需要到达的状态
        for _, v:= range d.Status {
            var tmpRollback []RollbackParams
            if nodeStatus, ok = w.MapNodeStatus[v]; !ok {
                return
            }

            // 改变状态
            nowWorkflowId, tmpRollback, err  = w.DefaultChangeStatus(operatorId, nowWorkflowId, docketCode, nodeStatus, params)
            rollbackAct = append(tmpRollback, rollbackAct...)
            if err != nil {
                return
            }

        }
    }
    return
}

func (w *Workflow)ArriveNodeStatus(operatorId int64, workflowId int64, docketCode string, nodeStatus *ConfNodeStatus,
    params map[string]interface{}) (nowWorkflowId int64, rollbackActList []RollbackParams, err error){
    nowWorkflowId = workflowId
    // 1. 获取节点状态配置
    if len(nodeStatus.Arrive)>0 {
        if workflowId, rollbackActList, err = w.RunAct(operatorId, workflowId, docketCode, nodeStatus.Arrive, params); err != nil {
            return
        }
    }

    // 2. 执行观察者

    // 3. 执行通知事件

    return

}

// 判断工作流状态是否满足允许的节点状态
func (w *Workflow) isAllowNodeStatus(workflowStatusList []string, allowNodeStatus []string) (err error) {
    // 如果当今工作流是空状态， 并且要求也是空状态，就返回正常
    if len(allowNodeStatus) == 0 {
        return nil
    }
    // 循环所有的需要判断的节点
    for _, v := range allowNodeStatus {
        if n, ok := w.MapNode[v]; ok { // 如果允许的是节点code ， 就判断节点满足
            if w.statusInNode(workflowStatusList, n.Code) { // 如果在节点里就返回正常
                return nil
            }
        } else if ns, ok := w.MapNodeStatus[v]; ok { // 如果允许的是节点状态code， 就判断节点状态满足
            if util.ListHave(workflowStatusList, ns.Code) { // 如果工作流状态满足节点状态，就返回正常
                return nil
            }
            // 判断所在节点状态是否满足 -- @todo 使用率底，以后再做

        } else { // 都不在就是配置错误
            return ecode.WorkflowWrongNodeStatus.Error(v)
        }
    }
    // 没有满足的， 返回匹配不了
    return ecode.WorkflowStatusNotAllowNodeStatus.Error(workflowStatusList, allowNodeStatus)
}

// 一批节点状态，是否有指定节点
func (w *Workflow) statusInNode(workflowStatusList []string, nodeCode string) bool{
    for _, v := range  workflowStatusList {
        if ns, ok := w.MapNodeStatus[v]; ok{ // 配置的节点状态存在
            if ns.NodeCode == nodeCode { // 如果节点匹配
                return true
            }
        }
    }
    return false
}

// 默认的节点状态改变
func (w *Workflow)DefaultChangeStatus(operatorId int64, workflowId int64, docketCode string, nodeStatus *ConfNodeStatus,
    params map[string]interface{}) (nowWorkflowId int64, rollbackActList []RollbackParams, err error){
    var rollback RollbackParams
    // 默认和传来的工作流ID一致
    nowWorkflowId = workflowId
    // 更新节点状态
    nowWorkflowId, rollback, err = w.ConfFn.ChangeNodeStatus(operatorId, nowWorkflowId, nodeStatus, params)
    if err != nil {
        return
    }
    if rollback.Name != "" {
        rollbackActList = append(rollbackActList, rollback)
    }


    //// 如果节点状态不要求隐藏状态，就更新工作流状态
    //if !nodeStatus.Hide {
    //    nowWorkflowId, rollback, err = w.ConfFn.ChangeWorkflowStatus(nowWorkflowId, nodeStatus, params)
    //    rollbackActList = append([]RollbackParams{rollback}, rollbackActList...)
    //}

    // 如果节点，有默认调整 就更新节点状态
    if len(nodeStatus.To)>0 {
        for _, v := range nodeStatus.To{
            ns := w.MapNodeStatus[v]
            // 更新节点状态
            nowWorkflowId, rollback, err = w.ConfFn.ChangeNodeStatus(operatorId, nowWorkflowId, ns, params)
            if err != nil {
                return
            }
            if rollback.Name != "" {
                rollbackActList = append(rollbackActList, rollback)
            }

            // 如果节点配置了到达行为
            if len(ns.Arrive) >0 {
                var tmpRollback []RollbackParams
                nowWorkflowId, tmpRollback , err = w.ArriveNodeStatus(operatorId, nowWorkflowId, docketCode, ns, params)
                if err != nil {
                    return
                }
                rollbackActList = append(tmpRollback, rollbackActList...)
            }
        }
    }


    return
}