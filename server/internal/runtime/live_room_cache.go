package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"your-project/internal/model"
)

type roomCache struct {
	audioTranscripts []model.ReportAudioTranscript
	chatByReceiver   map[uint][]model.ReportChatMessage
}

type LiveRoomStore struct {
	mu    sync.RWMutex
	rooms map[string]*roomCache
}

var globalLiveRoomStore = NewLiveRoomStore()

func NewLiveRoomStore() *LiveRoomStore {
	return &LiveRoomStore{rooms: make(map[string]*roomCache)}
}

func GetLiveRoomStore() *LiveRoomStore {
	return globalLiveRoomStore
}

func InterviewRoomCacheKey(interviewID uint) string {
	if interviewID == 0 {
		return ""
	}
	return "interview-" + strconv.FormatUint(uint64(interviewID), 10)
}

func NormalizeRoomCacheKey(roomID string, interviewID uint) string {
	trimmed := strings.TrimSpace(roomID)
	if trimmed != "" {
		return trimmed
	}
	return InterviewRoomCacheKey(interviewID)
}

func FormatChatPerspective(senderName string, senderID, receiverID uint, rawText string) string {
	text := strings.TrimSpace(rawText)
	if text == "" {
		return ""
	}

	if receiverID == senderID {
		return "我: " + text
	}

	name := strings.TrimSpace(senderName)
	if name == "" {
		name = fmt.Sprintf("用户%d", senderID)
	}
	return fmt.Sprintf("[%s]: %s", name, text)
}

func (s *LiveRoomStore) AppendAudioTranscript(roomID string, speakerID uint, content string) {
	key := strings.TrimSpace(roomID)
	text := strings.TrimSpace(content)
	if key == "" || speakerID == 0 || text == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	room := s.getOrCreateLocked(key)
	room.audioTranscripts = append(room.audioTranscripts, model.ReportAudioTranscript{
		SpeakerID: speakerID,
		Content:   text,
	})
	if len(room.audioTranscripts) > 3000 {
		room.audioTranscripts = append([]model.ReportAudioTranscript(nil), room.audioTranscripts[len(room.audioTranscripts)-3000:]...)
	}
}

func (s *LiveRoomStore) AppendChatMessageForReceiver(roomID string, receiverID, senderID uint, content string, sentAt time.Time) {
	key := strings.TrimSpace(roomID)
	text := strings.TrimSpace(content)
	if key == "" || receiverID == 0 || senderID == 0 || text == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	room := s.getOrCreateLocked(key)
	message := model.ReportChatMessage{
		SenderID: senderID,
		Content:  text,
	}
	if !sentAt.IsZero() {
		t := sentAt
		message.SentAt = &t
	}

	room.chatByReceiver[receiverID] = append(room.chatByReceiver[receiverID], message)
	if len(room.chatByReceiver[receiverID]) > 5000 {
		room.chatByReceiver[receiverID] = append([]model.ReportChatMessage(nil), room.chatByReceiver[receiverID][len(room.chatByReceiver[receiverID])-5000:]...)
	}
}

func (s *LiveRoomStore) Snapshot(roomID string) ([]model.ReportAudioTranscript, map[uint][]model.ReportChatMessage) {
	key := strings.TrimSpace(roomID)
	if key == "" {
		return []model.ReportAudioTranscript{}, map[uint][]model.ReportChatMessage{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	room, ok := s.rooms[key]
	if !ok {
		return []model.ReportAudioTranscript{}, map[uint][]model.ReportChatMessage{}
	}

	audio := append([]model.ReportAudioTranscript(nil), room.audioTranscripts...)
	chat := make(map[uint][]model.ReportChatMessage, len(room.chatByReceiver))
	for uid, messages := range room.chatByReceiver {
		chat[uid] = append([]model.ReportChatMessage(nil), messages...)
	}
	return audio, chat
}

func (s *LiveRoomStore) getOrCreateLocked(roomID string) *roomCache {
	if room, ok := s.rooms[roomID]; ok {
		return room
	}
	room := &roomCache{
		audioTranscripts: make([]model.ReportAudioTranscript, 0),
		chatByReceiver:   make(map[uint][]model.ReportChatMessage),
	}
	s.rooms[roomID] = room
	return room
}
