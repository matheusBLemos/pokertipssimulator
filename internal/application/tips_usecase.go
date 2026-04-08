package application

import (
	"context"

	"pokertipssimulator/internal/application/dto"
	"pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/domain/entity"
)

type TipsUseCase struct {
	repo port.RoomRepository
}

func NewTipsUseCase(repo port.RoomRepository) *TipsUseCase {
	return &TipsUseCase{repo: repo}
}

func (uc *TipsUseCase) TransferChips(ctx context.Context, roomID string, req dto.TransferChipsRequest) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.Mode != entity.RoomModeTips {
		return nil, entity.ErrWrongRoomMode
	}

	if req.Amount <= 0 {
		return nil, entity.ErrInvalidAmount
	}

	if req.FromPlayerID == req.ToPlayerID {
		return nil, entity.ErrSamePlayer
	}

	from := room.FindPlayer(req.FromPlayerID)
	if from == nil {
		return nil, entity.ErrPlayerNotFound
	}

	to := room.FindPlayer(req.ToPlayerID)
	if to == nil {
		return nil, entity.ErrPlayerNotFound
	}

	if from.Stack < req.Amount {
		return nil, entity.ErrInsufficientStack
	}

	from.Stack -= req.Amount
	to.Stack += req.Amount

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *TipsUseCase) AdvanceBlindLevel(ctx context.Context, roomID, playerID string) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.Mode != entity.RoomModeTips {
		return nil, entity.ErrWrongRoomMode
	}

	if room.HostPlayerID != playerID {
		return nil, entity.ErrNotHost
	}

	bs := &room.Config.BlindStructure
	if bs.CurrentLevel >= len(bs.Levels)-1 {
		return nil, entity.ErrNoBlindLevels
	}

	bs.CurrentLevel++

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *TipsUseCase) PauseTimer(ctx context.Context, roomID, playerID string) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.Mode != entity.RoomModeTips {
		return nil, entity.ErrWrongRoomMode
	}

	if room.HostPlayerID != playerID {
		return nil, entity.ErrNotHost
	}

	switch room.Status {
	case entity.RoomStatusPlaying:
		room.Status = entity.RoomStatusPaused
	case entity.RoomStatusPaused:
		room.Status = entity.RoomStatusPlaying
	case entity.RoomStatusWaiting:
		room.Status = entity.RoomStatusPlaying
	default:
		return nil, entity.ErrGameInProgress
	}

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *TipsUseCase) Rebuy(ctx context.Context, roomID, playerID, requestingPlayerID string, amount int) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.Mode != entity.RoomModeTips {
		return nil, entity.ErrWrongRoomMode
	}

	if room.HostPlayerID != requestingPlayerID {
		return nil, entity.ErrNotHost
	}

	if amount <= 0 {
		return nil, entity.ErrInvalidAmount
	}

	player := room.FindPlayer(playerID)
	if player == nil {
		return nil, entity.ErrPlayerNotFound
	}

	player.Stack += amount

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *TipsUseCase) KickPlayer(ctx context.Context, roomID, playerID, requestingPlayerID string) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.Mode != entity.RoomModeTips {
		return nil, entity.ErrWrongRoomMode
	}

	if room.HostPlayerID != requestingPlayerID {
		return nil, entity.ErrNotHost
	}

	if playerID == room.HostPlayerID {
		return nil, entity.ErrNotHost
	}

	idx := -1
	for i, p := range room.Players {
		if p.ID == playerID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, entity.ErrPlayerNotFound
	}

	room.Players = append(room.Players[:idx], room.Players[idx+1:]...)

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}
