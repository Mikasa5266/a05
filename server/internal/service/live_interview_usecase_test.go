package service

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"your-project/internal/model"
)

type fakeLiveInterviewRepo struct {
	invitationByID       map[uint]*model.HumanInterviewInvitation
	interviewByID        map[uint]*model.Interview
	invitationsByInvitee map[uint][]model.HumanInterviewInvitation
	participants         map[string]*model.HumanInterviewInvitationParticipant

	updatedInvitations []*model.HumanInterviewInvitation
	updatedInterviews  []*model.Interview
}

func participantKey(invitationID, userID uint) string {
	return fmt.Sprintf("%d-%d", invitationID, userID)
}

func (f *fakeLiveInterviewRepo) UpdateInvitation(invitation *model.HumanInterviewInvitation) error {
	if invitation == nil {
		return fmt.Errorf("invitation is nil")
	}
	copied := *invitation
	f.updatedInvitations = append(f.updatedInvitations, &copied)
	if f.invitationByID != nil {
		stored := copied
		f.invitationByID[invitation.ID] = &stored
	}
	return nil
}

func (f *fakeLiveInterviewRepo) GetInvitationByID(id uint) (*model.HumanInterviewInvitation, error) {
	invitation, ok := f.invitationByID[id]
	if !ok || invitation == nil {
		return nil, fmt.Errorf("not found")
	}
	return invitation, nil
}

func (f *fakeLiveInterviewRepo) GetInvitationByIDForParticipant(id, userID uint) (*model.HumanInterviewInvitation, error) {
	invitation, ok := f.invitationByID[id]
	if !ok || invitation == nil {
		return nil, fmt.Errorf("not found")
	}
	if invitation.StudentID != userID && invitation.InviteeUserID != userID {
		return nil, fmt.Errorf("forbidden")
	}
	return invitation, nil
}

func (f *fakeLiveInterviewRepo) GetInvitationParticipant(invitationID, userID uint) (*model.HumanInterviewInvitationParticipant, error) {
	if f.participants == nil {
		return nil, fmt.Errorf("not found")
	}
	participant, ok := f.participants[participantKey(invitationID, userID)]
	if !ok || participant == nil {
		return nil, fmt.Errorf("not found")
	}
	return participant, nil
}

