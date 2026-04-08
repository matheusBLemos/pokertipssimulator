package application

import (
	"context"
	"strings"
	"testing"

	"pokertipssimulator/internal/adapter/repository"
	"pokertipssimulator/internal/application/dto"
	"pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/domain/entity"
	"pokertipssimulator/internal/infrastructure/auth"
)

func newRoomTestDeps(t *testing.T) (port.RoomRepository, *auth.JWTService, *RoomUseCase) {
	t.Helper()
	db := repository.NewTestDB(t)
	repo := repository.NewSQLiteRoomRepository(db)
	jwt := auth.NewJWTService("test-secret")
	uc := NewRoomUseCase(repo, jwt)
	return repo, jwt, uc
}

func TestCreateRoom_Defaults(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	resp, err := uc.CreateRoom(ctx, dto.CreateRoomRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.RoomID == "" {
		t.Error("expected non-empty room ID")
	}
	if resp.Code == "" {
		t.Error("expected non-empty room code")
	}
	if len(resp.Code) != 6 {
		t.Errorf("expected 6-char code, got %d chars", len(resp.Code))
	}
	if resp.Token == "" {
		t.Error("expected non-empty JWT token")
	}
}

func TestCreateRoom_CustomValues(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	resp, err := uc.CreateRoom(ctx, dto.CreateRoomRequest{
		HostName:      "TestHost",
		GameMode:      "tournament",
		StartingStack: 5000,
		MaxPlayers:    6,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room, err := repo.FindByID(ctx, resp.RoomID)
	if err != nil {
		t.Fatalf("room not found: %v", err)
	}

	if room.Config.GameMode != entity.GameModeTournament {
		t.Errorf("expected tournament, got %s", room.Config.GameMode)
	}
	if room.Config.StartingStack != 5000 {
		t.Errorf("expected 5000 stack, got %d", room.Config.StartingStack)
	}
	if room.Config.MaxPlayers != 6 {
		t.Errorf("expected 6 max, got %d", room.Config.MaxPlayers)
	}
	if room.Players[0].Name != "TestHost" {
		t.Errorf("expected TestHost, got %s", room.Players[0].Name)
	}
	if room.Status != entity.RoomStatusWaiting {
		t.Errorf("expected waiting, got %s", room.Status)
	}
}

func TestCreateRoom_HostIsFirstPlayer(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	resp, err := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room, _ := repo.FindByID(ctx, resp.RoomID)
	if len(room.Players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(room.Players))
	}
	if room.HostPlayerID != room.Players[0].ID {
		t.Error("host player ID should match first player")
	}
	if room.Players[0].Stack != 1000 {
		t.Errorf("expected default stack 1000, got %d", room.Players[0].Stack)
	}
}

func TestCreateRoom_DefaultChipSet(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	resp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{})
	room, _ := repo.FindByID(ctx, resp.RoomID)

	if len(room.Config.ChipSet.Denominations) != 6 {
		t.Errorf("expected 6 chip denominations, got %d", len(room.Config.ChipSet.Denominations))
	}
}

func TestCreateRoom_DefaultBlindStructure(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	resp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{})
	room, _ := repo.FindByID(ctx, resp.RoomID)

	bs := room.Config.BlindStructure
	if len(bs.Levels) != 1 {
		t.Fatalf("expected 1 blind level, got %d", len(bs.Levels))
	}
	if bs.CurrentLevel != 0 {
		t.Errorf("expected level 0, got %d", bs.CurrentLevel)
	}
	if bs.Levels[0].SmallBlind != 5 || bs.Levels[0].BigBlind != 10 {
		t.Errorf("expected 5/10, got %d/%d", bs.Levels[0].SmallBlind, bs.Levels[0].BigBlind)
	}
}

func TestJoinRoom_Success(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})

	joinResp, err := uc.JoinRoom(ctx, dto.JoinRoomRequest{
		Code:       createResp.Code,
		PlayerName: "Guest",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if joinResp.RoomID != createResp.RoomID {
		t.Error("room IDs should match")
	}
	if joinResp.PlayerID == "" {
		t.Error("expected non-empty player ID")
	}
	if joinResp.Token == "" {
		t.Error("expected non-empty token")
	}

	room, _ := repo.FindByID(ctx, createResp.RoomID)
	if len(room.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(room.Players))
	}
	if room.Players[1].Name != "Guest" {
		t.Errorf("expected Guest, got %s", room.Players[1].Name)
	}
}

func TestJoinRoom_InvalidCode(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	tests := []struct {
		name string
		code string
	}{
		{"too short", "ABC"},
		{"too long", "ABCDEFGH"},
		{"empty", ""},
		{"with spaces", "AB CD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.JoinRoom(ctx, dto.JoinRoomRequest{Code: tt.code, PlayerName: "Guest"})
			if err == nil {
				t.Error("expected error for invalid code")
			}
		})
	}
}

