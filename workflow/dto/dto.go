package dto

import (
	"gorm.io/gorm"
)

type DTO struct {
	db              *gorm.DB
	ProcessDef      ProcessDef
	ProcessInstance ProcessInstance
	Task            Task
	IdentityLink    IdentityLink
}

func New(db *gorm.DB) DTO {
	return DTO{db: db,
		ProcessDef:      NewProcessDef(db),
		ProcessInstance: NewProcessInstance(db),
		Task:            NewTask(db),
		IdentityLink:    NewIdentityLink(db),
	}
}

// Begin 开事务
func (d DTO) Begin() *gorm.DB {
	return d.db.Begin()
}
