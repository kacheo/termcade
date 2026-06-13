package sudoku

import (
	"math/rand"
	"testing"
	"github.com/kacheo/tmvgs/internal/testutil"
)

// generateSeeded creates a deterministic board using a fixed seed.
func generateSeeded(diff Difficulty, seed int64) *Board {
	board := NewBoard()
	r := rand.New(rand.NewSource(seed))
	fillBoard(&board, r)
	removeClues(&board, difficultyClues[diff], r)
	return &board
}

func TestGoldenRenderSudokuInitialState(t *testing.T) {
	g := NewSudoku(DifficultyEasy, 0)
	g.board = generateSeeded(DifficultyEasy, 42)
	testutil.CheckGolden(t, "sudoku_initial", g.Render())
}
