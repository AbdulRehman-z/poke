package deck

import (
	"fmt"
	"math/rand"
)

type Suit int

func (s Suit) String() string {
	switch s {
	case Spades:
		return "SPADES"
	case Diamonds:
		return "DIAMONDS"
	case Clubs:
		return "CLUBS"
	case Hearts:
		return "HEARTS"
	default:
		return "Invalid Suit"
	}
}

const (
	Spades   = iota // 0
	Hearts          // 1
	Diamonds        // 2
	Clubs           // 3
)

type Card struct {
	suit Suit
	val  int
}

func (c Card) String() string {
	return fmt.Sprintf("%d of %s %s", c.val, c.suit, suitsToUnicode(c.suit))
}

// hello
type Deck [52]Card

func New() Deck {
	var (
		nSuits     = 4
		nCardsType = 13
		d          = [52]Card{}
	)
	for i := 0; i < nSuits; i++ {
		for j := 0; j < nCardsType; j++ {
			d[i*nCardsType+j] = Card{Suit(i), j + 1}
		}
	}
	return shuffle(d)
}

func shuffle(d Deck) Deck {
	for i := 0; i < len(d); i++ {
		r := rand.Intn(i + 1)
		if r != i {
			d[i], d[r] = d[r], d[i]
		}
	}
	return d
}

func suitsToUnicode(s Suit) string {
	switch s {
	case Spades:
		return "♠"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	case Hearts:
		return "♥"
	default:
		return "Invalid Suit"
	}
}
