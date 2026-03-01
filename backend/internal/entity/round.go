package entity

type Street string

const (
	StreetPreflop  Street = "preflop"
	StreetFlop     Street = "flop"
	StreetTurn     Street = "turn"
	StreetRiver    Street = "river"
	StreetShowdown Street = "showdown"
)

type PlayerState struct {
	PlayerID   string `bson:"player_id" json:"player_id"`
	Bet        int    `bson:"bet" json:"bet"`
	TotalBet   int    `bson:"total_bet" json:"total_bet"`
	HasActed   bool   `bson:"has_acted" json:"has_acted"`
	Folded     bool   `bson:"folded" json:"folded"`
	AllIn      bool   `bson:"all_in" json:"all_in"`
}

type Round struct {
	Number       int           `bson:"number" json:"number"`
	Street       Street        `bson:"street" json:"street"`
	DealerSeat   int           `bson:"dealer_seat" json:"dealer_seat"`
	SmallBlind   int           `bson:"small_blind" json:"small_blind"`
	BigBlind     int           `bson:"big_blind" json:"big_blind"`
	CurrentTurn  string        `bson:"current_turn" json:"current_turn"`
	CurrentBet   int           `bson:"current_bet" json:"current_bet"`
	MinRaise     int           `bson:"min_raise" json:"min_raise"`
	PlayerStates []PlayerState `bson:"player_states" json:"player_states"`
	Pots         []Pot         `bson:"pots" json:"pots"`
	Actions      []Action      `bson:"actions" json:"actions"`
	IsComplete   bool          `bson:"is_complete" json:"is_complete"`
}
