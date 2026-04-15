package handler

import (
	"errors"
	"strconv"

	groupws "your-project/pkg/groupws"
	livews "your-project/pkg/livews"
)

func ensureSingleLiveRoomConnection(identity *liveTokenIdentity, roomID string) error {
	if identity == nil {
		return nil
	}

	// Allow reconnect in the same room, reject cross-room concurrent sessions.
	activeClients := livews.GetLiveHub().GetClientsByUserID(strconv.FormatUint(uint64(identity.UserID), 10))
	for _, client := range activeClients {
		if client.GetRoomID() != roomID {
			return errors.New("您已在其他面试间中，请先退出")
		}
	}

	return nil
}

func ensureSingleGroupRoomConnection(identity *liveTokenIdentity, roomID string) error {
	if identity == nil {
		return nil
	}

	activeClients := groupws.GetGroupHub().GetClientsByUserID(strconv.FormatUint(uint64(identity.UserID), 10))
	for _, client := range activeClients {
		if client.GetRoomID() != roomID {
			return errors.New("您已在其他群面房间中，请先退出")
		}
	}

	return nil
}
