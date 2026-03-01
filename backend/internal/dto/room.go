package dto

import "pokertipssimulator/internal/entity"

type CreateRoomRequest struct {
	HostName      string `json:"host_name"`
	GameMode      string `json:"game_mode"`
	StartingStack int    `json:"starting_stack"`
	MaxPlayers    int    `json:"max_players"`
}

type CreateRoomResponse struct {
	RoomID string `json:"room_id"`
	Code   string `json:"code"`
	Token  string `json:"token"`
}

type JoinRoomRequest struct {
	Code       string `json:"code"`
	PlayerName string `json:"player_name"`
}

type JoinRoomResponse struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
	Token    string `json:"token"`
}

type PickSeatRequest struct {
	Seat int `json:"seat"`
}

type UpdateConfigRequest struct {
	GameMode       string                `json:"game_mode,omitempty"`
	StartingStack  int                   `json:"starting_stack,omitempty"`
	MaxPlayers     int                   `json:"max_players,omitempty"`
	BlindStructure *entity.BlindStructure `json:"blind_structure,omitempty"`
}

type ActionRequest struct {
	Type   string `json:"type"`
	Amount int    `json:"amount,omitempty"`
}

type SettleRequest struct {
	Winners []PotWinner `json:"winners"`
}

type PotWinner struct {
	PotIndex  int      `json:"pot_index"`
	PlayerIDs []string `json:"player_ids"`
}

type RebuyRequest struct {
	Amount int `json:"amount"`
}

type RoomResponse struct {
	*entity.Room
}

type ErrorResponse struct {
	Error string `json:"error"`
}
