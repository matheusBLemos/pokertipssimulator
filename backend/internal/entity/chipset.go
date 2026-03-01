package entity

type ChipDenomination struct {
	Value int    `bson:"value" json:"value"`
	Color string `bson:"color" json:"color"`
}

type ChipSet struct {
	Denominations []ChipDenomination `bson:"denominations" json:"denominations"`
}

func DefaultChipSet() ChipSet {
	return ChipSet{
		Denominations: []ChipDenomination{
			{Value: 1, Color: "#FFFFFF"},
			{Value: 5, Color: "#FF0000"},
			{Value: 10, Color: "#0000FF"},
			{Value: 25, Color: "#00FF00"},
			{Value: 100, Color: "#000000"},
			{Value: 500, Color: "#800080"},
		},
	}
}
