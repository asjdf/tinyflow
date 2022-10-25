package dto

import (
	"errors"
	"gorm.io/gorm"
	"tinyflow/workflow/model"
)

// ProcessDef 工作流定义
type ProcessDef struct {
	db *gorm.DB
}

func NewProcessDef(db *gorm.DB) ProcessDef {
	_ = db.AutoMigrate(&model.ProcessDefine{}, &model.ProcessDefineHistory{})
	return ProcessDef{db: db}
}

// Save 保存工作流定义 如果存在老版本工作流，则更新并保存定义的历史版本
func (d ProcessDef) Save(def *model.ProcessDefine) (*model.ProcessDefine, error) {
	// 检查是否存在老版本
	oldDef := &model.ProcessDefine{}
	err := d.db.Where("name_space = ? and name = ?", def.NameSpace, def.Name).First(oldDef).Error
	if err == gorm.ErrRecordNotFound {
		// 不存在老版本直接存
		newDef := &model.ProcessDefine{Name: def.Name, Nodes: def.Nodes, NameSpace: def.NameSpace}
		err = d.db.Save(newDef).Error
		return newDef, err
	} else if err != nil {
		return nil, err
	}

	// 存在老版本则把老版本移到历史记录表并且更新
	tx := d.db.Begin()
	if err := tx.Create(&model.ProcessDefineHistory{ProcessDefine: *oldDef}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if result := tx.Model(oldDef).Updates(def); result.Error != nil || result.RowsAffected != 1 {
		tx.Rollback()
		if result.RowsAffected != 1 {
			return nil, errors.New("乐观锁遇到冲突")
		}
		return nil, err
	}
	err = tx.Commit().Error
	return def, err
}

func (d ProcessDef) Get(id uint) (def *model.ProcessDefine, err error) {
	err = d.db.First(def, id).Error
	return
}