# tmvgs — Terminal Video Games

## Project Overview

**Name:** tmvgs (Terminal Video Games)
**Type:** Arcade game collection for the terminal
**Core Functionality:** A extensible framework for playing classic arcade games (starting with Tetris) in the terminal using Go and Bubbletea
**Target Users:** Terminal enthusiasts, roguelike fans, quick gaming sessions

---

## Concept

tmvgs is a collection of arcade games designed for the terminal. The first game is Tetris — a polished, guideline-compliant implementation. The framework is built to be extensible: adding new games (Snake, Pong, Space Invaders) means implementing a simple `Game` interface and dropping in a new package.

The aesthetic is refined terminal: box-drawing characters, proper color coding, smooth rendering via Bubbletea. No custom sprites needed — the beauty comes from consistent styling and classic gameplay.

---

## Architecture

### Directory Structure

```
tmvgs/
├── cmd/
│   └── main.go              # Entry point, game menu, Bubbletea app router
├── core/
│   ├── game.go              # Game interface + common types
│   ├── input.go             # Input handling abstraction
│   └── ui/
│       └── renderer.go      # Bubbletea grid rendering helpers
├── games/
│   ├── game.go              # Local interface alias (convenience)
│   └── tetris/              # Tetris implementation
│       ├── tetris.go        # Main game logic, implements Game interface
│       ├── pieces.go        # Tetromino definitions + SRS rotation + wall kicks
│       ├── board.go         # Board state, collision detection, line clears
│       ├── scoring.go       # Guideline scoring calculations
│       ├── randomizer.go    # 7-bag piece randomizer
│       └── config.go        # Game configuration (ghost, start level)
├── docs/
│   └── superpowers/
│       └── specs/
│           └── 2025-05-30-tetris-design.md
└── go.mod
```

### Game Interface

All games implement the `Game` interface:

```go
type Game interface {
    Update(delta time.Duration) error  # Game logic tick, delta = time since last update
    Render() string                   # Returns renderable content (tea.String)
    HandleInput(key string)           # Process keypress
    Name() string                     # Game name for menu
    Description() string              # Short description for menu
}
```

Games live in packages under `games/`. To add a new game, create `games/snake/snake.go` and implement `Game`.

---

## Tetris Specification

### Board

- **Dimensions:** 10 columns wide × 20 rows tall
- **Coordinate system:** Column 0-9 (left to right), Row 0-19 (top to bottom)
- **Cell states:** Empty (0) or filled with piece color (1-7)

### Tetrominoes

Seven standard pieces, each with 4 rotation states (SRS):

| Piece | Color | ASCII |
|-------|-------|-------|
| I | Cyan | I |
| O | Yellow | O |
| T | Purple | T |
| S | Green | S |
| Z | Red | Z |
| J | Blue | J |
| L | Orange | L |

**Rotation:** Super Rotation System (SRS) with wall kicks.
**O piece:** Rotates around center (no wall kicks needed).

**SRS Wall Kicks:** Standard 5-point kick data applied on failed rotation attempt.

### Randomizer

**7-bag randomizer:** Shuffle all 7 pieces, deal them out, repeat. Ensures fair distribution — no long droughts without a needed piece.

### Scoring (Guideline)

| Lines Cleared | Points |
|---------------|--------|
| 0 | 0 |
| 1 (Single) | 100 × level |
| 2 (Double) | 300 × level |
| 3 (Triple) | 500 × level |
| 4 (Tetris) | 800 × level |

**Drop bonuses:**
- Soft drop: 1 point per cell
- Hard drop: 2 points per cell

### Lock Delay

- **Duration:** 500ms after piece contacts ground
- **Reset:** Successful move or rotation resets the timer (max 15 resets)

### Ghost Piece

- Off by default
- Toggle via options menu
- Ghost shows where piece will land
- Rendered at 20% opacity of piece color

### Speed Curve (Guideline)

| Level | Drop Interval |
|-------|---------------|
| 0 | 800ms |
| 1 | 720ms |
| 2 | 630ms |
| 3 | 550ms |
| 4 | 480ms |
| 5 | 450ms |
| 6 | 380ms |
| 7 | 330ms |
| 8 | 280ms |
| 9 | 230ms |
| 10+ | 200ms → 50ms (decreases 20ms/level) |

**Level up:** Every 10 lines cleared.

### Controls

| Key | Action |
|-----|--------|
| `←` / `A` | Move left |
| `→` / `D` | Move right |
| `↑` / `W` | Rotate clockwise |
| `↓` / `S` | Soft drop (faster fall) |
| `Space` | Hard drop (instant) |
| `P` | Pause/unpause |
| `Q` | Quit (with confirmation) |

### Game Over

Game ends when a new piece spawns and immediately collides with existing blocks.

---

## UI Design

### Layout

