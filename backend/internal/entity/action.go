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
	PlayerID string     `bson:"player_id" json:"player_id"`
	Type     ActionType `bson:"type" json:"type"`
	Amount   int        `bson:"amount" json:"amount"`
	Street   Street     `bson:"street" json:"street"`
}
