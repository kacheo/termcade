package sudoku

import (
	"time"
)

type Sudoku struct {
	board       *Board
	difficulty  Difficulty
	cursorRow   int
	cursorCol   int
	pencilMode  bool
	startTime   time.Time
	elapsed     time.Duration
	score       int
	isPaused    bool
	isGameOver  bool
	won         bool
}

func NewSudoku(diff Difficulty) *Sudoku {
	return &Sudoku{
		board:      Generate(diff),
		difficulty: diff,
		startTime:  time.Now(),
		score:      10000,
	}
}

func (s *Sudoku) Name() string {
	return "Sudoku"
}

func (s *Sudoku) Description() string {
	return "Classic number puzzle"
}

func (s *Sudoku) Update(delta time.Duration) error {
	if s.isPaused || s.isGameOver {
		return nil
	}
	s.elapsed = time.Since(s.startTime)
	return nil
}

func (s *Sudoku) HandleInput(key string) {
	switch key {
	case "left":
		if s.cursorCol > 0 {
			s.cursorCol--
		}
	case "right":
		if s.cursorCol < 8 {
			s.cursorCol++
		}
	case "up":
		if s.cursorRow > 0 {
			s.cursorRow--
		}
	case "down":
		if s.cursorRow < 8 {
			s.cursorRow++
		}
	case " ":
		s.pencilMode = !s.pencilMode
	case "backspace", "delete":
		s.board.ClearCell(s.cursorRow, s.cursorCol)
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if s.pencilMode {
			s.togglePencilMark(int(key[0] - '1'))
		} else {
			s.setDigit(int(key[0] - '1'))
		}
	case "p":
		s.isPaused = true
	}
}

func (s *Sudoku) togglePencilMark(digit int) {
	cell := &s.board.cells[s.cursorRow][s.cursorCol]
	if cell.given || cell.value != 0 {
		return
	}
	cell.pencilMarks[digit] = !cell.pencilMarks[digit]
}

func (s *Sudoku) setDigit(digit int) {
	cell := &s.board.cells[s.cursorRow][s.cursorCol]
	if cell.given {
		return
	}
	cell.value = digit + 1
	s.updateConflicts()
	if s.board.IsComplete() {
		s.checkWin()
	}
}

func (s *Sudoku) updateConflicts() {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			s.board.cells[r][c].conflict = s.board.HasConflict(r, c)
		}
	}
}

func (s *Sudoku) checkWin() {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if s.board.HasConflict(r, c) {
				return
			}
		}
	}
	testBoard := *s.board
	if solved, _ := Solve(&testBoard); !solved {
		return
	}
	s.isGameOver = true
	s.won = true
}

func (s *Sudoku) IsPaused() bool       { return s.isPaused }
func (s *Sudoku) IsGameOver() bool      { return s.isGameOver }
func (s *Sudoku) GetScore() int        { return s.score }
func (s *Sudoku) GetLevel() int        { return int(s.difficulty) }
func (s *Sudoku) GetLines() int        { return 0 }