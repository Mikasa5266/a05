package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"your-project/internal/model"
	"your-project/internal/repository"
)

type LiveInterviewUseCase interface {
	EnsureInvitationSecurityFields(invitation *model.HumanInterviewInvitation) error
	ValidateJoinPermission(userID uint, userRole, userUUID string, invitation *model.HumanInterviewInvitation, invitationCode string) (string, error)
	JoinInterview(userID uint, userRole, userUUID string, invitationID uint, invitationCode string) (*LiveInterviewJoinResult, error)
	ValidateLiveRoomAccess(userID uint, userRole, userUUID, roomID, invitationCode string) (*LiveInterviewJoinResult, error)
	GetLiveInterviewWorkbench(userID uint, userRole string) (*LiveInterviewWorkbenchResult, error)
}

type liveInterviewRepository interface {
	UpdateInvitation(invitation *model.HumanInterviewInvitation) error
	GetInvitationByIDForParticipant(id, userID uint) (*model.HumanInterviewInvitation, error)
	GetByID(id uint) (*model.Interview, error)
	Update(interview *model.Interview) error
	GetInvitationsByInviteeID(inviteeUserID uint) ([]model.HumanInterviewInvitation, error)
	GetInterviewsByIDs(ids []uint) ([]model.Interview, error)
}

type liveInterviewUserRepository interface {
	GetByID(id uint) (*model.User, error)
	Update(user *model.User) error
}

type liveInterviewUseCase struct {
	interviewRepo liveInterviewRepository
	userRepo      liveInterviewUserRepository
}

var _ LiveInterviewUseCase = (*liveInterviewUseCase)(nil)

var _ liveInterviewRepository = (*repository.InterviewRepository)(nil)
var _ liveInterviewUserRepository = (*repository.UserRepository)(nil)

func NewLiveInterviewUseCase(interviewRepo *repository.InterviewRepository, userRepo *repository.UserRepository) LiveInterviewUseCase {
	return NewLiveInterviewUseCaseWithPorts(interviewRepo, userRepo)
}

func NewLiveInterviewUseCaseWithPorts(interviewRepo liveInterviewRepository, userRepo liveInterviewUserRepository) LiveInterviewUseCase {
	return &liveInterviewUseCase{
		interviewRepo: interviewRepo,
		userRepo:      userRepo,
	}
}

func (u *liveInterviewUseCase) EnsureInvitationSecurityFields(invitation *model.HumanInterviewInvitation) error {
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
		student, err := u.userRepo.GetByID(invitation.StudentID)
		if err != nil {
			return fmt.Errorf("读取学生身份失败: %w", err)
		}
		uuid, err := ensureUserUUID(u.userRepo, student)
		if err != nil {
			return fmt.Errorf("更新学生身份失败: %w", err)
		}
		invitation.StudentUUID = uuid
		changed = true
	}

	if strings.TrimSpace(invitation.InviteeUUID) == "" {
		invitee, err := u.userRepo.GetByID(invitation.InviteeUserID)
		if err != nil {
			return fmt.Errorf("读取面试官身份失败: %w", err)
		}
		uuid, err := ensureUserUUID(u.userRepo, invitee)
		if err != nil {
			return fmt.Errorf("更新面试官身份失败: %w", err)
		}
		invitation.InviteeUUID = uuid
		changed = true
	}

	if changed {
		if err := u.interviewRepo.UpdateInvitation(invitation); err != nil {
			return fmt.Errorf("更新邀请安全字段失败: %w", err)
		}
	}
	return nil
}

