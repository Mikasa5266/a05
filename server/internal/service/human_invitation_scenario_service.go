package service

import (
	"fmt"
	"strings"

	"your-project/internal/model"
)

const (
	groupInvitationMinParticipants = 2
	groupInvitationMaxParticipants = 5
	groupInvitationMinThreshold    = 2
)

func normalizeScenarioType(raw string) string {
	scenario := strings.ToLower(strings.TrimSpace(raw))
	if scenario == model.InvitationScenarioGroup {
		return model.InvitationScenarioGroup
	}
	return model.InvitationScenarioSingle
}

func normalizeGroupConfig(targetParticipants, startThreshold, participantCount int) (int, int) {
	target := targetParticipants
	if target < groupInvitationMinParticipants {
		target = groupInvitationMinParticipants
	}
	if target > groupInvitationMaxParticipants {
		target = groupInvitationMaxParticipants
	}
	if participantCount > target {
		target = participantCount
	}

	threshold := startThreshold
	if threshold < groupInvitationMinThreshold {
		threshold = groupInvitationMinThreshold
	}
	if threshold > target {
		threshold = target
	}

	return target, threshold
}

// ConfigureHumanInvitationScenario applies explicit scenario settings after invitation creation.
func ConfigureHumanInvitationScenario(invitationID uint, scenarioType string, targetParticipants, startThreshold int) (*model.HumanInterviewInvitation, error) {
	if invitationID == 0 {
		return nil, fmt.Errorf("invitation_id is required")
	}

	svc := NewInterviewService()
	invitation, err := svc.interviewRepo.GetInvitationByID(invitationID)
	if err != nil {
		return nil, fmt.Errorf("邀请不存在")
	}

	scenario := normalizeScenarioType(scenarioType)
	if scenario == model.InvitationScenarioSingle {
		invitation.ScenarioType = model.InvitationScenarioSingle
		invitation.TargetParticipants = 2
		invitation.StartThreshold = 2
	} else {
		participants, listErr := svc.interviewRepo.ListInvitationParticipants(invitation.ID)
		if listErr != nil {
			participants = []model.HumanInterviewInvitationParticipant{}
		}
		participantCount := len(participants)
		if participantCount < groupInvitationMinParticipants {
			participantCount = groupInvitationMinParticipants
		}

		normalizedTarget, normalizedThreshold := normalizeGroupConfig(targetParticipants, startThreshold, participantCount)
		invitation.ScenarioType = model.InvitationScenarioGroup
		invitation.TargetParticipants = normalizedTarget
		invitation.StartThreshold = normalizedThreshold
	}

	if err := svc.interviewRepo.UpdateInvitation(invitation); err != nil {
		return nil, fmt.Errorf("更新邀请场景失败: %w", err)
	}

	updated, err := svc.interviewRepo.GetInvitationByID(invitation.ID)
	if err != nil {
		return nil, fmt.Errorf("读取邀请详情失败: %w", err)
	}
	return updated, nil
}
