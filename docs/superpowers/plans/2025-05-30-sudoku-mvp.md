# Sudoku MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Sudoku game with procedural generation, pencil marks, real-time validation, timer, and scoring.

**Architecture:** MVU pattern with separate files for board operations (board.go), solver (solver.go), generator (generator.go), and game state (sudoku.go).

**Tech Stack:** Go, Bubble Tea, Lipgloss

---

### Task 1: Board Data Structure

**Files:**
- Create: `games/sudoku/board.go`
- Create: `games/sudoku/board_test.go`

- [ ] **Step 1: Write failing tests**

```go
package sudoku

import "testing"

func TestCellDefault(t *testing.T) {
    cell := NewCell()
    if cell.value != 0 {
        t.Errorf("expected value 0, got %d", cell.value)
    }
    if cell.given {
        t.Error("expected given false")
    }
    if len(cell.pencilMarks) != 9 {
        t.Errorf("expected 9 pencil marks, got %d", len(cell.pencilMarks))
    }
}

func TestBoardInit(t *testing.T) {
    board := NewBoard()
    if len(board.cells) != 9 {
        t.Errorf("expected 9 rows, got %d", len(board.cells))
    }
    for r := 0; r < 9; r++ {
        for c := 0; c < 9; c++ {
            if board.cells[r][c].value != 0 {
                t.Errorf("expected empty cell at [%d][%d]", r, c)
            }
        }
    }
}

func TestGetCandidates(t *testing.T) {
    board := NewBoard()
    board.cells[0][0].value = 5
    candidates := board.GetCandidates(1, 0)
    if candidates[4] { // 5 should be eliminated
        t.Error("5 should not be a candidate at [1][0]")
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./games/sudoku/... -v`
Expected: FAIL (functions not defined)

- [ ] **Step 3: Write minimal implementation**

```go
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
    // Check row
    for c := 0; c < 9; c++ {
        if v := b.cells[row][c].value; v != 0 {
            candidates[v-1] = false
        }
    }
    // Check column
    for r := 0; r < 9; r++ {
        if v := b.cells[r][col].value; v != 0 {
            candidates[v-1] = false
        }
    }
    // Check 3x3 box
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./games/sudoku/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add games/sudoku/board.go games/sudoku/board_test.go
git commit -m "feat(sudoku): add board data structure and candidate calculation"
```

---

### Task 2: Add Board Validation

**Files:**
- Modify: `games/sudoku/board.go`
- Modify: `games/sudoku/board_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestHasConflict(t *testing.T) {
    board := NewBoard()
    board.cells[0][0].value = 5
    board.cells[0][1].value = 5 // duplicate in row
    if !board.HasConflict(0, 1) {
        t.Error("expected conflict at [0][1] due to duplicate in row")
    }
}

func TestClearCell(t *testing.T) {
    board := NewBoard()
    board.cells[0][0].value = 5
    board.cells[0][0].given = true
    board.ClearCell(0, 0)
    // Given cells should not be clearable
    if board.cells[0][0].value != 5 {
        t.Error("given cell should not be cleared")
    }
}

func TestSetValue(t *testing.T) {
    board := NewBoard()
    board.SetValue(0, 0, 5, false)
    if board.cells[0][0].value != 5 {
        t.Error("expected value 5")
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./games/sudoku/... -v`
Expected: FAIL (HasConflict, ClearCell, SetValue not defined)

- [ ] **Step 3: Write minimal implementation**

```go
func (b *Board) HasConflict(row, col int) bool {
    val := b.cells[row][col].value
    if val == 0 {
        return false
    }
    // Check row
    for c := 0; c < 9; c++ {
        if c != col && b.cells[row][c].value == val {
            return true
        }
    }
    // Check column
    for r := 0; r < 9; r++ {
        if r != row && b.cells[r][col].value == val {
            return true
        }
    }
    // Check 3x3 box
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
    b.cells[row][col].value = val
    b.cells[row][col].given = given
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./games/sudoku/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add games/sudoku/board.go games/sudoku/board_test.go
git commit -m "feat(sudoku): add board validation and cell operations"
```

---

### Task 3: Backtracking Solver