```
╔══════════════════════════════════════════════════╗
║                                                  ║
║   NEXT                              ╔════════╗   ║
║   ┌─────┐                           ║        ║   ║
║   │  T  │                           ║        ║   ║
║   └─────┘                           ║  10x20 ║   ║
║                                    ║  BOARD ║   ║
║   HOLD                             ║        ║   ║
║   ┌─────┐                           ║        ║   ║
║   │     │                           ║        ║   ║
║   └─────┘                           ║        ║   ║
║                                    ╚════════╝   ║
║                                                  ║
║   ─────────────────────────────────────────────  ║
║                                                  ║
║   SCORE                                           ║
║   12,500                                          ║
║                                                  ║
║   LEVEL                                           ║
║   5                                               ║
║                                                  ║
║   LINES                                           ║
║   42                                              ║
║                                                  ║
║   [P] Pause   [Q] Quit                            ║
║                                                  ║
╚══════════════════════════════════════════════════╝
```

### Color Palette

| Element | Color (Hex) |
|---------|-------------|
| I piece | #00FFFF (Cyan) |
| O piece | #FFFF00 (Yellow) |
| T piece | #800080 (Purple) |
| S piece | #00FF00 (Green) |
| Z piece | #FF0000 (Red) |
| J piece | #0000FF (Blue) |
| L piece | #FF8000 (Orange) |
| Ghost | Same as piece at 20% opacity |
| Empty cell | #333333 (Dark gray) |
| Grid lines | #1A1A1A (Very dark) |
| Background | #0D0D0D (Near black) |
| Text | #CCCCCC (Light gray) |
| Border | #666666 (Medium gray) |

### Character Set

| Element | Character |
|---------|-----------|
| Placed block | `█` (full block) |
| Empty cell | `·` (middle dot) |
| Ghost block | `▒` (medium shade) |
| Box drawing | `╔═╗║╚╝╠══╣╚═══╝` |

---

## Menus

### Main Menu

```
╔═══════════════════════════════╗
║      tmvgs — Terminal Games    ║
╠═══════════════════════════════╣
║                               ║
║   ▶ Play Tetris               ║
║     Snake                     ║
║     Pong                      ║
║                               ║
║     Options                   ║
║     Quit                      ║
║                               ║
╚═══════════════════════════════╝
```

### Options Menu

```
╔═══════════════════════════════╗
║         Options               ║
╠═══════════════════════════════╣
║                               ║
║   Ghost Piece    [OFF] ◀ ▶    ║
║   Start Level    [  0 ] ◀ ▶   ║
║                               ║
║        [Back]                 ║
║                               ║
╚═══════════════════════════════╝
```

### Pause Menu

```
╔═══════════════════════════════╗
║         Paused               ║
╠═══════════════════════════════╣
║                               ║
║   ▶ Resume                    ║
║     Restart                   ║
║     Options                   ║
║     Quit                      ║
║                               ║
╚═══════════════════════════════╝
```

### Game Over Screen

```
╔═══════════════════════════════╗
║       Game Over!             ║
╠═══════════════════════════════╣
║                               ║
║   Final Score                 ║
║   12,500                      ║
║                               ║
║   Level Reached: 5            ║
║   Lines Cleared: 42           ║
║                               ║
║   ▶ Play Again                ║
║     Main Menu                 ║
║                               ║
╚═══════════════════════════════╝
```

---

## Technical Details

### Dependencies

- **github.com/charmbracelet/bubbletea** — TUI framework
- **github.com/charmbracelet/lipgloss** — Styling

### Key Implementation Notes

**Game Loop:**
- Bubbletea model runs at ~60fps
- `Update` called each frame with delta time
- Game logic ticks at speed determined by current level
- `Render` returns the full board state as a string

**Collision Detection:**
- Check if piece cells overlap with board cells or boundaries
- Used for movement validation, rotation validation, landing detection

**Line Clearing:**
- Scan board bottom to top
- Remove full rows, shift everything above down
- Award points based on lines cleared and current level

**Input Handling:**
- `HandleInput` receives keypress strings
- Immediately responsive for movement/rotation
- DAS (Delayed Auto Shift): initial 170ms, then 50ms repeat for left/right

---

## Extensibility

### Adding New Games

1. Create `games/<name>/<name>.go`
2. Implement the `Game` interface
3. Add to menu in `cmd/main.go`
4. Import and register in game router

### Future Games (Tentative)

- **Snake:** Classic snake with growing tail, food, collision
- **Pong:** Single player vs AI or 2-player hot-seat
- **Breakout:** Paddle, ball, bricks
- **Space Invaders:** Grid of aliens, shooting, increasing difficulty

---

## Success Criteria

1. Tetris is fully playable — all 7 pieces, rotation, scoring, levels, game over
2. Ghost piece works when enabled
3. 7-bag randomizer produces fair piece distribution
4. Lock delay behaves correctly (500ms, resets on move/rotate)
5. Speed curve follows guideline
6. Menu system works (main, options, pause, game over)
7. New games can be added by implementing `Game` interface
8. Code is clean, well-organized, idiomatic Go