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

func TestService(t *testing.T) {
	def := model.ProcessDefine{}
	err := json.Unmarshal([]byte(postData), &def)
	db, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	_ = service.New(dto.New(db))
	//s.StartProcessInstanceById()
	d := dto.New(db)
	save, err := d.ProcessDef.Save(&def)
	if err != nil {
		return
	}
	fmt.Println(save)
}
