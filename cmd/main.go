package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tmvgs/core"
	"tmvgs/games/pong"
	"tmvgs/games/tetris"
)

type tickMsg struct {
	time.Time
}

type menuState int

const (
	menuMain menuState = iota
	menuTetrisOptions
	menuPongOptions
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
	pongOpts struct {
		speedIncrease bool
		aiDifficulty  int
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
		case menuPongOptions:
			return m.updatePongOptions(msg)
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
	items := []string{"Play Tetris", "Play Pong", "", "Quit"}
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
			m.currentMenu = menuPongOptions
			m.selected = 0
		case 3:
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

func (m *model) updatePongOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if m.selected == 0 {
			m.pongOpts.speedIncrease = false
		} else if m.selected == 1 {
			if m.pongOpts.aiDifficulty > 0 {
				m.pongOpts.aiDifficulty--
			}
		}
	case "right", "l":
		if m.selected == 0 {
			m.pongOpts.speedIncrease = true
		} else if m.selected == 1 {
			if m.pongOpts.aiDifficulty < 2 {
				m.pongOpts.aiDifficulty++
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
			m.game = pong.NewPong(m.pongOpts.speedIncrease, m.pongOpts.aiDifficulty)
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
	case menuPongOptions:
		return m.renderPongOptions()
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
	items := []string{"Play Tetris", "Play Pong", "", "Quit"}
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

func (m *model) renderPongOptions() string {
	speedText := "OFF"
	if m.pongOpts.speedIncrease {
		speedText = "ON "
	}
	difficulty := []string{"Easy", "Medium", "Hard"}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  ╔═══════════════════════════════════════╗\n")
	sb.WriteString("  ║          Pong — Options              ║\n")
	sb.WriteString("  ╠═══════════════════════════════════════╣\n")
	sb.WriteString("  ║                                       ║\n")

	speedStr := fmt.Sprintf("  ║   Speed Increase   [ %s ]  ◀ ▶      ║", speedText)
	diffStr := fmt.Sprintf("  ║   AI Difficulty    [ %s ]  ◀ ▶     ║", difficulty[m.pongOpts.aiDifficulty])

	if m.selected == 0 {
		speedStr = "  ║  ▶ Speed Increase   [ " + speedText + " ]  ◀ ▶      ║"
		speedStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(speedStr)
	} else if m.selected == 1 {
		diffStr = "  ║  ▶ AI Difficulty    [ " + difficulty[m.pongOpts.aiDifficulty] + " ]  ◀ ▶     ║"
		diffStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(diffStr)
	}
	sb.WriteString(speedStr + "\n")
	sb.WriteString(diffStr + "\n")

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

	if m.game != nil {
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
		pongOpts: struct {
			speedIncrease bool
			aiDifficulty  int
		}{speedIncrease: false, aiDifficulty: 1},
	})
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
		os.Exit(1)
	}
}