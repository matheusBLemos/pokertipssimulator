package application

import (
	"context"
	"testing"

	"pokertipssimulator/internal/adapter/repository"
	"pokertipssimulator/internal/application/dto"
	"pokertipssimulator/internal/domain/entity"
	"pokertipssimulator/internal/infrastructure/auth"
)

func newTipsTestDeps(t *testing.T) (*RoomUseCase, *TipsUseCase) {
	t.Helper()
	db := repository.NewTestDB(t)
	repo := repository.NewSQLiteRoomRepository(db)
	jwt := auth.NewJWTService("test-secret")
	return NewRoomUseCase(repo, jwt), NewTipsUseCase(repo)
}

func createTipsRoom(t *testing.T, roomUC *RoomUseCase) *dto.CreateRoomResponse {
	t.Helper()
	resp, err := roomUC.CreateRoom(context.Background(), dto.CreateRoomRequest{
		HostName:      "Host",
		RoomMode:      "tips",
		StartingStack: 1000,
	})
	if err != nil {
		t.Fatalf("create tips room: %v", err)
	}
	return resp
}

func TestTransferChips_Success(t *testing.T) {
	roomUC, tipsUC := newTipsTestDeps(t)
	ctx := context.Background()

	createResp := createTipsRoom(t, roomUC)
	joinResp, _ := roomUC.JoinRoom(ctx, dto.JoinRoomRequest{
		Code: createResp.Code, PlayerName: "Guest",
	})
	room, _ := roomUC.GetRoom(ctx, createResp.RoomID)
	hostID := room.HostPlayerID

	updated, err := tipsUC.TransferChips(ctx, createResp.RoomID, dto.TransferChipsRequest{
		FromPlayerID: hostID,
		ToPlayerID:   joinResp.PlayerID,
		Amount:       200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	from := updated.FindPlayer(hostID)
	to := updated.FindPlayer(joinResp.PlayerID)
	if from.Stack != 800 {
		t.Errorf("expected from stack 800, got %d", from.Stack)
	}
	if to.Stack != 1200 {
		t.Errorf("expected to stack 1200, got %d", to.Stack)
	}
}

func TestTransferChips_InsufficientStack(t *testing.T) {
	roomUC, tipsUC := newTipsTestDeps(t)
	ctx := context.Background()

	createResp := createTipsRoom(t, roomUC)
	joinResp, _ := roomUC.JoinRoom(ctx, dto.JoinRoomRequest{
		Code: createResp.Code, PlayerName: "Guest",
	})
	room, _ := roomUC.GetRoom(ctx, createResp.RoomID)

	_, err := tipsUC.TransferChips(ctx, createResp.RoomID, dto.TransferChipsRequest{
		FromPlayerID: room.HostPlayerID,
		ToPlayerID:   joinResp.PlayerID,
		Amount:       5000,
	})
	if err != entity.ErrInsufficientStack {
		t.Errorf("expected ErrInsufficientStack, got %v", err)
	}
}

func TestTransferChips_SamePlayer(t *testing.T) {
	roomUC, tipsUC := newTipsTestDeps(t)
	ctx := context.Background()

	createResp := createTipsRoom(t, roomUC)
	room, _ := roomUC.GetRoom(ctx, createResp.RoomID)

	_, err := tipsUC.TransferChips(ctx, createResp.RoomID, dto.TransferChipsRequest{
		FromPlayerID: room.HostPlayerID,
		ToPlayerID:   room.HostPlayerID,
		Amount:       100,
	})
	if err != entity.ErrSamePlayer {
		t.Errorf("expected ErrSamePlayer, got %v", err)
	}
}

func TestTransferChips_InvalidAmount(t *testing.T) {
	roomUC, tipsUC := newTipsTestDeps(t)
	ctx := context.Background()

	createResp := createTipsRoom(t, roomUC)
	joinResp, _ := roomUC.JoinRoom(ctx, dto.JoinRoomRequest{
		Code: createResp.Code, PlayerName: "Guest",
	})
	room, _ := roomUC.GetRoom(ctx, createResp.RoomID)

	_, err := tipsUC.TransferChips(ctx, createResp.RoomID, dto.TransferChipsRequest{
		FromPlayerID: room.HostPlayerID,
		ToPlayerID:   joinResp.PlayerID,
		Amount:       0,
	})
	if err != entity.ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestTransferChips_WrongRoomMode(t *testing.T) {
	roomUC, tipsUC := newTipsTestDeps(t)
	ctx := context.Background()

	resp, _ := roomUC.CreateRoom(ctx, dto.CreateRoomRequest{
		HostName: "Host",
		RoomMode: "game",
	})
	joinResp, _ := roomUC.JoinRoom(ctx, dto.JoinRoomRequest{
		Code: resp.Code, PlayerName: "Guest",
	})
	room, _ := roomUC.GetRoom(ctx, resp.RoomID)

	_, err := tipsUC.TransferChips(ctx, resp.RoomID, dto.TransferChipsRequest{
		FromPlayerID: room.HostPlayerID,
		ToPlayerID:   joinResp.PlayerID,
		Amount:       100,
	})
	if err != entity.ErrWrongRoomMode {
		t.Errorf("expected ErrWrongRoomMode, got %v", err)
	}
}

func TestAdvanceBlindLevel_Success(t *testing.T) {
	roomUC, tipsUC := newTipsTestDeps(t)
	ctx := context.Background()

	createResp := createTipsRoom(t, roomUC)
	room, _ := roomUC.GetRoom(ctx, createResp.RoomID)

	_, _ = roomUC.UpdateConfig(ctx, room.ID, room.HostPlayerID, dto.UpdateConfigRequest{
		BlindStructure: &entity.BlindStructure{
			Levels: []entity.BlindLevel{
				{SmallBlind: 5, BigBlind: 10},
				{SmallBlind: 10, BigBlind: 20},
				{SmallBlind: 25, BigBlind: 50},
			},
			CurrentLevel: 0,
		},
	})

	updated, err := tipsUC.AdvanceBlindLevel(ctx, room.ID, room.HostPlayerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Config.BlindStructure.CurrentLevel != 1 {
		t.Errorf("expected level 1, got %d", updated.Config.BlindStructure.CurrentLevel)
	}
}

func TestAdvanceBlindLevel_NotHost(t *testing.T) {
	roomUC, tipsUC := newTipsTestDeps(t)
	ctx := context.Background()

	createResp := createTipsRoom(t, roomUC)
	joinResp, _ := roomUC.JoinRoom(ctx, dto.JoinRoomRequest{
		Code: createResp.Code, PlayerName: "Guest",
	})

	_, err := tipsUC.AdvanceBlindLevel(ctx, createResp.RoomID, joinResp.PlayerID)
	if err != entity.ErrNotHost {
		t.Errorf("expected ErrNotHost, got %v", err)
	}
}

func TestAdvanceBlindLevel_NoMoreLevels(t *testing.T) {
	roomUC, tipsUC := newTipsTestDeps(t)
	ctx := context.Background()

	createResp := createTipsRoom(t, roomUC)
	room, _ := roomUC.GetRoom(ctx, createResp.RoomID)

	_, err := tipsUC.AdvanceBlindLevel(ctx, room.ID, room.HostPlayerID)
	if err != entity.ErrNoBlindLevels {
		t.Errorf("expected ErrNoBlindLevels (only 1 level), got %v", err)
	}
}

func TestPauseTimer_Toggle(t *testing.T) {
	roomUC, tipsUC := newTipsTestDeps(t)
	ctx := context.Background()

	createResp := createTipsRoom(t, roomUC)
	room, _ := roomUC.GetRoom(ctx, createResp.RoomID)

	updated, err := tipsUC.PauseTimer(ctx, room.ID, room.HostPlayerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != entity.RoomStatusPlaying {
		t.Errorf("expected playing, got %s", updated.Status)
	}

	updated2, err := tipsUC.PauseTimer(ctx, room.ID, room.HostPlayerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated2.Status != entity.RoomStatusPaused {
		t.Errorf("expected paused, got %s", updated2.Status)
	}

	updated3, err := tipsUC.PauseTimer(ctx, room.ID, room.HostPlayerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated3.Status != entity.RoomStatusPlaying {
		t.Errorf("expected playing again, got %s", updated3.Status)
	}
}

func TestKickPlayer_Tips(t *testing.T) {
	roomUC, tipsUC := newTipsTestDeps(t)
	ctx := context.Background()

	createResp := createTipsRoom(t, roomUC)
	joinResp, _ := roomUC.JoinRoom(ctx, dto.JoinRoomRequest{
		Code: createResp.Code, PlayerName: "Guest",
	})
	room, _ := roomUC.GetRoom(ctx, createResp.RoomID)

	updated, err := tipsUC.KickPlayer(ctx, room.ID, joinResp.PlayerID, room.HostPlayerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updated.Players) != 1 {
		t.Errorf("expected 1 player after kick, got %d", len(updated.Players))
	}
}

func TestRebuy_Tips(t *testing.T) {
	roomUC, tipsUC := newTipsTestDeps(t)
	ctx := context.Background()

	createResp := createTipsRoom(t, roomUC)
	joinResp, _ := roomUC.JoinRoom(ctx, dto.JoinRoomRequest{
		Code: createResp.Code, PlayerName: "Guest",
	})
	room, _ := roomUC.GetRoom(ctx, createResp.RoomID)

	updated, err := tipsUC.Rebuy(ctx, room.ID, joinResp.PlayerID, room.HostPlayerID, 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	player := updated.FindPlayer(joinResp.PlayerID)
	if player.Stack != 1500 {
		t.Errorf("expected 1500 after rebuy, got %d", player.Stack)
	}
}
