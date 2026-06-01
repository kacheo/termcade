package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tmvgs/core"
	"tmvgs/games/snake"
	"tmvgs/games/sudoku"
	"tmvgs/games/tetris"
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
	menuPlaying
	menuPause
	menuGameOver
)

type gameKind int

const (
	gameKindTetris gameKind = iota
	gameKindSnake
	gameKindSudoku
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
	items := []string{"Play Tetris", "Play Snake", "Play Sudoku", "Pong (coming soon)", "", "Quit"}
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
		case 5:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) updateTetrisOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if m.selected == 0 {
			m.tetrisOpts.ghost = false
		} else if m.selected == 1 {
			if m.tetrisOpts.startLevel > 0 {
				m.tetrisOpts.startLevel--
			}
		}
	case "right", "l":
		if m.selected == 0 {
			m.tetrisOpts.ghost = true
		} else if m.selected == 1 {
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
		if m.selected == 2 {
			m.game = tetris.NewTetris(m.tetrisOpts.ghost, m.tetrisOpts.startLevel)
			m.activeGame = gameKindTetris
			m.currentMenu = menuPlaying
			m.gameOver = false
			m.selected = 0
		} else if m.selected == 3 {
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
		if m.selected == 0 {
			m.game = snake.NewSnake()
			m.activeGame = gameKindSnake
			m.currentMenu = menuPlaying
			m.gameOver = false
			m.selected = 0
		} else if m.selected == 1 {
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

func (m *model) restartGame() {
	switch m.activeGame {
	case gameKindTetris:
		m.game = tetris.NewTetris(m.tetrisOpts.ghost, m.tetrisOpts.startLevel)
	case gameKindSnake:
		m.game = snake.NewSnake()
	case gameKindSudoku:
		m.game = sudoku.NewSudoku(m.sudokuOpts.difficulty, m.sudokuOpts.highlightIdx)
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
	items := []string{"Play Tetris", "Play Snake", "Play Sudoku", "Pong (coming soon)", "", "Quit"}
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  ╔═══════════════════════════════════════╗\n")
	sb.WriteString("  ║        tmvgs — Terminal Games         ║\n")
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
		sb.WriteString(fmt.Sprintf("  ║  ▶ %-31s ║\n", item))
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

	if m.selected == 0 {
		ghostStr = "  ║  ▶ Ghost Piece     [ " + ghostText + " ]  ◀ ▶      ║"
		ghostStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(ghostStr)
	} else if m.selected == 1 {
		levelStr = "  ║  ▶ Start Level     [  " + levelText + "  ]  ◀ ▶     ║"
		levelStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(levelStr)
	}
	sb.WriteString(ghostStr + "\n")
	sb.WriteString(levelStr + "\n")

	sb.WriteString("  ║                                       ║\n")
	if m.selected == 2 {
		sb.WriteString(fmt.Sprintf("  ║  ▶ %-31s ║\n", "Start Game"))
	} else {
		sb.WriteString("  ║    Start Game                        ║\n")
	}
	if m.selected == 3 {
		sb.WriteString(fmt.Sprintf("  ║  ▶ %-31s ║\n", "Back"))
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
		sb.WriteString(fmt.Sprintf("  ║  ▶ %-31s ║\n", item))
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
		sb.WriteString(fmt.Sprintf("  ║  ▶ %-31s ║\n", "Start Game"))
	} else {
		sb.WriteString("  ║    Start Game                        ║\n")
	}
	if m.selected == 3 {
		sb.WriteString(fmt.Sprintf("  ║  ▶ %-31s ║\n", "Back"))
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
		sb.WriteString(fmt.Sprintf("  ║  ▶ %-31s ║\n", item))
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
		sb.WriteString(fmt.Sprintf("  ║%37s║\n", title))
		sb.WriteString("  ╠═══════════════════════════════════════╣\n")
		sb.WriteString("  ║                                       ║\n")
		minutes := int(s.GetElapsed().Seconds()) / 60
		seconds := int(s.GetElapsed().Seconds()) % 60
		sb.WriteString(fmt.Sprintf("  ║   Time: %02d:%02d                         ║\n", minutes, seconds))
		sb.WriteString(fmt.Sprintf("  ║   Difficulty: %-19s║\n", m.sudokuOpts.difficulty.String()))
	} else {
		sb.WriteString("\n")
		sb.WriteString("  ╔═══════════════════════════════════════╗\n")
		sb.WriteString("  ║            Game Over!                 ║\n")
		sb.WriteString("  ╠═══════════════════════════════════════╣\n")
		sb.WriteString("  ║                                       ║\n")
		sb.WriteString(fmt.Sprintf("  ║   Final Score: %-5d                 ║\n", m.game.GetScore()))
		sb.WriteString(fmt.Sprintf("  ║   Level Reached: %-3d                ║\n", m.game.GetLevel()))
		sb.WriteString(fmt.Sprintf("  ║   Lines Cleared: %-3d                ║\n", m.game.GetLines()))
	}

	sb.WriteString("  ║                                       ║\n")
	for i, item := range items {
		if i == m.selected {
			item = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(item)
		}
		sb.WriteString(fmt.Sprintf("  ║  ▶ %-31s ║\n", item))
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
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
		os.Exit(1)
	}
}