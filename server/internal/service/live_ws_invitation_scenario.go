package service

import (
	"fmt"
	"strings"

	"your-project/internal/model"
)

type invitationScenarioConfig struct {
	ScenarioType       string
	TargetParticipants int
	StartThreshold     int
}

func resolveInvitationScenarioConfigByRoomID(roomID string) (*invitationScenarioConfig, error) {
	invitationID, err := parseInvitationIDFromRoomID(roomID)
	if err != nil {
		return nil, err
	}

	svc := NewInterviewService()
	invitation, err := svc.interviewRepo.GetInvitationByID(invitationID)
	if err != nil {
		return nil, fmt.Errorf("邀请不存在")
	}

	scenario := strings.ToLower(strings.TrimSpace(invitation.ScenarioType))
	if scenario != model.InvitationScenarioGroup {
		scenario = model.InvitationScenarioSingle
	}

	targetParticipants := invitation.TargetParticipants
	if targetParticipants < groupInvitationMinParticipants {
		targetParticipants = groupInvitationMinParticipants
	}
	if targetParticipants > groupInvitationMaxParticipants {
		targetParticipants = groupInvitationMaxParticipants
	}

	startThreshold := invitation.StartThreshold
	if startThreshold < groupInvitationMinThreshold {
		startThreshold = groupInvitationMinThreshold
	}
	if startThreshold > targetParticipants {
		startThreshold = targetParticipants
	}

	return &invitationScenarioConfig{
		ScenarioType:       scenario,
		TargetParticipants: targetParticipants,
		StartThreshold:     startThreshold,
	}, nil
}