func (f *fakeLiveInterviewRepo) ListInvitationParticipants(invitationID uint) ([]model.HumanInterviewInvitationParticipant, error) {
	rows := make([]model.HumanInterviewInvitationParticipant, 0)
	for _, p := range f.participants {
		if p != nil && p.InvitationID == invitationID {
			rows = append(rows, *p)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	return rows, nil
}

func (f *fakeLiveInterviewRepo) UpsertInvitationParticipant(participant *model.HumanInterviewInvitationParticipant) error {
	if participant == nil {
		return fmt.Errorf("participant is nil")
	}
	if f.participants == nil {
		f.participants = map[string]*model.HumanInterviewInvitationParticipant{}
	}
	copyValue := *participant
	f.participants[participantKey(participant.InvitationID, participant.UserID)] = &copyValue
	return nil
}

func (f *fakeLiveInterviewRepo) MarkInvitationParticipantJoined(invitationID, userID uint, joinedAt time.Time) error {
	if f.participants == nil {
		f.participants = map[string]*model.HumanInterviewInvitationParticipant{}
	}
	key := participantKey(invitationID, userID)
	participant, ok := f.participants[key]
	if !ok || participant == nil {
		participant = &model.HumanInterviewInvitationParticipant{InvitationID: invitationID, UserID: userID}
		f.participants[key] = participant
	}
	participant.JoinedAt = &joinedAt
	participant.ResponseStatus = model.InvitationParticipantStatusJoined
	return nil
}

func (f *fakeLiveInterviewRepo) GetByID(id uint) (*model.Interview, error) {
	interview, ok := f.interviewByID[id]
	if !ok || interview == nil {
		return nil, fmt.Errorf("not found")
	}
	return interview, nil
}

func (f *fakeLiveInterviewRepo) Update(interview *model.Interview) error {
	if interview == nil {
		return fmt.Errorf("interview is nil")
	}
	copied := *interview
	f.updatedInterviews = append(f.updatedInterviews, &copied)
	if f.interviewByID != nil {
		stored := copied
		f.interviewByID[interview.ID] = &stored
	}
	return nil
}

func (f *fakeLiveInterviewRepo) GetInvitationsByInviteeID(inviteeUserID uint) ([]model.HumanInterviewInvitation, error) {
	if f.invitationsByInvitee == nil {
		return []model.HumanInterviewInvitation{}, nil
	}
	rows, ok := f.invitationsByInvitee[inviteeUserID]
	if !ok {
		return []model.HumanInterviewInvitation{}, nil
	}
	return rows, nil
}

func (f *fakeLiveInterviewRepo) GetInvitationsByParticipantOrInitiator(userID uint) ([]model.HumanInterviewInvitation, error) {
	rows := make([]model.HumanInterviewInvitation, 0)
	for _, invitation := range f.invitationByID {
		if invitation == nil {
			continue
		}
		if invitation.InitiatorUserID == userID || invitation.StudentID == userID || invitation.TargetUserID == userID || invitation.InviteeUserID == userID {
			rows = append(rows, *invitation)
			continue
		}
		if participant, ok := f.participants[participantKey(invitation.ID, userID)]; ok && participant != nil {
			rows = append(rows, *invitation)
		}
	}
	return rows, nil
}

func (f *fakeLiveInterviewRepo) GetInterviewsByIDs(ids []uint) ([]model.Interview, error) {
	rows := make([]model.Interview, 0, len(ids))
	for _, id := range ids {
		if interview, ok := f.interviewByID[id]; ok && interview != nil {
			rows = append(rows, *interview)
		}
	}
	return rows, nil
}

type fakeLiveInterviewUserRepo struct {
	users       map[uint]*model.User
	updateCount int
}

func (f *fakeLiveInterviewUserRepo) GetByID(id uint) (*model.User, error) {
	if f.users == nil {
		return nil, fmt.Errorf("not found")
	}
	user, ok := f.users[id]
	if !ok || user == nil {
		return nil, fmt.Errorf("not found")
	}
	return user, nil
}

func (f *fakeLiveInterviewUserRepo) Update(user *model.User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	f.updateCount++
	if f.users == nil {
		f.users = map[uint]*model.User{}
	}
	copied := *user
	f.users[user.ID] = &copied
	return nil
}

func TestValidateJoinPermissionAccepted(t *testing.T) {
	invitation := &model.HumanInterviewInvitation{
		ID:              1,
		InvitationCode:  "ABCD1234",
		InitiatorUserID: 10,
		TargetUserID:    20,
		InitiatorRole:   "student",
		TargetRole:      "enterprise",
		StudentID:       10,
		StudentUUID:     "student-uuid",
		InviteeUserID:   20,
		InviteeUUID:     "invitee-uuid",
		InviteeRole:     "enterprise",
		Status:          invitationStatusAccepted,
		CreatedAt:       time.Now(),
	}

	uc := NewLiveInterviewUseCaseWithPorts(&fakeLiveInterviewRepo{}, &fakeLiveInterviewUserRepo{users: map[uint]*model.User{
		10: {ID: 10, UUID: "student-uuid", Role: "student"},
		20: {ID: 20, UUID: "invitee-uuid", Role: "enterprise"},
	}})
	perm, err := uc.ValidateJoinPermission(10, "student", "student-uuid", invitation, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if perm.ParticipantUUID != "student-uuid" {
		t.Fatalf("unexpected uuid: %s", perm.ParticipantUUID)
	}
	if perm.ShouldAutoAccept {
		t.Fatalf("accepted invitation should not require auto accept")
	}
}

func TestValidateJoinPermissionPendingCanAutoAccept(t *testing.T) {
	invitation := &model.HumanInterviewInvitation{
		ID:              1,
		InvitationCode:  "ABCD1234",
		InitiatorUserID: 10,
		TargetUserID:    20,
		InitiatorRole:   "student",
		TargetRole:      "enterprise",
		StudentID:       10,
		StudentUUID:     "student-uuid",
		InviteeUserID:   20,
		InviteeUUID:     "invitee-uuid",
		InviteeRole:     "enterprise",
		Status:          invitationStatusPending,
		CreatedAt:       time.Now(),
	}

	uc := NewLiveInterviewUseCaseWithPorts(&fakeLiveInterviewRepo{}, &fakeLiveInterviewUserRepo{users: map[uint]*model.User{
		10: {ID: 10, UUID: "student-uuid", Role: "student"},
		20: {ID: 20, UUID: "invitee-uuid", Role: "enterprise"},
	}})
	perm, err := uc.ValidateJoinPermission(10, "student", "student-uuid", invitation, "")
	if err == nil {
		if !perm.ShouldAutoAccept {
			t.Fatalf("pending invitation should auto accept for direct participant")
		}
		return
	}
	if !strings.Contains(err.Error(), "尚未被接受") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJoinInterviewTransitionsPendingToAcceptedWithoutInterviewStart(t *testing.T) {
	interviewID := uint(99)
	invitation := &model.HumanInterviewInvitation{
		ID:              2,
		InvitationCode:  "JOIN-CODE",
		InitiatorUserID: 10,
		TargetUserID:    20,
		InitiatorRole:   "student",
		TargetRole:      "enterprise",
		StudentID:       10,
		StudentUUID:     "student-uuid",
		InviteeUserID:   20,
		InviteeUUID:     "invitee-uuid",
		InviteeRole:     "enterprise",
		Status:          invitationStatusPending,
		InterviewID:     &interviewID,
		CreatedAt:       time.Now(),
	}
	interview := &model.Interview{
		ID:     interviewID,
		Status: interviewStatusPending,
	}

	repo := &fakeLiveInterviewRepo{
		invitationByID: map[uint]*model.HumanInterviewInvitation{2: invitation},
		interviewByID:  map[uint]*model.Interview{interviewID: interview},
	}
	uc := NewLiveInterviewUseCaseWithPorts(repo, &fakeLiveInterviewUserRepo{users: map[uint]*model.User{
		10: {ID: 10, UUID: "student-uuid", Role: "student"},
		20: {ID: 20, UUID: "invitee-uuid", Role: "enterprise"},
	}})

	result, err := uc.JoinInterview(10, "student", "student-uuid", 2, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Status != invitationStatusAccepted {
		t.Fatalf("unexpected invitation status: %s", result.Status)
	}
	if len(repo.updatedInvitations) < 1 {
		t.Fatalf("expected at least one invitation update, got: %d", len(repo.updatedInvitations))
	}
	lastInvitation := repo.updatedInvitations[len(repo.updatedInvitations)-1]
	if lastInvitation.Status != invitationStatusAccepted {
		t.Fatalf("unexpected persisted invitation status: %s", lastInvitation.Status)
	}
	if len(repo.updatedInterviews) != 1 {
		t.Fatalf("expected one interview update, got: %d", len(repo.updatedInterviews))
	}
	if repo.updatedInterviews[0].Status != interviewStatusPending {
		t.Fatalf("unexpected persisted interview status: %s", repo.updatedInterviews[0].Status)
	}
	if repo.updatedInterviews[0].InvitationCode == nil || *repo.updatedInterviews[0].InvitationCode != "JOIN-CODE" {
		t.Fatalf("expected interview invitation code to be synced")
	}
}

func TestStartLiveInterviewByInitiator(t *testing.T) {
	interviewID := uint(77)
	invitation := &model.HumanInterviewInvitation{
		ID:              5,
		InvitationCode:  "START-CODE",
		InitiatorUserID: 10,
		TargetUserID:    20,
		InitiatorRole:   "student",
		TargetRole:      "enterprise",
		StudentID:       10,
		StudentUUID:     "student-uuid",
		InviteeUserID:   20,
		InviteeUUID:     "invitee-uuid",
		InviteeRole:     "enterprise",
		Status:          invitationStatusAccepted,
		InterviewID:     &interviewID,
		StartThreshold:  2,
		CreatedAt:       time.Now(),
	}
	interview := &model.Interview{ID: interviewID, Status: interviewStatusPending}

	repo := &fakeLiveInterviewRepo{
		invitationByID: map[uint]*model.HumanInterviewInvitation{5: invitation},
		interviewByID:  map[uint]*model.Interview{interviewID: interview},
		participants: map[string]*model.HumanInterviewInvitationParticipant{
			participantKey(5, 20): {
				InvitationID:    5,
				UserID:          20,
				ParticipantRole: model.InvitationParticipantRoleInvitee,
				ResponseStatus:  model.InvitationParticipantStatusAccepted,
			},
		},
	}

	uc := NewLiveInterviewUseCaseWithPorts(repo, &fakeLiveInterviewUserRepo{users: map[uint]*model.User{
		10: {ID: 10, UUID: "student-uuid", Role: "student"},
		20: {ID: 20, UUID: "invitee-uuid", Role: "enterprise"},
	}})

	result, err := uc.StartLiveInterview(10, "student", 5)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if normalizeStatusValue(result.Status) != invitationStatusInProgress {
		t.Fatalf("unexpected invitation status: %s", result.Status)
	}
	if normalizeStatusValue(result.InterviewStatus) != interviewStatusInProgress {
		t.Fatalf("unexpected interview status: %s", result.InterviewStatus)
	}
	if len(repo.updatedInvitations) < 1 {
		t.Fatalf("expected invitation update")
	}
	if len(repo.updatedInterviews) < 1 {
		t.Fatalf("expected interview update")
	}
}

func TestStartLiveInterviewRejectedForNonInitiator(t *testing.T) {
	invitation := &model.HumanInterviewInvitation{
		ID:              6,
		InvitationCode:  "START-CODE-2",
		InitiatorUserID: 10,
		TargetUserID:    20,
		InitiatorRole:   "student",
		TargetRole:      "enterprise",
		Status:          invitationStatusAccepted,
		StartThreshold:  2,
		CreatedAt:       time.Now(),
	}

	repo := &fakeLiveInterviewRepo{invitationByID: map[uint]*model.HumanInterviewInvitation{6: invitation}}
	uc := NewLiveInterviewUseCaseWithPorts(repo, &fakeLiveInterviewUserRepo{users: map[uint]*model.User{
		10: {ID: 10, UUID: "student-uuid", Role: "student"},
		20: {ID: 20, UUID: "invitee-uuid", Role: "enterprise"},
	}})

	_, err := uc.StartLiveInterview(20, "enterprise", 6)
	if err == nil {
		t.Fatalf("expected permission error for non-initiator")
	}
	if !strings.Contains(err.Error(), "仅发起方可开始面试") {
		t.Fatalf("unexpected error: %v", err)
	}
}
