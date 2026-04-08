package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"pokertipssimulator/internal/domain/entity"
)

type SQLiteRoomRepository struct {
	db *sql.DB
}

func NewSQLiteRoomRepository(db *sql.DB) *SQLiteRoomRepository {
	return &SQLiteRoomRepository{db: db}
}

func (r *SQLiteRoomRepository) Create(ctx context.Context, room *entity.Room) error {
	now := time.Now()
	room.CreatedAt = now
	room.UpdatedAt = now

	data, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("marshal room: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO rooms (id, code, data, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		room.ID, room.Code, string(data), now, now,
	)
	return err
}

func (r *SQLiteRoomRepository) FindByID(ctx context.Context, id string) (*entity.Room, error) {
	var data string
	err := r.db.QueryRowContext(ctx, "SELECT data FROM rooms WHERE id = ?", id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, entity.ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}

	var room entity.Room
	if err := json.Unmarshal([]byte(data), &room); err != nil {
		return nil, fmt.Errorf("unmarshal room: %w", err)
	}
	return &room, nil
}

func (r *SQLiteRoomRepository) FindByCode(ctx context.Context, code string) (*entity.Room, error) {
	var data string
	err := r.db.QueryRowContext(ctx, "SELECT data FROM rooms WHERE code = ?", code).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, entity.ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}

	var room entity.Room
	if err := json.Unmarshal([]byte(data), &room); err != nil {
		return nil, fmt.Errorf("unmarshal room: %w", err)
	}
	return &room, nil
}

func (r *SQLiteRoomRepository) Update(ctx context.Context, room *entity.Room) error {
	room.UpdatedAt = time.Now()

	data, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("marshal room: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		"UPDATE rooms SET code = ?, data = ?, updated_at = ? WHERE id = ?",
		room.Code, string(data), room.UpdatedAt, room.ID,
	)
	return err
}

func (r *SQLiteRoomRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM rooms WHERE id = ?", id)
	return err
}
