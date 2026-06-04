package snake

import (
	"math/rand"
	"testing"
	"tmvgs/internal/testutil"
)

func TestGoldenRenderSnakeInitialState(t *testing.T) {
	g := NewSnake()
	// Replace rng with a seeded source and respawn food so food position is deterministic.
	g.rng = rand.New(rand.NewSource(42))
	g.spawnFood()
	testutil.CheckGolden(t, "snake_initial", g.Render())
}
