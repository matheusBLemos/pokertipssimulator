package handler

import (
	"github.com/gofiber/fiber/v2"

	"pokertipssimulator/internal/application"
	"pokertipssimulator/internal/application/dto"
)

type RoomHandler struct {
	uc *application.RoomUseCase
}

func NewRoomHandler(uc *application.RoomUseCase) *RoomHandler {
	return &RoomHandler{uc: uc}
}

func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	var req dto.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	resp, err := h.uc.CreateRoom(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *RoomHandler) JoinRoom(c *fiber.Ctx) error {
	var req dto.JoinRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	resp, err := h.uc.JoinRoom(c.Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	room, err := h.uc.GetRoom(c.Context(), roomID)
	if err != nil {
		return err
	}
	playerID, _ := c.Locals("playerID").(string)
	return c.JSON(room.FilterForPlayer(playerID))
}

func (h *RoomHandler) UpdateConfig(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Locals("playerID").(string)

	var req dto.UpdateConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	room, err := h.uc.UpdateConfig(c.Context(), roomID, playerID, req)
	if err != nil {
		return err
	}
	return c.JSON(room)
}

func (h *RoomHandler) PickSeat(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	playerID := c.Params("playerId")

	var req dto.PickSeatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	room, err := h.uc.PickSeat(c.Context(), roomID, playerID, req.Seat)
	if err != nil {
		return err
	}
	return c.JSON(room)
}
