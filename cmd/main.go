package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kacheo/termcade/core"
	"github.com/kacheo/termcade/games/blackjack"
	"github.com/kacheo/termcade/games/poker"
	"github.com/kacheo/termcade/games/snake"
	"github.com/kacheo/termcade/games/sudoku"
	"github.com/kacheo/termcade/games/tetris"
)

type tickMsg struct {
	time.Time
}

type menuState int

const (
	menuMain menuState = iota
	menuTetrisOptions
	menuSnakeOptions
	menuSudokuOptions
	menuBlackjackOptions
	menuPokerOptions
	menuPlaying
	menuPause
	menuGameOver
)

type gameKind int

const (
	gameKindTetris gameKind = iota
	gameKindSnake
	gameKindSudoku
	gameKindBlackjack
	gameKindPoker
)

type model struct {
	currentMenu menuState
	selected    int
	game        core.Game
	activeGame  gameKind
	gameOver    bool
	lastTick    time.Time
	tetrisOpts  struct {
		ghost      bool
		startLevel int
	}
	sudokuOpts struct {
		difficulty   sudoku.Difficulty
		highlightIdx int
	}
	pokerOpts struct {
		seats      int
		difficulty int
	}
}

func (m *model) Init() tea.Cmd {
	m.lastTick = time.Now()
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return tickMsg{t}
	})
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		now := time.Now()
		if m.game != nil && m.currentMenu == menuPlaying && !m.game.IsPaused() {
			delta := now.Sub(m.lastTick)
			err := m.game.Update(delta)
			if err != nil {
				fmt.Printf("Game update error: %v\n", err)
			}
			if m.game.IsGameOver() {
				m.currentMenu = menuGameOver
				m.gameOver = true
				m.selected = 0
			}
		}
		m.lastTick = now
		return m, tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
			return tickMsg{t}
		})
	case tea.KeyMsg:
		switch m.currentMenu {
		case menuMain:
			return m.updateMainMenu(msg)
		case menuTetrisOptions:
			return m.updateTetrisOptions(msg)
		case menuSnakeOptions:
			return m.updateSnakeOptions(msg)
		case menuSudokuOptions:
			return m.updateSudokuOptions(msg)
		case menuBlackjackOptions:
			return m.updateBlackjackOptions(msg)
		case menuPokerOptions:
			return m.updatePokerOptions(msg)
		case menuPlaying:
			return m.updateGame(msg)
		case menuPause:
			return m.updatePauseMenu(msg)
		case menuGameOver:
			return m.updateGameOverMenu(msg)
		}
	}
	return m, nil
}

func (m *model) updateMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := []string{"Play Tetris", "Play Snake", "Play Sudoku", "Play Blackjack", "Play Poker", "", "Quit"}
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(items)-1 {
			m.selected++
		}
	case "enter", " ":
		switch m.selected {
		case 0:
			m.currentMenu = menuTetrisOptions
			m.selected = 0
		case 1:
			m.currentMenu = menuSnakeOptions
			m.selected = 0
		case 2:
			m.currentMenu = menuSudokuOptions
			m.selected = 0
		case 3:
			m.currentMenu = menuBlackjackOptions
			m.selected = 0
		case 4:
			m.currentMenu = menuPokerOptions
			m.selected = 0
		case 6:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) updateTetrisOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		switch m.selected {
		case 0:
			m.tetrisOpts.ghost = false
		case 1:
			if m.tetrisOpts.startLevel > 0 {
				m.tetrisOpts.startLevel--
			}
		}
	case "right", "l":
		switch m.selected {
		case 0:
			m.tetrisOpts.ghost = true
		case 1:
			if m.tetrisOpts.startLevel < 9 {
				m.tetrisOpts.startLevel++
			}
		}
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < 3 {
			m.selected++
		}
	case "enter", " ":
		switch m.selected {
		case 2:
			m.game = tetris.NewTetris(m.tetrisOpts.ghost, m.tetrisOpts.startLevel)
			m.activeGame = gameKindTetris
			m.currentMenu = menuPlaying
			m.gameOver = false
			m.selected = 0
		case 3:
			m.currentMenu = menuMain
			m.selected = 0
		}
	case "q":
		m.currentMenu = menuMain
		m.selected = 0
	}
	return m, nil
}

