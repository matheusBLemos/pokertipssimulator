package dto

import "pokertipssimulator/internal/domain/entity"

type CreateRoomRequest struct {
	HostName      string `json:"host_name"`
	RoomMode      string `json:"room_mode"`
	GameMode      string `json:"game_mode"`
	StartingStack int    `json:"starting_stack"`
	MaxPlayers    int    `json:"max_players"`
}

type JoinRoomRequest struct {
	Code       string `json:"code"`
	PlayerName string `json:"player_name"`
}

type PickSeatRequest struct {
	Seat int `json:"seat"`
}

type UpdateConfigRequest struct {
	GameMode       string                 `json:"game_mode,omitempty"`
	StartingStack  int                    `json:"starting_stack,omitempty"`
	MaxPlayers     int                    `json:"max_players,omitempty"`
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

type TransferChipsRequest struct {
	FromPlayerID string `json:"from_player_id"`
	ToPlayerID   string `json:"to_player_id"`
	Amount       int    `json:"amount"`
}
