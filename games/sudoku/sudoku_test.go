package sudoku

import (
	"strings"
	"testing"
	"time"
)

func TestNewSudoku(t *testing.T) {
	game := NewSudoku(DifficultyEasy, 4)
	if game.Name() != "Sudoku" {
		t.Errorf("expected name Sudoku, got %s", game.Name())
	}
	if game.GetScore() != 10000 {
		t.Errorf("expected initial score 10000, got %d", game.GetScore())
	}
}

func TestCursorMovement(t *testing.T) {
	game := NewSudoku(DifficultyEasy, 4)
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
	game := NewSudoku(DifficultyEasy, 4)
	if game.pencilMode {
		t.Error("pencil mode should be false initially")
	}
	game.HandleInput(" ")
	if !game.pencilMode {
		t.Error("pencil mode should be true after space")
	}
}

func TestRender(t *testing.T) {
	game := NewSudoku(DifficultyEasy, 4)
	output := game.Render()
	if len(output) == 0 {
		t.Error("render should return non-empty string")
	}
	if !strings.Contains(output, "SUDOKU") {
		t.Error("render should contain SUDOKU header")
	}
}

func TestSetDigitUndoFirstEntry(t *testing.T) {
	game := NewSudoku(DifficultyEasy, 4)
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if !game.board.cells[r][c].given && game.board.cells[r][c].value == 0 {
				game.cursorRow = r
				game.cursorCol = c
				goto found
			}
		}
	}
found:
	if game.board.cells[game.cursorRow][game.cursorCol].value != 0 {
		t.Fatal("cursor should be on empty non-given cell")
	}
	initialLen := len(game.undoStack)
	game.HandleInput("5")
	if len(game.undoStack) != initialLen+1 {
		t.Errorf("undo stack should grow when entering first digit, got stack len %d want %d", len(game.undoStack), initialLen+1)
	}
	if game.board.cells[game.cursorRow][game.cursorCol].value != 5 {
		t.Error("cell should have value 5")
	}
	game.HandleInput("u")
	if game.board.cells[game.cursorRow][game.cursorCol].value != 0 {
		t.Error("undo should restore empty cell after first digit entry")
	}
}

func TestSetDigitNoUndoSameValue(t *testing.T) {
	game := NewSudoku(DifficultyEasy, 4)
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if !game.board.cells[r][c].given && game.board.cells[r][c].value == 0 {
				game.cursorRow = r
				game.cursorCol = c
				goto found
			}
		}
	}
found:
	game.HandleInput("5")
	stackAfterFirst := len(game.undoStack)
	game.HandleInput("5")
	if len(game.undoStack) != stackAfterFirst {
		t.Error("re-entering same value should not push undo")
	}
}

func TestUndoStackLimit(t *testing.T) {
	game := NewSudoku(DifficultyEasy, 4)
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			game.board.cells[r][c].given = false
		}
	}
	game.cursorRow = 0
	game.cursorCol = 0
	for i := 0; i < MAX_UNDO+10; i++ {
		game.HandleInput("1")
		game.cursorCol = (game.cursorCol + 1) % 9
	}
	if len(game.undoStack) > MAX_UNDO {
		t.Errorf("undo stack should be limited to %d, got %d", MAX_UNDO, len(game.undoStack))
	}
}

func TestTimerPaused(t *testing.T) {
	game := NewSudoku(DifficultyEasy, 4)
	game.startTime = time.Now().Add(-10 * time.Second)
	if err := game.Update(0); err != nil {
		t.Fatal(err)
	}
	elapsedBeforePause := game.elapsed
	game.isPaused = true
	game.pausedAt = time.Now()
	time.Sleep(10 * time.Millisecond)
	if err := game.Update(0); err != nil {
		t.Fatal(err)
	}
	if game.elapsed != elapsedBeforePause {
		t.Error("elapsed should not increase while paused")
	}
}

func TestPencilMarksRender(t *testing.T) {
	game := NewSudoku(DifficultyEasy, 4)
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if !game.board.cells[r][c].given && game.board.cells[r][c].value == 0 {
				game.cursorRow = r
				game.cursorCol = c
				goto found
			}
		}
	}
found:
	game.pencilMode = true
	game.HandleInput("1")
	game.HandleInput("3")
	game.HandleInput("5")
	output := game.Render()
	if !strings.Contains(output, "135") && !strings.Contains(output, "1") {
		t.Error("pencil marks should be visible in render output")
	}
}