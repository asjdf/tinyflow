package service

import (
	"fmt"
	"tinyflow/workflow/model"
)

// 流程实例管理

func (s *Service) StartProcessInstanceById(processId uint, input model.ProcessInputVariable) {
	processDef, err := s.dto.ProcessDef.Get(processId)
	if err != nil {
		return
	}
	fmt.Println(processDef)
}
