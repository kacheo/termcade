package snake

import (
	"strings"
	"testing"
	"time"

	"tmvgs/core"
)

func advanceTicks(s *Snake, n int) {
	for range n {
		s.Update(s.tickInterval)
	}
}

func TestSnakeMetadata(t *testing.T) {
	s := NewSnake()
	if s.Name() != "Snake" {
		t.Errorf("Name() = %q, want %q", s.Name(), "Snake")
	}
	if s.Description() == "" {
		t.Error("Description() is empty")
	}
	if s.GetScore() != 0 {
		t.Errorf("initial GetScore() = %d, want 0", s.GetScore())
	}
	if s.GetLevel() != 1 {
		t.Errorf("initial GetLevel() = %d, want 1", s.GetLevel())
	}
	if s.GetLines() != 0 {
		t.Errorf("initial GetLines() = %d, want 0", s.GetLines())
	}
}

func TestSnakeInitialState(t *testing.T) {
	s := NewSnake()

	if s.IsGameOver() {
		t.Error("new snake should not be game over")
	}
	if s.IsPaused() {
		t.Error("new snake should not be paused")
	}
	if len(s.body) != 3 {
		t.Errorf("initial body length = %d, want 3", len(s.body))
	}

	cx, cy := BoardWidth/2, BoardHeight/2
	head := s.body[0]
	if head.X != cx || head.Y != cy {
		t.Errorf("head = {%d,%d}, want {%d,%d}", head.X, head.Y, cx, cy)
	}

	for _, p := range s.body {
		if s.food == p {
			t.Error("food spawned on snake body")
		}
	}

	if s.food.X < 0 || s.food.X >= BoardWidth || s.food.Y < 0 || s.food.Y >= BoardHeight {
		t.Errorf("food out of bounds: {%d,%d}", s.food.X, s.food.Y)
	}
}

func TestSnakeMove(t *testing.T) {
	s := NewSnake()
	origHead := s.body[0]

	// Initial direction is right; one tick should move head right by 1
	s.food = core.Position{X: -1, Y: -1} // ensure no food is eaten
	s.Update(s.tickInterval)

	newHead := s.body[0]
	if newHead.X != origHead.X+1 || newHead.Y != origHead.Y {
		t.Errorf("head after right-tick = {%d,%d}, want {%d,%d}", newHead.X, newHead.Y, origHead.X+1, origHead.Y)
	}
	if len(s.body) != 3 {
		t.Errorf("body length after move = %d, want 3", len(s.body))
	}
}

func TestSnakeWallCollision_Right(t *testing.T) {
	s := NewSnake()
	s.food = core.Position{X: -1, Y: -1}
	s.dir = dirRight
	s.nextDir = dirRight
	s.body = []core.Position{{X: BoardWidth - 1, Y: 5}, {X: BoardWidth - 2, Y: 5}}

	s.Update(s.tickInterval)

	if !s.IsGameOver() {
		t.Error("expected game over when hitting right wall")
	}
}

func TestSnakeWallCollision_Left(t *testing.T) {
	s := NewSnake()
	s.food = core.Position{X: -1, Y: -1}
	s.dir = dirLeft
	s.nextDir = dirLeft
	s.body = []core.Position{{X: 0, Y: 5}, {X: 1, Y: 5}}

	s.Update(s.tickInterval)

	if !s.IsGameOver() {
		t.Error("expected game over when hitting left wall")
	}
}

func TestSnakeWallCollision_Top(t *testing.T) {
	s := NewSnake()
	s.food = core.Position{X: -1, Y: -1}
	s.dir = dirUp
	s.nextDir = dirUp
	s.body = []core.Position{{X: 5, Y: 0}, {X: 5, Y: 1}}

	s.Update(s.tickInterval)

	if !s.IsGameOver() {
		t.Error("expected game over when hitting top wall")
	}
}

func TestSnakeWallCollision_Bottom(t *testing.T) {
	s := NewSnake()
	s.food = core.Position{X: -1, Y: -1}
	s.dir = dirDown
	s.nextDir = dirDown
	s.body = []core.Position{{X: 5, Y: BoardHeight - 1}, {X: 5, Y: BoardHeight - 2}}

	s.Update(s.tickInterval)

	if !s.IsGameOver() {
		t.Error("expected game over when hitting bottom wall")
	}
}

func TestSnakeSelfCollision(t *testing.T) {
	s := NewSnake()
	// Build a U-shape that the head will loop into:
	// Head at (5,5) going right, body wraps around
	s.body = []core.Position{
		{X: 5, Y: 5}, // head
		{X: 5, Y: 6},
		{X: 6, Y: 6},
		{X: 6, Y: 5}, // next right step lands here
	}
	s.dir = dirRight
	s.nextDir = dirRight
	s.food = core.Position{X: 0, Y: 0}

	s.Update(s.tickInterval)

	if !s.IsGameOver() {
		t.Error("expected game over on self collision")
	}
}

func TestSnakeCannotReverse(t *testing.T) {
	s := NewSnake()
	s.food = core.Position{X: -1, Y: -1}
	s.dir = dirRight
	s.nextDir = dirRight

	// Input the exact opposite direction — should be ignored
	s.HandleInput("left")
	s.Update(s.tickInterval)

	if s.IsGameOver() {
		t.Error("reversing into self should not cause game over (input ignored)")
	}
	if s.dir != dirRight {
		t.Errorf("direction after invalid reversal = %v, want dirRight", s.dir)
	}
}

