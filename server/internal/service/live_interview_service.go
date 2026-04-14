package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"your-project/internal/model"
)

const liveInterviewSessionDuration = 45 * time.Minute

type LiveInterviewJoinResult struct {
	InvitationID    uint   `json:"invitation_id"`
	InvitationCode  string `json:"invitation_code"`
	RoomID          string `json:"room_id"`
	InterviewID     *uint  `json:"interview_id,omitempty"`
	Status          string `json:"status"`
	Role            string `json:"role"`
	ParticipantUUID string `json:"participant_uuid"`
}

type LiveInterviewStartResult struct {
	InvitationID    uint       `json:"invitation_id"`
	InvitationCode  string     `json:"invitation_code"`
	InterviewID     *uint      `json:"interview_id,omitempty"`
	Status          string     `json:"status"`
	InterviewStatus string     `json:"interview_status,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
}

type LiveInterviewWorkbenchItem struct {
	ID               uint       `json:"id"`
	InvitationCode   string     `json:"invitation_code"`
	InitiatorUserID  uint       `json:"initiator_user_id"`
	InitiatorUUID    string     `json:"initiator_uuid"`
	InitiatorRole    string     `json:"initiator_role"`
	InitiatorName    string     `json:"initiator_name"`
	TargetUserID     uint       `json:"target_user_id"`
	TargetUUID       string     `json:"target_uuid"`
	TargetRole       string     `json:"target_role"`
	TargetName       string     `json:"target_name"`
	CounterpartID    uint       `json:"counterpart_id"`
	CounterpartRole  string     `json:"counterpart_role"`
	CounterpartName  string     `json:"counterpart_name"`
	CurrentUserRole  string     `json:"current_user_role"`
	StudentID        uint       `json:"student_id"`
	StudentUUID      string     `json:"student_uuid"`
	StudentName      string     `json:"student_name"`
	InviteeUserID    uint       `json:"invitee_user_id"`
	InviteeUUID      string     `json:"invitee_uuid"`
	InviteeRole      string     `json:"invitee_role"`
	Position         string     `json:"position"`
	Difficulty       string     `json:"difficulty"`
	Mode             string     `json:"mode"`
	Style            string     `json:"style"`
	Company          string     `json:"company"`
	Status           string     `json:"status"`
	ScheduledAt      *time.Time `json:"scheduled_at,omitempty"`
	Notes            string     `json:"notes,omitempty"`
	InterviewID      *uint      `json:"interview_id,omitempty"`
	InterviewStatus  string     `json:"interview_status,omitempty"`
	InterviewStartAt *time.Time `json:"interview_start_at,omitempty"`
	RemainingSeconds int64      `json:"remaining_seconds"`
	HumanScore       *int       `json:"human_score,omitempty"`
	HumanFeedback    string     `json:"human_feedback,omitempty"`
	CanJoin          bool       `json:"can_join"`
	CanRespond       bool       `json:"can_respond"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type LiveInterviewWorkbenchSummary struct {
	InviteCount     int `json:"invite_count"`
	PendingCount    int `json:"pending_count"`
	ProcessedCount  int `json:"processed_count"`
	InProgressCount int `json:"in_progress_count"`
	HistoryCount    int `json:"history_count"`
}

type LiveInterviewWorkbenchResult struct {
	InviteList []LiveInterviewWorkbenchItem  `json:"invite_list"`
	Pending    []LiveInterviewWorkbenchItem  `json:"pending"`
	Processed  []LiveInterviewWorkbenchItem  `json:"processed"`
	InProgress []LiveInterviewWorkbenchItem  `json:"in_progress"`
	History    []LiveInterviewWorkbenchItem  `json:"history"`
	Summary    LiveInterviewWorkbenchSummary `json:"summary"`
	ServerTime time.Time                     `json:"server_time"`
}

func buildInvitationRoomID(invitationID uint) string {
	return "invitation-" + strconv.FormatUint(uint64(invitationID), 10)
}

func parseInvitationIDFromRoomID(roomID string) (uint, error) {
	value := strings.TrimSpace(roomID)
	if !strings.HasPrefix(value, "invitation-") {
		return 0, fmt.Errorf("仅支持邀请房间")
	}
	raw := strings.TrimPrefix(value, "invitation-")
	parsed, err := strconv.ParseUint(raw, 10, 32)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("房间号无效")
	}
	return uint(parsed), nil
}

func normalizeActorRole(role string) string {
	r := strings.ToLower(strings.TrimSpace(role))
	switch r {
	case "teacher", "mentor", "faculty":
		return "university"
	case "hr", "interviewer", "recruiter":
		return "enterprise"
	case "student", "enterprise", "university":
		return r
	default:
		return ""
	}
}

func (s *InterviewService) liveInterviewUseCase() LiveInterviewUseCase {
	return NewLiveInterviewUseCase(s.interviewRepo, s.userRepo)
}

func (s *InterviewService) ensureInvitationSecurityFields(invitation *model.HumanInterviewInvitation) error {
	return s.liveInterviewUseCase().EnsureInvitationSecurityFields(invitation)
}

func (s *InterviewService) validateJoinPermission(userID uint, userRole, userUUID string, invitation *model.HumanInterviewInvitation, invitationCode string) (*liveJoinPermission, error) {
	return s.liveInterviewUseCase().ValidateJoinPermission(userID, userRole, userUUID, invitation, invitationCode)
}

func (s *InterviewService) JoinInterview(userID uint, userRole, userUUID string, invitationID uint, invitationCode string) (*LiveInterviewJoinResult, error) {
	return s.liveInterviewUseCase().JoinInterview(userID, userRole, userUUID, invitationID, invitationCode)
}

func JoinInterview(userID uint, userRole, userUUID string, invitationID uint, invitationCode string) (*LiveInterviewJoinResult, error) {
	svc := NewInterviewService()
	return svc.JoinInterview(userID, userRole, userUUID, invitationID, invitationCode)
}

func (s *InterviewService) StartLiveInterview(userID uint, userRole string, invitationID uint) (*LiveInterviewStartResult, error) {
	return s.liveInterviewUseCase().StartLiveInterview(userID, userRole, invitationID)
}

func StartLiveInterview(userID uint, userRole string, invitationID uint) (*LiveInterviewStartResult, error) {
	svc := NewInterviewService()
	return svc.StartLiveInterview(userID, userRole, invitationID)
}

func (s *InterviewService) ValidateLiveRoomAccess(userID uint, userRole, userUUID, roomID, invitationCode string) (*LiveInterviewJoinResult, error) {
	return s.liveInterviewUseCase().ValidateLiveRoomAccess(userID, userRole, userUUID, roomID, invitationCode)
}

func ValidateLiveRoomAccess(userID uint, userRole, userUUID, roomID, invitationCode string) (*LiveInterviewJoinResult, error) {
	svc := NewInterviewService()
	return svc.ValidateLiveRoomAccess(userID, userRole, userUUID, roomID, invitationCode)
}

func (s *InterviewService) GetLiveInterviewWorkbench(userID uint, userRole string) (*LiveInterviewWorkbenchResult, error) {
	return s.liveInterviewUseCase().GetLiveInterviewWorkbench(userID, userRole)
}

func GetLiveInterviewWorkbench(userID uint, userRole string) (*LiveInterviewWorkbenchResult, error) {
	svc := NewInterviewService()
	return svc.GetLiveInterviewWorkbench(userID, userRole)
}