func (m *model) updateSnakeOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := []string{"Start Game", "Back"}
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(items)-1 {
			m.selected++
		}
	case "enter", " ":
		switch m.selected {
		case 0:
			m.game = snake.NewSnake()
			m.activeGame = gameKindSnake
			m.currentMenu = menuPlaying
			m.gameOver = false
			m.selected = 0
		case 1:
			m.currentMenu = menuMain
			m.selected = 0
		}
	case "q":
		m.currentMenu = menuMain
		m.selected = 0
	}
	return m, nil
}

func (m *model) updateSudokuOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 4 items: 0=Difficulty, 1=Highlight, 2=Start Game, 3=Back
	n := len(sudoku.HighlightOptions)
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < 3 {
			m.selected++
		}
	case "left", "h":
		if m.selected == 0 {
			m.sudokuOpts.difficulty = (m.sudokuOpts.difficulty + 2) % 3
		}
		if m.selected == 1 {
			m.sudokuOpts.highlightIdx = (m.sudokuOpts.highlightIdx + n - 1) % n
		}
	case "right", "l":
		if m.selected == 0 {
			m.sudokuOpts.difficulty = (m.sudokuOpts.difficulty + 1) % 3
		}
		if m.selected == 1 {
			m.sudokuOpts.highlightIdx = (m.sudokuOpts.highlightIdx + 1) % n
		}
	case "enter", " ":
		switch m.selected {
		case 0:
			m.sudokuOpts.difficulty = (m.sudokuOpts.difficulty + 1) % 3
		case 1:
			m.sudokuOpts.highlightIdx = (m.sudokuOpts.highlightIdx + 1) % n
		case 2:
			m.game = sudoku.NewSudoku(m.sudokuOpts.difficulty, m.sudokuOpts.highlightIdx)
			m.activeGame = gameKindSudoku
			m.currentMenu = menuPlaying
			m.gameOver = false
			m.selected = 0
		case 3:
			m.currentMenu = menuMain
			m.selected = 0
		}
	case "q":
		m.currentMenu = menuMain
		m.selected = 0
	}
	return m, nil
}

func (m *model) updateBlackjackOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < 1 {
			m.selected++
		}
	case "enter", " ":
		switch m.selected {
		case 0:
			m.game = blackjack.NewBlackjack()
			m.activeGame = gameKindBlackjack
			m.currentMenu = menuPlaying
			m.gameOver = false
			m.selected = 0
		case 1:
			m.currentMenu = menuMain
			m.selected = 0
		}
	case "q":
		m.currentMenu = menuMain
		m.selected = 0
	}
	return m, nil
}

func (m *model) updatePokerOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pokerOpts.seats == 0 {
		m.pokerOpts.seats = 4
	}
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < 3 {
			m.selected++
		}
	case "left", "h":
		if m.selected == 0 && m.pokerOpts.seats > 3 {
			m.pokerOpts.seats--
		}
		if m.selected == 1 && m.pokerOpts.difficulty > 0 {
			m.pokerOpts.difficulty--
		}
	case "right", "l":
		if m.selected == 0 && m.pokerOpts.seats < 5 {
			m.pokerOpts.seats++
		}
		if m.selected == 1 && m.pokerOpts.difficulty < 2 {
			m.pokerOpts.difficulty++
		}
	case "enter", " ":
		switch m.selected {
		case 2:
			m.game = poker.NewPoker(m.pokerOpts.seats, poker.Difficulty(m.pokerOpts.difficulty))
			m.activeGame = gameKindPoker
			m.currentMenu = menuPlaying
			m.gameOver = false
			m.selected = 0
		case 3:
			m.currentMenu = menuMain
			m.selected = 0
		}
	case "q":
		m.currentMenu = menuMain
		m.selected = 0
	}
	return m, nil
}

