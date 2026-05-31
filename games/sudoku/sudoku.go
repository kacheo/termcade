package sudoku

import (
	"fmt"
	"strings"
	"time"

	"tmvgs/core/ui"

	"github.com/charmbracelet/lipgloss"
)

type Move struct {
	row, col   int
	prevValue  int
	prevGiven  bool
	prevPencil [9]bool
}

const MAX_UNDO = 100

type Sudoku struct {
	board          *Board
	difficulty     Difficulty
	cursorRow      int
	cursorCol      int
	pencilMode     bool
	startTime      time.Time
	elapsed        time.Duration
	score          int
	isPaused       bool
	isGameOver     bool
	won            bool
	quitRequested  bool
	undoStack      []Move
	pausedAt       time.Time
	totalPaused    time.Duration
}

func NewSudoku(diff Difficulty) *Sudoku {
	return &Sudoku{
		board:      Generate(diff),
		difficulty: diff,
		startTime:  time.Now(),
		score:      10000,
		undoStack:  make([]Move, 0),
	}
}

func (s *Sudoku) Name() string {
	return "Sudoku"
}

func (s *Sudoku) Description() string {
	return "Classic number puzzle"
}

func (s *Sudoku) Update(delta time.Duration) error {
	if s.isGameOver {
		return nil
	}
	if s.isPaused {
		return nil
	}
	s.elapsed = time.Since(s.startTime) - s.totalPaused
	return nil
}

func (s *Sudoku) HandleInput(key string) {
	if s.quitRequested && key != "esc" {
		s.quitRequested = false
		return
	}
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
		s.clearCurrentCell()
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if s.pencilMode {
			s.togglePencilMark(int(key[0] - '1'))
		} else {
			s.setDigit(int(key[0] - '1'))
		}
	case "p":
		if !s.isPaused {
			s.isPaused = true
			s.pausedAt = time.Now()
		}
	case "u":
		s.undo()
	case "esc":
		s.quitRequested = true
	}
}

func (s *Sudoku) togglePencilMark(digit int) {
	cell := &s.board.cells[s.cursorRow][s.cursorCol]
	if cell.given || cell.value != 0 {
		return
	}
	cell.pencilMarks[digit] = !cell.pencilMarks[digit]
}

func (s *Sudoku) pushUndo(row, col int) {
	if len(s.undoStack) >= MAX_UNDO {
		s.undoStack = s.undoStack[1:]
	}
	cell := &s.board.cells[row][col]
	move := Move{
		row:         row,
		col:         col,
		prevValue:   cell.value,
		prevGiven:   cell.given,
		prevPencil:  cell.pencilMarks,
	}
	s.undoStack = append(s.undoStack, move)
}

func (s *Sudoku) undo() {
	if len(s.undoStack) == 0 {
		return
	}
	move := s.undoStack[len(s.undoStack)-1]
	s.undoStack = s.undoStack[:len(s.undoStack)-1]
	cell := &s.board.cells[move.row][move.col]
	cell.value = move.prevValue
	cell.given = move.prevGiven
	cell.pencilMarks = move.prevPencil
	cell.conflict = s.board.HasConflict(move.row, move.col)
}

func (s *Sudoku) clearCurrentCell() {
	cell := &s.board.cells[s.cursorRow][s.cursorCol]
	if cell.given {
		return
	}
	if cell.value == 0 && !s.hasPencilMarks(cell) {
		return
	}
	s.pushUndo(s.cursorRow, s.cursorCol)
	s.board.ClearCell(s.cursorRow, s.cursorCol)
}

func (s *Sudoku) hasPencilMarks(cell *Cell) bool {
	for _, v := range cell.pencilMarks {
		if v {
			return true
		}
	}
	return false
}

func (s *Sudoku) setDigit(digit int) {
	cell := &s.board.cells[s.cursorRow][s.cursorCol]
	if cell.given {
		return
	}
	newValue := digit + 1
	if newValue == cell.value {
		return
	}
	s.pushUndo(s.cursorRow, s.cursorCol)
	cell.value = newValue
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
	s.isGameOver = true
	s.won = true
}

