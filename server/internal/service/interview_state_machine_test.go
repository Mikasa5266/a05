package service

import "testing"

func TestTransitionInterviewStatus(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		want    string
		wantErr bool
	}{
		{name: "pending to in_progress", from: interviewStatusPending, to: interviewStatusInProgress, want: interviewStatusInProgress},
		{name: "pending to completed", from: interviewStatusPending, to: interviewStatusCompleted, want: interviewStatusCompleted},
		{name: "in_progress to completed", from: interviewStatusInProgress, to: interviewStatusCompleted, want: interviewStatusCompleted},
		{name: "completed idempotent", from: interviewStatusCompleted, to: interviewStatusCompleted, want: interviewStatusCompleted},
		{name: "invalid rollback", from: interviewStatusCompleted, to: interviewStatusInProgress, wantErr: true},
		{name: "invalid reverse", from: interviewStatusInProgress, to: interviewStatusPending, wantErr: true},
		{name: "empty from", from: "", to: interviewStatusInProgress, wantErr: true},
		{name: "empty to", from: interviewStatusPending, to: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transitionInterviewStatus(tt.from, tt.to)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected status: got=%s want=%s", got, tt.want)
			}
		})
	}
}

func TestTransitionInvitationStatus(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		want    string
		wantErr bool
	}{
		{name: "pending to accepted", from: invitationStatusPending, to: invitationStatusAccepted, want: invitationStatusAccepted},
		{name: "pending to rejected", from: invitationStatusPending, to: invitationStatusRejected, want: invitationStatusRejected},
		{name: "accepted to in_progress", from: invitationStatusAccepted, to: invitationStatusInProgress, want: invitationStatusInProgress},
		{name: "in_progress to completed", from: invitationStatusInProgress, to: invitationStatusCompleted, want: invitationStatusCompleted},
		{name: "accepted idempotent", from: invitationStatusAccepted, to: invitationStatusAccepted, want: invitationStatusAccepted},
		{name: "terminal to accepted invalid", from: invitationStatusRejected, to: invitationStatusAccepted, wantErr: true},
		{name: "completed to in_progress invalid", from: invitationStatusCompleted, to: invitationStatusInProgress, wantErr: true},
		{name: "accepted to pending invalid", from: invitationStatusAccepted, to: invitationStatusPending, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transitionInvitationStatus(tt.from, tt.to)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected status: got=%s want=%s", got, tt.want)
			}
		})
	}
}

func TestIsInvitationJoinableStatus(t *testing.T) {
	if !isInvitationJoinableStatus(invitationStatusAccepted) {
		t.Fatalf("accepted should be joinable")
	}
	if !isInvitationJoinableStatus(invitationStatusInProgress) {
		t.Fatalf("in_progress should be joinable")
	}
	if !isInvitationJoinableStatus(" ACCEPTED ") {
		t.Fatalf("normalized accepted should be joinable")
	}
	if isInvitationJoinableStatus(invitationStatusPending) {
		t.Fatalf("pending should not be joinable")
	}
	if isInvitationJoinableStatus(invitationStatusRejected) {
		t.Fatalf("rejected should not be joinable")
	}
}