func TestJoinRoom_RoomFull(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{
		HostName:   "Host",
		MaxPlayers: 2,
	})

	_, _ = uc.JoinRoom(ctx, dto.JoinRoomRequest{Code: createResp.Code, PlayerName: "P2"})
	_, err := uc.JoinRoom(ctx, dto.JoinRoomRequest{Code: createResp.Code, PlayerName: "P3"})

	if err != entity.ErrRoomFull {
		t.Errorf("expected ErrRoomFull, got %v", err)
	}

	room, _ := repo.FindByID(ctx, createResp.RoomID)
	if len(room.Players) != 2 {
		t.Errorf("expected 2 players after full room, got %d", len(room.Players))
	}
}

func TestJoinRoom_NonexistentCode(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	_, err := uc.JoinRoom(ctx, dto.JoinRoomRequest{Code: "ZZZZZZ", PlayerName: "Guest"})
	if err == nil {
		t.Error("expected error for non-existent room")
	}
}

func TestJoinRoom_PlayerGetsStartingStack(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{
		HostName:      "Host",
		StartingStack: 3000,
	})

	joinResp, _ := uc.JoinRoom(ctx, dto.JoinRoomRequest{Code: createResp.Code, PlayerName: "Guest"})

	room, _ := repo.FindByID(ctx, createResp.RoomID)
	player := room.FindPlayer(joinResp.PlayerID)
	if player.Stack != 3000 {
		t.Errorf("expected 3000 starting stack, got %d", player.Stack)
	}
}

func TestGetRoom(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})

	room, err := uc.GetRoom(ctx, createResp.RoomID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if room.ID != createResp.RoomID {
		t.Error("room ID mismatch")
	}
}

