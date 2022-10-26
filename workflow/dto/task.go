package dto

import (
	"gorm.io/gorm"
	"tinyflow/workflow/model"
)

type Task struct {
	db *gorm.DB
}

func NewTask(db *gorm.DB) Task {
	_ = db.AutoMigrate(&model.Task{})
	return Task{db: db}
}

// Save 如果不传事务则直接存
func (t Task) Save(task *model.Task, tx ...*gorm.DB) error {
	if len(tx) != 0 {
		return tx[0].Save(task).Error
	}
	return t.db.Save(task).Error
}
