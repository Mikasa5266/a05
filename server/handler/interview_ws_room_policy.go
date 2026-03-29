package handler

import (
	"errors"
	"strconv"

	ws "your-project/pkg/websocket"
)

func ensureSingleRoomConnection(identity *liveTokenIdentity, roomID string) error {
	if identity == nil {
		return nil
	}

	// Allow reconnect in the same room, reject cross-room concurrent sessions.
	activeClients := ws.GetHub().GetClientsByUserID(strconv.FormatUint(uint64(identity.UserID), 10))
	for _, client := range activeClients {
		if client.GetInterviewID() != roomID {
			return errors.New("您已在其他面试间中，请先退出")
		}
	}

	return nil
}
