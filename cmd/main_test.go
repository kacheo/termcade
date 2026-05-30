package main

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ---- MockGame ---------------------------------------------------------------

type MockGame struct {
	pausedVal   bool
	gameOverVal bool
	scoreVal    int
	levelVal    int
	linesVal    int
	renderOut   string
	inputsSeen  []string
}

func (m *MockGame) Update(_ time.Duration) error { return nil }
func (m *MockGame) Render() string {
	if m.renderOut != "" {
		return m.renderOut
	}
	return "mock render"
}
func (m *MockGame) HandleInput(key string) { m.inputsSeen = append(m.inputsSeen, key) }
func (m *MockGame) Name() string           { return "Mock" }
func (m *MockGame) Description() string    { return "Mock game" }
func (m *MockGame) IsPaused() bool         { return m.pausedVal }
func (m *MockGame) IsGameOver() bool       { return m.gameOverVal }
func (m *MockGame) GetScore() int          { return m.scoreVal }
func (m *MockGame) GetLevel() int          { return m.levelVal }
func (m *MockGame) GetLines() int          { return m.linesVal }

// ---- helpers ----------------------------------------------------------------

func newModel() *model {
	return &model{
		currentMenu: menuMain,
		selected:    0,
		tetrisOpts: struct {
			ghost      bool
			startLevel int
		}{ghost: false, startLevel: 0},
	}
}

func keyMsg(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k), Alt: false}
}

