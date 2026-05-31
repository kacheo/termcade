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