func (m *model) restartGame() {
	switch m.activeGame {
	case gameKindTetris:
		m.game = tetris.NewTetris(m.tetrisOpts.ghost, m.tetrisOpts.startLevel)
	case gameKindSnake:
		m.game = snake.NewSnake()
	case gameKindSudoku:
		m.game = sudoku.NewSudoku(m.sudokuOpts.difficulty, m.sudokuOpts.highlightIdx)
	case gameKindBlackjack:
		m.game = blackjack.NewBlackjack()
	case gameKindPoker:
		m.game = poker.NewPoker(m.pokerOpts.seats, poker.Difficulty(m.pokerOpts.difficulty))
	}
}

func (m *model) updateGame(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		m.currentMenu = menuMain
		m.game = nil
	case "esc":
		if s, ok := m.game.(*sudoku.Sudoku); ok {
			if s.QuitRequested() {
				s.ClearQuitRequest()
				m.currentMenu = menuMain
				m.game = nil
				return m, nil
			}
			s.HandleInput("esc")
			return m, nil
		}
		m.game.HandleInput("esc")
	case "p":
		m.currentMenu = menuPause
		m.selected = 0
	default:
		if m.game != nil {
			key := convertKey(msg.String())
			m.game.HandleInput(key)
			if m.game.IsGameOver() {
				m.currentMenu = menuGameOver
				m.gameOver = true
				m.selected = 0
			}
		}
	}
	return m, nil
}

func (m *model) updatePauseMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := []string{"Resume", "Restart", "Main Menu"}
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(items)-1 {
			m.selected++
		}
	case "enter", " ":
		switch m.selected {
		case 0:
			if s, ok := m.game.(*sudoku.Sudoku); ok {
				s.Resume()
			}
			m.currentMenu = menuPlaying
		case 1:
			m.restartGame()
			m.currentMenu = menuPlaying
			m.gameOver = false
		case 2:
			m.currentMenu = menuMain
			m.game = nil
			m.selected = 0
		}
	case "p", "q":
		m.currentMenu = menuPlaying
	}
	return m, nil
}

func (m *model) updateGameOverMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := []string{"Play Again", "Main Menu"}
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(items)-1 {
			m.selected++
		}
	case "enter", " ":
		switch m.selected {
		case 0:
			m.restartGame()
			m.currentMenu = menuPlaying
			m.gameOver = false
			m.selected = 0
		case 1:
			m.currentMenu = menuMain
			m.game = nil
			m.selected = 0
		}
	}
	return m, nil
}

func convertKey(key string) string {
	switch key {
	case "left":
		return "left"
	case "right":
		return "right"
	case "up":
		return "up"
	case "down":
		return "down"
	case "ctrl+c", "q":
		return "q"
	case " ":
		return " "
	default:
		return key
	}
}

func (m *model) View() string {
	switch m.currentMenu {
	case menuMain:
		return m.renderMainMenu()
	case menuTetrisOptions:
		return m.renderTetrisOptions()
	case menuSnakeOptions:
		return m.renderSnakeOptions()
	case menuSudokuOptions:
		return m.renderSudokuOptions()
	case menuBlackjackOptions:
		return m.renderBlackjackOptions()
	case menuPokerOptions:
		return m.renderPokerOptions()
	case menuPlaying:
		return m.renderGame()
	case menuPause:
		return m.renderPauseMenu()
	case menuGameOver:
		return m.renderGameOverMenu()
	}
	return ""
}

