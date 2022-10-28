package service

import (
	"encoding/json"
	"time"
	"tinyflow/utils"
	"tinyflow/workflow/model"
)

// 流程实例管理

func (s *Service) StartProcessInstanceById(processId uint, userId string, input model.ProcessInputVariable) error {
	// 获取流程信息
	processDef, err := s.dto.ProcessDef.Get(processId)
	if err != nil {
		return err
	}

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

	// 下面是一坨屎 有空来处理
	execFlows := make([]model.NodeInfo, 0, ExecFlow.Len())
	utils.List2Array(ExecFlow, &execFlows)
	processInstance.NodeInfos, _ = json.Marshal(execFlows)

	if err := s.dto.ProcessInstance.Save(processInstance, tx); err != nil {
		tx.Rollback()
		return err
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

	if execFlows[0].ActType == "and" {
		task.UnCompleteNum = execFlows[0].MemberCount
		task.MemberCount = execFlows[0].MemberCount
	}
	if err := s.dto.Task.Save(task, tx); err != nil {
		tx.Rollback()
		return err
	}

	//开始工作流流转
	//这里还没写好
	if err := s.MoveStage(processInstance.ID, "启动流程", task.ID, 0, true, tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
