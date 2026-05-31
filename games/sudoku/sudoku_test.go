package sudoku

import (
	"testing"
)

func TestNewSudoku(t *testing.T) {
	game := NewSudoku(DifficultyEasy)
	if game.Name() != "Sudoku" {
		t.Errorf("expected name Sudoku, got %s", game.Name())
	}
	if game.GetScore() != 10000 {
		t.Errorf("expected initial score 10000, got %d", game.GetScore())
	}
}

func TestCursorMovement(t *testing.T) {
	game := NewSudoku(DifficultyEasy)
	game.HandleInput("right")
	if game.cursorCol != 1 {
		t.Error("cursor should move right")
	}
	game.HandleInput("down")
	if game.cursorRow != 1 {
		t.Error("cursor should move down")
	}
}

func TestPencilMode(t *testing.T) {
	game := NewSudoku(DifficultyEasy)
	if game.pencilMode {
		t.Error("pencil mode should be false initially")
	}
	game.HandleInput(" ")
	if !game.pencilMode {
		t.Error("pencil mode should be true after space")
	}
}