func (m *model) renderMainMenu() string {
	items := []string{"Play Tetris", "Play Snake", "Play Sudoku", "Play Blackjack", "Play Poker", "", "Quit"}
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  ╔═══════════════════════════════════════╗\n")
	sb.WriteString("  ║       termcade — Terminal Games       ║\n")
	sb.WriteString("  ╠═══════════════════════════════════════╣\n")
	sb.WriteString("  ║                                       ║\n")
	for i, item := range items {
		if item == "" {
			sb.WriteString("  ║                                       ║\n")
			continue
		}
		if i == m.selected {
			item = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(item)
		}
		fmt.Fprintf(&sb, "  ║  ▶ %-31s ║\n", item)
	}
	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ╚═══════════════════════════════════════╝\n")
	sb.WriteString("    ↑↓ Navigate   Enter Select   Q Quit\n")
	return sb.String()
}

func (m *model) renderTetrisOptions() string {
	ghostText := "OFF"
	if m.tetrisOpts.ghost {
		ghostText = "ON "
	}
	levelText := fmt.Sprintf("%d", m.tetrisOpts.startLevel)

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  ╔═══════════════════════════════════════╗\n")
	sb.WriteString("  ║        Tetris — Options              ║\n")
	sb.WriteString("  ╠═══════════════════════════════════════╣\n")
	sb.WriteString("  ║                                       ║\n")

	ghostStr := fmt.Sprintf("  ║   Ghost Piece     [ %s ]  ◀ ▶      ║", ghostText)
	levelStr := fmt.Sprintf("  ║   Start Level     [  %s  ]  ◀ ▶     ║", levelText)

	switch m.selected {
	case 0:
		ghostStr = "  ║  ▶ Ghost Piece     [ " + ghostText + " ]  ◀ ▶      ║"
		ghostStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(ghostStr)
	case 1:
		levelStr = "  ║  ▶ Start Level     [  " + levelText + "  ]  ◀ ▶     ║"
		levelStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(levelStr)
	}
	sb.WriteString(ghostStr + "\n")
	sb.WriteString(levelStr + "\n")

	sb.WriteString("  ║                                       ║\n")
	if m.selected == 2 {
		fmt.Fprintf(&sb, "  ║  ▶ %-31s ║\n", "Start Game")
	} else {
		sb.WriteString("  ║    Start Game                        ║\n")
	}
	if m.selected == 3 {
		fmt.Fprintf(&sb, "  ║  ▶ %-31s ║\n", "Back")
	} else {
		sb.WriteString("  ║    Back                             ║\n")
	}

	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ╚═══════════════════════════════════════╝\n")
	sb.WriteString("    ←→ Change   ↑↓ Select   Enter Start\n")
	return sb.String()
}

func (m *model) renderSnakeOptions() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  ╔═══════════════════════════════════════╗\n")
	sb.WriteString("  ║         Snake — Ready?                ║\n")
	sb.WriteString("  ╠═══════════════════════════════════════╣\n")
	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ║   Eat food to grow. Avoid walls and   ║\n")
	sb.WriteString("  ║   your own tail. Speed increases      ║\n")
	sb.WriteString("  ║   as you level up.                    ║\n")
	sb.WriteString("  ║                                       ║\n")
	items := []string{"Start Game", "Back"}
	for i, item := range items {
		if i == m.selected {
			item = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(item)
		}
		fmt.Fprintf(&sb, "  ║  ▶ %-31s ║\n", item)
	}
	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ╚═══════════════════════════════════════╝\n")
	sb.WriteString("    ↑↓ Navigate   Enter Select\n")
	return sb.String()
}

