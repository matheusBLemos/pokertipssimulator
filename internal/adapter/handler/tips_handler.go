package handler

import (
	"github.com/gofiber/fiber/v2"

	"pokertipssimulator/internal/application"
	"pokertipssimulator/internal/application/dto"
	"pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/domain/event"
)

type TipsHandler struct {
	uc          *application.TipsUseCase
	broadcaster port.WSBroadcaster
}

func NewTipsHandler(uc *application.TipsUseCase, broadcaster port.WSBroadcaster) *TipsHandler {
	return &TipsHandler{uc: uc, broadcaster: broadcaster}
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

	h.broadcaster.BroadcastToRoom(roomID, string(event.ChipsTransferred), room)
	return c.JSON(room)
}

func (h *TipsHandler) AdvanceBlind(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	room, err := h.uc.AdvanceBlindLevel(c.Context(), roomID, playerID)
	if err != nil {
		return err
	}

	h.broadcaster.BroadcastToRoom(roomID, string(event.BlindLevelChanged), room)
	return c.JSON(room)
}

func (h *TipsHandler) PauseTimer(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	room, err := h.uc.PauseTimer(c.Context(), roomID, playerID)
	if err != nil {
		return err
	}

	msgType := event.GamePaused
	if room.Status == "playing" {
		msgType = event.GameResumed
	}
	h.broadcaster.BroadcastToRoom(roomID, string(msgType), room)
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

	h.broadcaster.BroadcastToRoom(roomID, string(event.StackUpdated), room)
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

	h.broadcaster.BroadcastToRoom(roomID, string(event.PlayerLeft), map[string]string{"player_id": targetID})
	return c.JSON(room)
}
