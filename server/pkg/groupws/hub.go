package groupws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	internalruntime "your-project/internal/runtime"

	"github.com/gorilla/websocket"
)

type GroupHub struct {
	mu          sync.RWMutex
	clients     map[string]*GroupClient
	roundStates map[string]*groupRoundState
}

type GroupClient struct {
	id                      string
	hub                     *GroupHub
	conn                    *websocket.Conn
	send                    chan []byte
	userID                  string
	roomID                  string
	groupTargetParticipants int
	groupStartThreshold     int
}

type GroupSignalMessage struct {
	Type         string      `json:"type"`
	UserID       string      `json:"user_id,omitempty"`
	SenderUserID string      `json:"sender_user_id,omitempty"`
	TargetUserID string      `json:"target_user_id,omitempty"`
	RoomID       string      `json:"interview_id,omitempty"`
	Data         interface{} `json:"data"`
	Timestamp    time.Time   `json:"timestamp"`
}

type groupRoundState struct {
	queue            []string
	currentSpeakerID string
	roundDurationSec int
	roundStartedAt   time.Time
}

var groupUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewGroupHub() *GroupHub {
	return &GroupHub{
		clients:     make(map[string]*GroupClient),
		roundStates: make(map[string]*groupRoundState),
	}
}

func (h *GroupHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	roomID := strings.TrimSpace(r.URL.Query().Get("room_id"))
	if roomID == "" {
		roomID = strings.TrimSpace(r.URL.Query().Get("interview_id"))
	}

	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	if roomID == "" {
		http.Error(w, "room_id is required", http.StatusBadRequest)
		return
	}

	conn, err := groupUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("group ws upgrade failed: %v", err)
		return
	}

	targetParticipants := parsePositiveInt(r.URL.Query().Get("group_target_participants"), internalruntime.DefaultGroupTargetParticipants)
	if targetParticipants < internalruntime.DefaultGroupTargetParticipants {
		targetParticipants = internalruntime.DefaultGroupTargetParticipants
	}

	startThreshold := parsePositiveInt(r.URL.Query().Get("group_start_threshold"), internalruntime.GroupStartThresholdForTesting)
	if startThreshold < internalruntime.GroupStartThresholdForTesting {
		startThreshold = internalruntime.GroupStartThresholdForTesting
	}
	if startThreshold > targetParticipants {
		startThreshold = targetParticipants
	}

	client := &GroupClient{
		id:                      fmt.Sprintf("group:%s:%s:%d", userID, roomID, time.Now().UnixNano()),
		hub:                     h,
		conn:                    conn,
		send:                    make(chan []byte, 256),
		userID:                  userID,
		roomID:                  roomID,
		groupTargetParticipants: targetParticipants,
		groupStartThreshold:     startThreshold,
	}

	h.register(client)
	go client.writePump()
	go client.readPump()
}

func (h *GroupHub) register(client *GroupClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.id] = client
}

func (h *GroupHub) unregister(client *GroupClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client.id]; !ok {
		return
	}
	delete(h.clients, client.id)
	close(client.send)
}

func (c *GroupClient) readPump() {
	defer func() {
		c.hub.unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(1024 * 1024)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("group ws read error: %v", err)
			}
			return
		}

		var msg GroupSignalMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Printf("group ws invalid message: %v", err)
			continue
		}

		msg.UserID = c.userID
		msg.SenderUserID = c.userID
		if strings.TrimSpace(msg.RoomID) == "" {
			msg.RoomID = c.roomID
		}
		msg.TargetUserID = strings.TrimSpace(msg.TargetUserID)
		msg.Timestamp = time.Now()

		if err := c.hub.routeMessage(c, &msg); err != nil {
			log.Printf("group ws route message failed: type=%s room=%s user=%s err=%v", msg.Type, msg.RoomID, c.userID, err)
		}
	}
}

