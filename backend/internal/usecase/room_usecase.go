package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"pokertipssimulator/internal/dto"
	"pokertipssimulator/internal/entity"
	"pokertipssimulator/internal/repository"
	"pokertipssimulator/pkg/idgen"
)

type RoomUseCase struct {
	repo      repository.RoomRepository
	jwtSecret string
}

func NewRoomUseCase(repo repository.RoomRepository, jwtSecret string) *RoomUseCase {
	return &RoomUseCase{repo: repo, jwtSecret: jwtSecret}
}

func (uc *RoomUseCase) CreateRoom(ctx context.Context, req dto.CreateRoomRequest) (*dto.CreateRoomResponse, error) {
	if req.HostName == "" {
		req.HostName = "Host"
	}
	if req.StartingStack <= 0 {
		req.StartingStack = 1000
	}
	if req.MaxPlayers <= 0 {
		req.MaxPlayers = 10
	}

	gameMode := entity.GameModeCash
	if req.GameMode == "tournament" {
		gameMode = entity.GameModeTournament
	}

	hostID := idgen.NewID()
	roomID := idgen.NewID()
	code := idgen.NewRoomCode()

	room := &entity.Room{
		ID:           roomID,
		Code:         code,
		Status:       entity.RoomStatusWaiting,
		HostPlayerID: hostID,
		Config: entity.RoomConfig{
			GameMode:      gameMode,
			StartingStack: req.StartingStack,
			ChipSet:       entity.DefaultChipSet(),
			MaxPlayers:    req.MaxPlayers,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{
					{SmallBlind: 5, BigBlind: 10, Ante: 0, Duration: 0},
				},
				CurrentLevel: 0,
			},
		},
		Players: []entity.Player{
			{
				ID:     hostID,
				Name:   req.HostName,
				Seat:   0,
				Stack:  req.StartingStack,
				Status: entity.PlayerStatusWaiting,
			},
		},
	}

	if err := uc.repo.Create(ctx, room); err != nil {
		return nil, err
	}

	token, err := uc.generateToken(roomID, hostID, true)
	if err != nil {
		return nil, err
	}

	return &dto.CreateRoomResponse{
		RoomID: roomID,
		Code:   code,
		Token:  token,
	}, nil
}

func (uc *RoomUseCase) JoinRoom(ctx context.Context, req dto.JoinRoomRequest) (*dto.JoinRoomResponse, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if len(code) != 6 {
		return nil, entity.ErrInvalidCode
	}

	room, err := uc.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if len(room.Players) >= room.Config.MaxPlayers {
		return nil, entity.ErrRoomFull
	}

	playerID := idgen.NewID()
	player := entity.Player{
		ID:     playerID,
		Name:   req.PlayerName,
		Seat:   0,
		Stack:  room.Config.StartingStack,
		Status: entity.PlayerStatusWaiting,
	}

	room.Players = append(room.Players, player)
	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}

	token, err := uc.generateToken(room.ID, playerID, false)
	if err != nil {
		return nil, err
	}

	return &dto.JoinRoomResponse{
		RoomID:   room.ID,
		PlayerID: playerID,
		Token:    token,
	}, nil
}

func (uc *RoomUseCase) GetRoom(ctx context.Context, roomID string) (*entity.Room, error) {
	return uc.repo.FindByID(ctx, roomID)
}

func (uc *RoomUseCase) UpdateConfig(ctx context.Context, roomID, playerID string, req dto.UpdateConfigRequest) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.HostPlayerID != playerID {
		return nil, entity.ErrNotHost
	}

	if room.Status != entity.RoomStatusWaiting {
		return nil, entity.ErrGameInProgress
	}

	if req.GameMode != "" {
		room.Config.GameMode = entity.GameMode(req.GameMode)
	}
	if req.StartingStack > 0 {
		room.Config.StartingStack = req.StartingStack
	}
	if req.MaxPlayers > 0 {
		room.Config.MaxPlayers = req.MaxPlayers
	}
	if req.BlindStructure != nil {
		room.Config.BlindStructure = *req.BlindStructure
	}

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *RoomUseCase) PickSeat(ctx context.Context, roomID, playerID string, seat int) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if seat < 1 || seat > room.Config.MaxPlayers {
		return nil, entity.ErrSeatTaken
	}

	for _, p := range room.Players {
		if p.Seat == seat && p.ID != playerID {
			return nil, entity.ErrSeatTaken
		}
	}

	player := room.FindPlayer(playerID)
	if player == nil {
		return nil, entity.ErrPlayerNotFound
	}

	player.Seat = seat
	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *RoomUseCase) ValidateToken(tokenStr string) (roomID, playerID string, isHost bool, err error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(uc.jwtSecret), nil
	})
	if err != nil {
		return "", "", false, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", false, jwt.ErrSignatureInvalid
	}

	roomID, _ = claims["room_id"].(string)
	playerID, _ = claims["player_id"].(string)
	isHost, _ = claims["is_host"].(bool)
	return roomID, playerID, isHost, nil
}

func (uc *RoomUseCase) generateToken(roomID, playerID string, isHost bool) (string, error) {
	claims := jwt.MapClaims{
		"room_id":   roomID,
		"player_id": playerID,
		"is_host":   isHost,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}
