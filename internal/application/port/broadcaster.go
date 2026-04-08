package port

type WSBroadcaster interface {
	BroadcastToRoom(roomID string, msgType string, payload interface{})
	SendToPlayer(roomID, playerID string, msgType string, payload interface{})
	BroadcastPerPlayer(roomID string, msgType string, buildPayload func(playerID string) interface{})
}
