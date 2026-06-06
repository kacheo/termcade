//go:build regression

package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/bubbletea"

	"tmvgs/games/sudoku"
)

var updateGoldens = flag.Bool("update", false, "update regression golden files")

const regressionGoldenDir = "testdata/regression"

// checkGolden strips ANSI escape codes from got and compares against
// the golden file at testdata/regression/<name>.txt. With -update
// the file is written instead. Mirrors the pattern in devlog's
// tests/regression/regression_test.go.
func checkGolden(t *testing.T, name, got string) {
	t.Helper()

	clean := strings.TrimRight(ansi.Strip(got), "\n") + "\n"

	path := filepath.Join(regressionGoldenDir, name+".txt")

	if *updateGoldens {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("mkdir golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(clean), 0o600); err != nil {
			t.Fatalf("write golden %s: %v", name, err)
		}
		t.Logf("updated golden: %s", name)
		return
	}

	want, err := os.ReadFile(path) //nolint:gosec // path is constructed from test name, not user input
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("golden file not found: %s\nrun with -update to create (go test -tags=regression -run TestRegression -args -update ./cmd/...)", path)
		}
		t.Fatalf("read golden %s: %v", name, err)
	}

	if clean != string(want) {
		t.Errorf("golden mismatch: %s\n--- got ---\n%s\n--- want ---\n%s",
			name, clean, string(want))
	}
}

// ---- menuMain ---------------------------------------------------------------

func TestRegression_MainMenu_SelectedFirst(t *testing.T) {
	m := newModel()
	m.selected = 0
	checkGolden(t, "main_menu_selected_first", m.renderMainMenu())
}

func TestRegression_MainMenu_SelectedLast(t *testing.T) {
	m := newModel()
	m.selected = 6 // Quit
	checkGolden(t, "main_menu_selected_last", m.renderMainMenu())
}

// ---- menuTetrisOptions ------------------------------------------------------

func TestRegression_TetrisOptions_Default(t *testing.T) {
	m := newModel()
	m.currentMenu = menuTetrisOptions
	m.selected = 0
	checkGolden(t, "tetris_options_default", m.renderTetrisOptions())
}

func TestRegression_TetrisOptions_GhostOn_Level5(t *testing.T) {
	m := newModel()
	m.currentMenu = menuTetrisOptions
	m.selected = 1
	m.tetrisOpts.ghost = true
	m.tetrisOpts.startLevel = 5
	checkGolden(t, "tetris_options_ghost_on_level_5", m.renderTetrisOptions())
}

// ---- menuSnakeOptions -------------------------------------------------------

func TestRegression_SnakeOptions(t *testing.T) {
	m := newModel()
	m.currentMenu = menuSnakeOptions
	m.selected = 0
	checkGolden(t, "snake_options", m.renderSnakeOptions())
}

// ---- menuSudokuOptions ------------------------------------------------------

func TestRegression_SudokuOptions_MediumHighlight(t *testing.T) {
	m := newModel()
	m.currentMenu = menuSudokuOptions
	m.selected = 0
	m.sudokuOpts.difficulty = sudoku.DifficultyMedium
	m.sudokuOpts.highlightIdx = 0
	checkGolden(t, "sudoku_options_medium", m.renderSudokuOptions())
}

// ---- menuBlackjackOptions ---------------------------------------------------

func TestRegression_BlackjackOptions(t *testing.T) {
	m := newModel()
	m.currentMenu = menuBlackjackOptions
	m.selected = 0
	checkGolden(t, "blackjack_options", m.renderBlackjackOptions())
}

// ---- menuPokerOptions -------------------------------------------------------

func TestRegression_PokerOptions_4SeatsHard(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPokerOptions
	m.selected = 1
	m.pokerOpts.seats = 4
	m.pokerOpts.difficulty = 2 // Hard
	checkGolden(t, "poker_options_4seats_hard", m.renderPokerOptions())
}

// ---- menuPlaying ------------------------------------------------------------

func TestRegression_GameArea_NilGame(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	m.game = nil
	checkGolden(t, "game_area_loading", m.renderGame())
}

func TestRegression_GameArea_Tetris(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	m.game = &MockGame{renderOut: "[ tetris board ]"}
	m.activeGame = gameKindTetris
	checkGolden(t, "game_area_tetris", m.renderGame())
}

func TestRegression_GameArea_Snake(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	m.game = &MockGame{renderOut: "[ snake board ]"}
	m.activeGame = gameKindSnake
	checkGolden(t, "game_area_snake", m.renderGame())
}

func TestRegression_GameArea_Sudoku(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	m.game = &MockGame{renderOut: "[ sudoku board ]"}
	m.activeGame = gameKindSudoku
	checkGolden(t, "game_area_sudoku", m.renderGame())
}

func TestRegression_GameArea_Blackjack(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	m.game = &MockGame{renderOut: "[ blackjack table ]"}
	m.activeGame = gameKindBlackjack
	checkGolden(t, "game_area_blackjack", m.renderGame())
}

func TestRegression_GameArea_Poker(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	m.game = &MockGame{renderOut: "[ poker table ]"}
	m.activeGame = gameKindPoker
	checkGolden(t, "game_area_poker", m.renderGame())
}

// ---- menuPause --------------------------------------------------------------

func TestRegression_PauseMenu_Resume(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPause
	m.selected = 0
	m.game = &MockGame{renderOut: "game\n"}
	checkGolden(t, "pause_menu_resume", m.renderPauseMenu())
}

// ---- menuGameOver -----------------------------------------------------------

func TestRegression_GameOverMenu_NilGame(t *testing.T) {
	m := newModel()
	m.currentMenu = menuGameOver
	m.game = nil
	m.gameOver = true
	checkGolden(t, "game_over_menu_nil_game", m.renderGameOverMenu())
}

func TestRegression_GameOverMenu_WithMockScore(t *testing.T) {
	m := newModel()
	m.currentMenu = menuGameOver
	m.game = &MockGame{scoreVal: 4200, levelVal: 7, linesVal: 42}
	m.gameOver = true
	checkGolden(t, "game_over_menu_with_score", m.renderGameOverMenu())
}

// ---- View dispatch ----------------------------------------------------------

func TestRegression_View_AllMenuStates(t *testing.T) {
	// Smoke test: each menu state should produce a non-empty View() output
	// whose content matches its dedicated render*() function.
	cases := []struct {
		name  string
		state menuState
		setup func(m *model)
	}{
		{"main", menuMain, func(m *model) {}},
		{"tetris_opts", menuTetrisOptions, func(m *model) {}},
		{"snake_opts", menuSnakeOptions, func(m *model) {}},
		{"sudoku_opts", menuSudokuOptions, func(m *model) {}},
		{"blackjack_opts", menuBlackjackOptions, func(m *model) {}},
		{"poker_opts", menuPokerOptions, func(m *model) {}},
		{"playing", menuPlaying, func(m *model) {
			m.game = &MockGame{renderOut: "game"}
		}},
		{"pause", menuPause, func(m *model) {
			m.game = &MockGame{renderOut: "game"}
		}},
		{"gameover", menuGameOver, func(m *model) {
			m.game = &MockGame{scoreVal: 100, levelVal: 2, linesVal: 10}
			m.gameOver = true
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newModel()
			m.currentMenu = tc.state
			tc.setup(m)
			out := m.View()
			if strings.TrimSpace(ansi.Strip(out)) == "" {
				t.Errorf("View() for state %s returned empty output", tc.name)
			}
		})
	}
}

// ensure tea import is used (for future expansion without churn)
var _ tea.Model = (*model)(nil)
