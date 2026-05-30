package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tmvgs/core"
	"tmvgs/games/tetris"
)

type tickMsg struct {
	time.Time
}

type menuState int

const (
	menuMain menuState = iota
	menuTetrisOptions
	menuPlaying
	menuPause
	menuGameOver
)

type model struct {
	currentMenu menuState
	selected    int
	game        core.Game
	gameOver    bool
	lastTick    time.Time
	tetrisOpts  struct {
		ghost      bool
		startLevel int
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
		if m.game != nil && m.currentMenu == menuPlaying && !m.game.IsPaused() {
			now := time.Now()
			delta := now.Sub(m.lastTick)
			m.lastTick = now
			err := m.game.Update(delta)
			if err != nil {
				fmt.Printf("Game update error: %v\n", err)
			}
		}
		return m, tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
			return tickMsg{t}
		})
	case tea.KeyMsg:
		switch m.currentMenu {
		case menuMain:
			return m.updateMainMenu(msg)
		case menuTetrisOptions:
			return m.updateTetrisOptions(msg)
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
	items := []string{"Play Tetris", "Snake (coming soon)", "Pong (coming soon)", "", "Quit"}
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
		case 4:
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

func (m *model) updateGame(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		m.currentMenu = menuMain
		m.game = nil
	case "p":
		m.currentMenu = menuPause
		m.selected = 0
	default:
		if m.game != nil {
			key := convertKey(msg.String())
			m.game.HandleInput(key)
			if t, ok := m.game.(*tetris.Tetris); ok {
				if t.IsGameOver() {
					m.currentMenu = menuGameOver
					m.gameOver = true
					m.selected = 0
				}
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
			m.currentMenu = menuPlaying
		case 1:
			m.game = tetris.NewTetris(m.tetrisOpts.ghost, m.tetrisOpts.startLevel)
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
			m.game = tetris.NewTetris(m.tetrisOpts.ghost, m.tetrisOpts.startLevel)
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
	items := []string{"Play Tetris", "Snake (coming soon)", "Pong (coming soon)", "", "Quit"}
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
	sb.WriteString("\n")
	sb.WriteString("  ╔═══════════════════════════════════════╗\n")
	sb.WriteString("  ║            Game Over!                 ║\n")
	sb.WriteString("  ╠═══════════════════════════════════════╣\n")
	sb.WriteString("  ║                                       ║\n")

	if t, ok := m.game.(*tetris.Tetris); ok {
		sb.WriteString(fmt.Sprintf("  ║   Final Score: %-5d                 ║\n", t.GetScore()))
		sb.WriteString(fmt.Sprintf("  ║   Level Reached: %-3d                ║\n", t.GetLevel()))
		sb.WriteString(fmt.Sprintf("  ║   Lines Cleared: %-3d                ║\n", t.GetLines()))
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