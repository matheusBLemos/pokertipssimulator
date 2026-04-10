package ws

import "time"

type MessageType string

const (
	MsgTypePing MessageType = "ping"
	MsgTypePong MessageType = "pong"
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
