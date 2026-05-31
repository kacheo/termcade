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