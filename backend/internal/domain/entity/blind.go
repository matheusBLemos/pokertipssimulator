package entity

type BlindLevel struct {
	SmallBlind int `json:"small_blind"`
	BigBlind   int `json:"big_blind"`
	Ante       int `json:"ante"`
	Duration   int `json:"duration"`
}

type BlindStructure struct {
	Levels       []BlindLevel `json:"levels"`
	CurrentLevel int          `json:"current_level"`
}
