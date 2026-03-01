package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"pokertipssimulator/internal/entity"
)

type RoomRepository interface {
	Create(ctx context.Context, room *entity.Room) error
	FindByID(ctx context.Context, id string) (*entity.Room, error)
	FindByCode(ctx context.Context, code string) (*entity.Room, error)
	Update(ctx context.Context, room *entity.Room) error
	Delete(ctx context.Context, id string) error
}

type roomRepository struct {
	col *mongo.Collection
}

func NewRoomRepository(db *mongo.Database) RoomRepository {
	col := db.Collection("rooms")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &roomRepository{col: col}
}

func (r *roomRepository) Create(ctx context.Context, room *entity.Room) error {
	room.CreatedAt = time.Now()
	room.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, room)
	return err
}

func (r *roomRepository) FindByID(ctx context.Context, id string) (*entity.Room, error) {
	var room entity.Room
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&room)
	if err == mongo.ErrNoDocuments {
		return nil, entity.ErrRoomNotFound
	}
	return &room, err
}

func (r *roomRepository) FindByCode(ctx context.Context, code string) (*entity.Room, error) {
	var room entity.Room
	err := r.col.FindOne(ctx, bson.M{"code": code}).Decode(&room)
	if err == mongo.ErrNoDocuments {
		return nil, entity.ErrRoomNotFound
	}
	return &room, err
}

func (r *roomRepository) Update(ctx context.Context, room *entity.Room) error {
	room.UpdatedAt = time.Now()
	_, err := r.col.ReplaceOne(ctx, bson.M{"_id": room.ID}, room)
	return err
}

func (r *roomRepository) Delete(ctx context.Context, id string) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
