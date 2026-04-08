package ws

import "time"

type MessageType string

const (
	MsgTypeAction MessageType = "action"
	MsgTypePing   MessageType = "ping"

	MsgTypeRoomState         MessageType = "room_state"
	MsgTypePlayerJoined      MessageType = "player_joined"
	MsgTypePlayerLeft        MessageType = "player_left"
	MsgTypePlayerReconnected MessageType = "player_reconnected"
	MsgTypeRoundStarted      MessageType = "round_started"
	MsgTypeTurnChanged       MessageType = "turn_changed"
	MsgTypeActionPerformed   MessageType = "action_performed"
	MsgTypeStreetAdvanced    MessageType = "street_advanced"
	MsgTypePotsUpdated       MessageType = "pots_updated"
	MsgTypeStackUpdated      MessageType = "stack_updated"
	MsgTypeSettlement        MessageType = "settlement"
	MsgTypeRoundEnded        MessageType = "round_ended"
	MsgTypeBlindLevelChanged MessageType = "blind_level_changed"
	MsgTypeGamePaused        MessageType = "game_paused"
	MsgTypeGameResumed       MessageType = "game_resumed"
	MsgTypeChipsTransferred  MessageType = "chips_transferred"
	MsgTypeError             MessageType = "error"
	MsgTypePong              MessageType = "pong"
)

type Message struct {
	Type      MessageType `json:"type"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

func NewMessage(msgType MessageType, payload interface{}) Message {
	return Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
}

type InboundAction struct {
	Type   string `json:"type"`
	Amount int    `json:"amount,omitempty"`
}

type InboundMessage struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}
