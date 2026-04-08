package entity

import "sort"

type HandRank int

const (
	HandHighCard      HandRank = 0
	HandOnePair       HandRank = 1
	HandTwoPair       HandRank = 2
	HandThreeOfAKind  HandRank = 3
	HandStraight      HandRank = 4
	HandFlush         HandRank = 5
	HandFullHouse     HandRank = 6
	HandFourOfAKind   HandRank = 7
	HandStraightFlush HandRank = 8
	HandRoyalFlush    HandRank = 9
)

var handRankNames = map[HandRank]string{
	HandHighCard:      "High Card",
	HandOnePair:       "One Pair",
	HandTwoPair:       "Two Pair",
	HandThreeOfAKind:  "Three of a Kind",
	HandStraight:      "Straight",
	HandFlush:         "Flush",
	HandFullHouse:     "Full House",
	HandFourOfAKind:   "Four of a Kind",
	HandStraightFlush: "Straight Flush",
	HandRoyalFlush:    "Royal Flush",
}

type EvaluatedHand struct {
	Rank     HandRank `json:"rank"`
	RankName string   `json:"rank_name"`
	Kickers  []int    `json:"-"`
	Cards    []Card   `json:"cards"`
}

func (h EvaluatedHand) Beats(other EvaluatedHand) int {
	if h.Rank != other.Rank {
		if h.Rank > other.Rank {
			return 1
		}
		return -1
	}
	for i := 0; i < len(h.Kickers) && i < len(other.Kickers); i++ {
		if h.Kickers[i] != other.Kickers[i] {
			if h.Kickers[i] > other.Kickers[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}

// EvaluateBestHand finds the best 5-card poker hand from any number of cards (typically 7).
func EvaluateBestHand(cards []Card) EvaluatedHand {
	if len(cards) < 5 {
		return evaluate5(cards)
	}

	combos := combinations(cards, 5)
	var best EvaluatedHand
	first := true
	for _, combo := range combos {
		hand := evaluate5(combo)
		if first || hand.Beats(best) > 0 {
			best = hand
			first = false
		}
	}
	return best
}

func evaluate5(cards []Card) EvaluatedHand {
	sorted := make([]Card, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value() > sorted[j].Value()
	})

	isFlush := checkFlush(sorted)
	straight, highCard := checkStraight(sorted)

	values := make([]int, len(sorted))
	for i, c := range sorted {
		values[i] = c.Value()
	}

	groups := groupByRank(values)

	if isFlush && straight {
		if highCard == 14 {
			return EvaluatedHand{Rank: HandRoyalFlush, RankName: handRankNames[HandRoyalFlush], Kickers: []int{highCard}, Cards: sorted}
		}
		return EvaluatedHand{Rank: HandStraightFlush, RankName: handRankNames[HandStraightFlush], Kickers: []int{highCard}, Cards: sorted}
	}

	if groups[0].count == 4 {
		kicker := groups[1].value
		return EvaluatedHand{Rank: HandFourOfAKind, RankName: handRankNames[HandFourOfAKind], Kickers: []int{groups[0].value, kicker}, Cards: sorted}
	}

	if groups[0].count == 3 && len(groups) >= 2 && groups[1].count == 2 {
		return EvaluatedHand{Rank: HandFullHouse, RankName: handRankNames[HandFullHouse], Kickers: []int{groups[0].value, groups[1].value}, Cards: sorted}
	}

	if isFlush {
		return EvaluatedHand{Rank: HandFlush, RankName: handRankNames[HandFlush], Kickers: values, Cards: sorted}
	}

	if straight {
		return EvaluatedHand{Rank: HandStraight, RankName: handRankNames[HandStraight], Kickers: []int{highCard}, Cards: sorted}
	}

	if groups[0].count == 3 {
		kickers := []int{groups[0].value}
		for _, g := range groups[1:] {
			kickers = append(kickers, g.value)
		}
		return EvaluatedHand{Rank: HandThreeOfAKind, RankName: handRankNames[HandThreeOfAKind], Kickers: kickers, Cards: sorted}
	}

	if groups[0].count == 2 && len(groups) >= 2 && groups[1].count == 2 {
		kickers := []int{groups[0].value, groups[1].value}
		for _, g := range groups[2:] {
			kickers = append(kickers, g.value)
		}
		return EvaluatedHand{Rank: HandTwoPair, RankName: handRankNames[HandTwoPair], Kickers: kickers, Cards: sorted}
	}

	if groups[0].count == 2 {
		kickers := []int{groups[0].value}
		for _, g := range groups[1:] {
			kickers = append(kickers, g.value)
		}
		return EvaluatedHand{Rank: HandOnePair, RankName: handRankNames[HandOnePair], Kickers: kickers, Cards: sorted}
	}

	return EvaluatedHand{Rank: HandHighCard, RankName: handRankNames[HandHighCard], Kickers: values, Cards: sorted}
}

func checkFlush(cards []Card) bool {
	if len(cards) < 5 {
		return false
	}
	suit := cards[0].Suit
	for _, c := range cards[1:] {
		if c.Suit != suit {
			return false
		}
	}
	return true
}

func checkStraight(cards []Card) (bool, int) {
	if len(cards) < 5 {
		return false, 0
	}

	values := make([]int, len(cards))
	for i, c := range cards {
		values[i] = c.Value()
	}

	// Check A-2-3-4-5 (wheel)
	if values[0] == 14 && values[1] == 5 && values[2] == 4 && values[3] == 3 && values[4] == 2 {
		return true, 5
	}

	for i := 0; i < len(values)-1; i++ {
		if values[i]-values[i+1] != 1 {
			return false, 0
		}
	}
	return true, values[0]
}

type rankGroup struct {
	value int
	count int
}

func groupByRank(values []int) []rankGroup {
	freq := make(map[int]int)
	for _, v := range values {
		freq[v]++
	}

	groups := make([]rankGroup, 0, len(freq))
	for v, c := range freq {
		groups = append(groups, rankGroup{value: v, count: c})
	}

	sort.Slice(groups, func(i, j int) bool {
		if groups[i].count != groups[j].count {
			return groups[i].count > groups[j].count
		}
		return groups[i].value > groups[j].value
	})

	return groups
}

func combinations(cards []Card, k int) [][]Card {
	var result [][]Card
	var combo []Card
	var generate func(start int)
	generate = func(start int) {
		if len(combo) == k {
			tmp := make([]Card, k)
			copy(tmp, combo)
			result = append(result, tmp)
			return
		}
		for i := start; i < len(cards); i++ {
			combo = append(combo, cards[i])
			generate(i + 1)
			combo = combo[:len(combo)-1]
		}
	}
	generate(0)
	return result
}
