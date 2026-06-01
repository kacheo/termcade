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
		{Hand{{cards.Seven, cards.Spades}, {cards.Eight, cards.Hearts}}, 15},
		{Hand{{cards.King, cards.Spades}, {cards.Queen, cards.Hearts}}, 20},
		{Hand{{cards.Ace, cards.Spades}, {cards.King, cards.Hearts}}, 21},
		{Hand{{cards.Ace, cards.Spades}, {cards.Five, cards.Hearts}}, 16},
		{Hand{{cards.Ace, cards.Spades}, {cards.Ace, cards.Hearts}}, 12},
		{Hand{{cards.Ace, cards.Spades}, {cards.Ace, cards.Hearts}, {cards.Nine, cards.Clubs}}, 21},
		{Hand{{cards.Ten, cards.Spades}, {cards.Six, cards.Hearts}, {cards.Eight, cards.Clubs}}, 24},
	}
	for _, c := range cases {
		if got := c.hand.Value(); got != c.want {
			t.Errorf("hand %v: Value() = %d, want %d", c.hand, got, c.want)
		}
	}
}

func TestHandIsBust(t *testing.T) {
	if !(Hand{{cards.Ten, cards.Spades}, {cards.Six, cards.Hearts}, {cards.Eight, cards.Clubs}}.IsBust()) {
		t.Error("24 should be bust")
	}
	if (Hand{{cards.Ten, cards.Spades}, {cards.King, cards.Hearts}}.IsBust()) {
		t.Error("20 should not be bust")
	}
}

func TestHandIsBlackjack(t *testing.T) {
	if !(Hand{{cards.Ace, cards.Spades}, {cards.King, cards.Hearts}}.IsBlackjack()) {
		t.Error("A+K should be blackjack")
	}
	if (Hand{{cards.Nine, cards.Spades}, {cards.Two, cards.Hearts}, {cards.Ten, cards.Clubs}}.IsBlackjack()) {
		t.Error("3-card 21 is not blackjack")
	}
}

func TestHandIsSoft(t *testing.T) {
	if !(Hand{{cards.Ace, cards.Spades}, {cards.Six, cards.Hearts}}.IsSoft()) {
		t.Error("A+6 should be soft 17")
	}
	if (Hand{{cards.Ten, cards.Spades}, {cards.Seven, cards.Hearts}}.IsSoft()) {
		t.Error("T+7 should be hard")
	}
	if (Hand{{cards.Ace, cards.Spades}, {cards.King, cards.Hearts}, {cards.Five, cards.Clubs}}.IsSoft()) {
		t.Error("A+K+5=16 hard: ace forced to 1, not soft")
	}
}
