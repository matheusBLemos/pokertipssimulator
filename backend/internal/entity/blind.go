package entity

type BlindLevel struct {
	SmallBlind int `bson:"small_blind" json:"small_blind"`
	BigBlind   int `bson:"big_blind" json:"big_blind"`
	Ante       int `bson:"ante" json:"ante"`
	Duration   int `bson:"duration" json:"duration"` // seconds, 0 = manual
}

type BlindStructure struct {
	Levels       []BlindLevel `bson:"levels" json:"levels"`
	CurrentLevel int          `bson:"current_level" json:"current_level"`
}
