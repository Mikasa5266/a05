package handler

import (
	"net/http"
	"strconv"
	"strings"

	"your-project/internal/service"
	livews "your-project/pkg/livews"

	"github.com/gin-gonic/gin"
)

var liveWSSignalService = service.NewLiveWSSignalService()

type liveWSSignalHandshake struct {
	identity *liveTokenIdentity
	session  *service.LiveWSSignalSession
}

func resolveLiveWSSignalHandshake(c *gin.Context) (*liveWSSignalHandshake, int, string) {
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
	session, err := liveWSSignalService.Authorize(identity.UserID, identity.Role, identity.UserUUID, roomID, invitationCode)
	if err != nil {
		return nil, http.StatusForbidden, err.Error()
	}

	return &liveWSSignalHandshake{
		identity: identity,
		session:  session,
	}, http.StatusOK, ""
}

func proxyLiveSignalToDedicatedHub(c *gin.Context, handshake *liveWSSignalHandshake) {
	query := c.Request.URL.Query()
	query.Set("user_id", strconv.FormatUint(uint64(handshake.identity.UserID), 10))
	query.Set("room_id", handshake.session.RoomID)
	c.Request.URL.RawQuery = query.Encode()

	livews.GetLiveHub().HandleWebSocket(c.Writer, c.Request)
}

// InterviewLiveWS handles one-on-one live interview signaling endpoint.
func InterviewLiveWS(c *gin.Context) {
	handshake, statusCode, errMsg := resolveLiveWSSignalHandshake(c)
	if handshake == nil {
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	if err := ensureSingleLiveRoomConnection(handshake.identity, handshake.session.RoomID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	proxyLiveSignalToDedicatedHub(c, handshake)
}
