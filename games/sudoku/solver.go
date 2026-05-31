package sudoku

import (
	"fmt"
	"math/rand"
	"time"
)

func Solve(board *Board) (bool, error) {
	return solveRecursive(board, 0, 0, rand.New(rand.NewSource(time.Now().UnixNano())))
}

func solveRecursive(board *Board, row, col int, r *rand.Rand) (bool, error) {
	if row == 9 {
		return true, nil
	}
	if col == 9 {
		return solveRecursive(board, row+1, 0, r)
	}
	if board.cells[row][col].value != 0 {
		return solveRecursive(board, row, col+1, r)
	}
	candidates := board.GetCandidates(row, col)
	indices := make([]int, 0, 9)
	for i := 0; i < 9; i++ {
		if candidates[i] {
			indices = append(indices, i)
		}
	}
	if len(indices) == 0 {
		return false, fmt.Errorf("no candidates available for cell (%d,%d)", row, col)
	}
	for _, idx := range r.Perm(len(indices)) {
		i := indices[idx]
		board.cells[row][col].value = i + 1
		if board.HasConflict(row, col) {
			board.cells[row][col].value = 0
			continue
		}
		if solved, _ := solveRecursive(board, row, col+1, r); solved {
			return true, nil
		}
		board.cells[row][col].value = 0
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
