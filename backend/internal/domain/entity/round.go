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
	PlayerID string `json:"player_id"`
	Bet      int    `json:"bet"`
	TotalBet int    `json:"total_bet"`
	HasActed bool   `json:"has_acted"`
	Folded   bool   `json:"folded"`
	AllIn    bool   `json:"all_in"`
}

type Round struct {
	Number       int           `json:"number"`
	Street       Street        `json:"street"`
	DealerSeat   int           `json:"dealer_seat"`
	SmallBlind   int           `json:"small_blind"`
	BigBlind     int           `json:"big_blind"`
	CurrentTurn  string        `json:"current_turn"`
	CurrentBet   int           `json:"current_bet"`
	MinRaise     int           `json:"min_raise"`
	PlayerStates []PlayerState `json:"player_states"`
	Pots         []Pot         `json:"pots"`
	Actions      []Action      `json:"actions"`
	IsComplete   bool          `json:"is_complete"`
}
