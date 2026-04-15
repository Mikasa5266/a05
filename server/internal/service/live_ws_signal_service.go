package service

import (
	"fmt"
	"strings"
)

type LiveWSSignalSession struct {
	RoomID string
}

type LiveWSSignalService struct{}

func NewLiveWSSignalService() *LiveWSSignalService {
	return &LiveWSSignalService{}
}

func (s *LiveWSSignalService) Authorize(userID uint, role, userUUID, roomID, invitationCode string) (*LiveWSSignalSession, error) {
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
	if scenarioCfg.ScenarioType == "group" {
		return nil, fmt.Errorf("该房间为群面，请使用 /ws/interview/group")
	}

	return &LiveWSSignalSession{RoomID: trimmedRoomID}, nil
}
