package port

type WSBroadcaster interface {
	BroadcastToRoom(roomID string, msgType string, payload interface{})
	SendToPlayer(roomID, playerID string, msgType string, payload interface{})
}
