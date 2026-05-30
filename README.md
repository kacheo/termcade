# tmvgs — Terminal Video Games Suite

Classic arcade games in your terminal, built with Go and Bubble Tea.

**Stack:** Go · [Bubble Tea](https://github.com/charmbracelet/bubbletea) · [Lipgloss](https://github.com/charmbracelet/lipgloss)

---

## Getting Started

```bash
git clone https://github.com/kacheo/tmvgs
cd tmvgs
go run ./cmd
```

Or build a binary:

```bash
go build -o main ./cmd && ./main
```

Requires Go 1.21+.

---

## Games

### Tetris

Fully implemented with all standard mechanics.

**Options (before starting):**
- Ghost piece — shows a shadow where the current piece will land
- Start level — 0–9 (higher = faster drop speed)

**Controls:**

| Key | Action |
|-----|--------|
| `← →` | Move |
| `↑` | Rotate |
| `↓` | Soft drop |
| `Space` | Hard drop |
| `P` | Pause / Resume |
| `Q` | Quit to menu |

**Scoring:**

| Lines cleared | Points |
|---------------|--------|
| 1 | 100 × (level + 1) |
| 2 | 300 × (level + 1) |
| 3 | 500 × (level + 1) |
| 4 (Tetris!) | 800 × (level + 1) |

Soft drop: +1 pt/row · Hard drop: +2 pt/row · Level up every 10 lines

### Snake, Pong

Coming soon.

---

## Project Structure

```
cmd/            — entry point, menu state machine
core/           — Game interface, colour palette
  ui/           — lipgloss helpers
games/tetris/   — Tetris game logic and board
```

Adding a new game: implement `core.Game` and register it in `cmd/main.go`.
