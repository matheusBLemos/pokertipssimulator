package repository

import (
	"context"
	"testing"

	"pokertipssimulator/internal/domain/entity"
)

func TestSQLiteRepo_CreateAndFindByID(t *testing.T) {
	db := NewTestDB(t)
	repo := NewSQLiteRoomRepository(db)
	ctx := context.Background()

	room := &entity.Room{
		ID:   "room-1",
		Code: "ABC123",
		Config: entity.RoomConfig{
			StartingStack: 1000,
			MaxPlayers:    10,
			ChipSet:       entity.DefaultChipSet(),
		},
		Players: []entity.Player{
			{ID: "p1", Name: "Host", Stack: 1000, Status: entity.PlayerStatusWaiting},
		},
	}

	if err := repo.Create(ctx, room); err != nil {
		t.Fatalf("create: %v", err)
	}

	found, err := repo.FindByID(ctx, "room-1")
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}

	if found.ID != "room-1" {
		t.Errorf("expected room-1, got %s", found.ID)
	}
	if found.Code != "ABC123" {
		t.Errorf("expected ABC123, got %s", found.Code)
	}
	if len(found.Players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(found.Players))
	}
	if found.Players[0].Name != "Host" {
		t.Errorf("expected Host, got %s", found.Players[0].Name)
	}
}

func TestSQLiteRepo_FindByCode(t *testing.T) {
	db := NewTestDB(t)
	repo := NewSQLiteRoomRepository(db)
	ctx := context.Background()

	room := &entity.Room{
		ID:   "room-1",
		Code: "XYZ789",
	}
	_ = repo.Create(ctx, room)

	found, err := repo.FindByCode(ctx, "XYZ789")
	if err != nil {
		t.Fatalf("find by code: %v", err)
	}
	if found.ID != "room-1" {
		t.Errorf("expected room-1, got %s", found.ID)
	}
}

func TestSQLiteRepo_FindByID_NotFound(t *testing.T) {
	db := NewTestDB(t)
	repo := NewSQLiteRoomRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "nonexistent")
	if err != entity.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestSQLiteRepo_FindByCode_NotFound(t *testing.T) {
	db := NewTestDB(t)
	repo := NewSQLiteRoomRepository(db)
	ctx := context.Background()

	_, err := repo.FindByCode(ctx, "XXXXXX")
	if err != entity.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestSQLiteRepo_Update(t *testing.T) {
	db := NewTestDB(t)
	repo := NewSQLiteRoomRepository(db)
	ctx := context.Background()

	room := &entity.Room{
		ID:      "room-1",
		Code:    "ABC123",
		Status:  entity.RoomStatusWaiting,
		Players: []entity.Player{{ID: "p1", Stack: 1000}},
	}
	_ = repo.Create(ctx, room)

	room.Status = entity.RoomStatusPlaying
	room.Players[0].Stack = 500
	if err := repo.Update(ctx, room); err != nil {
		t.Fatalf("update: %v", err)
	}

	found, _ := repo.FindByID(ctx, "room-1")
	if found.Status != entity.RoomStatusPlaying {
		t.Errorf("expected playing, got %s", found.Status)
	}
	if found.Players[0].Stack != 500 {
		t.Errorf("expected 500, got %d", found.Players[0].Stack)
	}
}

func TestSQLiteRepo_Delete(t *testing.T) {
	db := NewTestDB(t)
	repo := NewSQLiteRoomRepository(db)
	ctx := context.Background()

	room := &entity.Room{ID: "room-1", Code: "ABC123"}
	_ = repo.Create(ctx, room)

	if err := repo.Delete(ctx, "room-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := repo.FindByID(ctx, "room-1")
	if err != entity.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound after delete, got %v", err)
	}
}

func TestSQLiteRepo_RoundTrip_WithRound(t *testing.T) {
	db := NewTestDB(t)
	repo := NewSQLiteRoomRepository(db)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "p1",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{
					{SmallBlind: 5, BigBlind: 10, Duration: 600},
				},
				CurrentLevel: 0,
			},
		},
		Players: []entity.Player{
			{ID: "p1", Seat: 1, Stack: 990, Status: entity.PlayerStatusActive},
			{ID: "p2", Seat: 2, Stack: 990, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetPreflop,
			DealerSeat:  1,
			SmallBlind:  5,
			BigBlind:    10,
			CurrentBet:  10,
			MinRaise:    10,
			CurrentTurn: "p2",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "p1", Bet: 5, TotalBet: 5},
				{PlayerID: "p2", Bet: 10, TotalBet: 10},
			},
			Pots: []entity.Pot{{Amount: 15, EligibleIDs: []string{"p1", "p2"}}},
			Actions: []entity.Action{
				{PlayerID: "p1", Type: entity.ActionCall, Amount: 5, Street: entity.StreetPreflop},
			},
		},
		RoundCount: 1,
	}
	_ = repo.Create(ctx, room)

	found, err := repo.FindByID(ctx, "room-1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}

	if found.Round == nil {
		t.Fatal("expected round to be preserved")
	}
	if found.Round.Street != entity.StreetPreflop {
		t.Errorf("expected preflop, got %s", found.Round.Street)
	}
	if found.Round.CurrentTurn != "p2" {
		t.Errorf("expected p2, got %s", found.Round.CurrentTurn)
	}
	if len(found.Round.PlayerStates) != 2 {
		t.Errorf("expected 2 player states, got %d", len(found.Round.PlayerStates))
	}
	if len(found.Round.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(found.Round.Actions))
	}
	if found.Round.Pots[0].Amount != 15 {
		t.Errorf("expected pot 15, got %d", found.Round.Pots[0].Amount)
	}
}

func TestSQLiteRepo_UniqueCode(t *testing.T) {
	db := NewTestDB(t)
	repo := NewSQLiteRoomRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.Room{ID: "room-1", Code: "ABC123"})
	err := repo.Create(ctx, &entity.Room{ID: "room-2", Code: "ABC123"})
	if err == nil {
		t.Error("expected error for duplicate code")
	}
}
