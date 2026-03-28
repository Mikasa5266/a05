package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"your-project/model"
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

type LiveInterviewWorkbenchItem struct {
	ID               uint       `json:"id"`
	InvitationCode   string     `json:"invitation_code"`
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
	case "student", "enterprise", "university":
		return r
	default:
		return ""
	}
}

func (s *InterviewService) ensureInvitationSecurityFields(invitation *model.HumanInterviewInvitation) error {
	if invitation == nil {
		return fmt.Errorf("邀请不存在")
	}

	changed := false
	if strings.TrimSpace(invitation.InvitationCode) == "" {
		code, err := generateInvitationCode()
		if err != nil {
			return fmt.Errorf("生成邀请码失败: %w", err)
		}
		invitation.InvitationCode = code
		changed = true
	}

	if strings.TrimSpace(invitation.StudentUUID) == "" {
		student, err := s.userRepo.GetByID(invitation.StudentID)
		if err != nil {
			return fmt.Errorf("读取学生身份失败: %w", err)
		}
		uuid, err := ensureUserUUID(s.userRepo, student)
		if err != nil {
			return fmt.Errorf("更新学生身份失败: %w", err)
		}
		invitation.StudentUUID = uuid
		changed = true
	}

	if strings.TrimSpace(invitation.InviteeUUID) == "" {
		invitee, err := s.userRepo.GetByID(invitation.InviteeUserID)
		if err != nil {
			return fmt.Errorf("读取面试官身份失败: %w", err)
		}
		uuid, err := ensureUserUUID(s.userRepo, invitee)
		if err != nil {
			return fmt.Errorf("更新面试官身份失败: %w", err)
		}
		invitation.InviteeUUID = uuid
		changed = true
	}

	if changed {
		if err := s.interviewRepo.UpdateInvitation(invitation); err != nil {
			return fmt.Errorf("更新邀请安全字段失败: %w", err)
		}
	}
	return nil
}

func (s *InterviewService) validateJoinPermission(userID uint, userRole, userUUID string, invitation *model.HumanInterviewInvitation, invitationCode string) (string, error) {
	if invitation == nil {
		return "", fmt.Errorf("邀请不存在")
	}
	if err := s.ensureInvitationSecurityFields(invitation); err != nil {
		return "", err
	}

	role := normalizeActorRole(userRole)
	if role == "" {
		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			return "", fmt.Errorf("用户不存在")
		}
		role = normalizeActorRole(user.Role)
		if role == "" {
			return "", fmt.Errorf("无效用户角色")
		}
	}

	normalizedUUID := strings.TrimSpace(userUUID)
	if normalizedUUID == "" {
		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			return "", fmt.Errorf("用户不存在")
		}
		uuid, err := ensureUserUUID(s.userRepo, user)
		if err != nil {
			return "", fmt.Errorf("用户身份标识校验失败: %w", err)
		}
		normalizedUUID = uuid
	}

	uuidGranted := normalizedUUID == strings.TrimSpace(invitation.StudentUUID) || normalizedUUID == strings.TrimSpace(invitation.InviteeUUID)
	if !uuidGranted {
		return "", fmt.Errorf("仅被邀请用户可进入该面试房间")
	}

	roleGranted := userID == invitation.StudentID || (userID == invitation.InviteeUserID && role == normalizeActorRole(invitation.InviteeRole))
	codeGranted := false
	providedCode := strings.ToUpper(strings.TrimSpace(invitationCode))
	if providedCode != "" && providedCode == strings.ToUpper(strings.TrimSpace(invitation.InvitationCode)) {
		codeGranted = true
	}
	if !roleGranted && !codeGranted {
		return "", fmt.Errorf("未通过邀请码或组织权限校验")
	}

	switch invitation.Status {
	case "accepted", "in_progress":
		return normalizedUUID, nil
	case "pending":
		return "", fmt.Errorf("邀请尚未被接受")
	case "rejected", "cancelled", "completed":
		return "", fmt.Errorf("邀请状态为 %s，无法进入房间", invitation.Status)
	default:
		return "", fmt.Errorf("邀请状态异常")
	}
}

