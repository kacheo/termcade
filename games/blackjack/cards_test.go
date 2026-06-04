package blackjack

import (
	"testing"

	"tmvgs/games/cards"
)

func TestHandValue(t *testing.T) {
	cases := []struct {
		hand Hand
		want int
	}{
		{Hand{cards.Card{Rank: cards.Seven, Suit: cards.Spades}, cards.Card{Rank: cards.Eight, Suit: cards.Hearts}}, 15},
		{Hand{cards.Card{Rank: cards.King, Suit: cards.Spades}, cards.Card{Rank: cards.Queen, Suit: cards.Hearts}}, 20},
		{Hand{cards.Card{Rank: cards.Ace, Suit: cards.Spades}, cards.Card{Rank: cards.King, Suit: cards.Hearts}}, 21},
		{Hand{cards.Card{Rank: cards.Ace, Suit: cards.Spades}, cards.Card{Rank: cards.Five, Suit: cards.Hearts}}, 16},
		{Hand{cards.Card{Rank: cards.Ace, Suit: cards.Spades}, cards.Card{Rank: cards.Ace, Suit: cards.Hearts}}, 12},
		{Hand{cards.Card{Rank: cards.Ace, Suit: cards.Spades}, cards.Card{Rank: cards.Ace, Suit: cards.Hearts}, cards.Card{Rank: cards.Nine, Suit: cards.Clubs}}, 21},
		{Hand{cards.Card{Rank: cards.Ten, Suit: cards.Spades}, cards.Card{Rank: cards.Six, Suit: cards.Hearts}, cards.Card{Rank: cards.Eight, Suit: cards.Clubs}}, 24},
	}
	for _, c := range cases {
		if got := c.hand.Value(); got != c.want {
			t.Errorf("hand %v: Value() = %d, want %d", c.hand, got, c.want)
		}
	}
}

func TestHandIsBust(t *testing.T) {
	if !(Hand{cards.Card{Rank: cards.Ten, Suit: cards.Spades}, cards.Card{Rank: cards.Six, Suit: cards.Hearts}, cards.Card{Rank: cards.Eight, Suit: cards.Clubs}}.IsBust()) {
		t.Error("24 should be bust")
	}
	if (Hand{cards.Card{Rank: cards.Ten, Suit: cards.Spades}, cards.Card{Rank: cards.King, Suit: cards.Hearts}}.IsBust()) {
		t.Error("20 should not be bust")
	}
}

func TestHandIsBlackjack(t *testing.T) {
	if !(Hand{cards.Card{Rank: cards.Ace, Suit: cards.Spades}, cards.Card{Rank: cards.King, Suit: cards.Hearts}}.IsBlackjack()) {
		t.Error("A+K should be blackjack")
	}
	if (Hand{cards.Card{Rank: cards.Nine, Suit: cards.Spades}, cards.Card{Rank: cards.Two, Suit: cards.Hearts}, cards.Card{Rank: cards.Ten, Suit: cards.Clubs}}.IsBlackjack()) {
		t.Error("3-card 21 is not blackjack")
	}
}

func TestHandIsSoft(t *testing.T) {
	if !(Hand{cards.Card{Rank: cards.Ace, Suit: cards.Spades}, cards.Card{Rank: cards.Six, Suit: cards.Hearts}}.IsSoft()) {
		t.Error("A+6 should be soft 17")
	}
	if (Hand{cards.Card{Rank: cards.Ten, Suit: cards.Spades}, cards.Card{Rank: cards.Seven, Suit: cards.Hearts}}.IsSoft()) {
		t.Error("T+7 should be hard")
	}
	if (Hand{cards.Card{Rank: cards.Ace, Suit: cards.Spades}, cards.Card{Rank: cards.King, Suit: cards.Hearts}, cards.Card{Rank: cards.Five, Suit: cards.Clubs}}.IsSoft()) {
		t.Error("A+K+5=16 hard: ace forced to 1, not soft")
	}
}
