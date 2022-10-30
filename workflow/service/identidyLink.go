package service

import (
	"gorm.io/gorm"
	"tinyflow/workflow/model"
	"tinyflow/workflow/model/identityType"
)

// AddParticipantTx 添加参与关系
func (s *Service) AddParticipantTx(userID, nameSpace, comment string, taskID, procInstID uint, step int, tx *gorm.DB) error {
	i := &model.IdentityLink{
		Type:       identityType.ToStr[identityType.PARTICIPANT],
		UserID:     userID,
		ProcInstID: procInstID,
		Step:       step,
		TaskID:     taskID,
		Comment:    comment,
		NameSpace:  nameSpace,
	}
	return s.dto.IdentityLink.Save(i, tx)
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
