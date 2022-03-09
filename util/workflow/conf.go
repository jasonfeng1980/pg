package workflow

// 工作流的配置
type Conf struct {
    Info ConfInfo           `json:"info"`    // 基本信息
    Node []*ConfNode        `json:"node"`    // 所有节点配置
    Act  []*ConfAct         `json:"act"`     // 所有行为配置
    Docket []*ConfDocket    `json:"docket"`  // 单据配置
    OB   []*ConfOB          `json:"ob"`      // 观察者
}
// 基本信息
type ConfInfo struct {
    Code    string  `json:"code"`       // 编码
    Name    string  `json:"name"`       // 名称
    Desc    string  `json:"desc"`       // 备注
}
// 节点配置
type ConfNode struct {
    Code        string  `json:"code"`           // 编码
    Name        string  `json:"name"`           // 名称
    Type        string  `json:"type"`           // 类型:node节点;gateway网关;
    Status      []*ConfNodeStatus `json:"status"` // 节点状态
}

// 节点状态
type ConfNodeStatus struct {
    NodeCode    string `json:"node_code"`       // 对应的节点编码
    Code        string  `json:"code"`           // 状态编码
    Name        string  `json:"name"`           // 状态名称
    End         int    `json:"end"`             // 是否是节点的完结状态；  默认0 不是; 1是
    Finish      int    `json:"finish"`          // 工作流是否完结
    Hide        bool    `json:"hide"`           // true 不修改工作流状态；默认false 修改工作流状态
    To          []string `json:"to"`            // Arrive的简写  ["status":"nodeStatus", ...]
    Arrive      []*ConfDo `json:"arrive"`       // 到达状态时，执行行为指令
    OB          []*ConfOB `json:"ob"`           // 节点状态的观察者
}

// 行为配置
type ConfAct struct {
    Code      string     `json:"code"`
    Name      string    `json:"name"`
    Status    []string  `json:"status"`
    Fn        string    `json:"fn"`
    Args      map[string]interface{} `json:"args"`
}
// 单据模板
type ConfDocket struct {
    Code    string      `json:"code"`       // 单据模板code
    Name    string      `json:"name"`       // 模板名称
    Allow   []string    `json:"allow"`      // 允许的当前节点或者状态
    Do      []*ConfDo   `json:"do"`         // 执行
}
// 执行的行为
type ConfDo struct {
    Check     []string  `json:"check"`      // 判断等待节点状态是否都完成，再执行方法
    Delay     int64     `json:"delay"`      // 延迟多少秒执行方法
    Plan      []string  `json:"plan"`       // 秒(0-59) 分(0-59) 时(0-23) 一个月中的第几天(1-31)  月(1-12) 星期几（0-6）
    Point     []int64   `json:"point"`      // 指定时间点执行  []时间戳

    Status    []string  `json:"status"`      // 默认行为：修改的状态
    Act       string    `json:"act"`        // 获取指定行为
}

// 行为配置
type ConfOB struct {
    Code      string     `json:"code"`             // 编码
    Name      string    `json:"name"`               // 名称
    Allow    []string  `json:"allow"`               // 监控的节点|节点状态
    Fn        string    `json:"fn"`                 // 绑定的方法标识
    Args      map[string]interface{} `json:"args"`  // 固定的参数
}


