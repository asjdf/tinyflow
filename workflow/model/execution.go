package model

// Execution 流程实例（执行流）表
// ProcInstID 流程实例ID
// ProcDefID 流程定义数据的ID
type Execution struct {
	Model
	Rev        int `json:"rev"`
	ProcInstID int `json:"procInstID"`
	ProcDefID  int `json:"procDefID"` // 流程定义ID
	// NodeInfos 执行流经过的所有节点
	NodeInfos string `gorm:"size:4000" json:"nodeInfos"`
	IsActive  int8   `json:"isActive"`
	StartTime string `json:"startTime"`
}
