package entity

type PlayerStatus string

const (
	PlayerStatusWaiting    PlayerStatus = "waiting"
	PlayerStatusActive     PlayerStatus = "active"
	PlayerStatusEliminated PlayerStatus = "eliminated"
)

type Player struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Seat   int          `json:"seat"`
	Stack  int          `json:"stack"`
	Status PlayerStatus `json:"status"`
}
