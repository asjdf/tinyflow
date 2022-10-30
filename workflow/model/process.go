package model

import (
	"gorm.io/datatypes"
	"gorm.io/plugin/optimisticlock"
	"time"
)

// ProcessDefine 流程定义
type ProcessDefine struct {
	Model

	Name string `json:"name"`

	Nodes *Node `json:"resource"` //todo: 这个变量名看看前端能不能换掉 有点阴间

	NameSpace string `json:"nameSpace"` // 多域

	Version optimisticlock.Version
}

// ProcessDefineHistory 历史流程定义
type ProcessDefineHistory struct {
	LogID uint `gorm:"primary_key;autoIncrement:true"` // todo: HackFix 解决主键冲突问题
	ProcessDefine
}

// ProcessInstance 流程实例
type ProcessInstance struct {
	Model
	// 流程定义ID
	ProcDefID uint `json:"procDefId"`
	// 流程定义名
	ProcDefName string `json:"procDefName"`
	// title 标题
	Title string `json:"title"` //暂时无用
	// 多域
	NameSpace string `json:"nameSpace"`
	// 当前节点
	NodeID string `json:"nodeID"`
	// 审批人
	Candidate string `json:"candidate"`
	// 当前任务
	TaskID      uint          `json:"taskID"`
	StartTime   time.Time     `json:"startTime"`
	EndTime     time.Time     `json:"endTime"`
	Duration    time.Duration `json:"duration"`
	StartUserID string        `json:"startUserId"`
	IsFinished  bool          `gorm:"default:false" json:"isFinished"`
	// 执行流
	NodeInfos datatypes.JSON `json:"nodeInfos"` // []NodeInfo

	Version optimisticlock.Version // 暂时乐观锁吧
}