**Files:**
- Create: `games/sudoku/solver.go`
- Create: `games/sudoku/solver_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestSolve(t *testing.T) {
    board := NewBoard()
    board.cells[0][0].value = 5
    board.cells[0][1].value = 3
    board.cells[0][4].value = 7
    // ... more given values for a solvable puzzle
    solved, err := Solve(&board)
    if err != nil {
        t.Errorf("solve error: %v", err)
    }
    if !solved {
        t.Error("expected board to be solved")
    }
}

func TestIsComplete(t *testing.T) {
    board := NewBoard()
    for r := 0; r < 9; r++ {
        for c := 0; c < 9; c++ {
            board.cells[r][c].value = (r*9 + c)%9 + 1
        }
    }
    if !board.IsComplete() {
        t.Error("expected complete board")
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./games/sudoku/... -v`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
package sudoku

func Solve(board *Board) (bool, error) {
    return solveRecursive(board, 0, 0)
}

func solveRecursive(board *Board, row, col int) (bool, error) {
    if row == 9 {
        return true, nil
    }
    if col == 9 {
        return solveRecursive(board, row+1, 0)
    }
    if board.cells[row][col].value != 0 {
        return solveRecursive(board, row, col+1)
    }
    candidates := board.GetCandidates(row, col)
    for i := 0; i < 9; i++ {
        if candidates[i] {
            board.cells[row][col].value = i + 1
            if solved, _ := solveRecursive(board, row, col+1); solved {
                return true, nil
            }
            board.cells[row][col].value = 0
        }
    }
    return false, nil
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./games/sudoku/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add games/sudoku/solver.go games/sudoku/solver_test.go
git commit -m "feat(sudoku): add backtracking solver"
```

---

### Task 4: Puzzle Generator

**Files:**
- Create: `games/sudoku/generator.go`
- Create: `games/sudoku/generator_test.go`

- [ ] **Step 1: Write failing tests**

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./games/sudoku/... -v`
Expected: FAIL (DifficultyEasy not defined, Generate not defined)

- [ ] **Step 3: Write minimal implementation**

```go
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
    DifficultyHard:    26,
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
        backup := board.cells[r][c].value
        board.cells[r][c].value = 0
        board.cells[r][c].given = false
        // For hard, ensure unique solution
        if diff := countSolutions(board); diff != 1 {
            board.cells[r][c].value = backup
            board.cells[r][c].given = true
            continue
        }
        removed++
    }
    // Mark remaining clues as given
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./games/sudoku/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add games/sudoku/generator.go games/sudoku/generator_test.go
git commit -m "feat(sudoku): add puzzle generator with difficulty levels"
```

---

### Task 5: Main Sudoku Game Implementation

**Files:**
- Create: `games/sudoku/sudoku.go`
- Create: `games/sudoku/sudoku_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestNewSudoku(t *testing.T) {
    game := NewSudoku(DifficultyEasy)
    if game.Name() != "Sudoku" {
        t.Errorf("expected name Sudoku, got %s", game.Name())
    }
    if game.GetScore() != 10000 {
        t.Errorf("expected initial score 10000, got %d", game.GetScore())
    }
}

func TestCursorMovement(t *testing.T) {
    game := NewSudoku(DifficultyEasy)
    game.HandleInput("right")
    if game.cursorCol != 1 {
        t.Error("cursor should move right")
    }
    game.HandleInput("down")
    if game.cursorRow != 1 {
        t.Error("cursor should move down")
    }
}

func TestPencilMode(t *testing.T) {
    game := NewSudoku(DifficultyEasy)
    if game.pencilMode {
        t.Error("pencil mode should be false initially")
    }
    game.HandleInput(" ")
    if !game.pencilMode {
        t.Error("pencil mode should be true after space")
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./games/sudoku/... -v`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
package sudoku

import (
    "time"
    "tmvgs/core"
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
    // Check if any conflicts remain
    for r := 0; r < 9; r++ {
        for c := 0; c < 9; c++ {
            if s.board.HasConflict(r, c) {
                return
            }
        }
    }
    // Verify solution
    testBoard := *s.board
    if solved, _ := Solve(&testBoard); !solved {
        return
    }
    s.isGameOver = true
    s.won = true
}

func (s *Sudoku) IsPaused() bool       { return s.isPaused }
func (s *Sudoku) IsGameOver() bool     { return s.isGameOver }
func (s *Sudoku) GetScore() int        { return s.score }
func (s *Sudoku) GetLevel() int        { return int(s.difficulty) }
func (s *Sudoku) GetLines() int        { return 0 }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./games/sudoku/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add games/sudoku/sudoku.go games/sudoku/sudoku_test.go
git commit -m "feat(sudoku): add main game implementation"
```

---

### Task 6: Add Render Method

**Files:**
- Modify: `games/sudoku/sudoku.go`
- Add tests for rendering (basic structure check)

- [ ] **Step 1: Write failing test for Render**

```go
func TestRender(t *testing.T) {
    game := NewSudoku(DifficultyEasy)
    output := game.Render()
    if len(output) == 0 {
        t.Error("render should return non-empty string")
    }
    if !strings.Contains(output, "SUDOKU") {
        t.Error("render should contain SUDOKU header")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./games/sudoku/... -v`
Expected: FAIL (no Render method)

- [ ] **Step 3: Write Render implementation**

```go
import (
    "strings"
    "tmvgs/core/ui"
    "github.com/charmbracelet/lipgloss"
)

var gridStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FFFFFF"))

func (s *Sudoku) Render() string {
    var b strings.Builder
    // Header
    minutes := int(s.elapsed.Seconds()) / 60
    seconds := int(s.elapsed.Seconds()) % 60
    b.WriteString("\n")
    b.WriteString(gridStyle.Render(fmt.Sprintf("  SUDOKU    Time: %02d:%02d   Score: %d",
        minutes, seconds, s.score)))
    b.WriteString("\n")
    b.WriteString("  ╔══════════════════════════════════════════╗\n")
    // Column numbers
    b.WriteString("  ║     1   2   3   4   5   6   7   8   9     ║\n")
    // Board
    for r := 0; r < 9; r++ {
        if r == 3 || r == 6 {
            b.WriteString("  ║   ─────────┼─────────┼─────────     ║\n")
        }
        b.WriteString(fmt.Sprintf("  ║ %d ", r+1))
        for c := 0; c < 9; c++ {
            if c == 3 || c == 6 {
                b.WriteString("│")
            }
            cell := s.board.cells[r][c]
            if cell.value == 0 {
                b.WriteString(" ·")
            } else {
                color := ui.GetPieceColor(byte('1' + cell.value - 1))
                if cell.conflict {
                    color = lipgloss.Color("196") // red
                }
                if cell.given {
                    b.WriteString(fmt.Sprintf("\x1b[1m%2d\x1b[0m", cell.value))
                } else {
                    b.WriteString(color.Render(fmt.Sprintf("%2d", cell.value)))
                }
            }
            b.WriteString(" ")
        }
        b.WriteString("║\n")
    }
    b.WriteString("  ╚══════════════════════════════════════════╝\n")
    // Mode indicator
    mode := "Normal"
    if s.pencilMode {
        mode = "Pencil"
    }
    b.WriteString(fmt.Sprintf("  Mode: [%s]   Arrow keys: move   1-9: enter   Space: toggle pencil\n", mode))
    return b.String()
}
```

Need to add `"fmt"` to imports.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./games/sudoku/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add games/sudoku/sudoku.go
git commit -m "feat(sudoku): add render method with board display"
```

---

### Task 7: Menu Integration

**Files:**
- Modify: `cmd/main.go`

- [ ] **Step 1: Add Sudoku to main menu**

In `updateMainMenu()`, add:
```go
if selected == len(games) {
    msg = "Sudoku"
}
```

- [ ] **Step 2: Add menu state for Sudoku options**

Add `menuSudokuOptions` state and handler:
```go
case "sudokuOptions":
    return updateSudokuOptions(msg, m)
```

- [ ] **Step 3: Handle navigation to Sudoku options**

In main menu selection:
```go
case selected == len(games): // Sudoku
    return tea.Batch(
        m.goToSudokuOptions(),
        nil,
    ), nil
```

- [ ] **Step 4: Handle Sudoku game start**

In options selection:
```go
diff := DifficultyEasy + selected
m.game = sudoku.NewSudoku(diff)
m.state = "playing"
```

- [ ] **Step 5: Build and test**

Run: `go build ./... && go run ./cmd`
Manual smoke test.

- [ ] **Step 6: Commit**

```bash
git add cmd/main.go
git commit -m "feat(sudoku): integrate with main menu"
```

---

## Execution Options

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**