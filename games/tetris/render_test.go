package tetris

import (
	"testing"
	"github.com/kacheo/termcade/internal/testutil"
)

func TestGoldenRenderTetrisInitialState(t *testing.T) {
	g := NewTetris(true, 1)
	// Override the piece bag with a fixed sequence so the rendered state is deterministic.
	// Piece bag order: I, O, T, S, Z, J, L
	g.rng = []byte{'I', 'O', 'T', 'S', 'Z', 'J', 'L'}
	g.rngIndex = 0
	g.current = nil
	g.next = nil
	g.spawnPiece()
	testutil.CheckGolden(t, "tetris_initial", g.Render())
}
