package sudoku

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

func countSolutions(board *Board) int {
	count := 0
	countRecursive(board, &count, 0, 0)
	return count
}

func countRecursive(board *Board, count *int, row, col int) {
	if *count > 1 {
		return
	}
	if row == 9 {
		*count++
		return
	}
	if col == 9 {
		countRecursive(board, count, row+1, 0)
		return
	}
	if board.cells[row][col].value != 0 {
		countRecursive(board, count, row, col+1)
		return
	}
	candidates := board.GetCandidates(row, col)
	for i := 0; i < 9; i++ {
		if candidates[i] {
			board.cells[row][col].value = i + 1
			countRecursive(board, count, row, col+1)
			board.cells[row][col].value = 0
		}
	}
}

func shuffle(cells [][2]int) {
	for i := len(cells) - 1; i > 0; i-- {
		j := (i*i*17 + i*7) % (i + 1)
		cells[i], cells[j] = cells[j], cells[i]
	}
}