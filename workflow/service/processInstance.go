package service

import (
	"fmt"
	"time"
	"tinyflow/utils"
	"tinyflow/workflow/model"
)

// 流程实例管理

func (s *Service) StartProcessInstanceById(processId uint, nameSpace string, userId string, input model.ProcessInputVariable) error {
	// 获取流程信息
	processDef, err := s.dto.ProcessDef.Get(processId)
	if err != nil {
		return err
	}
	fmt.Println(processDef)

	tx := s.dto.Begin()
	// 创建流程实例
	processInstance := &model.ProcessInstance{
		ProcDefID:   processDef.ID,
		ProcDefName: processDef.Name,
		Title:       "",
		NameSpace:   processDef.NameSpace,
		StartTime:   time.Now(),
		StartUserID: userId,
	}
	if err := s.dto.ProcessInstance.Save(processInstance, tx); err != nil {
		tx.Rollback()
		return err
	}

	exec := &model.Execution{
		ProcDefID:  processDef.ID,
		ProcInstID: processInstance.ID,
	}
	task := &model.Task{
		NodeID:        "开始",
		ProcInstID:    processInstance.ID,
		Assignee:      userId,
		IsFinished:    true,
		ClaimTime:     time.Now(),
		Step:          0, // 开始
		MemberCount:   1, // 直接过
		UnCompleteNum: 0,
		ActType:       "or",
		AgreeNum:      1,
	}

	ExecFlow, err := processDef.Nodes.GenExecFlow(input)
	if err != nil {
		return err
	}
	ExecFlow.PushFront(model.NodeInfo{
		NodeID:  "开始",
		Type:    model.NodeTypes[model.START],
		Aprover: userId,
	})
	ExecFlow.PushBack(model.NodeInfo{NodeID: "结束"})
	exec.NodeInfos = interface{}(utils.List2Array(ExecFlow)).([]model.NodeInfo)
	if err := s.dto.Execution.Save(exec, tx); err != nil {
		tx.Rollback()
		return err
	}

	if exec.NodeInfos[0].ActType == "and" {
		task.UnCompleteNum = exec.NodeInfos[0].MemberCount
		task.MemberCount = exec.NodeInfos[0].MemberCount
	}
	if err := s.dto.Task.Save(task, tx); err != nil {
		tx.Rollback()
		return err
	}

	//开始工作流流转
	//这里还没写好
	if err := s.MoveStage(exec, "启动流程", task.ID, 0, true, tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
