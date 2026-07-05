# Sudoku MVP Design

## Overview

**Project:** termcade Sudoku game
**Type:** Single-player puzzle game
**Status:** Approved

## Architecture

**MVU pattern** (Model-View-Update) via Bubble Tea:
- `games/sudoku/sudoku.go` — Game state machine, input routing, render orchestration
- `games/sudoku/board.go` — Board data structure, cell operations, validation
- `games/sudoku/generator.go` — Puzzle generation using backtracking + clue removal
- `games/sudoku/solver.go` — Backtracking solver (used by generator + validation)

## Data Model

### Cell Structure
```go
type Cell struct {
    value       int       // 0-9, 0 = empty
    given       bool      // true = pre-filled, immutable
    pencilMarks [9]bool   // candidate digits
    conflict    bool      // real-time error flag
}
```

### Board
```go
type Board [9][9]Cell
```

### Game State
- `difficulty` — easy/medium/hard
- `startTime`, `elapsed time.Duration`
- `score int`
- `isPaused bool`, `isGameOver bool`
- `cursor Row, Col int`
- `pencilMode bool`

## Generation Algorithm

1. **Generate solved board** — Backtracking fill from empty, produces valid complete grid
2. **Remove clues based on difficulty**:
   - Easy: ~38-42 givens
   - Medium: ~30-34 givens
   - Hard: ~24-28 givens
3. **Ensure uniqueness** (Hard only) — Verify single solution via solver

## Input Handling

| Key | Action |
|-----|--------|
| Arrow keys | Navigate cursor |
| 1-9 | Enter digit (or add pencil mark in pencil mode) |
| Space | Toggle pencil mode |
| Backspace/Delete | Clear cell |
| P | Pause |
| Escape | Return to menu (confirm if game in progress) |

## Real-time Validation

- On every digit entry, check row/column/box for conflicts
- Mark conflicting cells with visual indicator (red highlight)
- Conflicts don't prevent entry — player can self-correct
- On board completion, verify solution; if invalid, show error

## Scoring

- Base score = 10000
- Deduction for each conflict remaining at completion
- Time penalty: -1 point per second elapsed
- Best times stored per difficulty (in-memory for MVP)

## Render Layout

```
╔════════════════════════════════════════════════════════════╗
║  SUDOKU                          Time: 05:23   Score: 8472 ║
╠════════════════════════════════════════════════════════════╣
║                        1 2 3   4 5 6   7 8 9               ║
║                     1  █ █ █ | █ █ █ | █ █ █              ║
║                     2  █ 3 █ | █ █ 7 | █ █ █              ║
║ ...                                                            ║
╠════════════════════════════════════════════════════════════╣
║  Mode: [Normal]   Pencil marks: 3 5 | 1 7 | 2 6            ║
║  [Arrow] Navigate  [1-9] Enter  [Space] Pencil  [P] Pause  ║
╚════════════════════════════════════════════════════════════╝
```

## Menu Flow

```
menuMain → menuSudokuOptions (difficulty: Easy/Medium/Hard) → menuPlaying
                                                                ↓
                                                            menuPause
                                                                ↓
                                                         menuGameOver (win/lose)
```

## Interface Implementation

`core.Game` interface with all required methods:
- `Update(delta)` — Timer tick, conflict check
- `Render()` — Full board + sidebar + controls
- `HandleInput(key)` — All key handling above
- `Name()` → "Sudoku"
- `Description()` → "Classic number puzzle"
- `IsPaused()`, `IsGameOver()`, `GetScore()`, `GetLevel()`, `GetLines()`