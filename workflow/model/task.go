package model

import "time"

// Task 单个审批节点产生的任务
type Task struct {
	Model
	// 当前执行流所在的节点
	NodeID string `json:"nodeId"`
	Step   int    `json:"step"`
	// 流程实例id
	ProcInstID uint `json:"procInstID"`
	// Assignee 节点归属人 可考虑去除 根据ProcInst信息获取
	Assignee  string    `json:"assignee"`
	ClaimTime time.Time `json:"claimTime"`
	// 还未审批的用户数，等于0代表会签已经全部审批结束，默认值为1
	MemberCount   int8 `json:"memberCount" gorm:"default:1"`
	UnCompleteNum int8 `json:"unCompleteNum" gorm:"default:1"`
	//审批通过数
	AgreeNum int8 `json:"agreeNum"`
	// and 为会签，or为或签，默认为or
	ActType    string `json:"actType" gorm:"default:'or'"`
	IsFinished bool   `gorm:"default:false" json:"isFinished"`
}
