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
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fillBoard(&board, r)
	removeClues(&board, difficultyClues[diff], r)
	return &board
}

func fillBoard(board *Board, r *rand.Rand) {
	ok, err := solveRecursive(board, 0, 0, r)
	if err != nil || !ok {
		panic("sudoku: solver failed to fill empty board")
	}
}

func removeClues(board *Board, targetClues int, r *rand.Rand) {
	cells := make([][2]int, 81)
	idx := 0
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			cells[idx] = [2]int{row, col}
			idx++
		}
	}
	shuffleWith(cells, r)
	removed := 0
	for _, cell := range cells {
		if removed >= 81-targetClues {
			break
		}
		row, col := cell[0], cell[1]
		if board.cells[row][col].value == 0 {
			continue
		}
		board.cells[row][col].value = 0
		board.cells[row][col].given = false
		removed++
	}
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			if board.cells[row][col].value != 0 {
				board.cells[row][col].given = true
			}
		}
	}
}

func shuffleWith(cells [][2]int, r *rand.Rand) {
	for i := len(cells) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		cells[i], cells[j] = cells[j], cells[i]
	}
}