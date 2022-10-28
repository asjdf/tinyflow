package service

import (
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"time"
	"tinyflow/workflow/model"
	"tinyflow/workflow/model/identityType"
)

// 流程流转 向前向后

// MoveStage todo: 可以考虑隐函数
func (s *Service) MoveStage(instanceId uint, comment string, taskID uint, step int, pass bool, tx *gorm.DB) (err error) {
	procIns, err := s.dto.ProcessInstance.Get(instanceId, tx) // 真的可以考虑合并实例和执行流
	if err != nil {
		return err
	}
	procInstID := procIns.ID
	userID := procIns.StartUserID
	nameSpace := procIns.NameSpace
	nodeInfos := make([]model.NodeInfo, 0)
	_ = json.Unmarshal(procIns.NodeInfos, &nodeInfos)

	// 添加上一步的参与人
	if err := s.dto.IdentityLink.Save(&model.IdentityLink{
		Type:       identityType.ToStr[identityType.PARTICIPANT],
		UserID:     userID,
		NameSpace:  nameSpace,
		Comment:    comment,
		TaskID:     taskID,
		ProcInstID: procInstID,
		Step:       step,
	}, tx); err != nil {
		return err
	}
	if pass {
		step++
		if step-1 > len(nodeInfos) {
			return errors.New("已经结束无法流转到下一个节点")
		}
	} else {
		step--
		if step < 0 {
			return errors.New("处于开始位置，无法回退到上一个节点")
		}
	}

	// 判断下一流程： 如果是审批人是：抄送人
	// fmt.Printf("下一审批人类型：%s\n", nodeInfos[step].AproverType)
	if nodeInfos[step].AproverType == model.NodeTypes[model.NOTIFIER] {
		// 生成新的任务
		var task = model.Task{
			NodeID:     model.NodeTypes[model.NOTIFIER],
			Step:       step,
			ProcInstID: procInstID,
			IsFinished: true,
		}
		if err := s.dto.Task.Save(&task, tx); err != nil {
			return err
		}
		// 添加抄送人
		// todo: 应该检查一下是否已经抄送
		if err := s.dto.IdentityLink.Save(&model.IdentityLink{
			Type:       identityType.ToStr[identityType.NOTIFIER],
			Group:      nodeInfos[step].Aprover,
			NameSpace:  nameSpace,
			ProcInstID: procInstID,
			Step:       step,
		}, tx); err != nil {
			return err
		}
		return s.MoveStage(instanceId, comment, taskID, step, pass, tx) //todo: 这里应该可以优化 不然查太多次了
	}
	if pass {
		return s.MoveToNextStage(nodeInfos, nameSpace, procInstID, step, tx)
	}
	return s.MoveToPrevStage(nodeInfos, nameSpace, procInstID, step, tx)
}

// MoveToNextStage 通过
func (s *Service) MoveToNextStage(nodeInfos []model.NodeInfo, nameSpace string, procInstID uint, step int, tx *gorm.DB) error {
	task := &model.Task{ // 新任务
		NodeID:        nodeInfos[step].NodeID,
		Step:          step,
		ProcInstID:    procInstID,
		MemberCount:   nodeInfos[step].MemberCount,
		UnCompleteNum: nodeInfos[step].MemberCount,
		ActType:       nodeInfos[step].ActType,
	}
	procInst := &model.ProcessInstance{ // 流程实例要更新的字段
		NodeID:    nodeInfos[step].NodeID,
		Candidate: nodeInfos[step].Aprover,
	}
	procInst.ID = procInstID
	if (step + 1) != len(nodeInfos) { // 下一步不是【结束】
		// 生成新的任务
		if err := s.dto.Task.Save(task, tx); err != nil {
			return err
		}
		// 添加candidate group
		err := s.AddCandidateGroupTx(nodeInfos[step].Aprover, nameSpace, step, task.ID, procInstID, tx)
		if err != nil {
			return err
		}
		// 更新流程实例
		procInst.TaskID = task.ID
		if err := s.dto.ProcessInstance.Update(procInst, tx); err != nil {
			return err
		}
	} else { // 最后一步直接结束
		// 生成新的任务
		task.IsFinished = true
		task.ClaimTime = time.Now()
		if err := s.dto.Task.Save(task, tx); err != nil {
			return err
		}
		// 删除候选用户组
		if err := s.dto.IdentityLink.Del(
			&model.IdentityLink{
				ProcInstID: procInstID,
				Type:       identityType.ToStr[identityType.CANDIDATE],
			},
			tx); err != nil {
			return err
		}
		// 更新流程实例
		procInst.TaskID = task.ID
		procInst.EndTime = time.Now()
		procInst.IsFinished = true
		procInst.Candidate = "审批结束"
		if err := s.dto.ProcessInstance.Update(procInst, tx); err != nil {
			return err
		}
	}
	return nil
}

