package usecase

import (
	"context"
	"log"
	"sync"
	"time"

	"pokertipssimulator/internal/entity"
	"pokertipssimulator/internal/repository"
)

type BlindTimer struct {
	repo   repository.RoomRepository
	roomID string
	stopCh chan struct{}
	mu     sync.Mutex
	onTick func(room *entity.Room)
}

func NewBlindTimer(repo repository.RoomRepository, roomID string, onTick func(room *entity.Room)) *BlindTimer {
	return &BlindTimer{
		repo:   repo,
		roomID: roomID,
		stopCh: make(chan struct{}),
		onTick: onTick,
	}
}

func (bt *BlindTimer) Start() {
	go bt.run()
}

func (bt *BlindTimer) Stop() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	select {
	case <-bt.stopCh:
	default:
		close(bt.stopCh)
	}
}

func (bt *BlindTimer) run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastAdvance := time.Now()

	for {
		select {
		case <-bt.stopCh:
			return
		case <-ticker.C:
			ctx := context.Background()
			room, err := bt.repo.FindByID(ctx, bt.roomID)
			if err != nil {
				continue
			}

			if room.Config.GameMode != entity.GameModeTournament {
				continue
			}

			if room.Status != entity.RoomStatusPlaying {
				lastAdvance = time.Now()
				continue
			}

			bs := room.Config.BlindStructure
			if bs.CurrentLevel >= len(bs.Levels)-1 {
				continue
			}

			currentLevel := bs.Levels[bs.CurrentLevel]
			if currentLevel.Duration <= 0 {
				continue
			}

			if time.Since(lastAdvance) >= time.Duration(currentLevel.Duration)*time.Second {
				room.Config.BlindStructure.CurrentLevel++
				if err := bt.repo.Update(ctx, room); err != nil {
					log.Printf("blind timer: failed to update room: %v", err)
					continue
				}
				lastAdvance = time.Now()

				if bt.onTick != nil {
					bt.onTick(room)
				}
			}
		}
	}
}
