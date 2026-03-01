package entity

import "time"

type GameMode string

const (
	GameModeCash       GameMode = "cash"
	GameModeTournament GameMode = "tournament"
)

type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusPlaying  RoomStatus = "playing"
	RoomStatusPaused   RoomStatus = "paused"
	RoomStatusFinished RoomStatus = "finished"
)

type RoomConfig struct {
	GameMode       GameMode       `bson:"game_mode" json:"game_mode"`
	StartingStack  int            `bson:"starting_stack" json:"starting_stack"`
	ChipSet        ChipSet        `bson:"chip_set" json:"chip_set"`
	BlindStructure BlindStructure `bson:"blind_structure" json:"blind_structure"`
	MaxPlayers     int            `bson:"max_players" json:"max_players"`
	MaxRebuy       int            `bson:"max_rebuy" json:"max_rebuy"` // cash game only, 0 = unlimited
}

type Room struct {
	ID           string     `bson:"_id,omitempty" json:"id"`
	Code         string     `bson:"code" json:"code"`
	Status       RoomStatus `bson:"status" json:"status"`
	HostPlayerID string     `bson:"host_player_id" json:"host_player_id"`
	Config       RoomConfig `bson:"config" json:"config"`
	Players      []Player   `bson:"players" json:"players"`
	Round        *Round     `bson:"round,omitempty" json:"round,omitempty"`
	RoundCount   int        `bson:"round_count" json:"round_count"`
	CreatedAt    time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at" json:"updated_at"`
}

func (r *Room) FindPlayer(playerID string) *Player {
	for i := range r.Players {
		if r.Players[i].ID == playerID {
			return &r.Players[i]
		}
	}
	return nil
}

func (r *Room) FindPlayerState(playerID string) *PlayerState {
	if r.Round == nil {
		return nil
	}
	for i := range r.Round.PlayerStates {
		if r.Round.PlayerStates[i].PlayerID == playerID {
			return &r.Round.PlayerStates[i]
		}
	}
	return nil
}

func (r *Room) ActivePlayers() []Player {
	var active []Player
	for _, p := range r.Players {
		if p.Status == PlayerStatusActive {
			active = append(active, p)
		}
	}
	return active
}

func (r *Room) SeatedPlayers() []Player {
	var seated []Player
	for _, p := range r.Players {
		if p.Seat > 0 {
			seated = append(seated, p)
		}
	}
	return seated
}
