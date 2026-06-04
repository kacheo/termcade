package snake

import (
	"math/rand"
	"testing"
	"time"

	"tmvgs/core"
)

// TestScenarioSnakeMove verifies that calling Update(200ms) advances the head
// one step in the initial direction (right: X+1).
func TestScenarioSnakeMove(t *testing.T) {
	s := NewSnake()
	// Place food out of bounds so it won't be eaten.
	s.food = core.Position{X: -1, Y: -1}

	initialHead := s.body[0]

	if err := s.Update(200 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	newHead := s.body[0]
	if newHead.X != initialHead.X+1 || newHead.Y != initialHead.Y {
		t.Errorf("head after move = {%d,%d}, want {%d,%d}",
			newHead.X, newHead.Y, initialHead.X+1, initialHead.Y)
	}
	if s.IsGameOver() {
		t.Error("unexpected game over after normal move")
	}
}

// TestScenarioSnakeDirectionChange verifies that pressing "up" and then
// calling Update(200ms) moves the head upward (Y-1).
func TestScenarioSnakeDirectionChange(t *testing.T) {
	s := NewSnake()
	s.food = core.Position{X: -1, Y: -1}

	initialHead := s.body[0]

	s.HandleInput("up")

	if err := s.Update(200 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	newHead := s.body[0]
	if newHead.X != initialHead.X || newHead.Y != initialHead.Y-1 {
		t.Errorf("head after up-turn = {%d,%d}, want {%d,%d}",
			newHead.X, newHead.Y, initialHead.X, initialHead.Y-1)
	}
	if s.IsGameOver() {
		t.Error("unexpected game over after direction change move")
	}
}

// TestScenarioSnakeEatFood verifies that moving into food increments foodEaten
// and grows the snake body by 1.
func TestScenarioSnakeEatFood(t *testing.T) {
	s := NewSnake()
	// Use a deterministic RNG for food respawn after eating.
	s.rng = rand.New(rand.NewSource(42))

	initialLen := len(s.body)
	initialFoodEaten := s.foodEaten

	// Place food directly ahead of the head (snake starts moving right).
	head := s.body[0]
	s.food = core.Position{X: head.X + 1, Y: head.Y}

	if err := s.Update(200 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	if s.IsGameOver() {
		t.Fatal("unexpected game over after eating food")
	}
	if s.foodEaten != initialFoodEaten+1 {
		t.Errorf("foodEaten = %d, want %d", s.foodEaten, initialFoodEaten+1)
	}
	if len(s.body) != initialLen+1 {
		t.Errorf("body length after eating = %d, want %d", len(s.body), initialLen+1)
	}
}

// TestScenarioSnakeWallCollision verifies that driving the snake into a wall
// triggers game over.
func TestScenarioSnakeWallCollision(t *testing.T) {
	s := NewSnake()
	s.food = core.Position{X: -1, Y: -1}

	// Snake starts at (10,10) moving right; pressing "left" is the reverse
	// direction and is blocked, so the snake continues right.
	// From X=10, 9 steps reach X=19 (right edge); the 10th step would move to
	// X=20 which is out of bounds — game over.
	s.HandleInput("left") // ignored: reverse of right

	// 9 steps should NOT trigger game over.
	for i := 0; i < 9; i++ {
		if err := s.Update(200 * time.Millisecond); err != nil {
			t.Fatal(err)
		}
		if s.IsGameOver() {
			t.Fatalf("unexpected game over before wall: step %d", i+1)
		}
	}

	// 10th step hits the right wall.
	if err := s.Update(200 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	if !s.IsGameOver() {
		t.Error("expected game over after hitting the right wall")
	}
}

// TestScenarioSnakeSelfCollision verifies that steering the snake into its own
// body triggers game over.
func TestScenarioSnakeSelfCollision(t *testing.T) {
	s := NewSnake()
	// Set up a compact snake that forms a U-shape so the head can hit the body.
	// Head at (5,5) going right, body wraps so (6,5) is occupied (not the tail).
	//
	// Body layout (non-eating, so tail vacates):
	//   head  (5,5) →
	//         (5,6)
	//         (6,6)
	//   trap  (6,5)  ← head's next step
	//   tail  (7,5)  (will vacate, but (6,5) is NOT the tail so it's checked)
	s.body = []core.Position{
		{X: 5, Y: 5}, // head
		{X: 5, Y: 6},
		{X: 6, Y: 6},
		{X: 6, Y: 5}, // head will move here — not the tail
		{X: 7, Y: 5}, // tail
	}
	s.dir = dirRight
	s.nextDir = dirRight
	s.food = core.Position{X: 19, Y: 19} // food far away, not in path

	if err := s.Update(200 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	if !s.IsGameOver() {
		t.Error("expected game over after snake head entered its own body")
	}
}

// TestScenarioSnakeQuit verifies that pressing "q" sets gameOver immediately
// without requiring an Update call.
func TestScenarioSnakeQuit(t *testing.T) {
	s := NewSnake()

	if s.IsGameOver() {
		t.Fatal("game should not start as game over")
	}

	s.HandleInput("q")

	if !s.IsGameOver() {
		t.Error("expected IsGameOver()=true immediately after pressing 'q'")
	}
}
