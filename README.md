# termcade

Classic arcade games in your terminal, built with Go and Bubble Tea.

**Stack:** Go · [Bubble Tea](https://github.com/charmbracelet/bubbletea) · [Lipgloss](https://github.com/charmbracelet/lipgloss)

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kacheo/termcade)](https://goreportcard.com/report/github.com/kacheo/termcade)
[![CI](https://github.com/kacheo/termcade/actions/workflows/ci.yml/badge.svg)](https://github.com/kacheo/termcade/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kacheo/termcade)](https://github.com/kacheo/termcade/releases)
[![Homebrew](https://img.shields.io/badge/homebrew-tap-blue?logo=homebrew)](https://github.com/kacheo/homebrew-termcade)

---

## Install

### Homebrew (macOS/Linux)

```bash
brew install kacheo/termcade/termcade
```

### From source

```bash
git clone https://github.com/kacheo/termcade
cd termcade
go run ./cmd
```

Or build a binary:

```bash
go build -o termcade ./cmd && ./termcade
```

Requires Go 1.26+.

---

## Games

| Game | Description | Docs |
|------|-------------|------|
| Tetris | Classic block-stacking with ghost piece and configurable start level | [games/tetris/README.md](games/tetris/README.md) |
| Snake | Guide your snake to food, avoid walls and yourself | [games/snake/README.md](games/snake/README.md) |
| Sudoku | Procedurally generated number puzzles, three difficulty levels | [games/sudoku/README.md](games/sudoku/README.md) |
| Blackjack | Player vs dealer — hit or stand to beat 21 | [games/blackjack/README.md](games/blackjack/README.md) |
| Poker | Texas Hold'em with AI opponents and side pots | [games/poker/README.md](games/poker/README.md) |

---

## Project Structure

```
cmd/              — entry point, menu state machine
core/             — Game interface, colour palette, input helpers
  ui/             — lipgloss style helpers
games/tetris/     — Tetris (ghost piece, configurable start level)
games/snake/      — Snake (20×20 grid, 10 levels)
games/sudoku/     — Sudoku (procedurally generated, 3 difficulties)
games/blackjack/  — Blackjack (player vs dealer)
games/poker/      — Texas Hold'em (AI opponents, side pots)
games/cards/      — Shared card types and deck
```

Adding a new game: implement `core.Game` and register it in `cmd/main.go`.
