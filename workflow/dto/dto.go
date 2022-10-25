package dto

import (
	"gorm.io/gorm"
)

type DTO struct {
	db         *gorm.DB
	ProcessDef ProcessDef
}

func New(db *gorm.DB) DTO {
	return DTO{db: db,
		ProcessDef: NewProcessDef(db),
	}
}
