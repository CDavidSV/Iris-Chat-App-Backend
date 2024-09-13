package websocket

import "github.com/gofiber/fiber/v2/log"

type WebsocketMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type FriendStatus struct {
	Type   string `json:"type"`   // Can ber either REQUEST, ACCEPTED, or REMOVED
	UserID string `json:"userID"` // The user ID of the friend
}

func (ws *WebsocketServer) UpdateFriendStatus(recipientID string, status FriendStatus) error {
	conn, ok := ws.Connections[recipientID]

	if !ok {
		return ErrNoConnection
	}

	err := conn.Conn.WriteJSON(WebsocketMessage{
		Type: "FRIEND_STATUS",
		Data: status,
	})

	if err != nil {
		log.Error("Failed to send friend status update to user: ", recipientID)
		return err
	}

	return nil
}