func (u *liveInterviewUseCase) ValidateJoinPermission(userID uint, userRole, userUUID string, invitation *model.HumanInterviewInvitation, invitationCode string) (string, error) {
	if invitation == nil {
		return "", fmt.Errorf("邀请不存在")
	}
	if err := u.EnsureInvitationSecurityFields(invitation); err != nil {
		return "", err
	}

	role := normalizeActorRole(userRole)
	if role == "" {
		user, err := u.userRepo.GetByID(userID)
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
		user, err := u.userRepo.GetByID(userID)
		if err != nil {
			return "", fmt.Errorf("用户不存在")
		}
		uuid, err := ensureUserUUID(u.userRepo, user)
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

	switch normalizeStatusValue(invitation.Status) {
	case invitationStatusAccepted, invitationStatusInProgress:
		return normalizedUUID, nil
	case invitationStatusPending:
		return "", fmt.Errorf("邀请尚未被接受")
	case invitationStatusRejected, invitationStatusCancelled, invitationStatusCompleted:
		return "", fmt.Errorf("邀请状态为 %s，无法进入房间", invitation.Status)
	default:
		return "", fmt.Errorf("邀请状态异常")
	}
}

func (u *liveInterviewUseCase) JoinInterview(userID uint, userRole, userUUID string, invitationID uint, invitationCode string) (*LiveInterviewJoinResult, error) {
	if invitationID == 0 {
		return nil, fmt.Errorf("invitation_id is required")
	}

	invitation, err := u.interviewRepo.GetInvitationByIDForParticipant(invitationID, userID)
	if err != nil {
		return nil, fmt.Errorf("邀请不存在或无访问权限")
	}

	normalizedUUID, err := u.ValidateJoinPermission(userID, userRole, userUUID, invitation, invitationCode)
	if err != nil {
		return nil, err
	}

	if normalizeStatusValue(invitation.Status) == invitationStatusAccepted {
		nextStatus, transitionErr := transitionInvitationStatus(invitation.Status, invitationStatusInProgress)
		if transitionErr != nil {
			return nil, transitionErr
		}
		invitation.Status = nextStatus
		if err := u.interviewRepo.UpdateInvitation(invitation); err != nil {
			return nil, fmt.Errorf("更新邀请状态失败: %w", err)
		}
	}

	if invitation.InterviewID != nil {
		interview, interviewErr := u.interviewRepo.GetByID(*invitation.InterviewID)
		if interviewErr == nil && interview != nil {
			if normalizeStatusValue(interview.Status) == interviewStatusPending {
				nextStatus, transitionErr := transitionInterviewStatus(interview.Status, interviewStatusInProgress)
				if transitionErr != nil {
					return nil, transitionErr
				}
				interview.Status = nextStatus
			}
			if interview.InvitationCode == nil || strings.TrimSpace(*interview.InvitationCode) == "" {
				code := invitation.InvitationCode
				interview.InvitationCode = &code
			}
			if err := u.interviewRepo.Update(interview); err != nil {
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

func (u *liveInterviewUseCase) ValidateLiveRoomAccess(userID uint, userRole, userUUID, roomID, invitationCode string) (*LiveInterviewJoinResult, error) {
	invitationID, err := parseInvitationIDFromRoomID(roomID)
	if err != nil {
		return nil, err
	}
	result, err := u.JoinInterview(userID, userRole, userUUID, invitationID, invitationCode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(result.RoomID) != strings.TrimSpace(roomID) {
		return nil, fmt.Errorf("房间访问凭证无效")
	}
	return result, nil
}

func (u *liveInterviewUseCase) GetLiveInterviewWorkbench(userID uint, userRole string) (*LiveInterviewWorkbenchResult, error) {
	role := normalizeActorRole(userRole)
	if role != "enterprise" && role != "university" {
		return nil, fmt.Errorf("仅企业端或高校端可访问面试工作台")
	}

	invitations, err := u.interviewRepo.GetInvitationsByInviteeID(userID)
	if err != nil {
		return nil, err
	}

	interviewIDs := make([]uint, 0, len(invitations))
	for i := range invitations {
		inv := &invitations[i]
		if ensureErr := u.EnsureInvitationSecurityFields(inv); ensureErr != nil {
			continue
		}
		if inv.InterviewID != nil && *inv.InterviewID > 0 {
			interviewIDs = append(interviewIDs, *inv.InterviewID)
		}
	}

	interviewMap := map[uint]model.Interview{}
	if len(interviewIDs) > 0 {
		rows, rowsErr := u.interviewRepo.GetInterviewsByIDs(interviewIDs)
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
			CanJoin:        isInvitationJoinableStatus(inv.Status),
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

				if inv.Status == invitationStatusInProgress && !interview.StartTime.IsZero() {
					durationLeft := interview.StartTime.Add(liveInterviewSessionDuration).Sub(result.ServerTime)
					if durationLeft > 0 {
						item.RemainingSeconds = int64(durationLeft / time.Second)
					}
				}
			}
		}

		result.InviteList = append(result.InviteList, item)
		if item.Status == invitationStatusPending {
			result.Pending = append(result.Pending, item)
		} else {
			result.Processed = append(result.Processed, item)
		}
		if item.Status == invitationStatusInProgress {
			result.InProgress = append(result.InProgress, item)
		}
		if item.Status == invitationStatusCompleted {
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
