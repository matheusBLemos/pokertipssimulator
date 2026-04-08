package handler

import (
	"github.com/gofiber/fiber/v2"

	"pokertipssimulator/internal/adapter/ws"
	"pokertipssimulator/internal/application"
	"pokertipssimulator/internal/application/dto"
	"pokertipssimulator/internal/domain/entity"
)

type ActionHandler struct {
	uc  *application.ActionUseCase
	hub *ws.Hub
}

func NewActionHandler(uc *application.ActionUseCase, hub *ws.Hub) *ActionHandler {
	return &ActionHandler{uc: uc, hub: hub}
}

func (h *ActionHandler) PerformAction(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	var req dto.ActionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	room, err := h.uc.ProcessAction(c.Context(), roomID, playerID, entity.ActionType(req.Type), req.Amount)
	if err != nil {
		return err
	}

	h.hub.BroadcastToRoom(roomID, ws.NewMessage(ws.MsgTypeActionPerformed, map[string]interface{}{
		"player_id": playerID,
		"action":    req.Type,
		"amount":    req.Amount,
	}))

	h.hub.BroadcastToRoom(roomID, ws.NewMessage(ws.MsgTypeRoomState, room))

	return c.JSON(room)
}
