package sudoku

import "testing"

func TestSolve(t *testing.T) {
	board := NewBoard()
	board.cells[0][0].value = 5
	board.cells[0][1].value = 3
	board.cells[0][4].value = 7
	board.cells[1][0].value = 6
	board.cells[1][3].value = 1
	board.cells[1][4].value = 9
	board.cells[1][5].value = 5
	board.cells[2][1].value = 9
	board.cells[2][2].value = 8
	board.cells[2][7].value = 6
	board.cells[3][0].value = 8
	board.cells[3][4].value = 6
	board.cells[3][8].value = 3
	board.cells[4][0].value = 4
	board.cells[4][3].value = 8
	board.cells[4][7].value = 1
	board.cells[5][0].value = 7
	board.cells[5][4].value = 3
	board.cells[6][1].value = 6
	board.cells[7][3].value = 4
	board.cells[7][4].value = 2
	board.cells[7][5].value = 9
	board.cells[7][8].value = 7
	board.cells[8][4].value = 8

	solved, err := Solve(&board)
	if err != nil {
		t.Errorf("solve error: %v", err)
	}
	if !solved {
		t.Error("expected board to be solved")
	}
	if !board.IsComplete() {
		t.Error("expected board to be complete after solving")
	}
}

func TestIsComplete(t *testing.T) {
	board := NewBoard()
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			board.cells[r][c].value = (r*9 + c) % 9 + 1
		}
	}
	if !board.IsComplete() {
		t.Error("expected complete board")
	}
}

func TestIsCompleteFalse(t *testing.T) {
	board := NewBoard()
	board.cells[0][0].value = 5
	if board.IsComplete() {
		t.Error("expected incomplete board")
	}
}
