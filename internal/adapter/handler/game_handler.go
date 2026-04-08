package handler

import (
	"github.com/gofiber/fiber/v2"

	"pokertipssimulator/internal/application"
	"pokertipssimulator/internal/application/dto"
	"pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/domain/event"
)

type GameHandler struct {
	uc          *application.GameUseCase
	broadcaster port.WSBroadcaster
}

func NewGameHandler(uc *application.GameUseCase, broadcaster port.WSBroadcaster) *GameHandler {
	return &GameHandler{uc: uc, broadcaster: broadcaster}
}

func (h *GameHandler) StartRound(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	room, err := h.uc.StartRound(c.Context(), roomID, playerID)
	if err != nil {
		return err
	}

	h.broadcaster.BroadcastPerPlayer(roomID, string(event.RoundStarted), func(pid string) interface{} {
		return room.FilterForPlayer(pid)
	})
	return c.JSON(room.FilterForPlayer(playerID))
}

func (h *GameHandler) AdvanceStreet(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	room, err := h.uc.AdvanceStreet(c.Context(), roomID, playerID)
	if err != nil {
		return err
	}

	h.broadcaster.BroadcastPerPlayer(roomID, string(event.StreetAdvanced), func(pid string) interface{} {
		return room.FilterForPlayer(pid)
	})
	return c.JSON(room.FilterForPlayer(playerID))
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

	h.broadcaster.BroadcastToRoom(roomID, string(event.Settlement), room)
	return c.JSON(room)
}

func (h *GameHandler) AutoSettleRound(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	room, err := h.uc.SettleRound(c.Context(), roomID, playerID, dto.SettleRequest{})
	if err != nil {
		return err
	}

	h.broadcaster.BroadcastToRoom(roomID, string(event.Settlement), room)
	return c.JSON(room)
}

func (h *GameHandler) PauseGame(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	room, err := h.uc.PauseGame(c.Context(), roomID, playerID)
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

	h.broadcaster.BroadcastToRoom(roomID, string(event.StackUpdated), room)
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

	h.broadcaster.BroadcastToRoom(roomID, string(event.PlayerLeft), map[string]string{"player_id": targetID})
	return c.JSON(room)
}
