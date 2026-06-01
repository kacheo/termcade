package sudoku

import "testing"

func TestCellDefault(t *testing.T) {
    cell := NewCell()
    if cell.value != 0 {
        t.Errorf("expected value 0, got %d", cell.value)
    }
    if cell.given {
        t.Error("expected given false")
    }
    if len(cell.pencilMarks) != 9 {
        t.Errorf("expected 9 pencil marks, got %d", len(cell.pencilMarks))
    }
}

func TestBoardInit(t *testing.T) {
    board := NewBoard()
    if len(board.cells) != 9 {
        t.Errorf("expected 9 rows, got %d", len(board.cells))
    }
    for r := 0; r < 9; r++ {
        for c := 0; c < 9; c++ {
            if board.cells[r][c].value != 0 {
                t.Errorf("expected empty cell at [%d][%d]", r, c)
            }
        }
    }
}

func TestGetCandidates(t *testing.T) {
    board := NewBoard()
    board.cells[0][0].value = 5

    candidates := board.GetCandidates(1, 0)
    if candidates[4] {
        t.Error("5 should not be a candidate at [1][0] (column elimination)")
    }

    candidates = board.GetCandidates(0, 1)
    if candidates[4] {
        t.Error("5 should not be a candidate at [0][1] (row elimination)")
    }

    candidates = board.GetCandidates(1, 1)
    if candidates[4] {
        t.Error("5 should not be a candidate at [1][1] (box elimination)")
    }
}

func TestHasConflict(t *testing.T) {
    board := NewBoard()
    board.cells[0][0].value = 5
    board.cells[0][1].value = 5
    if !board.HasConflict(0, 1) {
        t.Error("expected conflict at [0][1] due to duplicate in row")
    }
}

func TestClearCell(t *testing.T) {
    board := NewBoard()
    board.cells[0][0].value = 5
    board.cells[0][0].given = true
    board.ClearCell(0, 0)
    if board.cells[0][0].value != 5 {
        t.Error("given cell should not be cleared")
    }
}

func TestSetValue(t *testing.T) {
	board := NewBoard()
	board.SetValue(0, 0, 5, false)
	if board.cells[0][0].value != 5 {
		t.Error("expected value 5")
	}
}

func TestIsDigitComplete(t *testing.T) {
	// place digit 1 in all 9 rows across different columns (no row/col/box conflicts)
	fill9 := func() Board {
		b := NewBoard()
		// Each row gets a 1 in a column that doesn't repeat in any box row
		// Rows 0-2 use cols 0,3,6; rows 3-5 use cols 1,4,7; rows 6-8 use cols 2,5,8
		positions := [][2]int{{0, 0}, {1, 3}, {2, 6}, {3, 1}, {4, 4}, {5, 7}, {6, 2}, {7, 5}, {8, 8}}
		for _, p := range positions {
			b.cells[p[0]][p[1]].value = 1
		}
		return b
	}

	t.Run("all 9 placed no conflict", func(t *testing.T) {
		b := fill9()
		if !b.IsDigitComplete(1) {
			t.Error("expected IsDigitComplete(1) = true")
		}
	})

	t.Run("one instance has conflict flag", func(t *testing.T) {
		b := fill9()
		b.cells[0][0].conflict = true
		if b.IsDigitComplete(1) {
			t.Error("expected IsDigitComplete(1) = false when a cell has conflict")
		}
	})

	t.Run("fewer than 9 placed", func(t *testing.T) {
		b := NewBoard()
		b.cells[0][0].value = 1
		if b.IsDigitComplete(1) {
			t.Error("expected IsDigitComplete(1) = false with only 1 instance placed")
		}
	})

	t.Run("digit 0 never complete", func(t *testing.T) {
		b := NewBoard()
		if b.IsDigitComplete(0) {
			t.Error("expected IsDigitComplete(0) = false (empty cells don't count)")
		}
	})
}