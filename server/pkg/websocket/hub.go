package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	internalruntime "your-project/internal/runtime"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[string]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Client struct {
	id                      string
	hub                     *Hub
	conn                    *websocket.Conn
	send                    chan []byte
	userID                  string
	interviewID             string
	groupTargetParticipants int
	groupStartThreshold     int
}

type Message struct {
	Type        string      `json:"type"`
	UserID      string      `json:"user_id,omitempty"`
	InterviewID string      `json:"interview_id,omitempty"`
	Data        interface{} `json:"data"`
	Timestamp   time.Time   `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.id] = client
			h.mu.Unlock()
			log.Printf("Client %s registered in room %s", client.userID, client.interviewID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				close(client.send)
				h.mu.Unlock()
				log.Printf("Client %s unregistered from room %s", client.userID, client.interviewID)
			} else {
				h.mu.Unlock()
			}

		case message := <-h.broadcast:
			var msg Message
			targetRoom := ""
			if err := json.Unmarshal(message, &msg); err == nil {
				targetRoom = strings.TrimSpace(msg.InterviewID)
			}
			h.broadcastRoomBytes(targetRoom, message)
		}
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	interviewID := strings.TrimSpace(r.URL.Query().Get("interview_id"))

	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	groupTargetParticipants := parsePositiveInt(r.URL.Query().Get("group_target_participants"), internalruntime.DefaultGroupTargetParticipants)
	if groupTargetParticipants < internalruntime.DefaultGroupTargetParticipants {
		groupTargetParticipants = internalruntime.DefaultGroupTargetParticipants
	}
	groupStartThreshold := parsePositiveInt(r.URL.Query().Get("group_start_threshold"), internalruntime.GroupStartThresholdForTesting)
	if groupStartThreshold < internalruntime.GroupStartThresholdForTesting {
		groupStartThreshold = internalruntime.GroupStartThresholdForTesting
	}

	client := &Client{
		id:                      fmt.Sprintf("%s:%s:%d", userID, interviewID, time.Now().UnixNano()),
		hub:                     h,
		conn:                    conn,
		send:                    make(chan []byte, 256),
		userID:                  userID,
		interviewID:             interviewID,
		groupTargetParticipants: groupTargetParticipants,
		groupStartThreshold:     groupStartThreshold,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// WebRTC SDP/ICE signaling payloads can exceed 512 bytes easily.
	// Keep a generous cap to avoid false disconnects while still bounded.
	c.conn.SetReadLimit(1024 * 1024)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		msg.Timestamp = time.Now()
		msg.UserID = c.userID
		if strings.TrimSpace(msg.InterviewID) == "" {
			msg.InterviewID = c.interviewID
		}

		switch strings.TrimSpace(msg.Type) {
		case "chat":
			if err := c.hub.BroadcastChatMessage(c, &msg); err != nil {
				log.Printf("broadcast chat failed: room=%s user=%s err=%v", msg.InterviewID, c.userID, err)
			}
			continue
		case "group_invite":
			if err := c.hub.BroadcastGroupInvite(c, &msg); err != nil {
				log.Printf("broadcast group invite failed: room=%s user=%s err=%v", msg.InterviewID, c.userID, err)
			}
			continue
		case "group_start_vote":
			if err := c.hub.BroadcastGroupStartVote(c, &msg); err != nil {
				log.Printf("broadcast group start vote failed: room=%s user=%s err=%v", msg.InterviewID, c.userID, err)
			}
			continue
		}

		processedMsg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
			continue
		}

		c.hub.broadcast <- processedMsg
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Hub) BroadcastChatMessage(sender *Client, msg *Message) error {
	if sender == nil || msg == nil {
		return nil
	}
	roomID := strings.TrimSpace(msg.InterviewID)
	if roomID == "" {
		return fmt.Errorf("room id is required")
	}

	data := normalizeMessageData(msg.Data)
	rawText := strings.TrimSpace(asString(data["text"]))
	if rawText == "" {
		return nil
	}
	senderID := asUintFromAny(msg.UserID)
	if senderID == 0 {
		return fmt.Errorf("sender id is invalid")
	}

	senderName := strings.TrimSpace(asString(data["sender_name"]))
	if senderName == "" {
		senderName = fmt.Sprintf("用户%d", senderID)
	}
	sentAt := msg.Timestamp
	if sentAt.IsZero() {
		sentAt = time.Now()
	}

	interviewCacheKey := internalruntime.InterviewRoomCacheKey(asUintFromAny(data["interview_id"]))
	clients := h.roomClients(roomID)
	for _, receiver := range clients {
		receiverID := asUintFromAny(receiver.userID)
		if receiverID == 0 {
			continue
		}

		displayText := internalruntime.FormatChatPerspective(senderName, senderID, receiverID, rawText)
		if displayText == "" {
			continue
		}

		internalruntime.GetLiveRoomStore().AppendChatMessageForReceiver(roomID, receiverID, senderID, displayText, sentAt)
		if interviewCacheKey != "" {
			internalruntime.GetLiveRoomStore().AppendChatMessageForReceiver(interviewCacheKey, receiverID, senderID, displayText, sentAt)
		}

		outData := cloneStringAnyMap(data)
		outData["sender_name"] = senderName
		outData["raw_text"] = rawText
		outData["text"] = displayText
		outData["display_text"] = displayText

		outMsg := Message{
			Type:        "chat",
			UserID:      msg.UserID,
			InterviewID: roomID,
			Data:        outData,
			Timestamp:   sentAt,
		}

		encoded, err := json.Marshal(outMsg)
		if err != nil {
			continue
		}
		h.sendToClient(receiver, encoded)
	}

	return nil
}

func (h *Hub) BroadcastGroupInvite(sender *Client, msg *Message) error {
	if sender == nil || msg == nil {
		return nil
	}
	roomID := strings.TrimSpace(msg.InterviewID)
	if roomID == "" {
		return fmt.Errorf("room id is required")
	}

	data := normalizeMessageData(msg.Data)
	senderName := strings.TrimSpace(asString(data["sender_name"]))
	if senderName == "" {
		senderName = fmt.Sprintf("用户%s", sender.userID)
	}

	targetParticipants := sender.groupTargetParticipants
	if targetParticipants < internalruntime.DefaultGroupTargetParticipants {
		targetParticipants = internalruntime.DefaultGroupTargetParticipants
	}
	startThreshold := sender.groupStartThreshold
	if startThreshold < internalruntime.GroupStartThresholdForTesting {
		startThreshold = internalruntime.GroupStartThresholdForTesting
	}

	outData := cloneStringAnyMap(data)
	outData["sender_name"] = senderName
	outData["target_participants"] = targetParticipants
	outData["start_threshold"] = startThreshold

	return h.broadcastStructuredToRoom(roomID, Message{
		Type:        "group_invite",
		UserID:      msg.UserID,
		InterviewID: roomID,
		Data:        outData,
		Timestamp:   time.Now(),
	})
}

func (h *Hub) BroadcastGroupStartVote(sender *Client, msg *Message) error {
	if sender == nil || msg == nil {
		return nil
	}
	roomID := strings.TrimSpace(msg.InterviewID)
	if roomID == "" {
		return fmt.Errorf("room id is required")
	}

	senderID := asUintFromAny(msg.UserID)
	if senderID == 0 {
		return fmt.Errorf("sender id is invalid")
	}

	state := internalruntime.GetLiveRoomStore().CastStartVote(roomID, senderID, sender.groupTargetParticipants, sender.groupStartThreshold)
	statusData := map[string]interface{}{
		"ready_count":         state.ReadyCount,
		"start_threshold":     state.StartThreshold,
		"target_participants": state.TargetParticipants,
		"can_start":           state.CanStart,
		"started":             state.Started,
		"voted_user_ids":      state.VotedUserIDs,
	}

	if err := h.broadcastStructuredToRoom(roomID, Message{
		Type:        "group_start_status",
		UserID:      msg.UserID,
		InterviewID: roomID,
		Data:        statusData,
		Timestamp:   time.Now(),
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
		if err := h.broadcastStructuredToRoom(roomID, Message{
			Type:        "group_start",
			UserID:      msg.UserID,
			InterviewID: roomID,
			Data:        startData,
			Timestamp:   time.Now(),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetInterviewID() string {
	return c.interviewID
}

func (h *Hub) GetClientsByUserID(userID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var clients []*Client
	for _, client := range h.clients {
		if client.userID == userID {
			clients = append(clients, client)
		}
	}
	return clients
}

func (h *Hub) SendToUser(userID string, messageType string, data interface{}) error {
	h.mu.RLock()
	clients := make([]*Client, 0)
	for _, client := range h.clients {
		if client.userID == userID {
			clients = append(clients, client)
		}
	}
	h.mu.RUnlock()

	if len(clients) == 0 {
		return fmt.Errorf("user %s not connected", userID)
	}

	msg := Message{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
	}

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	for _, client := range clients {
		if ok := h.sendToClient(client, jsonMsg); !ok {
			return fmt.Errorf("client send buffer full")
		}
	}

	return nil
}

func (h *Hub) BroadcastToInterview(interviewID string, messageType string, data interface{}) error {
	msg := Message{
		Type:        messageType,
		InterviewID: interviewID,
		Data:        data,
		Timestamp:   time.Now(),
	}

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	h.broadcast <- jsonMsg
	return nil
}

func (h *Hub) GetConnectedUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

func (h *Hub) roomClients(roomID string) []*Client {
	target := strings.TrimSpace(roomID)
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*Client, 0)
	for _, client := range h.clients {
		if target != "" && client.interviewID != target {
			continue
		}
		clients = append(clients, client)
	}
	return clients
}

func (h *Hub) broadcastRoomBytes(roomID string, payload []byte) {
	target := strings.TrimSpace(roomID)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.clients {
		if target != "" && client.interviewID != target {
			continue
		}
		h.sendToClient(client, payload)
	}
}

func (h *Hub) broadcastStructuredToRoom(roomID string, msg Message) error {
	encoded, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	h.broadcastRoomBytes(roomID, encoded)
	return nil
}

func (h *Hub) sendToClient(client *Client, payload []byte) bool {
	if client == nil {
		return false
	}
	select {
	case client.send <- payload:
		return true
	default:
		log.Printf("Client %s send buffer full, dropping", client.userID)
		return false
	}
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func normalizeMessageData(data interface{}) map[string]interface{} {
	if m, ok := data.(map[string]interface{}); ok && m != nil {
		return m
	}
	return map[string]interface{}{}
}

func asString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
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

func asUintFromAny(value interface{}) uint {
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

var globalHub *Hub

func init() {
	globalHub = NewHub()
	go globalHub.Run()
}

func GetHub() *Hub {
	return globalHub
}
