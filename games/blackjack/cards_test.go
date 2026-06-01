package blackjack

import (
	"math/rand"
	"testing"
)

func TestHandValue(t *testing.T) {
	cases := []struct {
		hand Hand
		want int
	}{
		{Hand{{Seven, Spades}, {Eight, Hearts}}, 15},
		{Hand{{King, Spades}, {Queen, Hearts}}, 20},
		{Hand{{Ace, Spades}, {King, Hearts}}, 21},
		{Hand{{Ace, Spades}, {Five, Hearts}}, 16},
		{Hand{{Ace, Spades}, {Ace, Hearts}}, 12},
		{Hand{{Ace, Spades}, {Ace, Hearts}, {Nine, Clubs}}, 21},
		{Hand{{Ten, Spades}, {Six, Hearts}, {Eight, Clubs}}, 24},
	}
	for _, c := range cases {
		if got := c.hand.Value(); got != c.want {
			t.Errorf("hand %v: Value() = %d, want %d", c.hand, got, c.want)
		}
	}
}

func TestHandIsBust(t *testing.T) {
	if !(Hand{{Ten, Spades}, {Six, Hearts}, {Eight, Clubs}}.IsBust()) {
		t.Error("24 should be bust")
	}
	if (Hand{{Ten, Spades}, {King, Hearts}}.IsBust()) {
		t.Error("20 should not be bust")
	}
}

func TestHandIsBlackjack(t *testing.T) {
	if !(Hand{{Ace, Spades}, {King, Hearts}}.IsBlackjack()) {
		t.Error("A+K should be blackjack")
	}
	if (Hand{{Nine, Spades}, {Two, Hearts}, {Ten, Clubs}}.IsBlackjack()) {
		t.Error("3-card 21 is not blackjack")
	}
}

func TestHandIsSoft(t *testing.T) {
	if !(Hand{{Ace, Spades}, {Six, Hearts}}.IsSoft()) {
		t.Error("A+6 should be soft 17")
	}
	if (Hand{{Ten, Spades}, {Seven, Hearts}}.IsSoft()) {
		t.Error("T+7 should be hard")
	}
	if (Hand{{Ace, Spades}, {King, Hearts}, {Five, Clubs}}.IsSoft()) {
		t.Error("A+K+5=16 hard: ace forced to 1, not soft")
	}
}

func TestCardSymbolAndSuit(t *testing.T) {
	cases := []struct {
		card       Card
		sym, suit string
	}{
		{Card{Ace, Spades}, "A", "♠"},
		{Card{Ten, Hearts}, "T", "♥"},
		{Card{King, Diamonds}, "K", "♦"},
		{Card{Two, Clubs}, "2", "♣"},
		{Card{Jack, Spades}, "J", "♠"},
	}
	for _, c := range cases {
		if c.card.Symbol() != c.sym {
			t.Errorf("Symbol() = %q, want %q", c.card.Symbol(), c.sym)
		}
		if c.card.SuitSymbol() != c.suit {
			t.Errorf("SuitSymbol() = %q, want %q", c.card.SuitSymbol(), c.suit)
		}
	}
}

func TestCardIsRed(t *testing.T) {
	if !(Card{Ace, Hearts}.IsRed()) {
		t.Error("Hearts should be red")
	}
	if !(Card{Ace, Diamonds}.IsRed()) {
		t.Error("Diamonds should be red")
	}
	if (Card{Ace, Spades}.IsRed()) {
		t.Error("Spades should not be red")
	}
}

func TestDeckSizeAndDraw(t *testing.T) {
	d := NewDeck()
	if len(d) != 52 {
		t.Fatalf("deck size = %d, want 52", len(d))
	}
	_ = d.Draw()
	if len(d) != 51 {
		t.Errorf("after Draw size = %d, want 51", len(d))
	}
}

func TestDeckShuffled(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	d1 := NewDeck()
	d2 := d1.Shuffled(r)
	if len(d2) != 52 {
		t.Errorf("shuffled size = %d, want 52", len(d2))
	}
	same := true
	for i := range d1 {
		if d1[i] != d2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("shuffled deck should differ from ordered deck")
	}
}
