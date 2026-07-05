package blackjack

import (
	"math/rand"
	"testing"

	"github.com/kacheo/termcade/internal/testutil"
)

func TestGoldenRenderBlackjackInitialState(t *testing.T) {
	g := NewBlackjack(6)
	g.shoe = NewShoe(6, rand.New(rand.NewSource(42)))
	g.startRound()
	testutil.CheckGolden(t, "blackjack_initial", g.Render())
}

func TestGoldenRenderBlackjackBetting(t *testing.T) {
	g := NewBlackjack(6)
	g.shoe = NewShoe(6, rand.New(rand.NewSource(42)))
	testutil.CheckGolden(t, "blackjack_betting", g.Render())
}

func TestGoldenRenderBlackjackCountOverlay(t *testing.T) {
	g := NewBlackjack(6)
	g.shoe = NewShoe(6, rand.New(rand.NewSource(42)))
	g.startRound()
	g.showCount = true
	testutil.CheckGolden(t, "blackjack_count_overlay", g.Render())
}
