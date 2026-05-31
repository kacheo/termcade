package sudoku

import (
	"math/rand"
	"time"
)

type Difficulty int

const (
	DifficultyEasy Difficulty = iota
	DifficultyMedium
	DifficultyHard
)

var difficultyClues = map[Difficulty]int{
	DifficultyEasy:   40,
	DifficultyMedium:  32,
	DifficultyHard:    28,
}

func (d Difficulty) String() string {
	switch d {
	case DifficultyEasy:
		return "Easy"
	case DifficultyMedium:
		return "Medium"
	case DifficultyHard:
		return "Hard"
	default:
		return "Unknown"
	}
}

func Generate(diff Difficulty) *Board {
	board := NewBoard()
	fillBoard(&board)
	clues := difficultyClues[diff]
	removeClues(&board, clues)
	return &board
}

func fillBoard(board *Board) {
	solveRecursive(board, 0, 0)
}

func removeClues(board *Board, targetClues int) {
	cells := make([][2]int, 81)
	idx := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			cells[idx] = [2]int{r, c}
			idx++
		}
	}
	shuffle(cells)
	removed := 0
	for _, cell := range cells {
		if removed >= 81-targetClues {
			break
		}
		r, c := cell[0], cell[1]
		if board.cells[r][c].value == 0 {
			continue
		}
		board.cells[r][c].value = 0
		board.cells[r][c].given = false
		removed++
	}
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board.cells[r][c].value != 0 {
				board.cells[r][c].given = true
			}
		}
	}
}

func shuffle(cells [][2]int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(cells) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		cells[i], cells[j] = cells[j], cells[i]
	}
}