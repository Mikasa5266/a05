package service

import (
	"fmt"
	"strings"
)

type GroupWSSignalSession struct {
	RoomID             string
	TargetParticipants int
	StartThreshold     int
}

type GroupWSSignalService struct{}

func NewGroupWSSignalService() *GroupWSSignalService {
	return &GroupWSSignalService{}
}

func (s *GroupWSSignalService) Authorize(userID uint, role, userUUID, roomID, invitationCode string) (*GroupWSSignalSession, error) {
	trimmedRoomID := strings.TrimSpace(roomID)
	if trimmedRoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	if _, err := ValidateLiveRoomAccess(userID, role, userUUID, trimmedRoomID, invitationCode); err != nil {
		return nil, err
	}

	scenarioCfg, err := resolveInvitationScenarioConfigByRoomID(trimmedRoomID)
	if err != nil {
		return nil, err
	}
	if scenarioCfg.ScenarioType != "group" {
		return nil, fmt.Errorf("该房间不是群面，请使用 /ws/interview/live")
	}

	return &GroupWSSignalSession{
		RoomID:             trimmedRoomID,
		TargetParticipants: scenarioCfg.TargetParticipants,
		StartThreshold:     scenarioCfg.StartThreshold,
	}, nil
}
