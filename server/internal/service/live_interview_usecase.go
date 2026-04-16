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
	ValidateJoinPermission(userID uint, userRole, userUUID string, invitation *model.HumanInterviewInvitation, invitationCode string) (*liveJoinPermission, error)
	JoinInterview(userID uint, userRole, userUUID string, invitationID uint, invitationCode string) (*LiveInterviewJoinResult, error)
	StartLiveInterview(userID uint, userRole string, invitationID uint) (*LiveInterviewStartResult, error)
	ValidateLiveRoomAccess(userID uint, userRole, userUUID, roomID, invitationCode string) (*LiveInterviewJoinResult, error)
	GetLiveInterviewWorkbench(userID uint, userRole string) (*LiveInterviewWorkbenchResult, error)
}

type liveInterviewRepository interface {
	UpdateInvitation(invitation *model.HumanInterviewInvitation) error
	GetInvitationByID(id uint) (*model.HumanInterviewInvitation, error)
	GetInvitationByIDForParticipant(id, userID uint) (*model.HumanInterviewInvitation, error)
	GetInvitationParticipant(invitationID, userID uint) (*model.HumanInterviewInvitationParticipant, error)
	ListInvitationParticipants(invitationID uint) ([]model.HumanInterviewInvitationParticipant, error)
	UpsertInvitationParticipant(participant *model.HumanInterviewInvitationParticipant) error
	MarkInvitationParticipantJoined(invitationID, userID uint, joinedAt time.Time) error
	GetByID(id uint) (*model.Interview, error)
	Update(interview *model.Interview) error
	GetInvitationsByParticipantOrInitiator(userID uint) ([]model.HumanInterviewInvitation, error)
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

type liveJoinPermission struct {
	ParticipantUUID     string
	SessionRole         string
	ParticipantRole     string
	IsDirectParticipant bool
	ShouldAutoAccept    bool
}

const (
	liveJoinEarlyAllowance      = 30 * time.Minute
	liveJoinLateAllowance       = 6 * time.Hour
	liveJoinDefaultExpiryWindow = 72 * time.Hour
)

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

func invitationInitiatorID(invitation *model.HumanInterviewInvitation) uint {
	if invitation == nil {
		return 0
	}
	if invitation.InitiatorUserID > 0 {
		return invitation.InitiatorUserID
	}
	return invitation.StudentID
}

func invitationTargetID(invitation *model.HumanInterviewInvitation) uint {
	if invitation == nil {
		return 0
	}
	if invitation.TargetUserID > 0 {
		return invitation.TargetUserID
	}
	return invitation.InviteeUserID
}

func invitationInitiatorRole(invitation *model.HumanInterviewInvitation) string {
	if invitation == nil {
		return ""
	}
	role := normalizeActorRole(invitation.InitiatorRole)
	if role != "" {
		return role
	}
	if normalizeActorRole(invitation.Student.Role) != "" {
		return normalizeActorRole(invitation.Student.Role)
	}
	return "student"
}

func invitationTargetRole(invitation *model.HumanInterviewInvitation) string {
	if invitation == nil {
		return ""
	}
	role := normalizeActorRole(invitation.TargetRole)
	if role != "" {
		return role
	}
	if normalizeActorRole(invitation.InviteeRole) != "" {
		return normalizeActorRole(invitation.InviteeRole)
	}
	return normalizeActorRole(invitation.Invitee.Role)
}

func validateInvitationJoinWindow(invitation *model.HumanInterviewInvitation, now time.Time) error {
	if invitation == nil {
		return fmt.Errorf("邀请不存在")
	}

	if invitation.ScheduledAt != nil && !invitation.ScheduledAt.IsZero() {
		scheduledAt := *invitation.ScheduledAt
		if now.Before(scheduledAt.Add(-liveJoinEarlyAllowance)) {
			return fmt.Errorf("尚未到可入会时间")
		}
		if now.After(scheduledAt.Add(liveJoinLateAllowance)) {
			return fmt.Errorf("该房间已过可入会时间")
		}
		return nil
	}

	if !invitation.CreatedAt.IsZero() && now.After(invitation.CreatedAt.Add(liveJoinDefaultExpiryWindow)) {
		return fmt.Errorf("该房间已过期")
	}

	return nil
}

func (u *liveInterviewUseCase) EnsureInvitationSecurityFields(invitation *model.HumanInterviewInvitation) error {
	if invitation == nil {
		return fmt.Errorf("邀请不存在")
	}

	changed := false
	if invitation.InitiatorUserID == 0 && invitation.StudentID > 0 {
		invitation.InitiatorUserID = invitation.StudentID
		changed = true
	}
	if invitation.TargetUserID == 0 && invitation.InviteeUserID > 0 {
		invitation.TargetUserID = invitation.InviteeUserID
		changed = true
	}
	if invitation.StudentID == 0 && invitation.InitiatorUserID > 0 {
		invitation.StudentID = invitation.InitiatorUserID
		changed = true
	}
	if invitation.InviteeUserID == 0 && invitation.TargetUserID > 0 {
		invitation.InviteeUserID = invitation.TargetUserID
		changed = true
	}
	if strings.TrimSpace(invitation.InitiatorRole) == "" {
		invitation.InitiatorRole = invitationInitiatorRole(invitation)
		changed = true
	}
	if strings.TrimSpace(invitation.TargetRole) == "" {
		invitation.TargetRole = invitationTargetRole(invitation)
		changed = true
	}
	if strings.TrimSpace(invitation.ScenarioType) == "" {
		invitation.ScenarioType = model.InvitationScenarioSingle
		changed = true
	}
	if invitation.TargetParticipants < 2 {
		invitation.TargetParticipants = 2
		changed = true
	}
	if invitation.StartThreshold < 1 {
		invitation.StartThreshold = 2
		changed = true
	}
	if strings.TrimSpace(invitation.InvitationCode) == "" {
		code, err := generateInvitationCode()
		if err != nil {
			return fmt.Errorf("生成邀请码失败: %w", err)
		}
		invitation.InvitationCode = code
		changed = true
	}

	if strings.TrimSpace(invitation.StudentUUID) == "" || strings.TrimSpace(invitation.InitiatorUUID) == "" {
		student, err := u.userRepo.GetByID(invitationInitiatorID(invitation))
		if err != nil {
			return fmt.Errorf("读取发起人身份失败: %w", err)
		}
		uuid, err := ensureUserUUID(u.userRepo, student)
		if err != nil {
			return fmt.Errorf("更新发起人身份失败: %w", err)
		}
		invitation.StudentUUID = uuid
		invitation.InitiatorUUID = uuid
		changed = true
	}

	if strings.TrimSpace(invitation.InviteeUUID) == "" || strings.TrimSpace(invitation.TargetUUID) == "" {
		invitee, err := u.userRepo.GetByID(invitationTargetID(invitation))
		if err != nil {
			return fmt.Errorf("读取目标用户身份失败: %w", err)
		}
		uuid, err := ensureUserUUID(u.userRepo, invitee)
		if err != nil {
			return fmt.Errorf("更新目标用户身份失败: %w", err)
		}
		invitation.InviteeUUID = uuid
		invitation.TargetUUID = uuid
		changed = true
	}

	if changed {
		if err := u.interviewRepo.UpdateInvitation(invitation); err != nil {
			return fmt.Errorf("更新邀请安全字段失败: %w", err)
		}
	}
	return nil
}

func (u *liveInterviewUseCase) ValidateJoinPermission(userID uint, userRole, userUUID string, invitation *model.HumanInterviewInvitation, invitationCode string) (*liveJoinPermission, error) {
	if invitation == nil {
		return nil, fmt.Errorf("邀请不存在")
	}
	if err := u.EnsureInvitationSecurityFields(invitation); err != nil {
		return nil, err
	}

	role := normalizeActorRole(userRole)
	if role == "" {
		user, err := u.userRepo.GetByID(userID)
		if err != nil {
			return nil, fmt.Errorf("用户不存在")
		}
		role = normalizeActorRole(user.Role)
		if role == "" {
			return nil, fmt.Errorf("无效用户角色")
		}
	}

	normalizedUUID := strings.TrimSpace(userUUID)
	if normalizedUUID == "" {
		user, err := u.userRepo.GetByID(userID)
		if err != nil {
			return nil, fmt.Errorf("用户不存在")
		}
		uuid, err := ensureUserUUID(u.userRepo, user)
		if err != nil {
			return nil, fmt.Errorf("用户身份标识校验失败: %w", err)
		}
		normalizedUUID = uuid
	}

	if err := validateInvitationJoinWindow(invitation, time.Now()); err != nil {
		return nil, err
	}

	initiatorID := invitationInitiatorID(invitation)
	targetID := invitationTargetID(invitation)
	directParticipant := userID == initiatorID || userID == targetID || userID == invitation.StudentID || userID == invitation.InviteeUserID
	participantRole := model.InvitationParticipantRoleInvitee
	if userID == initiatorID || userID == invitation.StudentID {
		participantRole = model.InvitationParticipantRoleInitiator
	}

	participant, participantErr := u.interviewRepo.GetInvitationParticipant(invitation.ID, userID)
	if participantErr == nil && participant != nil {
		directParticipant = true
		if strings.TrimSpace(participant.ParticipantRole) != "" {
			participantRole = strings.TrimSpace(participant.ParticipantRole)
		}
		status := normalizeStatusValue(participant.ResponseStatus)
		if status == invitationStatusRejected || status == invitationStatusCancelled {
			return nil, fmt.Errorf("你已拒绝该邀请，无法进入房间")
		}
	}

	codeGranted := false
	providedCode := strings.ToUpper(strings.TrimSpace(invitationCode))
	if providedCode != "" && providedCode == strings.ToUpper(strings.TrimSpace(invitation.InvitationCode)) {
		codeGranted = true
	}
	if !directParticipant && !codeGranted {
		return nil, fmt.Errorf("未通过邀请码或参与者权限校验")
	}

	status := normalizeStatusValue(invitation.Status)
	autoAccept := false
	switch normalizeStatusValue(invitation.Status) {
	case invitationStatusAccepted, invitationStatusInProgress:
		// joinable
	case invitationStatusPending:
		if directParticipant {
			autoAccept = true
		} else {
			return nil, fmt.Errorf("邀请尚未被接受")
		}
	case invitationStatusRejected, invitationStatusCancelled, invitationStatusCompleted:
		return nil, fmt.Errorf("邀请状态为 %s，无法进入房间", invitation.Status)
	default:
		return nil, fmt.Errorf("邀请状态异常")
	}

	sessionRole := role
	if !directParticipant && codeGranted {
		if status == invitationStatusPending {
			return nil, fmt.Errorf("邀请尚未被接受")
		}
		sessionRole = model.InvitationParticipantRoleObserver
	}

	return &liveJoinPermission{
		ParticipantUUID:     normalizedUUID,
		SessionRole:         sessionRole,
		ParticipantRole:     participantRole,
		IsDirectParticipant: directParticipant,
		ShouldAutoAccept:    autoAccept,
	}, nil
}

func (u *liveInterviewUseCase) JoinInterview(userID uint, userRole, userUUID string, invitationID uint, invitationCode string) (*LiveInterviewJoinResult, error) {
	if invitationID == 0 {
		return nil, fmt.Errorf("invitation_id is required")
	}

	invitation, err := u.interviewRepo.GetInvitationByID(invitationID)
	if err != nil {
		return nil, fmt.Errorf("邀请不存在")
	}

	permission, err := u.ValidateJoinPermission(userID, userRole, userUUID, invitation, invitationCode)
	if err != nil {
		return nil, err
	}

	if permission.ShouldAutoAccept && normalizeStatusValue(invitation.Status) == invitationStatusPending {
		nextStatus, transitionErr := transitionInvitationStatus(invitation.Status, invitationStatusAccepted)
		if transitionErr != nil {
			return nil, transitionErr
		}
		invitation.Status = nextStatus
	}

	if permission.IsDirectParticipant {
		participantUserRole := normalizeActorRole(userRole)
		if participantUserRole == "" {
			participantUserRole = normalizeActorRole(permission.SessionRole)
		}
		participant := &model.HumanInterviewInvitationParticipant{
			InvitationID:    invitation.ID,
			UserID:          userID,
			UserUUID:        strings.TrimSpace(permission.ParticipantUUID),
			UserRole:        participantUserRole,
			ParticipantRole: permission.ParticipantRole,
			ResponseStatus:  model.InvitationParticipantStatusAccepted,
		}
		now := time.Now()
		participant.RespondedAt = &now
		if err := u.interviewRepo.UpsertInvitationParticipant(participant); err != nil {
			return nil, fmt.Errorf("更新入会参与者状态失败: %w", err)
		}
		if err := u.interviewRepo.MarkInvitationParticipantJoined(invitation.ID, userID, now); err != nil {
			return nil, fmt.Errorf("记录入会状态失败: %w", err)
		}
	}

	if err := u.interviewRepo.UpdateInvitation(invitation); err != nil {
		return nil, fmt.Errorf("更新邀请状态失败: %w", err)
	}

	if invitation.InterviewID != nil {
		interview, interviewErr := u.interviewRepo.GetByID(*invitation.InterviewID)
		if interviewErr == nil && interview != nil {
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
		Role:            permission.SessionRole,
		ParticipantUUID: permission.ParticipantUUID,
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

func (u *liveInterviewUseCase) StartLiveInterview(userID uint, userRole string, invitationID uint) (*LiveInterviewStartResult, error) {
	if invitationID == 0 {
		return nil, fmt.Errorf("invitation_id is required")
	}

	role := normalizeActorRole(userRole)
	if role == "" {
		user, err := u.userRepo.GetByID(userID)
		if err != nil {
			return nil, fmt.Errorf("用户不存在")
		}
		role = normalizeActorRole(user.Role)
		if role == "" {
			return nil, fmt.Errorf("无效用户角色")
		}
	}

	invitation, err := u.interviewRepo.GetInvitationByID(invitationID)
	if err != nil {
		return nil, fmt.Errorf("邀请不存在")
	}
	if err := u.EnsureInvitationSecurityFields(invitation); err != nil {
		return nil, err
	}
	if err := validateInvitationJoinWindow(invitation, time.Now()); err != nil {
		return nil, err
	}

	initiatorID := invitationInitiatorID(invitation)
	if initiatorID == 0 || userID != initiatorID {
		return nil, fmt.Errorf("仅发起方可开始面试")
	}

	status := normalizeStatusValue(invitation.Status)
	if status == invitationStatusRejected || status == invitationStatusCancelled || status == invitationStatusCompleted {
		return nil, fmt.Errorf("当前邀请状态为 %s，无法开始", invitation.Status)
	}

	participants, listErr := u.interviewRepo.ListInvitationParticipants(invitation.ID)
	if listErr != nil {
		participants = []model.HumanInterviewInvitationParticipant{}
	}

	activeParticipants := 1 // 发起方
	for i := range participants {
		p := participants[i]
		if p.ParticipantRole != model.InvitationParticipantRoleInvitee {
			continue
		}
		s := normalizeStatusValue(p.ResponseStatus)
		if s == model.InvitationParticipantStatusAccepted || s == model.InvitationParticipantStatusJoined {
			activeParticipants++
		}
	}

	threshold := invitation.StartThreshold
	if threshold < 1 {
		threshold = 2
	}
	if activeParticipants < threshold {
		return nil, fmt.Errorf("当前可用参与者不足，至少需要 %d 人", threshold)
	}

	if status == invitationStatusPending {
		nextStatus, transitionErr := transitionInvitationStatus(invitation.Status, invitationStatusAccepted)
		if transitionErr != nil {
			return nil, transitionErr
		}
		invitation.Status = nextStatus
		status = invitationStatusAccepted
	}

	if status == invitationStatusAccepted {
		nextStatus, transitionErr := transitionInvitationStatus(invitation.Status, invitationStatusInProgress)
		if transitionErr != nil {
			return nil, transitionErr
		}
		invitation.Status = nextStatus
	}

	if err := u.interviewRepo.UpdateInvitation(invitation); err != nil {
		return nil, fmt.Errorf("更新邀请状态失败: %w", err)
	}

	result := &LiveInterviewStartResult{
		InvitationID:    invitation.ID,
		InvitationCode:  invitation.InvitationCode,
		InterviewID:     invitation.InterviewID,
		Status:          invitation.Status,
		InterviewStatus: "",
		StartedAt:       nil,
	}

	if invitation.InterviewID != nil && *invitation.InterviewID > 0 {
		interview, interviewErr := u.interviewRepo.GetByID(*invitation.InterviewID)
		if interviewErr == nil && interview != nil {
			if normalizeStatusValue(interview.Status) == interviewStatusPending {
				nextStatus, transitionErr := transitionInterviewStatus(interview.Status, interviewStatusInProgress)
				if transitionErr != nil {
					return nil, transitionErr
				}
				interview.Status = nextStatus
				if interview.StartTime.IsZero() {
					interview.StartTime = time.Now()
				}
			}
			if interview.StartTime.IsZero() {
				interview.StartTime = time.Now()
			}
			if err := u.interviewRepo.Update(interview); err != nil {
				return nil, fmt.Errorf("更新面试会话状态失败: %w", err)
			}
			started := interview.StartTime
			result.StartedAt = &started
			result.InterviewStatus = interview.Status
		}
	}

	return result, nil
}

func (u *liveInterviewUseCase) GetLiveInterviewWorkbench(userID uint, userRole string) (*LiveInterviewWorkbenchResult, error) {
	role := normalizeActorRole(userRole)
	if role == "" {
		return nil, fmt.Errorf("无效用户角色")
	}

	invitations, err := u.interviewRepo.GetInvitationsByParticipantOrInitiator(userID)
	if err != nil {
		return nil, err
	}

	interviewIDs := make([]uint, 0, len(invitations))
	for i := range invitations {
		inv := &invitations[i]
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
		initiatorID := invitationInitiatorID(&inv)
		targetID := invitationTargetID(&inv)
		initiatorName := strings.TrimSpace(inv.Initiator.Username)
		if initiatorName == "" {
			initiatorName = strings.TrimSpace(inv.Student.Username)
		}
		targetName := strings.TrimSpace(inv.Target.Username)
		if targetName == "" {
			targetName = strings.TrimSpace(inv.Invitee.Username)
		}
		currentIsInitiator := userID == initiatorID || userID == inv.StudentID

		counterpartID := initiatorID
		counterpartRole := invitationInitiatorRole(&inv)
		counterpartName := initiatorName
		if currentIsInitiator {
			counterpartID = targetID
			counterpartRole = invitationTargetRole(&inv)
			counterpartName = targetName
		}

		status := normalizeStatusValue(inv.Status)
		canRespond := !currentIsInitiator && status == invitationStatusPending
		canJoin := status == invitationStatusPending || isInvitationJoinableStatus(inv.Status)

		item := LiveInterviewWorkbenchItem{
			ID:              inv.ID,
			InvitationCode:  inv.InvitationCode,
			InitiatorUserID: initiatorID,
			InitiatorUUID:   strings.TrimSpace(inv.InitiatorUUID),
			InitiatorRole:   invitationInitiatorRole(&inv),
			InitiatorName:   initiatorName,
			TargetUserID:    targetID,
			TargetUUID:      strings.TrimSpace(inv.TargetUUID),
			TargetRole:      invitationTargetRole(&inv),
			TargetName:      targetName,
			CounterpartID:   counterpartID,
			CounterpartRole: counterpartRole,
			CounterpartName: counterpartName,
			CurrentUserRole: role,
			StudentID:       inv.StudentID,
			StudentUUID:     inv.StudentUUID,
			StudentName:     initiatorName,
			InviteeUserID:   inv.InviteeUserID,
			InviteeUUID:     inv.InviteeUUID,
			InviteeRole:     inv.InviteeRole,
			ScenarioType:    normalizeScenarioType(inv.ScenarioType),
			TargetParticipants: inv.TargetParticipants,
			StartThreshold:  inv.StartThreshold,
			Position:        inv.Position,
			Difficulty:      inv.Difficulty,
			Mode:            inv.Mode,
			Style:           inv.Style,
			Company:         inv.Company,
			Status:          inv.Status,
			ScheduledAt:     inv.ScheduledAt,
			Notes:           inv.Notes,
			InterviewID:     inv.InterviewID,
			CanJoin:         canJoin,
			CanRespond:      canRespond,
			CreatedAt:       inv.CreatedAt,
			UpdatedAt:       inv.UpdatedAt,
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

				if normalizeStatusValue(interview.Status) == interviewStatusInProgress && !interview.StartTime.IsZero() {
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
		if normalizeStatusValue(item.Status) == invitationStatusInProgress || normalizeStatusValue(item.InterviewStatus) == interviewStatusInProgress {
			result.InProgress = append(result.InProgress, item)
		}
		if normalizeStatusValue(item.Status) == invitationStatusCompleted || normalizeStatusValue(item.InterviewStatus) == interviewStatusCompleted {
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