func (s *InterviewService) JoinInterview(userID uint, userRole, userUUID string, invitationID uint, invitationCode string) (*LiveInterviewJoinResult, error) {
	if invitationID == 0 {
		return nil, fmt.Errorf("invitation_id is required")
	}

	invitation, err := s.interviewRepo.GetInvitationByIDForParticipant(invitationID, userID)
	if err != nil {
		return nil, fmt.Errorf("邀请不存在或无访问权限")
	}

	normalizedUUID, err := s.validateJoinPermission(userID, userRole, userUUID, invitation, invitationCode)
	if err != nil {
		return nil, err
	}

	if invitation.Status == "accepted" {
		invitation.Status = "in_progress"
		if err := s.interviewRepo.UpdateInvitation(invitation); err != nil {
			return nil, fmt.Errorf("更新邀请状态失败: %w", err)
		}
	}

	if invitation.InterviewID != nil {
		interview, interviewErr := s.interviewRepo.GetByID(*invitation.InterviewID)
		if interviewErr == nil && interview != nil {
			if strings.TrimSpace(interview.Status) == "pending" {
				interview.Status = "in_progress"
			}
			if interview.InvitationCode == nil || strings.TrimSpace(*interview.InvitationCode) == "" {
				code := invitation.InvitationCode
				interview.InvitationCode = &code
			}
			if err := s.interviewRepo.Update(interview); err != nil {
				return nil, fmt.Errorf("更新面试会话状态失败: %w", err)
			}
		}
	}

	return &LiveInterviewJoinResult{
		InvitationID:    invitation.ID,
		InvitationCode:  invitation.InvitationCode,
		RoomID:          buildInvitationRoomID(invitation.ID),
		InterviewID:     invitation.InterviewID,
		Status:          invitation.Status,
		Role:            normalizeActorRole(userRole),
		ParticipantUUID: normalizedUUID,
	}, nil
}

func JoinInterview(userID uint, userRole, userUUID string, invitationID uint, invitationCode string) (*LiveInterviewJoinResult, error) {
	svc := NewInterviewService()
	return svc.JoinInterview(userID, userRole, userUUID, invitationID, invitationCode)
}

