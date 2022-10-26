package dto

import (
	"gorm.io/gorm"
	"tinyflow/workflow/model"
)

type Execution struct {
	db *gorm.DB
}

func NewExecution(db *gorm.DB) Execution {
	_ = db.AutoMigrate(&model.Execution{})
	return Execution{db: db}
}

// Save 如果不传事务则直接存
func (e Execution) Save(exec *model.Execution, tx ...*gorm.DB) error {
	if len(tx) != 0 {
		return tx[0].Save(exec).Error
	}
	return e.db.Save(exec).Error
}
