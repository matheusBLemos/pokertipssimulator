package handler

import (
	"github.com/gofiber/fiber/v2"

	"pokertipssimulator/internal/application"
	"pokertipssimulator/internal/application/dto"
	"pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/domain/entity"
	"pokertipssimulator/internal/domain/event"
)

type ActionHandler struct {
	uc          *application.ActionUseCase
	broadcaster port.WSBroadcaster
}

func NewActionHandler(uc *application.ActionUseCase, broadcaster port.WSBroadcaster) *ActionHandler {
	return &ActionHandler{uc: uc, broadcaster: broadcaster}
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

	h.broadcaster.BroadcastToRoom(roomID, string(event.ActionPerformed), map[string]interface{}{
		"player_id": playerID,
		"action":    req.Type,
		"amount":    req.Amount,
	})

	h.broadcaster.BroadcastPerPlayer(roomID, string(event.RoomStateChanged), func(pid string) interface{} {
		return room.FilterForPlayer(pid)
	})

	return c.JSON(room.FilterForPlayer(playerID))
}
