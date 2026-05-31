package blackjack

import "testing"

func TestNewBlackjack_Metadata(t *testing.T) {
	g := NewBlackjack(2)
	if g.Name() != "Blackjack" {
		t.Errorf("Name() = %q, want Blackjack", g.Name())
	}
	if g.IsGameOver() {
		t.Error("should not be game over at start")
	}
	if g.IsPaused() {
		t.Error("should not be paused at start")
	}
	if g.GetScore() != 0 {
		t.Errorf("GetScore() = %d, want 0", g.GetScore())
	}
	if g.GetLevel() != 0 {
		t.Errorf("GetLevel() = %d, want 0", g.GetLevel())
	}
	if g.GetLines() != 1 {
		t.Errorf("GetLines() = %d, want 1 (first round)", g.GetLines())
	}
}

func TestNewBlackjack_Players(t *testing.T) {
	cases := []struct{ ai, total int }{{0, 1}, {1, 2}, {3, 4}, {4, 4}}
	for _, c := range cases {
		g := NewBlackjack(c.ai)
		if len(g.players) != c.total {
			t.Errorf("aiCount=%d: players=%d, want %d", c.ai, len(g.players), c.total)
		}
	}
	g := NewBlackjack(3)
	if g.players[0].isAI {
		t.Error("players[0] should be human")
	}
	for i := 1; i <= 3; i++ {
		if !g.players[i].isAI {
			t.Errorf("players[%d] should be AI", i)
		}
	}
}

func TestInitialDeal(t *testing.T) {
	g := NewBlackjack(2)
	if len(g.dealer) != 2 {
		t.Errorf("dealer cards = %d, want 2", len(g.dealer))
	}
	for i, p := range g.players {
		if len(p.hand) != 2 {
			t.Errorf("player[%d] cards = %d, want 2", i, len(p.hand))
		}
	}
	if g.phase != phaseDealing {
		t.Errorf("initial phase = %v, want phaseDealing", g.phase)
	}
}
