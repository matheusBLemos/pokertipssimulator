package dto

import "pokertipssimulator/internal/domain/entity"

type CreateRoomResponse struct {
	RoomID string `json:"room_id"`
	Code   string `json:"code"`
	Token  string `json:"token"`
}

type JoinRoomResponse struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
	Token    string `json:"token"`
}

type RoomResponse struct {
	*entity.Room
}

type ErrorResponse struct {
	Error string `json:"error"`
}