func (m *model) renderSudokuOptions() string {
	difficultyText := "Easy"
	switch m.sudokuOpts.difficulty {
	case sudoku.DifficultyMedium:
		difficultyText = "Medium"
	case sudoku.DifficultyHard:
		difficultyText = "Hard"
	}
	highlightText := sudoku.HighlightOptions[m.sudokuOpts.highlightIdx].Name

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  ╔═══════════════════════════════════════╗\n")
	sb.WriteString("  ║        Sudoku — Options              ║\n")
	sb.WriteString("  ╠═══════════════════════════════════════╣\n")
	sb.WriteString("  ║                                       ║\n")

	hlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))

	diffStr := fmt.Sprintf("  ║   Difficulty    [ %-8s ]            ║", difficultyText)
	if m.selected == 0 {
		diffStr = fmt.Sprintf("  ║  ▶ Difficulty    [ %-8s ]            ║", difficultyText)
		diffStr = hlStyle.Render(diffStr)
	}
	sb.WriteString(diffStr + "\n")

	tintStr := fmt.Sprintf("  ║   Highlight     [ %-8s ]            ║", highlightText)
	if m.selected == 1 {
		tintStr = fmt.Sprintf("  ║  ▶ Highlight     [ %-8s ]            ║", highlightText)
		tintStr = hlStyle.Render(tintStr)
	}
	sb.WriteString(tintStr + "\n")

	sb.WriteString("  ║                                       ║\n")
	if m.selected == 2 {
		fmt.Fprintf(&sb, "  ║  ▶ %-31s ║\n", "Start Game")
	} else {
		sb.WriteString("  ║    Start Game                        ║\n")
	}
	if m.selected == 3 {
		fmt.Fprintf(&sb, "  ║  ▶ %-31s ║\n", "Back")
	} else {
		sb.WriteString("  ║    Back                             ║\n")
	}

	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ╚═══════════════════════════════════════╝\n")
	sb.WriteString("    ←→ Change   ↑↓ Select   Enter Confirm   Q Back\n")
	return sb.String()
}

func (m *model) renderBlackjackOptions() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  ╔═══════════════════════════════════════╗\n")
	sb.WriteString("  ║       Blackjack — Ready?              ║\n")
	sb.WriteString("  ╠═══════════════════════════════════════╣\n")
	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ║   Beat the dealer. You and 3 AI       ║\n")
	sb.WriteString("  ║   players vs the house. Hit or        ║\n")
	sb.WriteString("  ║   stand — closest to 21 wins.         ║\n")
	sb.WriteString("  ║                                       ║\n")
	items := []string{"Start Game", "Back"}
	for i, item := range items {
		if i == m.selected {
			item = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(item)
		}
		fmt.Fprintf(&sb, "  ║  ▶ %-31s ║\n", item)
	}
	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ╚═══════════════════════════════════════╝\n")
	sb.WriteString("    ↑↓ Navigate   Enter Select\n")
	return sb.String()
}

func (m *model) renderPokerOptions() string {
	seatsText := fmt.Sprintf("%d", m.pokerOpts.seats)
	diffText := "Medium"
	switch m.pokerOpts.difficulty {
	case 0:
		diffText = "Easy"
	case 2:
		diffText = "Hard"
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  ╔═══════════════════════════════════════╗\n")
	sb.WriteString("  ║        Poker — Options                ║\n")
	sb.WriteString("  ╠═══════════════════════════════════════╣\n")
	sb.WriteString("  ║                                       ║\n")

	hlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))

	seatsStr := fmt.Sprintf("  ║   Seats        [ %s ]  ◀ ▶           ║", seatsText)
	if m.selected == 0 {
		seatsStr = fmt.Sprintf("  ║  ▶ Seats        [ %s ]  ◀ ▶           ║", seatsText)
		seatsStr = hlStyle.Render(seatsStr)
	}
	sb.WriteString(seatsStr + "\n")

	diffStr := fmt.Sprintf("  ║   Difficulty   [ %-8s ]          ║", diffText)
	if m.selected == 1 {
		diffStr = fmt.Sprintf("  ║  ▶ Difficulty   [ %-8s ]          ║", diffText)
		diffStr = hlStyle.Render(diffStr)
	}
	sb.WriteString(diffStr + "\n")

	sb.WriteString("  ║                                       ║\n")
	if m.selected == 2 {
		fmt.Fprintf(&sb, "  ║  ▶ %-31s ║\n", "Start Game")
	} else {
		sb.WriteString("  ║    Start Game                        ║\n")
	}
	if m.selected == 3 {
		fmt.Fprintf(&sb, "  ║  ▶ %-31s ║\n", "Back")
	} else {
		sb.WriteString("  ║    Back                             ║\n")
	}

	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ╚═══════════════════════════════════════╝\n")
	sb.WriteString("    ←→ Change   ↑↓ Select   Enter Confirm   Q Back\n")
	return sb.String()
}

