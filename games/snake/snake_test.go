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
	// Head at (5,5) going right; body wraps so (6,5) is a middle segment (not the tail).
	// newHead = (6,5) which is in body[:4] → game over.
	s.body = []core.Position{
		{X: 5, Y: 5}, // head
		{X: 5, Y: 6},
		{X: 6, Y: 6},
		{X: 6, Y: 5}, // next right step lands here — NOT the tail
		{X: 7, Y: 5}, // tail
	}
	s.dir = dirRight
	s.nextDir = dirRight
	s.food = core.Position{X: 0, Y: 0}

	s.Update(s.tickInterval)

	if !s.IsGameOver() {
		t.Error("expected game over when head enters a non-tail body segment")
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

	// Pause is managed externally by the menu; set s.paused directly.
	s.paused = true
	if !s.IsPaused() {
		t.Error("expected IsPaused() true when s.paused is set")
	}

	s.Update(s.tickInterval * 5)
	if s.body[0] != origHead {
		t.Error("snake moved while paused")
	}

	s.paused = false
	if s.IsPaused() {
		t.Error("expected IsPaused() false after clearing s.paused")
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

func TestSnakeSpawnFoodFullBoard(t *testing.T) {
	s := NewSnake()
	// Fill the entire board with body segments
	body := make([]core.Position, 0, BoardWidth*BoardHeight)
	for y := range BoardHeight {
		for x := range BoardWidth {
			body = append(body, core.Position{X: x, Y: y})
		}
	}
	s.body = body
	s.gameOver = false

	s.spawnFood()

	if !s.IsGameOver() {
		t.Error("expected game over when board is completely full")
	}
}

func TestSnakeTailChase(t *testing.T) {
	s := NewSnake()
	// Snake going right: head at (5,5), body extends left.
	// Tail is at (3,5). newHead moving right would be (6,5) — not the tail.
	// Instead, set direction left and place tail one step ahead of head.
	// Head at (5,5) going up, body going down: tail at (5,8).
	// Move up: newHead = (5,4) — not the tail. Let's use a simpler shape:
	// Head at (5,5), body: (5,5),(5,6),(5,7). Direction = right.
	// Tail = (5,7). newHead = (6,5) — not tail. Hard to chase right tail.
	//
	// Classic tail-chase: head going right, then turn down, then left toward tail.
	// Simpler: set up body so head's next step == current tail position.
	// Head (3,5) going right, body: (3,5),(2,5),(2,6),(3,6),(4,6),(4,5) — tail at (4,5).
	// newHead = (4,5) == tail. Without the fix, game over. With fix, not game over.
	s.body = []core.Position{
		{X: 3, Y: 5}, // head
		{X: 2, Y: 5},
		{X: 2, Y: 6},
		{X: 3, Y: 6},
		{X: 4, Y: 6},
		{X: 4, Y: 5}, // tail — will vacate this step
	}
	s.dir = dirRight
	s.nextDir = dirRight
	s.food = core.Position{X: 0, Y: 0} // food far away, not eating

	s.Update(s.tickInterval)

	if s.IsGameOver() {
		t.Error("chasing the tail should not cause game over (tail vacates its cell)")
	}
}
