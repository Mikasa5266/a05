package service

import (
	"fmt"
	"strings"
	"testing"

	"your-project/model"
)

type fakeLiveInterviewRepo struct {
	invitationByID       map[uint]*model.HumanInterviewInvitation
	interviewByID        map[uint]*model.Interview
	invitationsByInvitee map[uint][]model.HumanInterviewInvitation

	updatedInvitations []*model.HumanInterviewInvitation
	updatedInterviews  []*model.Interview
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
		ID:             1,
		InvitationCode: "ABCD1234",
		StudentID:      10,
		StudentUUID:    "student-uuid",
		InviteeUserID:  20,
		InviteeUUID:    "invitee-uuid",
		InviteeRole:    "enterprise",
		Status:         invitationStatusAccepted,
	}

	uc := NewLiveInterviewUseCaseWithPorts(&fakeLiveInterviewRepo{}, &fakeLiveInterviewUserRepo{})
	uuid, err := uc.ValidateJoinPermission(10, "student", "student-uuid", invitation, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if uuid != "student-uuid" {
		t.Fatalf("unexpected uuid: %s", uuid)
	}
}

func TestValidateJoinPermissionPendingRejected(t *testing.T) {
	invitation := &model.HumanInterviewInvitation{
		ID:             1,
		InvitationCode: "ABCD1234",
		StudentID:      10,
		StudentUUID:    "student-uuid",
		InviteeUserID:  20,
		InviteeUUID:    "invitee-uuid",
		InviteeRole:    "enterprise",
		Status:         invitationStatusPending,
	}

	uc := NewLiveInterviewUseCaseWithPorts(&fakeLiveInterviewRepo{}, &fakeLiveInterviewUserRepo{})
	_, err := uc.ValidateJoinPermission(10, "student", "student-uuid", invitation, "")
	if err == nil {
		t.Fatalf("expected pending invitation to be rejected")
	}
	if !strings.Contains(err.Error(), "尚未被接受") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJoinInterviewTransitionsAcceptedToInProgress(t *testing.T) {
	interviewID := uint(99)
	invitation := &model.HumanInterviewInvitation{
		ID:             2,
		InvitationCode: "JOIN-CODE",
		StudentID:      10,
		StudentUUID:    "student-uuid",
		InviteeUserID:  20,
		InviteeUUID:    "invitee-uuid",
		InviteeRole:    "enterprise",
		Status:         invitationStatusAccepted,
		InterviewID:    &interviewID,
	}
	interview := &model.Interview{
		ID:     interviewID,
		Status: interviewStatusPending,
	}

	repo := &fakeLiveInterviewRepo{
		invitationByID: map[uint]*model.HumanInterviewInvitation{2: invitation},
		interviewByID:  map[uint]*model.Interview{interviewID: interview},
	}
	uc := NewLiveInterviewUseCaseWithPorts(repo, &fakeLiveInterviewUserRepo{})

	result, err := uc.JoinInterview(10, "student", "student-uuid", 2, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Status != invitationStatusInProgress {
		t.Fatalf("unexpected invitation status: %s", result.Status)
	}
	if len(repo.updatedInvitations) != 1 {
		t.Fatalf("expected one invitation update, got: %d", len(repo.updatedInvitations))
	}
	if repo.updatedInvitations[0].Status != invitationStatusInProgress {
		t.Fatalf("unexpected persisted invitation status: %s", repo.updatedInvitations[0].Status)
	}
	if len(repo.updatedInterviews) != 1 {
		t.Fatalf("expected one interview update, got: %d", len(repo.updatedInterviews))
	}
	if repo.updatedInterviews[0].Status != interviewStatusInProgress {
		t.Fatalf("unexpected persisted interview status: %s", repo.updatedInterviews[0].Status)
	}
	if repo.updatedInterviews[0].InvitationCode == nil || *repo.updatedInterviews[0].InvitationCode != "JOIN-CODE" {
		t.Fatalf("expected interview invitation code to be synced")
	}
}
