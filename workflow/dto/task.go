package dto

import (
	"errors"
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

func (t Task) Get(id uint, tx ...*gorm.DB) (*model.Task, error) {
	ins := &model.Task{}
	if len(tx) != 0 {
		err := tx[0].First(ins, id).Error
		return ins, err
	}
	err := t.db.First(ins, id).Error
	return ins, err
}

func (t Task) Find(where *model.Task) ([]*model.Task, error) {
	list := make([]*model.Task, 0)
	err := t.db.Where(where).Find(&list).Error
	return list, err
}

func (t Task) Update(instance *model.Task, tx ...*gorm.DB) error {
	var result *gorm.DB
	if len(tx) != 0 {
		result = tx[0].Model(instance).Updates(instance)
	} else {
		result = t.db.Model(instance).Updates(instance)
	}
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected != 1 {
		return errors.New("update locked by optimistic lock")
	}
	return nil
}