func (s *Sudoku) IsPaused() bool       { return s.isPaused }
func (s *Sudoku) IsGameOver() bool      { return s.isGameOver }
func (s *Sudoku) GetScore() int        { return s.score }
func (s *Sudoku) GetLevel() int        { return int(s.difficulty) }
func (s *Sudoku) GetLines() int        { return 0 }
func (s *Sudoku) QuitRequested() bool   { return s.quitRequested }
func (s *Sudoku) ClearQuitRequest()    { s.quitRequested = false }
func (s *Sudoku) GetElapsed() time.Duration { return s.elapsed }
func (s *Sudoku) Won() bool            { return s.won }
func (s *Sudoku) Resume() {
	if s.isPaused {
		s.totalPaused += time.Since(s.pausedAt)
		s.isPaused = false
	}
}

var gridStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFFFFF"))

var (
	cursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Background(lipgloss.Color("#333333"))
	pencilMarkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	givenStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
)

func (s *Sudoku) Render() string {
	var b strings.Builder
	minutes := int(s.elapsed.Seconds()) / 60
	seconds := int(s.elapsed.Seconds()) % 60
	b.WriteString("\n")
	b.WriteString(gridStyle.Render(fmt.Sprintf("  SUDOKU    Time: %02d:%02d   Score: %d",
		minutes, seconds, s.score)))
	b.WriteString("\n")
	b.WriteString("  ╔══════════════════════════════╗\n")
	for r := 0; r < 9; r++ {
		if r == 3 || r == 6 {
			b.WriteString("  ║ ─────────┼─────────┼─────────║\n")
		}
		b.WriteString("  ║ ")
		for c := 0; c < 9; c++ {
			if c == 3 || c == 6 {
				b.WriteString("│")
			}
			cell := &s.board.cells[r][c]
			isCursor := r == s.cursorRow && c == s.cursorCol
			s.renderCell(&b, cell, isCursor)
			b.WriteString(" ")
		}
		b.WriteString("║\n")
	}
	b.WriteString("  ╚══════════════════════════════╝\n")
	mode := "Normal"
	if s.pencilMode {
		mode = "Pencil"
	}
	b.WriteString(fmt.Sprintf("  Mode: [%s]   [↑↓←→] Move  [1-9] Enter  [Space] Pencil  [U] Undo  [P] Pause\n", mode))
	if s.quitRequested {
		b.WriteString("  *** Press Esc again to quit, or any other key to cancel ***\n")
	}
	return b.String()
}

func (s *Sudoku) renderCell(b *strings.Builder, cell *Cell, isCursor bool) {
	if cell.value == 0 {
		if s.hasPencilMarks(cell) {
			s.renderPencilMarks(b, cell, isCursor)
		} else {
			if isCursor {
				b.WriteString(cursorStyle.Render("  "))
			} else {
				b.WriteString(" ·")
			}
		}
	} else {
		color := ui.GetPieceColor(byte('1' + cell.value - 1))
		if cell.conflict {
			color = lipgloss.Color("196")
		}
		if cell.given {
			content := givenStyle.Render(fmt.Sprintf("%2d", cell.value))
			if isCursor {
				b.WriteString(cursorStyle.Render(fmt.Sprintf("%2d", cell.value)))
			} else {
				b.WriteString(content)
			}
		} else {
			if isCursor {
				b.WriteString(cursorStyle.Render(fmt.Sprintf("%2d", cell.value)))
			} else {
				b.WriteString(lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%2d", cell.value)))
			}
		}
	}
}

func (s *Sudoku) renderPencilMarks(b *strings.Builder, cell *Cell, isCursor bool) {
	marks := make([]string, 0)
	for i := 0; i < 9; i++ {
		if cell.pencilMarks[i] {
			marks = append(marks, fmt.Sprintf("%d", i+1))
		}
	}
	content := strings.Join(marks, "")

	if isCursor {
		bgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Background(lipgloss.Color("#333333")).
			Width(2)
		if len(content) == 1 {
			b.WriteString(bgStyle.Render(" " + content))
		} else {
			b.WriteString(bgStyle.Render(content))
		}
	} else {
		b.WriteString(pencilMarkStyle.Render(fmt.Sprintf("%-2s", content)))
	}
}