package sudoku

import (
	"testing"
)

func TestGenerateEasy(t *testing.T) {
	board := Generate(DifficultyEasy)
	givenCount := countGivens(board)
	if givenCount < 38 || givenCount > 42 {
		t.Errorf("expected 38-42 givens for easy, got %d", givenCount)
	}
}

func TestGenerateMedium(t *testing.T) {
	board := Generate(DifficultyMedium)
	givenCount := countGivens(board)
	if givenCount < 30 || givenCount > 34 {
		t.Errorf("expected 30-34 givens for medium, got %d", givenCount)
	}
}

func TestGenerateHard(t *testing.T) {
	board := Generate(DifficultyHard)
	givenCount := countGivens(board)
	if givenCount < 24 || givenCount > 28 {
		t.Errorf("expected 24-28 givens for hard, got %d", givenCount)
	}
}

func countGivens(board *Board) int {
	count := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board.cells[r][c].given {
				count++
			}
		}
	}
	return count
}