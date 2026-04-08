package application

import (
	"context"
	"sync"
	"testing"
	"time"

	"pokertipssimulator/internal/application/mock"
	"pokertipssimulator/internal/domain/entity"
)

func TestBlindTimer_AdvancesLevel(t *testing.T) {
	repo := mock.NewRoomRepository()
	ctx := context.Background()

	room := &entity.Room{
		ID:     "room-1",
		Code:   "ABC123",
		Status: entity.RoomStatusPlaying,
		Config: entity.RoomConfig{
			GameMode: entity.GameModeTournament,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{
					{SmallBlind: 5, BigBlind: 10, Duration: 1},
					{SmallBlind: 10, BigBlind: 20, Duration: 1},
					{SmallBlind: 25, BigBlind: 50, Duration: 1},
				},
				CurrentLevel: 0,
			},
		},
	}
	repo.Seed(room)

	var mu sync.Mutex
	ticks := 0
	timer := NewBlindTimer(repo, "room-1", func(r *entity.Room) {
		mu.Lock()
		ticks++
		mu.Unlock()
	})

	timer.Start()
	time.Sleep(2500 * time.Millisecond)
	timer.Stop()

	updated, _ := repo.FindByID(ctx, "room-1")

	mu.Lock()
	tickCount := ticks
	mu.Unlock()

	if tickCount == 0 {
		t.Error("expected at least one tick/level advance")
	}
	if updated.Config.BlindStructure.CurrentLevel == 0 {
		t.Error("expected blind level to advance from 0")
	}
}

func TestBlindTimer_StopsAtLastLevel(t *testing.T) {
	repo := mock.NewRoomRepository()
	ctx := context.Background()

	room := &entity.Room{
		ID:     "room-1",
		Code:   "ABC123",
		Status: entity.RoomStatusPlaying,
		Config: entity.RoomConfig{
			GameMode: entity.GameModeTournament,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{
					{SmallBlind: 5, BigBlind: 10, Duration: 1},
				},
				CurrentLevel: 0,
			},
		},
	}
	repo.Seed(room)

	timer := NewBlindTimer(repo, "room-1", nil)
	timer.Start()
	time.Sleep(2 * time.Second)
	timer.Stop()

	updated, _ := repo.FindByID(ctx, "room-1")
	if updated.Config.BlindStructure.CurrentLevel != 0 {
		t.Errorf("should not advance past last level, got %d", updated.Config.BlindStructure.CurrentLevel)
	}
}

func TestBlindTimer_SkipsCashGame(t *testing.T) {
	repo := mock.NewRoomRepository()
	ctx := context.Background()

	room := &entity.Room{
		ID:     "room-1",
		Code:   "ABC123",
		Status: entity.RoomStatusPlaying,
		Config: entity.RoomConfig{
			GameMode: entity.GameModeCash,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{
					{SmallBlind: 5, BigBlind: 10, Duration: 1},
					{SmallBlind: 10, BigBlind: 20, Duration: 1},
				},
				CurrentLevel: 0,
			},
		},
	}
	repo.Seed(room)

	timer := NewBlindTimer(repo, "room-1", nil)
	timer.Start()
	time.Sleep(2 * time.Second)
	timer.Stop()

	updated, _ := repo.FindByID(ctx, "room-1")
	if updated.Config.BlindStructure.CurrentLevel != 0 {
		t.Error("cash game blinds should not auto-advance")
	}
}

func TestBlindTimer_PausedGameDoesNotAdvance(t *testing.T) {
	repo := mock.NewRoomRepository()
	ctx := context.Background()

	room := &entity.Room{
		ID:     "room-1",
		Code:   "ABC123",
		Status: entity.RoomStatusPaused,
		Config: entity.RoomConfig{
			GameMode: entity.GameModeTournament,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{
					{SmallBlind: 5, BigBlind: 10, Duration: 1},
					{SmallBlind: 10, BigBlind: 20, Duration: 1},
				},
				CurrentLevel: 0,
			},
		},
	}
	repo.Seed(room)

	timer := NewBlindTimer(repo, "room-1", nil)
	timer.Start()
	time.Sleep(2 * time.Second)
	timer.Stop()

	updated, _ := repo.FindByID(ctx, "room-1")
	if updated.Config.BlindStructure.CurrentLevel != 0 {
		t.Error("paused game blinds should not advance")
	}
}

func TestBlindTimer_ManualDurationSkipsAdvance(t *testing.T) {
	repo := mock.NewRoomRepository()
	ctx := context.Background()

	room := &entity.Room{
		ID:     "room-1",
		Code:   "ABC123",
		Status: entity.RoomStatusPlaying,
		Config: entity.RoomConfig{
			GameMode: entity.GameModeTournament,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{
					{SmallBlind: 5, BigBlind: 10, Duration: 0},
					{SmallBlind: 10, BigBlind: 20, Duration: 0},
				},
				CurrentLevel: 0,
			},
		},
	}
	repo.Seed(room)

	timer := NewBlindTimer(repo, "room-1", nil)
	timer.Start()
	time.Sleep(2 * time.Second)
	timer.Stop()

	updated, _ := repo.FindByID(ctx, "room-1")
	if updated.Config.BlindStructure.CurrentLevel != 0 {
		t.Error("manual duration (0) should not auto-advance")
	}
}

func TestBlindTimer_StopIsIdempotent(t *testing.T) {
	repo := mock.NewRoomRepository()

	room := &entity.Room{
		ID:     "room-1",
		Code:   "ABC123",
		Status: entity.RoomStatusPlaying,
		Config: entity.RoomConfig{
			GameMode: entity.GameModeTournament,
			BlindStructure: entity.BlindStructure{
				Levels:       []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10, Duration: 10}},
				CurrentLevel: 0,
			},
		},
	}
	repo.Seed(room)

	timer := NewBlindTimer(repo, "room-1", nil)
	timer.Start()

	timer.Stop()
	timer.Stop()
	timer.Stop()
}
