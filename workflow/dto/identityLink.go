package dto

import (
	"gorm.io/gorm"
	"tinyflow/workflow/model"
)

type IdentityLink struct {
	db *gorm.DB
}

func NewIdentityLink(db *gorm.DB) IdentityLink {
	_ = db.AutoMigrate(&model.IdentityLink{})
	return IdentityLink{db: db}
}

// Save 如果不传事务则直接存
func (i IdentityLink) Save(data *model.IdentityLink, tx ...*gorm.DB) error {
	if len(tx) != 0 {
		return tx[0].Save(data).Error
	}
	return i.db.Save(data).Error
}

func (i IdentityLink) Del(data *model.IdentityLink, tx ...*gorm.DB) error {
	if len(tx) != 0 {
		return tx[0].Where(data).Delete(&model.IdentityLink{}).Error
	}
	return i.db.Where(data).Delete(&model.IdentityLink{}).Error
}
