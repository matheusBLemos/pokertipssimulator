package entity

type Pot struct {
	Amount      int      `bson:"amount" json:"amount"`
	EligibleIDs []string `bson:"eligible_ids" json:"eligible_ids"`
}
