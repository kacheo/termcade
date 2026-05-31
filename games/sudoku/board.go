package sudoku

type Cell struct {
    value       int
    given       bool
    pencilMarks [9]bool
    conflict    bool
}

type Board struct {
    cells [9][9]Cell
}

func NewCell() Cell {
    return Cell{}
}

func NewBoard() Board {
    var board Board
    for r := 0; r < 9; r++ {
        for c := 0; c < 9; c++ {
            board.cells[r][c] = NewCell()
        }
    }
    return board
}

func (b *Board) GetCandidates(row, col int) [9]bool {
    var candidates [9]bool
    for i := 0; i < 9; i++ {
        candidates[i] = true
    }
    for c := 0; c < 9; c++ {
        if v := b.cells[row][c].value; v != 0 {
            candidates[v-1] = false
        }
    }
    for r := 0; r < 9; r++ {
        if v := b.cells[r][col].value; v != 0 {
            candidates[v-1] = false
        }
    }
    boxRow := (row / 3) * 3
    boxCol := (col / 3) * 3
    for r := boxRow; r < boxRow+3; r++ {
        for c := boxCol; c < boxCol+3; c++ {
            if v := b.cells[r][c].value; v != 0 {
                candidates[v-1] = false
            }
        }
    }
    return candidates
}

func (b *Board) HasConflict(row, col int) bool {
	if row < 0 || row >= 9 || col < 0 || col >= 9 {
		return false
	}
	val := b.cells[row][col].value
    if val == 0 {
        return false
    }
    for c := 0; c < 9; c++ {
        if c != col && b.cells[row][c].value == val {
            return true
        }
    }
    for r := 0; r < 9; r++ {
        if r != row && b.cells[r][col].value == val {
            return true
        }
    }
    boxRow := (row / 3) * 3
    boxCol := (col / 3) * 3
    for r := boxRow; r < boxRow+3; r++ {
        for c := boxCol; c < boxCol+3; c++ {
            if (r != row || c != col) && b.cells[r][c].value == val {
                return true
            }
        }
    }
    return false
}

func (b *Board) ClearCell(row, col int) {
	if row < 0 || row >= 9 || col < 0 || col >= 9 {
		return
	}
	if b.cells[row][col].given {
		return
	}
	b.cells[row][col].value = 0
    b.cells[row][col].conflict = false
    for i := 0; i < 9; i++ {
        b.cells[row][col].pencilMarks[i] = false
    }
}

func (b *Board) SetValue(row, col, val int, given bool) {
	if row < 0 || row >= 9 || col < 0 || col >= 9 {
		return
	}
	b.cells[row][col].value = val
	b.cells[row][col].given = given
	for i := 0; i < 9; i++ {
		b.cells[row][col].pencilMarks[i] = false
	}
	b.cells[row][col].conflict = b.HasConflict(row, col)
}