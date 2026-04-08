package handler

import (
	"github.com/gofiber/fiber/v2"

	"pokertipssimulator/internal/adapter/ws"
	"pokertipssimulator/internal/application"
	"pokertipssimulator/internal/application/dto"
)

type GameHandler struct {
	uc  *application.GameUseCase
	hub *ws.Hub
}

func NewGameHandler(uc *application.GameUseCase, hub *ws.Hub) *GameHandler {
	return &GameHandler{uc: uc, hub: hub}
}

func (h *GameHandler) StartRound(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	room, err := h.uc.StartRound(c.Context(), roomID, playerID)
	if err != nil {
		return err
	}

	h.hub.BroadcastToRoom(roomID, ws.NewMessage(ws.MsgTypeRoundStarted, room))
	return c.JSON(room)
}

func (h *GameHandler) AdvanceStreet(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	room, err := h.uc.AdvanceStreet(c.Context(), roomID, playerID)
	if err != nil {
		return err
	}

	h.hub.BroadcastToRoom(roomID, ws.NewMessage(ws.MsgTypeStreetAdvanced, room))
	return c.JSON(room)
}

func (h *GameHandler) SettleRound(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	var req dto.SettleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	room, err := h.uc.SettleRound(c.Context(), roomID, playerID, req)
	if err != nil {
		return err
	}

	h.hub.BroadcastToRoom(roomID, ws.NewMessage(ws.MsgTypeSettlement, room))
	return c.JSON(room)
}

func (h *GameHandler) PauseGame(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	room, err := h.uc.PauseGame(c.Context(), roomID, playerID)
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

func (h *GameHandler) Rebuy(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Params("playerId")

	var req dto.RebuyRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	room, err := h.uc.Rebuy(c.Context(), roomID, playerID, req.Amount)
	if err != nil {
		return err
	}

	h.hub.BroadcastToRoom(roomID, ws.NewMessage(ws.MsgTypeStackUpdated, room))
	return c.JSON(room)
}

func (h *GameHandler) KickPlayer(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	hostID := c.Locals("playerID").(string)
	targetID := c.Params("playerId")

	room, err := h.uc.KickPlayer(c.Context(), roomID, hostID, targetID)
	if err != nil {
		return err
	}

	h.hub.BroadcastToRoom(roomID, ws.NewMessage(ws.MsgTypePlayerLeft, map[string]string{"player_id": targetID}))
	return c.JSON(room)
}
