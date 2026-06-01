package sudoku

import (
	"fmt"
	"strings"
	"time"

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
	highlightBg    lipgloss.Style
	matchHighlight bool
}

type HighlightOption struct {
	Name  string
	Color lipgloss.Color
}

var HighlightOptions = []HighlightOption{
	{"Blue", lipgloss.Color("#3A3A5A")},
	{"Green", lipgloss.Color("#1A4A2A")},
	{"Red", lipgloss.Color("#4A2020")},
	{"Gray", lipgloss.Color("#3A3A3A")},
	{"None", lipgloss.Color("")},
}

func NewSudoku(diff Difficulty, highlightIdx int) *Sudoku {
	var hl lipgloss.Style
	if opt := HighlightOptions[highlightIdx]; opt.Color != "" {
		hl = lipgloss.NewStyle().Background(opt.Color)
	}
	return &Sudoku{
		board:          Generate(diff),
		difficulty:     diff,
		startTime:      time.Now(),
		score:          10000,
		undoStack:      make([]Move, 0),
		highlightBg:    hl,
		matchHighlight: true,
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
	case "h":
		s.matchHighlight = !s.matchHighlight
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
		row:        row,
		col:        col,
		prevValue:  cell.value,
		prevGiven:  cell.given,
		prevPencil: cell.pencilMarks,
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

func (s *Sudoku) IsPaused() bool            { return s.isPaused }
func (s *Sudoku) IsGameOver() bool          { return s.isGameOver }
func (s *Sudoku) GetScore() int             { return s.score }
func (s *Sudoku) GetLevel() int             { return int(s.difficulty) }
func (s *Sudoku) GetLines() int             { return 0 }
func (s *Sudoku) QuitRequested() bool       { return s.quitRequested }
func (s *Sudoku) ClearQuitRequest()         { s.quitRequested = false }
func (s *Sudoku) GetElapsed() time.Duration { return s.elapsed }
func (s *Sudoku) Won() bool                 { return s.won }
func (s *Sudoku) Resume() {
	if s.isPaused {
		s.totalPaused += time.Since(s.pausedAt)
		s.isPaused = false
	}
}

var gridStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFFFFF"))

var digitColors = [9]lipgloss.Color{
	"#5599FF", // 1 — sky blue
	"#FFCC44", // 2 — amber
	"#CC66FF", // 3 — lavender
	"#44CC88", // 4 — mint
	"#FF7766", // 5 — coral
	"#44AAFF", // 6 — azure
	"#FF9944", // 7 — orange
	"#88CC44", // 8 — lime
	"#FF66AA", // 9 — rose
}

var (
	cursorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Background(lipgloss.Color("#333333"))
	pencilMarkStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	givenStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	completedColor   = lipgloss.Color("#5A5A5A")
	matchHighlightBg = lipgloss.NewStyle().Background(lipgloss.Color("#3A3300"))
)

func (s *Sudoku) Render() string {
	var b strings.Builder
	minutes := int(s.elapsed.Seconds()) / 60
	seconds := int(s.elapsed.Seconds()) % 60
	b.WriteString("\n")
	b.WriteString(gridStyle.Render(fmt.Sprintf("  SUDOKU    Time: %02d:%02d   Score: %d",
		minutes, seconds, s.score)))
	b.WriteString("\n")
	var completedDigits [9]bool
	for d := 1; d <= 9; d++ {
		completedDigits[d-1] = s.board.IsDigitComplete(d)
	}
	cursorDigit := s.board.cells[s.cursorRow][s.cursorCol].value

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
			isHighlighted := !isCursor && (r == s.cursorRow || c == s.cursorCol)
			isMatchingDigit := s.matchHighlight && cursorDigit != 0 && cell.value == cursorDigit && !isCursor
			isComplete := cell.value != 0 && completedDigits[cell.value-1]
			s.renderCell(&b, cell, isCursor, isHighlighted, isMatchingDigit, isComplete)
			b.WriteString(" ")
		}
		b.WriteString("║\n")
	}
	b.WriteString("  ╚══════════════════════════════╝\n")
	mode := "Normal"
	if s.pencilMode {
		mode = "Pencil"
	}
	b.WriteString(fmt.Sprintf("  Mode: [%s]   [↑↓←→] Move  [1-9] Enter  [Space] Pencil  [U] Undo  [P] Pause  [H] Match\n", mode))
	if s.quitRequested {
		b.WriteString("  *** Press Esc again to quit, or any other key to cancel ***\n")
	}
	return b.String()
}

func (s *Sudoku) renderCell(b *strings.Builder, cell *Cell, isCursor, isHighlighted, isMatchingDigit, isComplete bool) {
	if cell.value == 0 {
		if s.hasPencilMarks(cell) {
			s.renderPencilMarks(b, cell, isCursor, isHighlighted)
		} else {
			if isCursor {
				b.WriteString(cursorStyle.Render("  "))
			} else if isHighlighted {
				b.WriteString(s.highlightBg.Render(" ·"))
			} else {
				b.WriteString(" ·")
			}
		}
	} else {
		color := digitColors[cell.value-1]
		if cell.conflict {
			color = lipgloss.Color("196")
		} else if isComplete {
			color = completedColor
		}
		if isCursor {
			b.WriteString(cursorStyle.Render(fmt.Sprintf("%2d", cell.value)))
		} else if isMatchingDigit {
			if cell.given {
				b.WriteString(matchHighlightBg.Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Render(fmt.Sprintf("%2d", cell.value)))
			} else {
				b.WriteString(matchHighlightBg.Foreground(color).Render(fmt.Sprintf("%2d", cell.value)))
			}
		} else if isHighlighted {
			if cell.given {
				b.WriteString(s.highlightBg.Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Render(fmt.Sprintf("%2d", cell.value)))
			} else {
				b.WriteString(s.highlightBg.Foreground(color).Render(fmt.Sprintf("%2d", cell.value)))
			}
		} else {
			if cell.given {
				b.WriteString(givenStyle.Render(fmt.Sprintf("%2d", cell.value)))
			} else {
				b.WriteString(lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%2d", cell.value)))
			}
		}
	}
}

func (s *Sudoku) renderPencilMarks(b *strings.Builder, cell *Cell, isCursor bool, isHighlighted bool) {
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
	} else if isHighlighted {
		b.WriteString(s.highlightBg.Foreground(lipgloss.Color("#888888")).Render(fmt.Sprintf("%-2s", content)))
	} else {
		b.WriteString(pencilMarkStyle.Render(fmt.Sprintf("%-2s", content)))
	}
}
