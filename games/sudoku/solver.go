package sudoku

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
	for i := 0; i < 9; i++ {
		if candidates[i] {
			board.cells[row][col].value = i + 1
			if solved, _ := solveRecursive(board, row, col+1); solved {
				return true, nil
			}
			board.cells[row][col].value = 0
		}
	}
	return false, nil
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
