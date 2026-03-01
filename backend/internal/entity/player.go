package entity

type PlayerStatus string

const (
	PlayerStatusWaiting      PlayerStatus = "waiting"
	PlayerStatusActive       PlayerStatus = "active"
	PlayerStatusSittingOut   PlayerStatus = "sitting_out"
	PlayerStatusEliminated   PlayerStatus = "eliminated"
	PlayerStatusDisconnected PlayerStatus = "disconnected"
)

type Player struct {
	ID     string       `bson:"id" json:"id"`
	Name   string       `bson:"name" json:"name"`
	Seat   int          `bson:"seat" json:"seat"`
	Stack  int          `bson:"stack" json:"stack"`
	Status PlayerStatus `bson:"status" json:"status"`
}
