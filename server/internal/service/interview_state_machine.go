package service

import (
	"fmt"
	"strings"
)

const (
	interviewStatusPending    = "pending"
	interviewStatusInProgress = "in_progress"
	interviewStatusCompleted  = "completed"
)

const (
	invitationStatusPending    = "pending"
	invitationStatusAccepted   = "accepted"
	invitationStatusRejected   = "rejected"
	invitationStatusInProgress = "in_progress"
	invitationStatusCompleted  = "completed"
	invitationStatusCancelled  = "cancelled"
)

func normalizeStatusValue(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func transitionInterviewStatus(current, target string) (string, error) {
	from := normalizeStatusValue(current)
	to := normalizeStatusValue(target)

	if from == "" || to == "" {
		return "", fmt.Errorf("面试状态不能为空")
	}
	if from == to {
		return to, nil
	}

	switch from {
	case interviewStatusPending:
		if to == interviewStatusInProgress || to == interviewStatusCompleted {
			return to, nil
		}
	case interviewStatusInProgress:
		if to == interviewStatusCompleted {
			return to, nil
		}
	case interviewStatusCompleted:
		if to == interviewStatusCompleted {
			return to, nil
		}
	}

	return "", fmt.Errorf("非法面试状态流转: %s -> %s", current, target)
}

func transitionInvitationStatus(current, target string) (string, error) {
	from := normalizeStatusValue(current)
	to := normalizeStatusValue(target)

	if from == "" || to == "" {
		return "", fmt.Errorf("邀请状态不能为空")
	}
	if from == to {
		return to, nil
	}

	switch from {
	case invitationStatusPending:
		if to == invitationStatusAccepted || to == invitationStatusRejected || to == invitationStatusCancelled || to == invitationStatusCompleted {
			return to, nil
		}
	case invitationStatusAccepted:
		if to == invitationStatusInProgress || to == invitationStatusCompleted || to == invitationStatusCancelled {
			return to, nil
		}
	case invitationStatusInProgress:
		if to == invitationStatusCompleted || to == invitationStatusCancelled {
			return to, nil
		}
	case invitationStatusRejected, invitationStatusCancelled, invitationStatusCompleted:
		// Terminal states only allow idempotent writes handled above.
	}

	return "", fmt.Errorf("非法邀请状态流转: %s -> %s", current, target)
}

func isInvitationJoinableStatus(status string) bool {
	normalized := normalizeStatusValue(status)
	return normalized == invitationStatusPending || normalized == invitationStatusAccepted || normalized == invitationStatusInProgress
}
