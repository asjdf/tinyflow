package model

// IdentityLink 用户组同任务的关系
type IdentityLink struct {
	Model
	Group      string `json:"group,omitempty"`
	Type       string `json:"type,omitempty"`
	UserID     string `json:"userid,omitempty"`
	TaskID     uint   `json:"taskID,omitempty"`
	Step       int    `json:"step"`
	ProcInstID uint   `json:"procInstID,omitempty"`
	NameSpace  string `json:"nameSpace,omitempty"`
	Comment    string `json:"comment,omitempty"`
}
