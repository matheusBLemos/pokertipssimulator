package entity

import "time"

type GameMode string

const (
	GameModeCash       GameMode = "cash"
	GameModeTournament GameMode = "tournament"
)

type RoomMode string

const (
	RoomModeGame RoomMode = "game"
	RoomModeTips RoomMode = "tips"
)

type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusPlaying  RoomStatus = "playing"
	RoomStatusPaused   RoomStatus = "paused"
	RoomStatusFinished RoomStatus = "finished"
)

type RoomConfig struct {
	GameMode       GameMode       `json:"game_mode"`
	StartingStack  int            `json:"starting_stack"`
	ChipSet        ChipSet        `json:"chip_set"`
	BlindStructure BlindStructure `json:"blind_structure"`
	MaxPlayers     int            `json:"max_players"`
	MaxRebuy       int            `json:"max_rebuy"`
}

type Room struct {
	ID           string     `json:"id"`
	Code         string     `json:"code"`
	Mode         RoomMode   `json:"mode"`
	Status       RoomStatus `json:"status"`
	HostPlayerID string     `json:"host_player_id"`
	Config       RoomConfig `json:"config"`
	Players      []Player   `json:"players"`
	Round        *Round     `json:"round,omitempty"`
	RoundCount   int        `json:"round_count"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
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

// FilterForPlayer returns a copy of the room with sensitive data stripped:
// the deck is removed, and other players' hole cards are hidden (unless showdown).
func (r *Room) FilterForPlayer(playerID string) *Room {
	if r.Round == nil {
		return r
	}

	copy := *r
	roundCopy := *r.Round
	copy.Round = &roundCopy

	// Never send the deck to clients
	roundCopy.Deck = nil

	isShowdown := roundCopy.Street == StreetShowdown || roundCopy.IsComplete

	states := make([]PlayerState, len(roundCopy.PlayerStates))
	for i, ps := range roundCopy.PlayerStates {
		states[i] = ps
		if ps.PlayerID != playerID && !isShowdown {
			states[i].HoleCards = nil
		}
	}
	roundCopy.PlayerStates = states

	return &copy
}
