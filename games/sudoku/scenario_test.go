package sudoku

import (
	"testing"
)

// findEmptyCell returns the first non-given, empty cell on the board.
func findEmptyCell(s *Sudoku) (row, col int, found bool) {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if !s.board.cells[r][c].given && s.board.cells[r][c].value == 0 {
				return r, c, true
			}
		}
	}
	return 0, 0, false
}

// TestScenarioCursorMovement exercises full cursor movement including boundary clamping.
func TestScenarioCursorMovement(t *testing.T) {
	s := NewSudoku(DifficultyEasy, 0)

	// Initial position should be (0, 0).
	if s.cursorRow != 0 || s.cursorCol != 0 {
		t.Fatalf("expected initial cursor at (0,0), got (%d,%d)", s.cursorRow, s.cursorCol)
	}

	// Boundary: moving up/left from (0,0) should clamp.
	s.HandleInput("up")
	if s.cursorRow != 0 {
		t.Errorf("moving up from row 0 should clamp: got row %d", s.cursorRow)
	}
	s.HandleInput("left")
	if s.cursorCol != 0 {
		t.Errorf("moving left from col 0 should clamp: got col %d", s.cursorCol)
	}

	// Move right to col 1.
	s.HandleInput("right")
	if s.cursorCol != 1 {
		t.Errorf("expected col 1 after right, got %d", s.cursorCol)
	}

	// Move down to row 1.
	s.HandleInput("down")
	if s.cursorRow != 1 {
		t.Errorf("expected row 1 after down, got %d", s.cursorRow)
	}

	// Move to row/col 8 (boundary).
	for i := 0; i < 7; i++ {
		s.HandleInput("right")
		s.HandleInput("down")
	}
	if s.cursorRow != 8 {
		t.Errorf("expected row 8 after moving to bottom boundary, got %d", s.cursorRow)
	}
	if s.cursorCol != 8 {
		t.Errorf("expected col 8 after moving to right boundary, got %d", s.cursorCol)
	}

	// Boundary: moving right/down from (8,8) should clamp.
	s.HandleInput("right")
	if s.cursorCol != 8 {
		t.Errorf("moving right from col 8 should clamp: got col %d", s.cursorCol)
	}
	s.HandleInput("down")
	if s.cursorRow != 8 {
		t.Errorf("moving down from row 8 should clamp: got row %d", s.cursorRow)
	}

	// Move left and up to verify reverse directions.
	s.HandleInput("left")
	if s.cursorCol != 7 {
		t.Errorf("expected col 7 after left from 8, got %d", s.cursorCol)
	}
	s.HandleInput("up")
	if s.cursorRow != 7 {
		t.Errorf("expected row 7 after up from 8, got %d", s.cursorRow)
	}
}

// TestScenarioPlaceAndClearDigit places a digit in a non-given cell and then clears it.
func TestScenarioPlaceAndClearDigit(t *testing.T) {
	s := NewSudoku(DifficultyEasy, 0)

	row, col, found := findEmptyCell(s)
	if !found {
		t.Fatal("no empty non-given cell found on easy board")
	}

	// Navigate cursor to the target cell directly (white-box access).
	s.cursorRow = row
	s.cursorCol = col

	// Place digit "5".
	s.HandleInput("5")
	cell := &s.board.cells[row][col]
	if cell.value != 5 {
		t.Errorf("expected cell value 5 after placing digit, got %d", cell.value)
	}

	// Clear the cell with backspace.
	s.HandleInput("backspace")
	if cell.value != 0 {
		t.Errorf("expected cell value 0 after backspace, got %d", cell.value)
	}
}

// TestScenarioUndo places a digit then undoes it; verifies the undo stack bug fix from PR #12.
func TestScenarioUndo(t *testing.T) {
	s := NewSudoku(DifficultyEasy, 0)

	row, col, found := findEmptyCell(s)
	if !found {
		t.Fatal("no empty non-given cell found on easy board")
	}

	s.cursorRow = row
	s.cursorCol = col

	initialStackLen := len(s.undoStack)

	// Place a digit — undo stack should grow.
	s.HandleInput("3")
	if s.board.cells[row][col].value != 3 {
		t.Fatalf("expected value 3 before undo, got %d", s.board.cells[row][col].value)
	}
	if len(s.undoStack) != initialStackLen+1 {
		t.Errorf("expected undo stack length %d, got %d", initialStackLen+1, len(s.undoStack))
	}

	// Undo — digit should be gone, stack should shrink back.
	s.HandleInput("u")
	if s.board.cells[row][col].value != 0 {
		t.Errorf("expected value 0 after undo, got %d", s.board.cells[row][col].value)
	}
	if len(s.undoStack) != initialStackLen {
		t.Errorf("expected undo stack length %d after undo, got %d", initialStackLen, len(s.undoStack))
	}

	// Undoing on empty stack should not panic.
	s.HandleInput("u")
}

// TestScenarioPencilMode toggles pencil mode and verifies marks are set without changing cell value.
func TestScenarioPencilMode(t *testing.T) {
	s := NewSudoku(DifficultyEasy, 0)

	row, col, found := findEmptyCell(s)
	if !found {
		t.Fatal("no empty non-given cell found on easy board")
	}

	s.cursorRow = row
	s.cursorCol = col

	// Pencil mode should be off initially.
	if s.pencilMode {
		t.Fatal("pencil mode should be false initially")
	}

	// Toggle pencil mode on.
	s.HandleInput(" ")
	if !s.pencilMode {
		t.Fatal("pencil mode should be true after pressing space")
	}

	// Place digit "7" in pencil mode.
	s.HandleInput("7")

	cell := &s.board.cells[row][col]
	// Cell value must remain 0 in pencil mode.
	if cell.value != 0 {
		t.Errorf("cell value should be 0 in pencil mode, got %d", cell.value)
	}
	// Pencil mark at index 6 (digit 7 → index 6) should be set.
	if !cell.pencilMarks[6] {
		t.Error("pencil mark for digit 7 (index 6) should be set")
	}

	// Toggle pencil mode off.
	s.HandleInput(" ")
	if s.pencilMode {
		t.Error("pencil mode should be false after second space press")
	}
}

// TestScenarioPauseResume verifies that pressing "p" pauses the game.
// Note: HandleInput("p") only pauses (not unpauses); Resume() is called to unpause.
func TestScenarioPauseResume(t *testing.T) {
	s := NewSudoku(DifficultyEasy, 0)

	if s.isPaused {
		t.Fatal("game should not be paused initially")
	}

	// Pause via HandleInput.
	s.HandleInput("p")
	if !s.isPaused {
		t.Error("game should be paused after pressing p")
	}

	// A second "p" while already paused should be a no-op (stays paused).
	s.HandleInput("p")
	if !s.isPaused {
		t.Error("game should remain paused after pressing p a second time")
	}

	// Resume via the Resume() method.
	s.Resume()
	if s.isPaused {
		t.Error("game should not be paused after calling Resume()")
	}
}
