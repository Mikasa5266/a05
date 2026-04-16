package runtime

import (
	"sort"
	"strings"
	"sync"
)

const (
	DefaultGroupTargetParticipants = 4
	GroupStartThresholdForTesting  = 2
)

type GroupStartState struct {
	ReadyCount         int    `json:"ready_count"`
	StartThreshold     int    `json:"start_threshold"`
	TargetParticipants int    `json:"target_participants"`
	CanStart           bool   `json:"can_start"`
	Started            bool   `json:"started"`
	JustStarted        bool   `json:"just_started"`
	VotedUserIDs       []uint `json:"voted_user_ids"`
}

type groupRoomCache struct {
	startVotes map[uint]struct{}
	started    bool
}

type GroupRoomStore struct {
	mu    sync.RWMutex
	rooms map[string]*groupRoomCache
}

var globalGroupRoomStore = NewGroupRoomStore()

func NewGroupRoomStore() *GroupRoomStore {
	return &GroupRoomStore{rooms: make(map[string]*groupRoomCache)}
}

func GetGroupRoomStore() *GroupRoomStore {
	return globalGroupRoomStore
}

func (s *GroupRoomStore) CastStartVote(roomID string, userID uint, targetParticipants, startThreshold int) GroupStartState {
	key := strings.TrimSpace(roomID)
	if key == "" || userID == 0 {
		return GroupStartState{}
	}

	target := targetParticipants
	if target < 2 {
		target = DefaultGroupTargetParticipants
	}
	threshold := startThreshold
	if threshold < 1 {
		threshold = GroupStartThresholdForTesting
	}
	if threshold > target {
		threshold = target
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	room := s.getOrCreateLocked(key)
	room.startVotes[userID] = struct{}{}

	voteIDs := make([]uint, 0, len(room.startVotes))
	for uid := range room.startVotes {
		voteIDs = append(voteIDs, uid)
	}
	sort.Slice(voteIDs, func(i, j int) bool { return voteIDs[i] < voteIDs[j] })

	readyCount := len(voteIDs)
	canStart := readyCount >= threshold
	justStarted := false
	if canStart && !room.started {
		room.started = true
		justStarted = true
	}

	return GroupStartState{
		ReadyCount:         readyCount,
		StartThreshold:     threshold,
		TargetParticipants: target,
		CanStart:           canStart,
		Started:            room.started,
		JustStarted:        justStarted,
		VotedUserIDs:       voteIDs,
	}
}

func (s *GroupRoomStore) getOrCreateLocked(roomID string) *groupRoomCache {
	if room, ok := s.rooms[roomID]; ok {
		return room
	}
	room := &groupRoomCache{startVotes: make(map[uint]struct{})}
	s.rooms[roomID] = room
	return room
}
