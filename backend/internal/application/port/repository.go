package port

import (
	"context"

	"pokertipssimulator/internal/domain/entity"
)

type RoomRepository interface {
	Create(ctx context.Context, room *entity.Room) error
	FindByID(ctx context.Context, id string) (*entity.Room, error)
	FindByCode(ctx context.Context, code string) (*entity.Room, error)
	Update(ctx context.Context, room *entity.Room) error
	Delete(ctx context.Context, id string) error
}
