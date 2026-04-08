package mock

import "sync"

type BroadcastCall struct {
	RoomID  string
	Type    string
	Payload interface{}
}

type SendCall struct {
	RoomID   string
	PlayerID string
	Type     string
	Payload  interface{}
}

type WSBroadcaster struct {
	mu             sync.Mutex
	BroadcastCalls []BroadcastCall
	SendCalls      []SendCall
}

func NewWSBroadcaster() *WSBroadcaster {
	return &WSBroadcaster{}
}

func (b *WSBroadcaster) BroadcastToRoom(roomID string, msgType string, payload interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.BroadcastCalls = append(b.BroadcastCalls, BroadcastCall{
		RoomID:  roomID,
		Type:    msgType,
		Payload: payload,
	})
}

func (b *WSBroadcaster) SendToPlayer(roomID, playerID string, msgType string, payload interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.SendCalls = append(b.SendCalls, SendCall{
		RoomID:   roomID,
		PlayerID: playerID,
		Type:     msgType,
		Payload:  payload,
	})
}

func (b *WSBroadcaster) BroadcastPerPlayer(roomID string, msgType string, buildPayload func(playerID string) interface{}) {
	// no-op in mock: individual payloads depend on connected clients
}

func (b *WSBroadcaster) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.BroadcastCalls = nil
	b.SendCalls = nil
}
