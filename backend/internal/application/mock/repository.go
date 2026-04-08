package mock

import (
	"context"
	"sync"

	"pokertipssimulator/internal/domain/entity"
)

type RoomRepository struct {
	mu    sync.RWMutex
	rooms map[string]*entity.Room
	codes map[string]string

	CreateFunc   func(ctx context.Context, room *entity.Room) error
	FindByIDFunc func(ctx context.Context, id string) (*entity.Room, error)
	FindByCodeFunc func(ctx context.Context, code string) (*entity.Room, error)
	UpdateFunc   func(ctx context.Context, room *entity.Room) error
	DeleteFunc   func(ctx context.Context, id string) error
}

func NewRoomRepository() *RoomRepository {
	r := &RoomRepository{
		rooms: make(map[string]*entity.Room),
		codes: make(map[string]string),
	}
	r.CreateFunc = r.defaultCreate
	r.FindByIDFunc = r.defaultFindByID
	r.FindByCodeFunc = r.defaultFindByCode
	r.UpdateFunc = r.defaultUpdate
	r.DeleteFunc = r.defaultDelete
	return r
}

func (r *RoomRepository) Create(ctx context.Context, room *entity.Room) error {
	return r.CreateFunc(ctx, room)
}

func (r *RoomRepository) FindByID(ctx context.Context, id string) (*entity.Room, error) {
	return r.FindByIDFunc(ctx, id)
}

func (r *RoomRepository) FindByCode(ctx context.Context, code string) (*entity.Room, error) {
	return r.FindByCodeFunc(ctx, code)
}

func (r *RoomRepository) Update(ctx context.Context, room *entity.Room) error {
	return r.UpdateFunc(ctx, room)
}

func (r *RoomRepository) Delete(ctx context.Context, id string) error {
	return r.DeleteFunc(ctx, id)
}

func (r *RoomRepository) defaultCreate(_ context.Context, room *entity.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rooms[room.ID] = room
	r.codes[room.Code] = room.ID
	return nil
}

func (r *RoomRepository) defaultFindByID(_ context.Context, id string) (*entity.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	room, ok := r.rooms[id]
	if !ok {
		return nil, entity.ErrRoomNotFound
	}
	return room, nil
}

func (r *RoomRepository) defaultFindByCode(_ context.Context, code string) (*entity.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.codes[code]
	if !ok {
		return nil, entity.ErrRoomNotFound
	}
	room, ok := r.rooms[id]
	if !ok {
		return nil, entity.ErrRoomNotFound
	}
	return room, nil
}

func (r *RoomRepository) defaultUpdate(_ context.Context, room *entity.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rooms[room.ID] = room
	r.codes[room.Code] = room.ID
	return nil
}

func (r *RoomRepository) defaultDelete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if room, ok := r.rooms[id]; ok {
		delete(r.codes, room.Code)
	}
	delete(r.rooms, id)
	return nil
}

func (r *RoomRepository) Seed(room *entity.Room) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rooms[room.ID] = room
	r.codes[room.Code] = room.ID
}
