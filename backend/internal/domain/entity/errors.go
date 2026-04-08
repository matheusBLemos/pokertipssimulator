package entity

import "errors"

var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrPlayerNotFound    = errors.New("player not found")
	ErrRoomFull          = errors.New("room is full")
	ErrSeatTaken         = errors.New("seat is already taken")
	ErrInvalidAction     = errors.New("invalid action")
	ErrNotYourTurn       = errors.New("not your turn")
	ErrInsufficientStack = errors.New("insufficient stack")
	ErrGameInProgress    = errors.New("game is in progress")
	ErrGameNotStarted    = errors.New("game has not started")
	ErrNotHost           = errors.New("only the host can perform this action")
	ErrInvalidAmount     = errors.New("invalid bet amount")
	ErrNotEnoughPlayers  = errors.New("not enough players to start")
	ErrAlreadyJoined     = errors.New("player already in room")
	ErrInvalidCode       = errors.New("invalid room code")
	ErrRoundComplete     = errors.New("round is already complete")
	ErrInvalidStreet     = errors.New("invalid street advancement")
	ErrPlayerEliminated  = errors.New("player is eliminated")
)
