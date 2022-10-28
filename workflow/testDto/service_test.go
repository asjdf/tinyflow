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

func TestDto(t *testing.T) {
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

func TestService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	s := service.New(dto.New(db))
	s.StartProcessInstanceById(1, "1", &map[string]string{"DDHolidayField-J2BWEN12__options": "年假", "DDHolidayField-J2BWEN12__duration": "8"})

}
