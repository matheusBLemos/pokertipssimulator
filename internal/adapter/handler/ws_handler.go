package handler

import (
	"context"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"pokertipssimulator/internal/adapter/ws"
	"pokertipssimulator/internal/application"
	"pokertipssimulator/internal/domain/event"
)

type WSHandler struct {
	hub *ws.Hub
	uc  *application.RoomUseCase
}

func NewWSHandler(hub *ws.Hub, uc *application.RoomUseCase) *WSHandler {
	return &WSHandler{hub: hub, uc: uc}
}

func (h *WSHandler) Upgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func (h *WSHandler) Handle(c *websocket.Conn) {
	token := c.Query("token")
	if token == "" {
		log.Println("ws: missing token")
		c.Close()
		return
	}

	roomID, playerID, isHost, err := h.uc.ValidateToken(token)
	if err != nil {
		log.Printf("ws: invalid token: %v", err)
		c.Close()
		return
	}

	client := &ws.Client{
		Hub:      h.hub,
		Conn:     c,
		Send:     make(chan []byte, 256),
		RoomID:   roomID,
		PlayerID: playerID,
		IsHost:   isHost,
	}

	h.hub.Register <- client

	room, err := h.uc.GetRoom(context.Background(), roomID)
	if err == nil {
		h.hub.SendToPlayer(roomID, playerID, string(event.RoomStateChanged), room.FilterForPlayer(playerID))
	}

	go client.WritePump()
	client.ReadPump()
}
