package handler

import (
	"github.com/gofiber/fiber/v2"

	"pokertipssimulator/internal/adapter/ws"
	"pokertipssimulator/internal/application"
	"pokertipssimulator/internal/application/dto"
)

type TipsHandler struct {
	uc  *application.TipsUseCase
	hub *ws.Hub
}

func NewTipsHandler(uc *application.TipsUseCase, hub *ws.Hub) *TipsHandler {
	return &TipsHandler{uc: uc, hub: hub}
}

func (h *TipsHandler) TransferChips(c *fiber.Ctx) error {
	roomID := c.Params("roomId")

	var req dto.TransferChipsRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	room, err := h.uc.TransferChips(c.Context(), roomID, req)
	if err != nil {
		return err
	}

	h.hub.BroadcastToRoom(roomID, ws.NewMessage(ws.MsgTypeChipsTransferred, room))
	return c.JSON(room)
}

func (h *TipsHandler) AdvanceBlind(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	room, err := h.uc.AdvanceBlindLevel(c.Context(), roomID, playerID)
	if err != nil {
		return err
	}

	h.hub.BroadcastToRoom(roomID, ws.NewMessage(ws.MsgTypeBlindLevelChanged, room))
	return c.JSON(room)
}

func (h *TipsHandler) PauseTimer(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	room, err := h.uc.PauseTimer(c.Context(), roomID, playerID)
	if err != nil {
		return err
	}

	msgType := ws.MsgTypeGamePaused
	if room.Status == "playing" {
		msgType = ws.MsgTypeGameResumed
	}
	h.hub.BroadcastToRoom(roomID, ws.NewMessage(msgType, room))
	return c.JSON(room)
}

func (h *TipsHandler) Rebuy(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Params("playerId")
	requestingPlayerID := c.Locals("playerID").(string)

	var req dto.RebuyRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	room, err := h.uc.Rebuy(c.Context(), roomID, playerID, requestingPlayerID, req.Amount)
	if err != nil {
		return err
	}

	h.hub.BroadcastToRoom(roomID, ws.NewMessage(ws.MsgTypeStackUpdated, room))
	return c.JSON(room)
}

func (h *TipsHandler) KickPlayer(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	hostID := c.Locals("playerID").(string)
	targetID := c.Params("playerId")

	room, err := h.uc.KickPlayer(c.Context(), roomID, targetID, hostID)
	if err != nil {
		return err
	}

	h.hub.BroadcastToRoom(roomID, ws.NewMessage(ws.MsgTypePlayerLeft, map[string]string{"player_id": targetID}))
	return c.JSON(room)
}
