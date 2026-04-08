package entity

type Pot struct {
	Amount      int      `json:"amount"`
	EligibleIDs []string `json:"eligible_ids"`
}
