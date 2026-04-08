package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"pokertipssimulator/internal/application/dto"
	"pokertipssimulator/internal/domain/entity"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	switch {
	case errors.Is(err, entity.ErrRoomNotFound), errors.Is(err, entity.ErrPlayerNotFound):
		code = fiber.StatusNotFound
	case errors.Is(err, entity.ErrRoomFull), errors.Is(err, entity.ErrSeatTaken),
		errors.Is(err, entity.ErrAlreadyJoined):
		code = fiber.StatusConflict
	case errors.Is(err, entity.ErrInvalidAction), errors.Is(err, entity.ErrNotYourTurn),
		errors.Is(err, entity.ErrInsufficientStack), errors.Is(err, entity.ErrInvalidAmount),
		errors.Is(err, entity.ErrInvalidCode), errors.Is(err, entity.ErrNotEnoughPlayers),
		errors.Is(err, entity.ErrRoundComplete), errors.Is(err, entity.ErrInvalidStreet),
		errors.Is(err, entity.ErrWrongRoomMode), errors.Is(err, entity.ErrSamePlayer),
		errors.Is(err, entity.ErrNoBlindLevels):
		code = fiber.StatusBadRequest
	case errors.Is(err, entity.ErrNotHost):
		code = fiber.StatusForbidden
	case errors.Is(err, entity.ErrGameInProgress), errors.Is(err, entity.ErrGameNotStarted):
		code = fiber.StatusConflict
	}

	return c.Status(code).JSON(dto.ErrorResponse{Error: err.Error()})
}
