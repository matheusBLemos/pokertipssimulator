package entity

import (
	"fmt"
	"math/rand"
)

type Suit string

const (
	SuitSpades   Suit = "s"
	SuitHearts   Suit = "h"
	SuitDiamonds Suit = "d"
	SuitClubs    Suit = "c"
)

type Rank string

const (
	RankTwo   Rank = "2"
	RankThree Rank = "3"
	RankFour  Rank = "4"
	RankFive  Rank = "5"
	RankSix   Rank = "6"
	RankSeven Rank = "7"
	RankEight Rank = "8"
	RankNine  Rank = "9"
	RankTen   Rank = "T"
	RankJack  Rank = "J"
	RankQueen Rank = "Q"
	RankKing  Rank = "K"
	RankAce   Rank = "A"
)

var allRanks = []Rank{
	RankTwo, RankThree, RankFour, RankFive, RankSix, RankSeven,
	RankEight, RankNine, RankTen, RankJack, RankQueen, RankKing, RankAce,
}

var allSuits = []Suit{SuitSpades, SuitHearts, SuitDiamonds, SuitClubs}

var rankValue = map[Rank]int{
	RankTwo: 2, RankThree: 3, RankFour: 4, RankFive: 5, RankSix: 6,
	RankSeven: 7, RankEight: 8, RankNine: 9, RankTen: 10,
	RankJack: 11, RankQueen: 12, RankKing: 13, RankAce: 14,
}

type Card struct {
	Rank Rank `json:"rank"`
	Suit Suit `json:"suit"`
}

func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Rank, c.Suit)
}

func (c Card) Value() int {
	return rankValue[c.Rank]
}

type Deck struct {
	Cards []Card `json:"cards"`
}

func NewDeck() *Deck {
	cards := make([]Card, 0, 52)
	for _, suit := range allSuits {
		for _, rank := range allRanks {
			cards = append(cards, Card{Rank: rank, Suit: suit})
		}
	}
	return &Deck{Cards: cards}
}

func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

func (d *Deck) Deal(n int) []Card {
	if n > len(d.Cards) {
		n = len(d.Cards)
	}
	dealt := make([]Card, n)
	copy(dealt, d.Cards[:n])
	d.Cards = d.Cards[n:]
	return dealt
}

func (d *Deck) DealOne() Card {
	cards := d.Deal(1)
	return cards[0]
}

func (d *Deck) Remaining() int {
	return len(d.Cards)
}
