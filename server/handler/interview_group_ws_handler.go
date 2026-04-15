package handler

import (
	"net/http"
	"strconv"
	"strings"

	"your-project/internal/service"
	groupws "your-project/pkg/groupws"

	"github.com/gin-gonic/gin"
)

var groupWSSignalService = service.NewGroupWSSignalService()

type groupWSSignalHandshake struct {
	identity *liveTokenIdentity
	session  *service.GroupWSSignalSession
}

func resolveGroupWSSignalHandshake(c *gin.Context) (*groupWSSignalHandshake, int, string) {
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
	session, err := groupWSSignalService.Authorize(identity.UserID, identity.Role, identity.UserUUID, roomID, invitationCode)
	if err != nil {
		return nil, http.StatusForbidden, err.Error()
	}

	return &groupWSSignalHandshake{
		identity: identity,
		session:  session,
	}, http.StatusOK, ""
}

func proxyGroupSignalToDedicatedHub(c *gin.Context, handshake *groupWSSignalHandshake) {
	query := c.Request.URL.Query()
	query.Set("user_id", strconv.FormatUint(uint64(handshake.identity.UserID), 10))
	query.Set("room_id", handshake.session.RoomID)
	query.Set("group_target_participants", strconv.Itoa(handshake.session.TargetParticipants))
	query.Set("group_start_threshold", strconv.Itoa(handshake.session.StartThreshold))
	c.Request.URL.RawQuery = query.Encode()

	groupws.GetGroupHub().HandleWebSocket(c.Writer, c.Request)
}

// InterviewGroupWS handles group interview signaling endpoint.
func InterviewGroupWS(c *gin.Context) {
	handshake, statusCode, errMsg := resolveGroupWSSignalHandshake(c)
	if handshake == nil {
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	if err := ensureSingleGroupRoomConnection(handshake.identity, handshake.session.RoomID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	proxyGroupSignalToDedicatedHub(c, handshake)
}
