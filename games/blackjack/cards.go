package blackjack

import (
	"fmt"
	"math/rand"
)

type Suit int

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

type Rank int

const (
	Ace Rank = iota + 1
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

type Card struct {
	Rank Rank
	Suit Suit
}

func (c Card) BaseValue() int {
	if c.Rank >= Ten {
		return 10
	}
	return int(c.Rank)
}

func (c Card) Symbol() string {
	switch c.Rank {
	case Ace:
		return "A"
	case Ten:
		return "T"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	default:
		return fmt.Sprintf("%d", int(c.Rank))
	}
}

func (c Card) SuitSymbol() string {
	switch c.Suit {
	case Spades:
		return "♠"
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	}
	return "?"
}

func (c Card) IsRed() bool {
	return c.Suit == Hearts || c.Suit == Diamonds
}

type Hand []Card

func (h Hand) Value() int {
	total, aces := 0, 0
	for _, c := range h {
		if c.Rank == Ace {
			aces++
			total += 11
		} else {
			total += c.BaseValue()
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

func (h Hand) IsBust() bool { return h.Value() > 21 }

func (h Hand) IsBlackjack() bool { return len(h) == 2 && h.Value() == 21 }

func (h Hand) IsSoft() bool {
	total, aces := 0, 0
	for _, c := range h {
		if c.Rank == Ace {
			aces++
			total += 11
		} else {
			total += c.BaseValue()
		}
	}
	reduced := 0
	for total > 21 && reduced < aces {
		total -= 10
		reduced++
	}
	return aces > reduced && total <= 21
}

type Deck []Card

func NewDeck() Deck {
	var d Deck
	for suit := Spades; suit <= Clubs; suit++ {
		for rank := Ace; rank <= King; rank++ {
			d = append(d, Card{Rank: rank, Suit: suit})
		}
	}
	return d
}

func (d Deck) Shuffled(r *rand.Rand) Deck {
	out := make(Deck, len(d))
	copy(out, d)
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// Draw removes and returns the top card. Panics if the deck is empty.
func (d *Deck) Draw() Card {
	if len(*d) == 0 {
		panic("blackjack: draw from empty deck")
	}
	c := (*d)[0]
	*d = (*d)[1:]
	return c
}
