package testDto

import (
	"encoding/json"
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"testing"
	"tinyflow/workflow/dto"
	"tinyflow/workflow/model"
	"tinyflow/workflow/service"
)

func TestDTOAndParseProcessDef(t *testing.T) {
	def := model.ProcessDefine{}
	err := json.Unmarshal([]byte(postData), &def)
	db, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	d := dto.New(db)
	save, err := d.ProcessDef.Save(&def)
	if err != nil {
		return
	}
	fmt.Println(save)
}

func TestStartProcess(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	s := service.New(dto.New(db))
	s.StartProcessInstanceById(1, "1", &map[string]string{"DDHolidayField-J2BWEN12__options": "年假", "DDHolidayField-J2BWEN12__duration": "8"})
}

func TestPassTask(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	s := service.New(dto.New(db))
	err = s.PassTask(2, "abc", "A公司", "通过你的申请", "", true)
	if err != nil {
		return
	}
	return
}

func TestWithdrawTask(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	s := service.New(dto.New(db))
	err = s.WithdrawTask(2, "abc", "A公司", "驳回你的申请")
	if err != nil {
		return
	}
	return
}
