package blackjack

import (
	"math/rand"
	"testing"
	"github.com/kacheo/tmvgs/internal/testutil"
)

func TestGoldenRenderBlackjackInitialState(t *testing.T) {
	g := NewBlackjack()
	g.rng = rand.New(rand.NewSource(42))
	g.startRound()
	testutil.CheckGolden(t, "blackjack_initial", g.Render())
}