func TestGetRoom_NotFound(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	_, err := uc.GetRoom(ctx, "nonexistent")
	if err != entity.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestUpdateConfig_HostOnly(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	room, _ := repo.FindByID(ctx, createResp.RoomID)

	_, err := uc.UpdateConfig(ctx, room.ID, "wrong-player", dto.UpdateConfigRequest{
		StartingStack: 5000,
	})
	if err != entity.ErrNotHost {
		t.Errorf("expected ErrNotHost, got %v", err)
	}
}

func TestUpdateConfig_OnlyWhenWaiting(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	room, _ := repo.FindByID(ctx, createResp.RoomID)
	room.Status = entity.RoomStatusPlaying
	_ = repo.Update(ctx, room)

	_, err := uc.UpdateConfig(ctx, room.ID, room.HostPlayerID, dto.UpdateConfigRequest{
		StartingStack: 5000,
	})
	if err != entity.ErrGameInProgress {
		t.Errorf("expected ErrGameInProgress, got %v", err)
	}
}

func TestUpdateConfig_Success(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	room, _ := repo.FindByID(ctx, createResp.RoomID)

	updated, err := uc.UpdateConfig(ctx, room.ID, room.HostPlayerID, dto.UpdateConfigRequest{
		GameMode:      "tournament",
		StartingStack: 5000,
		MaxPlayers:    6,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Config.GameMode != entity.GameModeTournament {
		t.Errorf("expected tournament, got %s", updated.Config.GameMode)
	}
	if updated.Config.StartingStack != 5000 {
		t.Errorf("expected 5000, got %d", updated.Config.StartingStack)
	}
	if updated.Config.MaxPlayers != 6 {
		t.Errorf("expected 6, got %d", updated.Config.MaxPlayers)
	}
}

func TestUpdateConfig_BlindStructure(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	room, _ := repo.FindByID(ctx, createResp.RoomID)

	newBlinds := &entity.BlindStructure{
		Levels: []entity.BlindLevel{
			{SmallBlind: 10, BigBlind: 20, Duration: 300},
			{SmallBlind: 25, BigBlind: 50, Duration: 300},
		},
		CurrentLevel: 0,
	}

	updated, err := uc.UpdateConfig(ctx, room.ID, room.HostPlayerID, dto.UpdateConfigRequest{
		BlindStructure: newBlinds,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updated.Config.BlindStructure.Levels) != 2 {
		t.Errorf("expected 2 blind levels, got %d", len(updated.Config.BlindStructure.Levels))
	}
}

func TestPickSeat_Success(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	room, _ := repo.FindByID(ctx, createResp.RoomID)

	updated, err := uc.PickSeat(ctx, room.ID, room.HostPlayerID, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	player := updated.FindPlayer(room.HostPlayerID)
	if player.Seat != 3 {
		t.Errorf("expected seat 3, got %d", player.Seat)
	}
}

func TestPickSeat_InvalidRange(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	room, _ := repo.FindByID(ctx, createResp.RoomID)

	_, err := uc.PickSeat(ctx, room.ID, room.HostPlayerID, 0)
	if err != entity.ErrSeatTaken {
		t.Errorf("expected ErrSeatTaken for seat 0, got %v", err)
	}

	_, err = uc.PickSeat(ctx, room.ID, room.HostPlayerID, room.Config.MaxPlayers+1)
	if err != entity.ErrSeatTaken {
		t.Errorf("expected ErrSeatTaken for out of range, got %v", err)
	}
}

func TestPickSeat_AlreadyTaken(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	joinResp, _ := uc.JoinRoom(ctx, dto.JoinRoomRequest{Code: createResp.Code, PlayerName: "Guest"})

	room, _ := repo.FindByID(ctx, createResp.RoomID)
	_, _ = uc.PickSeat(ctx, room.ID, room.HostPlayerID, 1)

	_, err := uc.PickSeat(ctx, room.ID, joinResp.PlayerID, 1)
	if err != entity.ErrSeatTaken {
		t.Errorf("expected ErrSeatTaken, got %v", err)
	}
}

func TestPickSeat_PlayerCanChangeSeat(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	room, _ := repo.FindByID(ctx, createResp.RoomID)

	_, _ = uc.PickSeat(ctx, room.ID, room.HostPlayerID, 1)
	updated, err := uc.PickSeat(ctx, room.ID, room.HostPlayerID, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	player := updated.FindPlayer(room.HostPlayerID)
	if player.Seat != 5 {
		t.Errorf("expected seat 5, got %d", player.Seat)
	}
}

func TestPickSeat_PlayerNotFound(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	room, _ := repo.FindByID(ctx, createResp.RoomID)

	_, err := uc.PickSeat(ctx, room.ID, "nonexistent-player", 1)
	if err != entity.ErrPlayerNotFound {
		t.Errorf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestValidateToken(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})

	roomID, playerID, isHost, err := uc.ValidateToken(createResp.Token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if roomID != createResp.RoomID {
		t.Error("room ID mismatch")
	}
	if playerID == "" {
		t.Error("expected non-empty player ID")
	}
	if !isHost {
		t.Error("expected isHost to be true")
	}
}

func TestValidateToken_JoinedPlayerNotHost(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	joinResp, _ := uc.JoinRoom(ctx, dto.JoinRoomRequest{Code: createResp.Code, PlayerName: "Guest"})

	_, _, isHost, err := uc.ValidateToken(joinResp.Token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isHost {
		t.Error("joined player should not be host")
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)

	_, _, _, err := uc.ValidateToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestJoinRoom_LowercaseCodeNormalized(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})

	lowerCode := strings.ToLower(createResp.Code)
	joinResp, err := uc.JoinRoom(ctx, dto.JoinRoomRequest{
		Code:       lowerCode,
		PlayerName: "Guest",
	})
	if err != nil {
		t.Fatalf("expected lowercase code to work, got: %v", err)
	}
	if joinResp.RoomID != createResp.RoomID {
		t.Error("room IDs should match when joining with lowercase code")
	}
}

func TestJoinRoom_WhitespaceCodeTrimmed(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})

	joinResp, err := uc.JoinRoom(ctx, dto.JoinRoomRequest{
		Code:       "  " + createResp.Code + "  ",
		PlayerName: "Guest",
	})
	if err != nil {
		t.Fatalf("expected whitespace-padded code to work, got: %v", err)
	}
	if joinResp.RoomID != createResp.RoomID {
		t.Error("room IDs should match when joining with whitespace-padded code")
	}
}

func TestUpdateConfig_PartialUpdate_OnlyGameMode(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{
		HostName:      "Host",
		StartingStack: 2000,
		MaxPlayers:    8,
	})
	room, _ := repo.FindByID(ctx, createResp.RoomID)

	updated, err := uc.UpdateConfig(ctx, room.ID, room.HostPlayerID, dto.UpdateConfigRequest{
		GameMode: "tournament",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Config.GameMode != entity.GameModeTournament {
		t.Errorf("expected tournament, got %s", updated.Config.GameMode)
	}
	if updated.Config.StartingStack != 2000 {
		t.Errorf("starting stack should be unchanged at 2000, got %d", updated.Config.StartingStack)
	}
	if updated.Config.MaxPlayers != 8 {
		t.Errorf("max players should be unchanged at 8, got %d", updated.Config.MaxPlayers)
	}
}

func TestCreateRoom_TokenIsHostToken(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	resp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})

	roomID, _, isHost, err := uc.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isHost {
		t.Error("create room token should be host token")
	}
	if roomID != resp.RoomID {
		t.Error("token room ID should match created room ID")
	}
}

func TestJoinRoom_TokenIsNonHostToken(t *testing.T) {
	_, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	joinResp, _ := uc.JoinRoom(ctx, dto.JoinRoomRequest{Code: createResp.Code, PlayerName: "Guest"})

	_, _, isHost, err := uc.ValidateToken(joinResp.Token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isHost {
		t.Error("join room token should NOT be host token")
	}
}

func TestPickSeat_SamePlayerCanReSitSameSeat(t *testing.T) {
	repo, _, uc := newRoomTestDeps(t)
	ctx := context.Background()

	createResp, _ := uc.CreateRoom(ctx, dto.CreateRoomRequest{HostName: "Host"})
	room, _ := repo.FindByID(ctx, createResp.RoomID)

	_, _ = uc.PickSeat(ctx, room.ID, room.HostPlayerID, 3)
	updated, err := uc.PickSeat(ctx, room.ID, room.HostPlayerID, 3)
	if err != nil {
		t.Fatalf("same player re-picking same seat should succeed, got: %v", err)
	}
	if updated.FindPlayer(room.HostPlayerID).Seat != 3 {
		t.Error("seat should still be 3")
	}
}
