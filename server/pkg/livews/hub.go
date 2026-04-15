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
	Type      string      `json:"type"`
	UserID    string      `json:"user_id,omitempty"`
	RoomID    string      `json:"interview_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
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
		if strings.TrimSpace(msg.RoomID) == "" {
			msg.RoomID = c.roomID
		}
		msg.Timestamp = time.Now()

		encoded, err := json.Marshal(msg)
		if err != nil {
			log.Printf("live ws marshal message failed: %v", err)
			continue
		}

		c.hub.broadcastToRoom(msg.RoomID, encoded)
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