func specialKey(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

// ---- convertKey -------------------------------------------------------------

func TestConvertKey(t *testing.T) {
	cases := []struct{ in, want string }{
		{"left", "left"},
		{"right", "right"},
		{"up", "up"},
		{"down", "down"},
		{"ctrl+c", "q"},
		{"q", "q"},
		{" ", " "},
		{"p", "p"},
		{"anything", "anything"},
	}
	for _, tc := range cases {
		got := convertKey(tc.in)
		if got != tc.want {
			t.Errorf("convertKey(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ---- updateMainMenu ---------------------------------------------------------

func TestUpdateMainMenu_NavigateDown(t *testing.T) {
	m := newModel()
	m.updateMainMenu(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 1 {
		t.Errorf("down: got selected=%d, want 1", m.selected)
	}
}

func TestUpdateMainMenu_NavigateUp_AtZero(t *testing.T) {
	m := newModel()
	m.updateMainMenu(tea.KeyMsg{Type: tea.KeyUp})
	if m.selected != 0 {
		t.Errorf("up at 0 should stay 0, got %d", m.selected)
	}
}

func TestUpdateMainMenu_NavigateBoundary(t *testing.T) {
	m := newModel()
	items := []string{"Play Tetris", "Snake (coming soon)", "Pong (coming soon)", "", "Quit"}
	// Drive to last item
	for i := 0; i < len(items)*2; i++ {
		m.updateMainMenu(tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.selected >= len(items) {
		t.Errorf("selected %d should not exceed item count %d", m.selected, len(items))
	}
}

func TestUpdateMainMenu_SelectTetris(t *testing.T) {
	m := newModel()
	m.selected = 0
	m.updateMainMenu(tea.KeyMsg{Type: tea.KeyEnter})
	if m.currentMenu != menuTetrisOptions {
		t.Errorf("selecting Tetris: got menu %v, want menuTetrisOptions", m.currentMenu)
	}
	if m.selected != 0 {
		t.Errorf("selection should reset to 0, got %d", m.selected)
	}
}

func TestUpdateMainMenu_Quit(t *testing.T) {
	m := newModel()
	m.selected = 4
	_, cmd := m.updateMainMenu(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("selecting Quit should return a non-nil tea.Cmd")
	}
}

func TestUpdateMainMenu_VimKeys(t *testing.T) {
	m := newModel()
	m.updateMainMenu(keyMsg("j")) // down
	if m.selected != 1 {
		t.Errorf("j key: got selected=%d, want 1", m.selected)
	}
	m.updateMainMenu(keyMsg("k")) // up
	if m.selected != 0 {
		t.Errorf("k key: got selected=%d, want 0", m.selected)
	}
}

// ---- updateTetrisOptions ----------------------------------------------------

func TestUpdateTetrisOptions_GhostToggle(t *testing.T) {
	m := newModel()
	m.currentMenu = menuTetrisOptions
	m.selected = 0

	m.updateTetrisOptions(tea.KeyMsg{Type: tea.KeyRight})
	if !m.tetrisOpts.ghost {
		t.Error("right on ghost should set ghost=true")
	}
	m.updateTetrisOptions(tea.KeyMsg{Type: tea.KeyLeft})
	if m.tetrisOpts.ghost {
		t.Error("left on ghost should set ghost=false")
	}
}

func TestUpdateTetrisOptions_LevelIncrement(t *testing.T) {
	m := newModel()
	m.selected = 1
	m.updateTetrisOptions(tea.KeyMsg{Type: tea.KeyRight})
	if m.tetrisOpts.startLevel != 1 {
		t.Errorf("right on level: got %d, want 1", m.tetrisOpts.startLevel)
	}
}

func TestUpdateTetrisOptions_LevelDecrement(t *testing.T) {
	m := newModel()
	m.selected = 1
	m.tetrisOpts.startLevel = 5
	m.updateTetrisOptions(tea.KeyMsg{Type: tea.KeyLeft})
	if m.tetrisOpts.startLevel != 4 {
		t.Errorf("left on level: got %d, want 4", m.tetrisOpts.startLevel)
	}
}

func TestUpdateTetrisOptions_LevelLowerBound(t *testing.T) {
	m := newModel()
	m.selected = 1
	m.tetrisOpts.startLevel = 0
	m.updateTetrisOptions(tea.KeyMsg{Type: tea.KeyLeft})
	if m.tetrisOpts.startLevel != 0 {
		t.Errorf("level should not go below 0, got %d", m.tetrisOpts.startLevel)
	}
}

func TestUpdateTetrisOptions_LevelUpperBound(t *testing.T) {
	m := newModel()
	m.selected = 1
	m.tetrisOpts.startLevel = 9
	m.updateTetrisOptions(tea.KeyMsg{Type: tea.KeyRight})
	if m.tetrisOpts.startLevel != 9 {
		t.Errorf("level should not exceed 9, got %d", m.tetrisOpts.startLevel)
	}
}

func TestUpdateTetrisOptions_StartGame(t *testing.T) {
	m := newModel()
	m.selected = 2
	m.updateTetrisOptions(tea.KeyMsg{Type: tea.KeyEnter})
	if m.currentMenu != menuPlaying {
		t.Errorf("Start Game: got menu %v, want menuPlaying", m.currentMenu)
	}
	if m.game == nil {
		t.Error("Start Game should create a game")
	}
}

func TestUpdateTetrisOptions_Back(t *testing.T) {
	m := newModel()
	m.currentMenu = menuTetrisOptions
	m.selected = 3
	m.updateTetrisOptions(tea.KeyMsg{Type: tea.KeyEnter})
	if m.currentMenu != menuMain {
		t.Errorf("Back: got menu %v, want menuMain", m.currentMenu)
	}
}

func TestUpdateTetrisOptions_QBack(t *testing.T) {
	m := newModel()
	m.currentMenu = menuTetrisOptions
	m.updateTetrisOptions(keyMsg("q"))
	if m.currentMenu != menuMain {
		t.Errorf("q in options: got menu %v, want menuMain", m.currentMenu)
	}
}

func TestUpdateTetrisOptions_Navigation(t *testing.T) {
	m := newModel()
	m.selected = 0
	m.updateTetrisOptions(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 1 {
		t.Errorf("down: got %d, want 1", m.selected)
	}
	m.updateTetrisOptions(tea.KeyMsg{Type: tea.KeyUp})
	if m.selected != 0 {
		t.Errorf("up: got %d, want 0", m.selected)
	}
}

// ---- updateGame -------------------------------------------------------------

func TestUpdateGame_Quit(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	m.game = &MockGame{}
	m.updateGame(keyMsg("q"))
	if m.currentMenu != menuMain {
		t.Errorf("q in game: got menu %v, want menuMain", m.currentMenu)
	}
	if m.game != nil {
		t.Error("q should nil the game")
	}
}

func TestUpdateGame_Pause(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	m.game = &MockGame{}
	m.updateGame(keyMsg("p"))
	if m.currentMenu != menuPause {
		t.Errorf("p in game: got menu %v, want menuPause", m.currentMenu)
	}
}

func TestUpdateGame_GameOver(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	mock := &MockGame{gameOverVal: true}
	m.game = mock
	m.updateGame(keyMsg("x")) // any key triggers HandleInput which sets game over
	if m.currentMenu != menuGameOver {
		t.Errorf("game over: got menu %v, want menuGameOver", m.currentMenu)
	}
	if !m.gameOver {
		t.Error("m.gameOver should be true")
	}
}

func TestUpdateGame_ForwardsInputToGame(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	mock := &MockGame{}
	m.game = mock
	m.updateGame(tea.KeyMsg{Type: tea.KeyLeft})
	if len(mock.inputsSeen) == 0 {
		t.Error("game should have received input")
	}
}

// ---- updatePauseMenu --------------------------------------------------------

func TestUpdatePauseMenu_Resume(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPause
	m.game = &MockGame{}
	m.selected = 0
	m.updatePauseMenu(tea.KeyMsg{Type: tea.KeyEnter})
	if m.currentMenu != menuPlaying {
		t.Errorf("Resume: got %v, want menuPlaying", m.currentMenu)
	}
}

func TestUpdatePauseMenu_Restart(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPause
	m.selected = 1
	m.updatePauseMenu(tea.KeyMsg{Type: tea.KeyEnter})
	if m.currentMenu != menuPlaying {
		t.Errorf("Restart: got %v, want menuPlaying", m.currentMenu)
	}
	if m.game == nil {
		t.Error("Restart should create a new game")
	}
}

func TestUpdatePauseMenu_MainMenu(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPause
	m.selected = 2
	m.updatePauseMenu(tea.KeyMsg{Type: tea.KeyEnter})
	if m.currentMenu != menuMain {
		t.Errorf("Main Menu: got %v, want menuMain", m.currentMenu)
	}
}

func TestUpdatePauseMenu_PResumeKey(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPause
	m.updatePauseMenu(keyMsg("p"))
	if m.currentMenu != menuPlaying {
		t.Errorf("p in pause: got %v, want menuPlaying", m.currentMenu)
	}
}

func TestUpdatePauseMenu_Navigation(t *testing.T) {
	m := newModel()
	m.selected = 0
	m.updatePauseMenu(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 1 {
		t.Errorf("down: got %d, want 1", m.selected)
	}
	m.updatePauseMenu(tea.KeyMsg{Type: tea.KeyUp})
	if m.selected != 0 {
		t.Errorf("up: got %d, want 0", m.selected)
	}
}

// ---- updateGameOverMenu -----------------------------------------------------

func TestUpdateGameOverMenu_PlayAgain(t *testing.T) {
	m := newModel()
	m.currentMenu = menuGameOver
	m.selected = 0
	m.updateGameOverMenu(tea.KeyMsg{Type: tea.KeyEnter})
	if m.currentMenu != menuPlaying {
		t.Errorf("Play Again: got %v, want menuPlaying", m.currentMenu)
	}
	if m.game == nil {
		t.Error("Play Again should create a new game")
	}
	if m.gameOver {
		t.Error("gameOver should be reset to false")
	}
}

func TestUpdateGameOverMenu_MainMenu(t *testing.T) {
	m := newModel()
	m.currentMenu = menuGameOver
	m.selected = 1
	m.updateGameOverMenu(tea.KeyMsg{Type: tea.KeyEnter})
	if m.currentMenu != menuMain {
		t.Errorf("Main Menu: got %v, want menuMain", m.currentMenu)
	}
}

func TestUpdateGameOverMenu_Navigation(t *testing.T) {
	m := newModel()
	m.selected = 0
	m.updateGameOverMenu(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 1 {
		t.Errorf("down: got %d, want 1", m.selected)
	}
	m.updateGameOverMenu(tea.KeyMsg{Type: tea.KeyUp})
	if m.selected != 0 {
		t.Errorf("up: got %d, want 0", m.selected)
	}
}

// ---- renderMainMenu ---------------------------------------------------------

func TestRenderMainMenu_ContainsItems(t *testing.T) {
	m := newModel()
	out := m.renderMainMenu()
	for _, item := range []string{"Play Tetris", "Quit"} {
		if !strings.Contains(out, item) {
			t.Errorf("renderMainMenu should contain %q", item)
		}
	}
}

func TestRenderMainMenu_SelectionMarker(t *testing.T) {
	items := []string{"Play Tetris", "Snake (coming soon)", "Pong (coming soon)", "Quit"}
	for sel, item := range []int{0, 1, 2, 4} {
		m := newModel()
		m.selected = item
		out := m.renderMainMenu()
		if !strings.Contains(out, "▶") {
			t.Errorf("selected=%d: render should contain selection marker ▶", sel)
		}
		_ = items
	}
}

// ---- renderTetrisOptions ----------------------------------------------------

func TestRenderTetrisOptions_GhostOff(t *testing.T) {
	m := newModel()
	m.tetrisOpts.ghost = false
	out := m.renderTetrisOptions()
	if !strings.Contains(out, "OFF") {
		t.Error("ghost=false should show OFF")
	}
}

func TestRenderTetrisOptions_GhostOn(t *testing.T) {
	m := newModel()
	m.tetrisOpts.ghost = true
	out := m.renderTetrisOptions()
	if !strings.Contains(out, "ON") {
		t.Error("ghost=true should show ON")
	}
}

func TestRenderTetrisOptions_Level(t *testing.T) {
	m := newModel()
	m.tetrisOpts.startLevel = 7
	out := m.renderTetrisOptions()
	if !strings.Contains(out, "7") {
		t.Error("renderTetrisOptions should show start level 7")
	}
}

func TestRenderTetrisOptions_AllSelectionsRender(t *testing.T) {
	for sel := 0; sel <= 3; sel++ {
		m := newModel()
		m.selected = sel
		out := m.renderTetrisOptions()
		if out == "" {
			t.Errorf("selected=%d: renderTetrisOptions returned empty string", sel)
		}
	}
}

// ---- renderGame -------------------------------------------------------------

func TestRenderGame_NilGame(t *testing.T) {
	m := newModel()
	m.game = nil
	out := m.renderGame()
	if out != "Loading..." {
		t.Errorf("nil game: got %q, want Loading...", out)
	}
}

func TestRenderGame_WithMock(t *testing.T) {
	m := newModel()
	m.game = &MockGame{renderOut: "mock board"}
	out := m.renderGame()
	if out != "mock board" {
		t.Errorf("renderGame: got %q, want mock board", out)
	}
}

// ---- renderPauseMenu --------------------------------------------------------

func TestRenderPauseMenu_ContainsItems(t *testing.T) {
	m := newModel()
	out := m.renderPauseMenu()
	for _, item := range []string{"Paused", "Resume", "Restart", "Main Menu"} {
		if !strings.Contains(out, item) {
			t.Errorf("renderPauseMenu should contain %q", item)
		}
	}
}

func TestRenderPauseMenu_AllSelections(t *testing.T) {
	for sel := 0; sel <= 2; sel++ {
		m := newModel()
		m.selected = sel
		out := m.renderPauseMenu()
		if !strings.Contains(out, "▶") {
			t.Errorf("selected=%d: should contain ▶", sel)
		}
	}
}

// ---- renderGameOverMenu -----------------------------------------------------

func TestRenderGameOverMenu_ContainsStats(t *testing.T) {
	m := newModel()
	m.game = &MockGame{scoreVal: 1234, levelVal: 3, linesVal: 27}
	out := m.renderGameOverMenu()
	for _, want := range []string{"1234", "3", "27", "Game Over"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderGameOverMenu should contain %q", want)
		}
	}
}

func TestRenderGameOverMenu_NilGame(t *testing.T) {
	m := newModel()
	m.game = nil
	// Should not panic; stats section is skipped
	out := m.renderGameOverMenu()
	if !strings.Contains(out, "Game Over") {
		t.Error("renderGameOverMenu should contain 'Game Over' even with nil game")
	}
}

func TestRenderGameOverMenu_AllSelections(t *testing.T) {
	for sel := 0; sel <= 1; sel++ {
		m := newModel()
		m.selected = sel
		m.game = &MockGame{}
		out := m.renderGameOverMenu()
		if !strings.Contains(out, "▶") {
			t.Errorf("selected=%d: should contain ▶", sel)
		}
	}
}

// ---- Init / Update(tickMsg) -------------------------------------------------

func TestInit_SetsLastTick(t *testing.T) {
	m := newModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a non-nil tick command")
	}
	if m.lastTick.IsZero() {
		t.Error("Init should set lastTick")
	}
}

func TestUpdate_TickMsg_AdvancesGame(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	mock := &MockGame{}
	m.game = mock
	m.lastTick = time.Now().Add(-100 * time.Millisecond)

	result, cmd := m.Update(tickMsg{time.Now()})
	if result == nil {
		t.Error("Update should return a model")
	}
	if cmd == nil {
		t.Error("Update tickMsg should schedule next tick")
	}
}

func TestUpdate_TickMsg_WhenPaused(t *testing.T) {
	m := newModel()
	m.currentMenu = menuPlaying
	m.game = &MockGame{pausedVal: true}

	_, cmd := m.Update(tickMsg{time.Now()})
	if cmd == nil {
		t.Error("paused game Update should still schedule next tick")
	}
}

func TestUpdate_TickMsg_NoGame(t *testing.T) {
	m := newModel()
	m.currentMenu = menuMain
	// No game set; tickMsg while not in playing state
	_, cmd := m.Update(tickMsg{time.Now()})
	if cmd == nil {
		t.Error("Update tickMsg should schedule next tick even outside playing state")
	}
}

func TestUpdate_KeyMsg_Routing(t *testing.T) {
	m := newModel()
	m.currentMenu = menuMain
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if result == nil {
		t.Error("Update KeyMsg should return model")
	}
	if m.selected != 1 {
		t.Errorf("down key in main menu: got selected=%d, want 1", m.selected)
	}
}

// ---- View dispatch ----------------------------------------------------------

func TestView_AllMenuStates(t *testing.T) {
	states := []menuState{menuMain, menuTetrisOptions, menuPlaying, menuPause, menuGameOver}
	for _, state := range states {
		m := newModel()
		m.currentMenu = state
		if state == menuPlaying || state == menuPause || state == menuGameOver {
			m.game = &MockGame{renderOut: "game"}
		}
		out := m.View()
		if out == "" {
			t.Errorf("View() for state %v returned empty string", state)
		}
	}
}