func (c *GroupClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *GroupHub) routeMessage(sender *GroupClient, msg *GroupSignalMessage) error {
	switch strings.TrimSpace(msg.Type) {
	case "chat":
		return h.broadcastChat(sender, msg)
	case "offer", "answer", "candidate":
		targetUserID := strings.TrimSpace(msg.TargetUserID)
		if targetUserID == "" {
			return fmt.Errorf("target user id is required for %s", msg.Type)
		}
		if sender != nil && targetUserID == sender.userID {
			return fmt.Errorf("cannot send %s to self", msg.Type)
		}
		msg.UserID = sender.userID
		msg.SenderUserID = sender.userID
		return h.sendToTargetUser(msg.RoomID, targetUserID, msg)
	case "group_invite":
		return h.broadcastGroupInvite(sender, msg)
	case "group_start_vote":
		return h.broadcastGroupStartVote(sender, msg)
	case "group_claim_mic":
		return h.claimMicRound(sender, msg)
	case "group_round_next":
		return h.nextRound(sender, msg)
	case "group_round_sync_request":
		return h.syncRoundToRoom(msg.RoomID)
	default:
		return h.broadcastRawToRoom(msg.RoomID, msg)
	}
}

func (h *GroupHub) broadcastChat(sender *GroupClient, msg *GroupSignalMessage) error {
	if sender == nil || msg == nil {
		return nil
	}
	roomID := strings.TrimSpace(msg.RoomID)
	if roomID == "" {
		return fmt.Errorf("room id is required")
	}

	data := normalizeDataMap(msg.Data)
	rawText := strings.TrimSpace(asString(data["text"]))
	if rawText == "" {
		return nil
	}

	senderID := asUint(msg.UserID)
	if senderID == 0 {
		return fmt.Errorf("sender id is invalid")
	}

	senderName := strings.TrimSpace(asString(data["sender_name"]))
	if senderName == "" {
		senderName = fmt.Sprintf("用户%s", sender.userID)
	}

	clients := h.roomClients(roomID)
	for _, receiver := range clients {
		receiverID := asUint(receiver.userID)
		if receiverID == 0 {
			continue
		}

		displayText := internalruntime.FormatChatPerspective(senderName, senderID, receiverID, rawText)
		if displayText == "" {
			continue
		}

		outData := cloneStringAnyMap(data)
		outData["sender_name"] = senderName
		outData["raw_text"] = rawText
		outData["text"] = displayText
		outData["display_text"] = displayText

		out := GroupSignalMessage{
			Type:         "chat",
			UserID:       msg.UserID,
			SenderUserID: msg.UserID,
			RoomID:       roomID,
			Data:         outData,
			Timestamp:    time.Now(),
		}
		h.sendStructured(receiver, out)
	}

	return nil
}

func (h *GroupHub) broadcastGroupInvite(sender *GroupClient, msg *GroupSignalMessage) error {
	data := normalizeDataMap(msg.Data)
	senderName := strings.TrimSpace(asString(data["sender_name"]))
	if senderName == "" {
		senderName = fmt.Sprintf("用户%s", sender.userID)
	}

	outData := cloneStringAnyMap(data)
	outData["sender_name"] = senderName
	outData["target_participants"] = sender.groupTargetParticipants
	outData["start_threshold"] = sender.groupStartThreshold

	return h.broadcastRawToRoom(msg.RoomID, &GroupSignalMessage{
		Type:         "group_invite",
		UserID:       msg.UserID,
		SenderUserID: msg.UserID,
		RoomID:       msg.RoomID,
		Data:         outData,
		Timestamp:    time.Now(),
	})
}

