package blackjack

import (
	"math/rand"
	"testing"

	cardpkg "github.com/kacheo/termcade/games/cards"
)

func TestHiLoValue(t *testing.T) {
	cases := []struct {
		rank cardpkg.Rank
		want int
	}{
		{cardpkg.Ace, -1},
		{cardpkg.Two, 1},
		{cardpkg.Three, 1},
		{cardpkg.Four, 1},
		{cardpkg.Five, 1},
		{cardpkg.Six, 1},
		{cardpkg.Seven, 0},
		{cardpkg.Eight, 0},
		{cardpkg.Nine, 0},
		{cardpkg.Ten, -1},
		{cardpkg.Jack, -1},
		{cardpkg.Queen, -1},
		{cardpkg.King, -1},
	}
	for _, c := range cases {
		if got := hiLoValue(c.rank); got != c.want {
			t.Errorf("hiLoValue(%v) = %d, want %d", c.rank, got, c.want)
		}
	}
}

func TestNewShoe_ClampsDeckCount(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	if s := NewShoe(0, r); s.numDecks != 1 {
		t.Errorf("numDecks = %d, want 1 (clamped up)", s.numDecks)
	}
	if s := NewShoe(20, r); s.numDecks != 8 {
		t.Errorf("numDecks = %d, want 8 (clamped down)", s.numDecks)
	}
	if s := NewShoe(6, r); s.CardsRemaining() != 6*52 {
		t.Errorf("CardsRemaining() = %d, want %d", s.CardsRemaining(), 6*52)
	}
}

func TestShoe_TrueCountArithmetic(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	s := NewShoe(2, r) // 104 cards

	// Draw down to 52 cards remaining (1 deck) with a known running count.
	s.cards = s.cards[:52]
	s.runningCount = 4

	if got := s.DecksRemaining(); got != 1.0 {
		t.Errorf("DecksRemaining() = %v, want 1.0", got)
	}
	if got := s.TrueCount(); got != 4.0 {
		t.Errorf("TrueCount() = %v, want 4.0", got)
	}
}

func TestShoe_DecksRemainingFloor(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	s := NewShoe(1, r)
	s.cards = s.cards[:2] // nearly empty
	if got := s.DecksRemaining(); got != 0.25 {
		t.Errorf("DecksRemaining() = %v, want floor of 0.25", got)
	}
}

func TestShoe_PenetrationBoundary(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	s := NewShoe(1, r) // 52 cards, threshold = 52*0.25 = 13

	s.cards = s.cards[:14]
	s.checkPenetration()
	if s.NeedsReshuffle() {
		t.Error("reshufflePending should be false with 14 cards remaining (above 25% threshold)")
	}

	s.cards = s.cards[:13]
	s.checkPenetration()
	if !s.NeedsReshuffle() {
		t.Error("reshufflePending should be true with 13 cards remaining (at/below 25% threshold)")
	}
}

func TestShoe_Reshuffle_ResetsState(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	s := NewShoe(1, r)
	s.cards = s.cards[:5]
	s.runningCount = 7
	s.reshufflePending = true

	s.Reshuffle()

	if s.CardsRemaining() != 52 {
		t.Errorf("CardsRemaining() = %d, want 52 after reshuffle", s.CardsRemaining())
	}
	if s.RunningCount() != 0 {
		t.Errorf("RunningCount() = %d, want 0 after reshuffle", s.RunningCount())
	}
	if s.NeedsReshuffle() {
		t.Error("reshufflePending should be false immediately after reshuffle")
	}
}

func TestShoe_Draw_EmergencyReshuffleOnEmpty(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	s := NewShoe(1, r)
	s.cards = cardpkg.Deck{}

	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("Draw() panicked on empty shoe: %v", rec)
		}
	}()
	c := s.Draw()
	_ = c
	if s.CardsRemaining() != 51 {
		t.Errorf("CardsRemaining() = %d, want 51 after emergency reshuffle + draw", s.CardsRemaining())
	}
}

func TestShoe_CountCard_UpdatesRunningCount(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	s := NewShoe(1, r)
	s.CountCard(cardpkg.Card{Rank: cardpkg.Five, Suit: cardpkg.Spades})
	s.CountCard(cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Hearts})
	s.CountCard(cardpkg.Card{Rank: cardpkg.Eight, Suit: cardpkg.Clubs})
	if s.RunningCount() != 0 {
		t.Errorf("RunningCount() = %d, want 0 (+1 -1 +0)", s.RunningCount())
	}
}