// MoveToPrevStage 驳回
func (s *Service) MoveToPrevStage(nodeInfos []model.NodeInfo, nameSpace string, procInstID uint, step int, tx *gorm.DB) error {
	// 生成新的任务
	task := &model.Task{ // 新任务
		NodeID:        nodeInfos[step].NodeID,
		Step:          step,
		ProcInstID:    procInstID,
		MemberCount:   nodeInfos[step].MemberCount,
		UnCompleteNum: nodeInfos[step].MemberCount,
		ActType:       nodeInfos[step].ActType,
	}
	if err := s.dto.Task.Save(task, tx); err != nil {
		return err
	}
	procInst := &model.ProcessInstance{ // 流程实例要更新的字段
		NodeID:    nodeInfos[step].NodeID,
		Candidate: nodeInfos[step].Aprover,
	}
	procInst.ID = procInstID
	if err := s.dto.ProcessInstance.Update(procInst, tx); err != nil {
		return err
	}
	if step == 0 { // 流程回到起始位置，注意起始位置为0,
		if err := s.AddCandidateUserTx(nodeInfos[step].Aprover, nameSpace, step, task.ID, procInstID, tx); err != nil {
			return err
		}
		return nil
	}
	// 添加candidate group
	if err := s.AddCandidateGroupTx(nodeInfos[step].Aprover, nameSpace, step, task.ID, procInstID, tx); err != nil {
		return err
	}
	return nil
}

// AddCandidateUserTx 添加候选用户 两个差别不大 看看能不能合并
func (s *Service) AddCandidateUserTx(userID, nameSpace string, step int, taskID, procInstID uint, tx *gorm.DB) error {
	if err := s.dto.IdentityLink.Del(
		&model.IdentityLink{
			ProcInstID: procInstID,
			Type:       identityType.ToStr[identityType.CANDIDATE],
		}, tx); err != nil {
		return err
	}
	i := &model.IdentityLink{
		UserID:     userID,
		Type:       identityType.ToStr[identityType.CANDIDATE],
		TaskID:     taskID,
		Step:       step,
		ProcInstID: procInstID,
		NameSpace:  nameSpace,
	}
	return s.dto.IdentityLink.Save(i, tx)
}

// AddCandidateGroupTx 添加候选用户组
func (s *Service) AddCandidateGroupTx(group, nameSpace string, step int, taskID, procInstID uint, tx *gorm.DB) error {
	if err := s.dto.IdentityLink.Del(
		&model.IdentityLink{
			ProcInstID: procInstID,
			Type:       identityType.ToStr[identityType.CANDIDATE],
		}, tx); err != nil {
		return err
	}
	i := &model.IdentityLink{
		Group:      group,
		Type:       identityType.ToStr[identityType.CANDIDATE],
		TaskID:     taskID,
		Step:       step,
		ProcInstID: procInstID,
		NameSpace:  nameSpace,
	}
	return s.dto.IdentityLink.Save(i, tx)
}
