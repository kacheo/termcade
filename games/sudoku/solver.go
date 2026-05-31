package sudoku

import "fmt"

func Solve(board *Board) (bool, error) {
	return solveRecursive(board, 0, 0)
}

func solveRecursive(board *Board, row, col int) (bool, error) {
	if row == 9 {
		return true, nil
	}
	if col == 9 {
		return solveRecursive(board, row+1, 0)
	}
	if board.cells[row][col].value != 0 {
		return solveRecursive(board, row, col+1)
	}
	candidates := board.GetCandidates(row, col)
	placed := false
	for i := 0; i < 9; i++ {
		if candidates[i] {
			placed = true
			board.cells[row][col].value = i + 1
			if board.HasConflict(row, col) {
				board.cells[row][col].value = 0
				continue
			}
			if solved, _ := solveRecursive(board, row, col+1); solved {
				return true, nil
			}
			board.cells[row][col].value = 0
		}
	}
	if !placed {
		return false, fmt.Errorf("no candidates available for cell (%d,%d)", row, col)
	}
	return false, fmt.Errorf("no solution found for cell (%d,%d)", row, col)
}

func (b *Board) IsComplete() bool {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if b.cells[r][c].value == 0 {
				return false
			}
		}
	}
	return true
}