func TestSnakeFoodEat(t *testing.T) {
	s := NewSnake()
	origLen := len(s.body)
	origScore := s.GetScore()

	// Place food directly in front of head
	head := s.body[0]
	s.food = core.Position{X: head.X + 1, Y: head.Y}
	s.dir = dirRight
	s.nextDir = dirRight

	s.Update(s.tickInterval)

	if len(s.body) != origLen+1 {
		t.Errorf("body length after eating = %d, want %d", len(s.body), origLen+1)
	}
	if s.GetScore() <= origScore {
		t.Errorf("score did not increase after eating food")
	}
	if s.GetLines() != 1 {
		t.Errorf("foodEaten = %d, want 1", s.GetLines())
	}
}

func TestSnakeLevelUp(t *testing.T) {
	s := NewSnake()
	if s.GetLevel() != 1 {
		t.Fatalf("initial level = %d, want 1", s.GetLevel())
	}

	// Eat foodPerLevel foods directly
	for range foodPerLevel {
		head := s.body[0]
		d := dirDelta[s.dir]
		s.food = core.Position{X: head.X + d.X, Y: head.Y + d.Y}
		s.Update(s.tickInterval)
		if s.IsGameOver() {
			t.Fatal("game over while eating food for level test")
		}
	}

	if s.GetLevel() != 2 {
		t.Errorf("GetLevel() = %d after %d foods, want 2", s.GetLevel(), foodPerLevel)
	}
}

func TestSnakeTickIntervalDecreases(t *testing.T) {
	s := NewSnake()
	initialInterval := s.tickInterval

	for range foodPerLevel {
		head := s.body[0]
		d := dirDelta[s.dir]
		s.food = core.Position{X: head.X + d.X, Y: head.Y + d.Y}
		s.Update(s.tickInterval)
	}

	if s.tickInterval >= initialInterval {
		t.Errorf("tickInterval did not decrease: %v → %v", initialInterval, s.tickInterval)
	}
}

func TestSnakePause(t *testing.T) {
	s := NewSnake()
	s.food = core.Position{X: -1, Y: -1}
	origHead := s.body[0]

	s.HandleInput("p")
	if !s.IsPaused() {
		t.Error("expected paused after 'p'")
	}

	s.Update(s.tickInterval * 5)
	if s.body[0] != origHead {
		t.Error("snake moved while paused")
	}

	s.HandleInput("p")
	if s.IsPaused() {
		t.Error("expected unpaused after second 'p'")
	}
}

func TestSnakeQuit(t *testing.T) {
	s := NewSnake()
	s.HandleInput("q")
	if !s.IsGameOver() {
		t.Error("expected game over after 'q'")
	}
}

func TestSnakeHandleInput_AllDirections(t *testing.T) {
	tests := []struct {
		key     string
		wantDir direction
	}{
		{"up", dirUp},
		{"down", dirDown},
		{"left", dirLeft},
		{"right", dirRight},
	}
	for _, tt := range tests {
		s := NewSnake()
		s.dir = dirRight // start right so left is invalid, but we'll just check nextDir
		s.HandleInput(tt.key)
		if s.nextDir != tt.wantDir {
			t.Errorf("HandleInput(%q) nextDir = %v, want %v", tt.key, s.nextDir, tt.wantDir)
		}
	}
}

func TestSnakeRender(t *testing.T) {
	s := NewSnake()
	out := s.Render()
	if out == "" {
		t.Error("Render() returned empty string")
	}
	for _, want := range []string{"SCORE", "LEVEL", "LENGTH"} {
		if !strings.Contains(out, want) {
			t.Errorf("Render() missing %q", want)
		}
	}
}

func TestSnakeGameOverRender(t *testing.T) {
	s := NewSnake()
	s.gameOver = true
	out := s.Render()
	if out == "" {
		t.Error("Render() returned empty string when game over")
	}
}

func TestSnakeElapsedAccumulator(t *testing.T) {
	s := NewSnake()
	s.food = core.Position{X: -1, Y: -1}
	origHead := s.body[0]

	// half a tick should not move
	s.Update(s.tickInterval / 2)
	if s.body[0] != origHead {
		t.Error("snake moved on sub-tick update")
	}

	// another half tick should trigger the step
	s.Update(s.tickInterval / 2)
	if s.body[0] == origHead {
		t.Error("snake did not move after full tick worth of delta")
	}
}

func TestSnakeLevelCap(t *testing.T) {
	s := NewSnake()
	s.foodEaten = foodPerLevel * (maxLevel + 5) // way past max level
	if s.GetLevel() != maxLevel {
		t.Errorf("GetLevel() = %d, want %d (capped)", s.GetLevel(), maxLevel)
	}
}

func TestSnakeGetLines(t *testing.T) {
	s := NewSnake()
	s.foodEaten = 7
	if s.GetLines() != 7 {
		t.Errorf("GetLines() = %d, want 7", s.GetLines())
	}
}

func TestSnakeNoUpdateWhenGameOver(t *testing.T) {
	s := NewSnake()
	s.gameOver = true
	s.body = []core.Position{{X: 5, Y: 5}, {X: 4, Y: 5}}
	s.dir = dirRight
	snap := s.body[0]

	s.Update(s.tickInterval * 10)
	if s.body[0] != snap {
		t.Error("snake moved after game over")
	}
}

func TestSnakeTickIntervalFor(t *testing.T) {
	s := NewSnake()
	tests := []struct {
		level    int
		wantMS   int
	}{
		{1, baseTickMS},
		{2, baseTickMS - speedupMS},
		{10, baseTickMS - 9*speedupMS},
	}
	for _, tt := range tests {
		got := s.tickInterval_for(tt.level)
		want := time.Duration(tt.wantMS) * time.Millisecond
		if got != want {
			t.Errorf("tickInterval_for(%d) = %v, want %v", tt.level, got, want)
		}
	}
}
