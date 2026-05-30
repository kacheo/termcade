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

type menuItem struct {
	label    string
	action   string
	submenu  string
}

type model struct {
	currentMenu string
	selected    int
	options     struct {
		ghost      bool
		startLevel int
	}
	game     core.Game
	gameOver bool
	lastTick time.Time
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
		if m.game != nil && m.currentMenu == "playing" && !m.game.IsPaused() {
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
		case "main":
			return m.updateMainMenu(msg)
		case "options":
			return m.updateOptionsMenu(msg)
		case "playing":
			return m.updateGame(msg)
		case "pause":
			return m.updatePauseMenu(msg)
		case "gameover":
			return m.updateGameOverMenu(msg)
		}
	}
	return m, nil
}

func (m *model) updateMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := []string{"Play Tetris", "Snake (coming soon)", "Pong (coming soon)", "", "Options", "Quit"}
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
			m.currentMenu = "playing"
			m.game = tetris.NewTetris()
			m.gameOver = false
		case 4:
			m.currentMenu = "options"
			m.selected = 0
		case 5:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) updateOptionsMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if m.selected == 0 {
			m.options.ghost = !m.options.ghost
		} else if m.selected == 1 {
			if m.options.startLevel > 0 {
				m.options.startLevel--
			}
		}
	case "right", "l":
		if m.selected == 0 {
			m.options.ghost = !m.options.ghost
		} else if m.selected == 1 {
			if m.options.startLevel < 9 {
				m.options.startLevel++
			}
		}
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < 2 {
			m.selected++
		}
	case "enter", " ", "q":
		m.currentMenu = "main"
		m.selected = 0
	}
	return m, nil
}

func (m *model) updateGame(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		m.currentMenu = "main"
		m.game = nil
	case "p":
		m.currentMenu = "pause"
		m.selected = 0
	default:
		if m.game != nil {
			key := convertKey(msg.String())
			m.game.HandleInput(key)
			if t, ok := m.game.(*tetris.Tetris); ok {
				if t.IsGameOver() {
					m.currentMenu = "gameover"
					m.gameOver = true
					m.selected = 0
				}
			}
		}
	}
	return m, nil
}

func (m *model) updatePauseMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := []string{"Resume", "Restart", "Options", "Main Menu"}
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
			m.currentMenu = "playing"
		case 1:
			m.game = tetris.NewTetris()
			m.currentMenu = "playing"
			m.gameOver = false
		case 2:
			m.currentMenu = "options"
			m.selected = 0
		case 3:
			m.currentMenu = "main"
			m.game = nil
			m.selected = 0
		}
	case "p", "q":
		m.currentMenu = "playing"
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
			m.game = tetris.NewTetris()
			m.currentMenu = "playing"
			m.gameOver = false
			m.selected = 0
		case 1:
			m.currentMenu = "main"
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
	case "main":
		return m.renderMainMenu()
	case "options":
		return m.renderOptionsMenu()
	case "playing":
		return m.renderGame()
	case "pause":
		return m.renderPauseMenu()
	case "gameover":
		return m.renderGameOverMenu()
	}
	return ""
}

func (m *model) renderMainMenu() string {
	items := []string{"Play Tetris", "Snake (coming soon)", "Pong (coming soon)", "", "Options", "Quit"}
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

func (m *model) renderOptionsMenu() string {
	ghostText := "OFF"
	if m.options.ghost {
		ghostText = "ON "
	}
	levelText := fmt.Sprintf("%d", m.options.startLevel)

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  ╔═══════════════════════════════════════╗\n")
	sb.WriteString("  ║              Options                 ║\n")
	sb.WriteString("  ╠═══════════════════════════════════════╣\n")
	sb.WriteString("  ║                                       ║\n")

	ghostStr := fmt.Sprintf("  ║   Ghost Piece     [ %s ]  ◀ ▶      ║", ghostText)
	optionsStr := fmt.Sprintf("  ║   Start Level     [  %s  ]  ◀ ▶     ║", levelText)

	if m.selected == 0 {
		ghostStr = "  ║  ▶ Ghost Piece     [ " + ghostText + " ]  ◀ ▶      ║"
		ghostStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(ghostStr)
	} else if m.selected == 1 {
		optionsStr = "  ║  ▶ Start Level     [  " + levelText + "  ]  ◀ ▶     ║"
		optionsStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(optionsStr)
	}
	sb.WriteString(ghostStr + "\n")
	sb.WriteString(optionsStr + "\n")

	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ║           [Back]                      ║\n")
	sb.WriteString("  ║                                       ║\n")
	sb.WriteString("  ╚═══════════════════════════════════════╝\n")
	sb.WriteString("    ←→ Change   ↑↓ Select   Q/Enter Back\n")
	return sb.String()
}

func (m *model) renderGame() string {
	if m.game == nil {
		return "Loading..."
	}
	return m.game.Render()
}

func (m *model) renderPauseMenu() string {
	items := []string{"Resume", "Restart", "Options", "Main Menu"}
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
		currentMenu: "main",
		selected:    0,
		options: struct {
			ghost      bool
			startLevel int
		}{ghost: false, startLevel: 0},
	})
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
		os.Exit(1)
	}
}