package model

// Execution 流程实例（执行流）表
// ProcInstID 流程实例ID
// ProcDefID 流程定义数据的ID
// todo: 评估是否可以直接合并到实例 目前看起来可合并
type Execution struct {
	Model
	Rev        int  `json:"rev"`
	ProcInstID uint `json:"procInstID"`
	ProcDefID  uint `json:"procDefID"` // 流程定义ID
	// NodeInfos 执行流经过的所有节点
	NodeInfos []NodeInfo `json:"nodeInfos"`
	IsActive  int8       `json:"isActive"`
	StartTime string     `json:"startTime"`
}
