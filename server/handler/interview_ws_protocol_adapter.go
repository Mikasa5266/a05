package handler

import (
	"net/http"
	"strconv"
	"strings"

	"your-project/internal/service"
	ws "your-project/pkg/websocket"

	"github.com/gin-gonic/gin"
)

const (
	groupRoomTargetParticipants = 4
	groupRoomStartThreshold     = 2
)

type liveSignalHandshake struct {
	identity       *liveTokenIdentity
	roomID         string
	invitationCode string
}

func resolveLiveSignalHandshake(c *gin.Context) (*liveSignalHandshake, int, string) {
	tokenString := strings.TrimSpace(c.Query("token"))
	if tokenString == "" {
		return nil, http.StatusUnauthorized, "token is required"
	}

	roomID := strings.TrimSpace(c.Query("room_id"))
	if roomID == "" {
		return nil, http.StatusBadRequest, "room_id is required"
	}

	identity, err := parseLiveIdentityFromToken(tokenString)
	if err != nil {
		return nil, http.StatusUnauthorized, "invalid token"
	}

	invitationCode := strings.TrimSpace(c.Query("invitation_code"))
	if _, err := service.ValidateLiveRoomAccess(identity.UserID, identity.Role, identity.UserUUID, roomID, invitationCode); err != nil {
		return nil, http.StatusForbidden, err.Error()
	}

	return &liveSignalHandshake{
		identity:       identity,
		roomID:         roomID,
		invitationCode: invitationCode,
	}, http.StatusOK, ""
}

func proxyLiveSignalToHub(c *gin.Context, identity *liveTokenIdentity, roomID string) {
	query := c.Request.URL.Query()
	query.Set("user_id", strconv.FormatUint(uint64(identity.UserID), 10))
	query.Set("interview_id", roomID)
	query.Set("group_target_participants", strconv.Itoa(groupRoomTargetParticipants))
	query.Set("group_start_threshold", strconv.Itoa(groupRoomStartThreshold))
	c.Request.URL.RawQuery = query.Encode()

	ws.GetHub().HandleWebSocket(c.Writer, c.Request)
}
