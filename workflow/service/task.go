package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math"
	"sort"
	"time"
	"tinyflow/workflow/model"
	"tinyflow/workflow/model/identityType"
)

// 流程流转 向前向后

// PassTask 通过审批
func (s *Service) PassTask(taskID uint, userID, nameSpace, comment, candidate string, pass bool) error {
	tx := s.dto.Begin()
	//更新任务
	task, err := s.updateTaskWhenComplete(taskID, userID, pass, tx)
	if err != nil {
		return err
	}
	// 如果是会签
	if task.ActType == "and" {
		// 判断用户是否已参与过审批（存在会签的情况）
		yes, err := s.ifParticipantTask(taskID, userID)
		if err != nil {
			tx.Rollback()
			return err
		}
		if yes {
			tx.Rollback()
			return errors.New("已审批过")
		}
	}

	// 查看任务的未审批人数是否为0，不为0就不流转
	if task.UnCompleteNum > 0 && pass == true { // 默认是全部通过
		// 添加参与人
		err := s.AddParticipantTx(userID, nameSpace, comment, task.ID, task.ProcInstID, task.Step, tx)
		if err != nil {
			tx.Rollback()
			return err
		}
		return nil
	}
	// 流转到下一流程
	err = s.MoveStage(task.ProcInstID, userID, comment, candidate, task.ID, task.Step, pass, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// WithdrawTask 撤回审批 todo: 上锁 另外检查流程有没有问题
func (s *Service) WithdrawTask(taskID uint, userID string, nameSpace string, comment string) error {
	currentTask, err := s.dto.Task.Get(taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("审批任务不存在")
		}
		return err
	}
	if currentTask.Step == 0 {
		return errors.New("开始位置无法撤回")
	}

	// 找前一审批任务
	passedTasks, err := s.dto.Task.Find(&model.Task{ProcInstID: currentTask.ProcInstID, IsFinished: true})
	if err != nil {
		return err
	} else if len(passedTasks) == 0 {
		return errors.New("找不到前一审批节点，撤回失败")
	}

	sort.Slice(passedTasks, func(i, j int) bool {
		return passedTasks[i].ClaimTime.After(passedTasks[j].ClaimTime)
	})

	latestPassedTask := passedTasks[0]
	if latestPassedTask.Assignee != userID {
		return errors.New("只能撤回本人审批过的任务")
	}
	if currentTask.IsFinished {
		return errors.New("已经审批结束已无法撤回")
	}
	if currentTask.UnCompleteNum != currentTask.MemberCount {
		return errors.New("已经有其他人审批过无法撤回")
	}
	sub := currentTask.Step - latestPassedTask.Step
	if math.Abs(float64(sub)) != 1 {
		return errors.New("只能撤回相邻的任务")
	}
	var pass = false
	if sub < 0 {
		pass = true
	}
	tx := s.dto.Begin()
	// 更新当前的任务
	currentTask.IsFinished = true

	err = s.dto.Task.Update(currentTask, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	// 撤回
	err = s.MoveStage(currentTask.ProcInstID, userID, comment, "", taskID, currentTask.Step, pass, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (s *Service) updateTaskWhenComplete(taskID uint, userID string, pass bool, tx *gorm.DB) (*model.Task, error) {
	// 查询任务
	task, err := s.dto.Task.Get(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("任务【" + fmt.Sprintf("%d", task.ID) + "】不存在")
	}
	// 判断是否已经结束
	if task.IsFinished == true {
		if task.NodeID == "结束" {
			return nil, errors.New("流程已经结束")
		}
		return nil, errors.New("任务【" + fmt.Sprintf("%d", taskID) + "】已经被审批过了！！")
	}
	// 设置处理人和处理时间
	task.Assignee = userID
	task.ClaimTime = time.Now()

	// 会签 （默认全部通过才结束），只要存在一个不通过，就结束，然后流转到上一步
	if pass {
		task.AgreeNum++
	} else {
		task.IsFinished = true
	}
	// 未审批人数减一
	task.UnCompleteNum--
	// 判断是否结束
	if task.UnCompleteNum == 0 {
		task.IsFinished = true
	}
	err = s.dto.Task.Update(task, tx)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// ifParticipantTask 检查是否已参与单个节点审批
func (s *Service) ifParticipantTask(taskID uint, userID string) (bool, error) {
	list, err := s.dto.IdentityLink.Find(&model.IdentityLink{TaskID: taskID, UserID: userID})
	if err != nil {
		return false, err
	}
	return len(list) != 0, err
}

// MoveStage 负责流程流转
// candidate 可为空 是为了指定下一步审批人 可为空 还没做好其他东西
func (s *Service) MoveStage(instanceId uint, userID, comment, candidate string, taskID uint, step int, pass bool, tx *gorm.DB) (err error) {
	procIns, err := s.dto.ProcessInstance.Get(instanceId, tx) // 真的可以考虑合并实例和执行流
	if err != nil {
		return err
	}
	procInstID := procIns.ID
	nameSpace := procIns.NameSpace
	nodeInfos := make([]model.NodeInfo, 0)
	_ = json.Unmarshal(procIns.NodeInfos, &nodeInfos)

	// 添加上一步的参与人 也就是当前审批的这一步
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

	if candidate != "" {
		nodeInfos[step].Aprover = candidate
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
		return s.MoveStage(instanceId, userID, comment, candidate, taskID, step, pass, tx) //todo: 这里应该可以优化 不然查太多次了
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
