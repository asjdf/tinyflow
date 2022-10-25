package model

import "gorm.io/plugin/optimisticlock"

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