func (m *model) renderGame() string {
	if m.game == nil {
		return "Loading..."
	}
	return m.game.Render()
}

func (m *model) renderPauseMenu() string {
	items := []string{"Resume", "Restart", "Main Menu"}
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  ╔═══════════════════════════════════════╗\n")
	sb.WriteString("  ║              Paused                   ║\n")
	sb.WriteString("  ╠═══════════════════════════════════════╣\n")
	sb.WriteString("  ║                                       ║\n")
	for i, item := range items {
		if i == m.selected {
			item = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(item)
		}
		fmt.Fprintf(&sb, "  ║  ▶ %-31s ║\n", item)
	}
	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ╚═══════════════════════════════════════╝\n")
	sb.WriteString("    ↑↓ Navigate   Enter Select   P Resume\n")
	return sb.String()
}

func (m *model) renderGameOverMenu() string {
	items := []string{"Play Again", "Main Menu"}
	var sb strings.Builder

	if m.game == nil {
		sb.WriteString("\n")
		sb.WriteString("  ╔═══════════════════════════════════════╗\n")
		sb.WriteString("  ║            Game Over!                 ║\n")
		sb.WriteString("  ╠═══════════════════════════════════════╣\n")
		sb.WriteString("  ║                                       ║\n")
	} else if s, ok := m.game.(*sudoku.Sudoku); ok {
		title := "Game Over!"
		if s.Won() {
			title = "   You Win!  "
		}
		sb.WriteString("\n")
		sb.WriteString("  ╔═══════════════════════════════════════╗\n")
		fmt.Fprintf(&sb, "  ║%37s║\n", title)
		sb.WriteString("  ╠═══════════════════════════════════════╣\n")
		sb.WriteString("  ║                                       ║\n")
		minutes := int(s.GetElapsed().Seconds()) / 60
		seconds := int(s.GetElapsed().Seconds()) % 60
		fmt.Fprintf(&sb, "  ║   Time: %02d:%02d                         ║\n", minutes, seconds)
		fmt.Fprintf(&sb, "  ║   Difficulty: %-19s║\n", m.sudokuOpts.difficulty.String())
	} else {
		sb.WriteString("\n")
		sb.WriteString("  ╔═══════════════════════════════════════╗\n")
		sb.WriteString("  ║            Game Over!                 ║\n")
		sb.WriteString("  ╠═══════════════════════════════════════╣\n")
		sb.WriteString("  ║                                       ║\n")
		fmt.Fprintf(&sb, "  ║   Final Score: %-5d                 ║\n", m.game.GetScore())
		fmt.Fprintf(&sb, "  ║   Level Reached: %-3d                ║\n", m.game.GetLevel())
		fmt.Fprintf(&sb, "  ║   Lines Cleared: %-3d                ║\n", m.game.GetLines())
	}

	sb.WriteString("  ║                                       ║\n")
	for i, item := range items {
		if i == m.selected {
			item = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(item)
		}
		fmt.Fprintf(&sb, "  ║  ▶ %-31s ║\n", item)
	}
	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ╚═══════════════════════════════════════╝\n")
	sb.WriteString("    ↑↓ Navigate   Enter Select\n")
	return sb.String()
}

func main() {
	p := tea.NewProgram(&model{
		currentMenu: menuMain,
		selected:    0,
		tetrisOpts: struct {
			ghost      bool
			startLevel int
		}{ghost: false, startLevel: 0},
	})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
		os.Exit(1)
	}
}
