package entity

type ActionType string

const (
	ActionFold  ActionType = "fold"
	ActionCheck ActionType = "check"
	ActionCall  ActionType = "call"
	ActionBet   ActionType = "bet"
	ActionRaise ActionType = "raise"
	ActionAllIn ActionType = "allin"
)

type Action struct {
	PlayerID string     `json:"player_id"`
	Type     ActionType `json:"type"`
	Amount   int        `json:"amount"`
	Street   Street     `json:"street"`
}