func (h *GroupHub) broadcastGroupStartVote(sender *GroupClient, msg *GroupSignalMessage) error {
	senderID := asUint(msg.UserID)
	if senderID == 0 {
		return fmt.Errorf("sender id is invalid")
	}

	state := internalruntime.GetGroupRoomStore().CastStartVote(
		msg.RoomID,
		senderID,
		sender.groupTargetParticipants,
		sender.groupStartThreshold,
	)

	statusData := map[string]interface{}{
		"ready_count":         state.ReadyCount,
		"start_threshold":     state.StartThreshold,
		"target_participants": state.TargetParticipants,
		"can_start":           state.CanStart,
		"started":             state.Started,
		"voted_user_ids":      state.VotedUserIDs,
	}

	if err := h.broadcastRawToRoom(msg.RoomID, &GroupSignalMessage{
		Type:         "group_start_status",
		UserID:       msg.UserID,
		SenderUserID: msg.UserID,
		RoomID:       msg.RoomID,
		Data:         statusData,
		Timestamp:    time.Now(),
	}); err != nil {
		return err
	}

	if state.JustStarted {
		startData := map[string]interface{}{
			"message":             "人数达到测试开考阈值，已进入群面流程",
			"ready_count":         state.ReadyCount,
			"start_threshold":     state.StartThreshold,
			"target_participants": state.TargetParticipants,
		}
		if err := h.broadcastRawToRoom(msg.RoomID, &GroupSignalMessage{
			Type:         "group_start",
			UserID:       msg.UserID,
			SenderUserID: msg.UserID,
			RoomID:       msg.RoomID,
			Data:         startData,
			Timestamp:    time.Now(),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (h *GroupHub) claimMicRound(sender *GroupClient, msg *GroupSignalMessage) error {
	roomID := strings.TrimSpace(msg.RoomID)
	if roomID == "" {
		return fmt.Errorf("room id is required")
	}

	data := normalizeDataMap(msg.Data)
	roundDuration := parsePositiveInt(asString(data["round_duration_sec"]), 90)

	h.mu.Lock()
	state := h.getOrCreateRoundStateLocked(roomID)
	if roundDuration > 0 {
		state.roundDurationSec = roundDuration
	}
	if state.roundDurationSec <= 0 {
		state.roundDurationSec = 90
	}

	alreadyQueued := state.currentSpeakerID == sender.userID
	if !alreadyQueued {
		for _, uid := range state.queue {
			if uid == sender.userID {
				alreadyQueued = true
				break
			}
		}
	}
	if !alreadyQueued {
		state.queue = append(state.queue, sender.userID)
	}

	if state.currentSpeakerID == "" {
		h.advanceRoundLocked(state)
	}
	payload := h.buildRoundPayloadLocked(state)
	h.mu.Unlock()

	return h.broadcastRawToRoom(roomID, &GroupSignalMessage{
		Type:         "group_round_sync",
		UserID:       msg.UserID,
		SenderUserID: msg.UserID,
		RoomID:       roomID,
		Data:         payload,
		Timestamp:    time.Now(),
	})
}

func (h *GroupHub) nextRound(sender *GroupClient, msg *GroupSignalMessage) error {
	roomID := strings.TrimSpace(msg.RoomID)
	if roomID == "" {
		return fmt.Errorf("room id is required")
	}

	h.mu.Lock()
	state := h.getOrCreateRoundStateLocked(roomID)
	h.advanceRoundLocked(state)
	payload := h.buildRoundPayloadLocked(state)
	h.mu.Unlock()

	return h.broadcastRawToRoom(roomID, &GroupSignalMessage{
		Type:         "group_round_sync",
		UserID:       msg.UserID,
		SenderUserID: msg.UserID,
		RoomID:       roomID,
		Data:         payload,
		Timestamp:    time.Now(),
	})
}

func (h *GroupHub) syncRoundToRoom(roomID string) error {
	targetRoom := strings.TrimSpace(roomID)
	if targetRoom == "" {
		return nil
	}

	h.mu.Lock()
	state := h.getOrCreateRoundStateLocked(targetRoom)
	payload := h.buildRoundPayloadLocked(state)
	h.mu.Unlock()

	return h.broadcastRawToRoom(targetRoom, &GroupSignalMessage{
		Type:      "group_round_sync",
		RoomID:    targetRoom,
		Data:      payload,
		Timestamp: time.Now(),
	})
}

func (h *GroupHub) sendToTargetUser(roomID, targetUserID string, msg *GroupSignalMessage) error {
	targetRoom := strings.TrimSpace(roomID)
	targetUser := strings.TrimSpace(targetUserID)
	if msg == nil {
		return nil
	}
	if targetRoom == "" {
		return fmt.Errorf("room id is required")
	}
	if targetUser == "" {
		return fmt.Errorf("target user id is required")
	}
	if strings.TrimSpace(msg.SenderUserID) != "" && targetUser == strings.TrimSpace(msg.SenderUserID) {
		return fmt.Errorf("cannot send %s to self", msg.Type)
	}

	msg.TargetUserID = targetUser
	if strings.TrimSpace(msg.SenderUserID) == "" {
		msg.SenderUserID = strings.TrimSpace(msg.UserID)
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	encoded, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	matched := false
	for _, client := range h.roomClients(targetRoom) {
		if client.userID != targetUser {
			continue
		}
		matched = true
		select {
		case client.send <- encoded:
		default:
			log.Printf("group ws client send buffer full, user=%s room=%s", client.userID, client.roomID)
		}
	}

	if !matched {
		return fmt.Errorf("target user %s is not connected in room %s", targetUser, targetRoom)
	}
	return nil
}

func (h *GroupHub) advanceRoundLocked(state *groupRoundState) {
	if state == nil {
		return
	}

	if len(state.queue) == 0 {
		state.currentSpeakerID = ""
		state.roundStartedAt = time.Time{}
		return
	}

	state.currentSpeakerID = state.queue[0]
	state.queue = append([]string{}, state.queue[1:]...)
	if state.roundDurationSec <= 0 {
		state.roundDurationSec = 90
	}
	state.roundStartedAt = time.Now()
}

func (h *GroupHub) buildRoundPayloadLocked(state *groupRoundState) map[string]interface{} {
	if state == nil {
		return map[string]interface{}{}
	}

	countdown := 0
	startedAt := ""
	if !state.roundStartedAt.IsZero() && state.roundDurationSec > 0 {
		startedAt = state.roundStartedAt.Format(time.RFC3339)
		elapsed := int(time.Since(state.roundStartedAt).Seconds())
		countdown = state.roundDurationSec - elapsed
		if countdown < 0 {
			countdown = 0
		}
	}

	queue := append([]string(nil), state.queue...)
	return map[string]interface{}{
		"queue_user_ids":          queue,
		"current_speaker_user_id": state.currentSpeakerID,
		"round_duration_sec":      state.roundDurationSec,
		"round_started_at":        startedAt,
		"countdown_seconds":       countdown,
	}
}

func (h *GroupHub) getOrCreateRoundStateLocked(roomID string) *groupRoundState {
	if state, ok := h.roundStates[roomID]; ok {
		return state
	}
	state := &groupRoundState{
		queue:            make([]string, 0),
		currentSpeakerID: "",
		roundDurationSec: 90,
	}
	h.roundStates[roomID] = state
	return state
}

func (h *GroupHub) roomClients(roomID string) []*GroupClient {
	target := strings.TrimSpace(roomID)
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*GroupClient, 0)
	for _, client := range h.clients {
		if target != "" && client.roomID != target {
			continue
		}
		clients = append(clients, client)
	}

	sort.Slice(clients, func(i, j int) bool {
		return clients[i].id < clients[j].id
	})
	return clients
}

func (h *GroupHub) broadcastRawToRoom(roomID string, msg *GroupSignalMessage) error {
	if msg == nil {
		return nil
	}
	encoded, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	clients := h.roomClients(roomID)
	for _, client := range clients {
		select {
		case client.send <- encoded:
		default:
			log.Printf("group ws client send buffer full, user=%s room=%s", client.userID, client.roomID)
		}
	}
	return nil
}

func (h *GroupHub) sendStructured(client *GroupClient, msg GroupSignalMessage) {
	if client == nil {
		return
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case client.send <- payload:
	default:
		log.Printf("group ws client send buffer full, user=%s room=%s", client.userID, client.roomID)
	}
}

func normalizeDataMap(data interface{}) map[string]interface{} {
	if typed, ok := data.(map[string]interface{}); ok && typed != nil {
		return typed
	}
	return map[string]interface{}{}
}

func cloneStringAnyMap(src map[string]interface{}) map[string]interface{} {
	if len(src) == 0 {
		return map[string]interface{}{}
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func asString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return ""
	}
}

func asUint(value interface{}) uint {
	switch v := value.(type) {
	case uint:
		return v
	case uint64:
		return uint(v)
	case int:
		if v > 0 {
			return uint(v)
		}
	case int64:
		if v > 0 {
			return uint(v)
		}
	case float64:
		if v > 0 {
			return uint(v)
		}
	case string:
		parsed, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64)
		if err == nil {
			return uint(parsed)
		}
	}
	return 0
}

func (h *GroupHub) GetClientsByUserID(userID string) []*GroupClient {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*GroupClient, 0)
	for _, client := range h.clients {
		if client.userID == userID {
			clients = append(clients, client)
		}
	}
	return clients
}

func (c *GroupClient) GetRoomID() string {
	return c.roomID
}

var globalGroupHub = NewGroupHub()

func GetGroupHub() *GroupHub {
	return globalGroupHub
}
