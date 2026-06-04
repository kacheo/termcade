package poker

import (
	"math/rand"
	"testing"
	"tmvgs/internal/testutil"
)

func TestGoldenRenderPokerInitialState(t *testing.T) {
	g := NewPoker(3, Easy)
	// Reset state for a fully deterministic render:
	// - seed the rng so deck shuffle is reproducible
	// - reset chips (NewPoker already posted blinds once)
	// - fix dealer position so startHand advances to a known position
	g.rng = rand.New(rand.NewSource(42))
	for i := range g.players {
		g.players[i].chips = 1000
	}
	g.pot = 0
	g.dealer = 0 // startHand will advance dealer to 1 (AI-1)
	g.startHand()
	testutil.CheckGolden(t, "poker_initial", g.Render())
}
