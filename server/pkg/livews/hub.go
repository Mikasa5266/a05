package livews

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type LiveHub struct {
	mu      sync.RWMutex
	clients map[string]*LiveClient
}

type LiveClient struct {
	id     string
	hub    *LiveHub
	conn   *websocket.Conn
	send   chan []byte
	userID string
	roomID string
}

type LiveSignalMessage struct {
	Type         string      `json:"type"`
	UserID       string      `json:"user_id,omitempty"`
	SenderUserID string      `json:"sender_user_id,omitempty"`
	TargetUserID string      `json:"target_user_id,omitempty"`
	RoomID       string      `json:"interview_id,omitempty"`
	Data         interface{} `json:"data"`
	Timestamp    time.Time   `json:"timestamp"`
}

var liveUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewLiveHub() *LiveHub {
	return &LiveHub{clients: make(map[string]*LiveClient)}
}

func (h *LiveHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
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

	conn, err := liveUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("live ws upgrade failed: %v", err)
		return
	}

	client := &LiveClient{
		id:     fmt.Sprintf("live:%s:%s:%d", userID, roomID, time.Now().UnixNano()),
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		roomID: roomID,
	}

	h.register(client)
	go client.writePump()
	go client.readPump()
}

func (h *LiveHub) register(client *LiveClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.id] = client
}

func (h *LiveHub) unregister(client *LiveClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client.id]; !ok {
		return
	}
	delete(h.clients, client.id)
	close(client.send)
}

func (c *LiveClient) readPump() {
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
				log.Printf("live ws read error: %v", err)
			}
			return
		}

		var msg LiveSignalMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Printf("live ws invalid message: %v", err)
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
			log.Printf("live ws route message failed: type=%s room=%s user=%s err=%v", msg.Type, msg.RoomID, c.userID, err)
		}
	}
}

func (c *LiveClient) writePump() {
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

func (h *LiveHub) broadcastToRoom(roomID string, payload []byte) {
	target := strings.TrimSpace(roomID)
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.roomID != target {
			continue
		}
		select {
		case client.send <- payload:
		default:
			log.Printf("live ws client send buffer full, user=%s room=%s", client.userID, client.roomID)
		}
	}
}

func (h *LiveHub) routeMessage(sender *LiveClient, msg *LiveSignalMessage) error {
	if sender == nil || msg == nil {
		return nil
	}

	switch strings.TrimSpace(msg.Type) {
	case "offer", "answer", "candidate":
		targetUserID := strings.TrimSpace(msg.TargetUserID)
		if targetUserID == "" {
			return fmt.Errorf("target user id is required for %s", msg.Type)
		}
		if targetUserID == sender.userID {
			return fmt.Errorf("cannot send %s to self", msg.Type)
		}
		msg.UserID = sender.userID
		msg.SenderUserID = sender.userID
		return h.sendToTargetUser(msg.RoomID, targetUserID, msg)
	default:
		msg.UserID = sender.userID
		msg.SenderUserID = sender.userID
		return h.broadcastRawToRoom(msg.RoomID, msg)
	}
}

func (h *LiveHub) broadcastRawToRoom(roomID string, msg *LiveSignalMessage) error {
	if msg == nil {
		return nil
	}
	encoded, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	h.broadcastToRoom(roomID, encoded)
	return nil
}

func (h *LiveHub) sendToTargetUser(roomID, targetUserID string, msg *LiveSignalMessage) error {
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

	h.mu.RLock()
	defer h.mu.RUnlock()

	matched := false
	for _, client := range h.clients {
		if client.roomID != targetRoom || client.userID != targetUser {
			continue
		}
		matched = true
		select {
		case client.send <- encoded:
		default:
			log.Printf("live ws client send buffer full, user=%s room=%s", client.userID, client.roomID)
		}
	}

	if !matched {
		return fmt.Errorf("target user %s is not connected in room %s", targetUser, targetRoom)
	}

	return nil
}

func (h *LiveHub) GetClientsByUserID(userID string) []*LiveClient {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*LiveClient, 0)
	for _, client := range h.clients {
		if client.userID == userID {
			clients = append(clients, client)
		}
	}
	return clients
}

func (c *LiveClient) GetRoomID() string {
	return c.roomID
}

var globalLiveHub = NewLiveHub()

func GetLiveHub() *LiveHub {
	return globalLiveHub
}