func (s *InterviewService) ValidateLiveRoomAccess(userID uint, userRole, userUUID, roomID, invitationCode string) (*LiveInterviewJoinResult, error) {
	invitationID, err := parseInvitationIDFromRoomID(roomID)
	if err != nil {
		return nil, err
	}
	result, err := s.JoinInterview(userID, userRole, userUUID, invitationID, invitationCode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(result.RoomID) != strings.TrimSpace(roomID) {
		return nil, fmt.Errorf("房间访问凭证无效")
	}
	return result, nil
}

func ValidateLiveRoomAccess(userID uint, userRole, userUUID, roomID, invitationCode string) (*LiveInterviewJoinResult, error) {
	svc := NewInterviewService()
	return svc.ValidateLiveRoomAccess(userID, userRole, userUUID, roomID, invitationCode)
}

func (s *InterviewService) GetLiveInterviewWorkbench(userID uint, userRole string) (*LiveInterviewWorkbenchResult, error) {
	role := normalizeActorRole(userRole)
	if role != "enterprise" && role != "university" {
		return nil, fmt.Errorf("仅企业端或高校端可访问面试工作台")
	}

	invitations, err := s.interviewRepo.GetInvitationsByInviteeID(userID)
	if err != nil {
		return nil, err
	}

	interviewIDs := make([]uint, 0, len(invitations))
	for i := range invitations {
		inv := &invitations[i]
		if ensureErr := s.ensureInvitationSecurityFields(inv); ensureErr != nil {
			continue
		}
		if inv.InterviewID != nil && *inv.InterviewID > 0 {
			interviewIDs = append(interviewIDs, *inv.InterviewID)
		}
	}

	interviewMap := map[uint]model.Interview{}
	if len(interviewIDs) > 0 {
		rows, rowsErr := s.interviewRepo.GetInterviewsByIDs(interviewIDs)
		if rowsErr == nil {
			for i := range rows {
				interviewMap[rows[i].ID] = rows[i]
			}
		}
	}

	result := &LiveInterviewWorkbenchResult{
		InviteList: make([]LiveInterviewWorkbenchItem, 0, len(invitations)),
		Pending:    make([]LiveInterviewWorkbenchItem, 0),
		Processed:  make([]LiveInterviewWorkbenchItem, 0),
		InProgress: make([]LiveInterviewWorkbenchItem, 0),
		History:    make([]LiveInterviewWorkbenchItem, 0),
		ServerTime: time.Now(),
	}

	for i := range invitations {
		inv := invitations[i]
		item := LiveInterviewWorkbenchItem{
			ID:             inv.ID,
			InvitationCode: inv.InvitationCode,
			StudentID:      inv.StudentID,
			StudentUUID:    inv.StudentUUID,
			StudentName:    strings.TrimSpace(inv.Student.Username),
			InviteeUserID:  inv.InviteeUserID,
			InviteeUUID:    inv.InviteeUUID,
			InviteeRole:    inv.InviteeRole,
			Position:       inv.Position,
			Difficulty:     inv.Difficulty,
			Mode:           inv.Mode,
			Style:          inv.Style,
			Company:        inv.Company,
			Status:         inv.Status,
			ScheduledAt:    inv.ScheduledAt,
			Notes:          inv.Notes,
			InterviewID:    inv.InterviewID,
			CanJoin:        inv.Status == "accepted" || inv.Status == "in_progress",
			CreatedAt:      inv.CreatedAt,
			UpdatedAt:      inv.UpdatedAt,
		}

		if inv.InterviewID != nil {
			if interview, ok := interviewMap[*inv.InterviewID]; ok {
				item.InterviewStatus = interview.Status
				if !interview.StartTime.IsZero() {
					start := interview.StartTime
					item.InterviewStartAt = &start
				}
				item.HumanScore = interview.HumanScore
				item.HumanFeedback = interview.HumanFeedback

				if inv.Status == "in_progress" && !interview.StartTime.IsZero() {
					durationLeft := interview.StartTime.Add(liveInterviewSessionDuration).Sub(result.ServerTime)
					if durationLeft > 0 {
						item.RemainingSeconds = int64(durationLeft / time.Second)
					}
				}
			}
		}

		result.InviteList = append(result.InviteList, item)
		if item.Status == "pending" {
			result.Pending = append(result.Pending, item)
		} else {
			result.Processed = append(result.Processed, item)
		}
		if item.Status == "in_progress" {
			result.InProgress = append(result.InProgress, item)
		}
		if item.Status == "completed" {
			result.History = append(result.History, item)
		}
	}

	sort.Slice(result.InviteList, func(i, j int) bool { return result.InviteList[i].CreatedAt.After(result.InviteList[j].CreatedAt) })
	sort.Slice(result.Pending, func(i, j int) bool { return result.Pending[i].CreatedAt.After(result.Pending[j].CreatedAt) })
	sort.Slice(result.Processed, func(i, j int) bool { return result.Processed[i].UpdatedAt.After(result.Processed[j].UpdatedAt) })
	sort.Slice(result.InProgress, func(i, j int) bool { return result.InProgress[i].UpdatedAt.After(result.InProgress[j].UpdatedAt) })
	sort.Slice(result.History, func(i, j int) bool { return result.History[i].UpdatedAt.After(result.History[j].UpdatedAt) })

	result.Summary = LiveInterviewWorkbenchSummary{
		InviteCount:     len(result.InviteList),
		PendingCount:    len(result.Pending),
		ProcessedCount:  len(result.Processed),
		InProgressCount: len(result.InProgress),
		HistoryCount:    len(result.History),
	}

	return result, nil
}

func GetLiveInterviewWorkbench(userID uint, userRole string) (*LiveInterviewWorkbenchResult, error) {
	svc := NewInterviewService()
	return svc.GetLiveInterviewWorkbench(userID, userRole)
}
