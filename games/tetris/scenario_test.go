package tetris

import (
	"testing"
)

// TestScenarioTetrisHardDrop verifies that hard-dropping a piece scores points
// (+2 per row fallen) and causes a new piece to spawn (or the game to end).
func TestScenarioTetrisHardDrop(t *testing.T) {
	g := NewTetris(false, 0)

	startScore := g.GetScore()
	startPiece := g.current

	// Compute how many rows the piece will fall.
	ghostRows := g.ghostY() - g.current.Y

	g.HandleInput(" ") // hard drop

	// After a hard drop the game either ends (board full) or a new piece appears.
	if g.IsGameOver() {
		// Valid outcome – game ended immediately because the spawn position was
		// already blocked after locking. Nothing more to check.
		return
	}

	// A new piece must have been spawned (different pointer or at least at Y=0).
	if g.current == startPiece {
		t.Error("hard drop: current piece pointer should change after lock")
	}
	if g.current.Y != 0 {
		t.Errorf("hard drop: new piece should spawn at Y=0, got Y=%d", g.current.Y)
	}

	// Score should have increased by 2 per row dropped (hard-drop bonus).
	if ghostRows > 0 {
		wantScore := startScore + ghostRows*2
		if g.GetScore() != wantScore {
			t.Errorf("hard drop %d rows: got score %d, want %d",
				ghostRows, g.GetScore(), wantScore)
		}
	}
}

// TestScenarioTetrisPauseUnpause verifies that pressing "p" toggles the paused state.
func TestScenarioTetrisPauseUnpause(t *testing.T) {
	g := NewTetris(false, 0)

	if g.paused {
		t.Fatal("game should not start paused")
	}

	g.HandleInput("p")
	if !g.paused {
		t.Error("after first 'p': expected paused=true")
	}

	g.HandleInput("p")
	if g.paused {
		t.Error("after second 'p': expected paused=false")
	}
}

// TestScenarioTetrisQuit verifies that pressing "q" sets gameOver and that
// subsequent input is ignored.
func TestScenarioTetrisQuit(t *testing.T) {
	g := NewTetris(false, 0)

	g.HandleInput("q")

	if !g.gameOver {
		t.Error("after 'q': expected gameOver=true")
	}
	if !g.IsGameOver() {
		t.Error("IsGameOver() should return true after 'q'")
	}

	// Subsequent input must be ignored.
	scoreBefore := g.GetScore()
	xBefore := g.current.X
	g.HandleInput("left")
	g.HandleInput("right")
	g.HandleInput("down")
	g.HandleInput(" ")
	if g.GetScore() != scoreBefore {
		t.Error("score should not change after game over")
	}
	if g.current.X != xBefore {
		t.Error("piece should not move after game over")
	}
}

// TestScenarioTetrisMoveAndRotate verifies that left/right movement and rotation
// change position/rotation without panicking.
func TestScenarioTetrisMoveAndRotate(t *testing.T) {
	g := NewTetris(false, 0)

	startX := g.current.X
	startRot := g.current.Rotation

	// Move left and verify position decreases (or stays at wall).
	g.HandleInput("left")
	if g.current.X > startX {
		t.Errorf("move left: X increased from %d to %d", startX, g.current.X)
	}

	// Move right twice and verify position is at least as far right as the
	// original X (we moved left once already, so two rights bring us back or
	// further right).
	g.HandleInput("right")
	g.HandleInput("right")
	if g.current.X < startX {
		t.Errorf("after left+right+right, X %d is less than start %d", g.current.X, startX)
	}

	// Rotate and verify rotation changed (or stayed the same if wall-kicked back).
	beforeRot := g.current.Rotation
	g.HandleInput("up")
	afterRot := g.current.Rotation
	expectedRot := (startRot + 1) % 4
	// Rotation should have advanced by 1 (wall kick may still succeed).
	if afterRot != expectedRot {
		// Only fail if we started at rotation 0 and expected rotation 1 but got
		// something else entirely.
		if beforeRot == 0 && afterRot != 1 {
			t.Errorf("rotate from 0: expected rotation 1, got %d", afterRot)
		}
	}

	// Piece must not collide with the board after all the moves.
	if g.board.Collides(g.current) {
		t.Error("piece collides with board after move/rotate sequence")
	}
}

// TestScenarioTetrisGameOverOnFill verifies that repeatedly hard-dropping pieces
// eventually fills the board and triggers game over.
func TestScenarioTetrisGameOverOnFill(t *testing.T) {
	g := NewTetris(false, 0)

	// Override the RNG to always supply I pieces so the board fills predictably.
	g.rng = []byte{'I', 'I', 'I', 'I', 'I', 'I', 'I'}
	g.rngIndex = 0

	const maxDrops = 300
	for i := 0; i < maxDrops; i++ {
		if g.IsGameOver() {
			break
		}
		g.HandleInput(" ") // hard drop
	}

	if !g.IsGameOver() {
		t.Errorf("expected game over after %d hard drops, but game is still running", maxDrops)
	}

	// Once game over, further input must be ignored.
	scoreBefore := g.GetScore()
	g.HandleInput(" ")
	g.HandleInput("left")
	if g.GetScore() != scoreBefore {
		t.Error("input should be ignored after game over")
	}
}
