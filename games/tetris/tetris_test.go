package tetris

import (
	"strings"
	"testing"
	"time"
)

// ---- helpers ----------------------------------------------------------------

// forceCurrentType keeps spawning until the current piece matches the given
// type, swapping in a known next piece via the unexported bag.
func newTetrisWithPiece(pieceType byte) *Tetris {
	for {
		t := NewTetris(false, 0)
		if t.current.Type == pieceType {
			return t
		}
	}
}

// dropToFloor moves the current piece straight down until it locks.
func dropToFloor(t *Tetris) {
	t.HandleInput(" ")
}

// ---- NewTetris / metadata ---------------------------------------------------

func TestNewTetris(t *testing.T) {
	g := NewTetris(true, 5)
	if g.current == nil {
		t.Error("current piece should not be nil after NewTetris")
	}
	if g.next == nil {
		t.Error("next piece should not be nil after NewTetris")
	}
	if g.level != 5 {
		t.Errorf("startLevel=5: got level %d", g.level)
	}
	if g.paused {
		t.Error("should not start paused")
	}
	if g.gameOver {
		t.Error("should not start game over")
	}
	if !g.ghost {
		t.Error("ghost=true should be stored")
	}
}

func TestTetrisMetadata(t *testing.T) {
	g := NewTetris(false, 0)
	if g.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if g.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestTetrisGetters(t *testing.T) {
	g := NewTetris(false, 3)
	if g.GetScore() != 0 {
		t.Errorf("initial score: got %d, want 0", g.GetScore())
	}
	if g.GetLevel() != 3 {
		t.Errorf("initial level: got %d, want 3", g.GetLevel())
	}
	if g.GetLines() != 0 {
		t.Errorf("initial lines: got %d, want 0", g.GetLines())
	}
	if g.IsPaused() {
		t.Error("should not be paused initially")
	}
	if g.IsGameOver() {
		t.Error("should not be game over initially")
	}
}

// ---- 7-bag RNG --------------------------------------------------------------

func TestTetrisBagSequence(t *testing.T) {
	// White-box test: reset the bag and verify a single full bag contains all 7 types.
	g := NewTetris(false, 0)
	g.shuffleRNG()

	seen := make(map[byte]bool)
	for i := 0; i < 7; i++ {
		pt := g.nextPieceType()
		if seen[pt] {
			t.Errorf("duplicate piece type %c at position %d in bag", pt, i)
		}
		seen[pt] = true
	}
	if len(seen) != 7 {
		t.Errorf("bag should contain all 7 unique piece types, got %d: %v", len(seen), seen)
	}

	// After exhausting the bag, calling nextPieceType again reshuffles automatically
	pt := g.nextPieceType()
	allTypes := map[byte]bool{'I': true, 'O': true, 'T': true, 'S': true, 'Z': true, 'J': true, 'L': true}
	if !allTypes[pt] {
		t.Errorf("piece type after reshuffle %c is not a valid piece type", pt)
	}
}

// ---- spawnPiece promotes next -----------------------------------------------

func TestTetrisSpawnPromotesNext(t *testing.T) {
	g := NewTetris(false, 0)
	expectedNext := g.next.Type
	dropToFloor(g)
	if g.IsGameOver() {
		t.Skip("game over before second piece")
	}
	if g.current.Type != expectedNext {
		t.Errorf("after lock, current should be old next (%c), got %c", expectedNext, g.current.Type)
	}
}

// ---- move -------------------------------------------------------------------

func TestTetrisMove_Left(t *testing.T) {
	g := NewTetris(false, 0)
	origX := g.current.X
	g.HandleInput("left")
	if g.current.X != origX-1 {
		t.Errorf("move left: got x=%d, want %d", g.current.X, origX-1)
	}
}

func TestTetrisMove_Right(t *testing.T) {
	g := NewTetris(false, 0)
	origX := g.current.X
	g.HandleInput("right")
	if g.current.X != origX+1 {
		t.Errorf("move right: got x=%d, want %d", g.current.X, origX+1)
	}
}

func TestTetrisMove_WallLeft(t *testing.T) {
	g := NewTetris(false, 0)
	// Drive current piece to the left wall
	for i := 0; i < BoardWidth; i++ {
		g.HandleInput("left")
	}
	xAtWall := g.current.X
	g.HandleInput("left")
	if g.current.X != xAtWall {
		t.Errorf("piece should stop at left wall, but x changed from %d to %d", xAtWall, g.current.X)
	}
}

func TestTetrisMove_WallRight(t *testing.T) {
	g := NewTetris(false, 0)
	for i := 0; i < BoardWidth; i++ {
		g.HandleInput("right")
	}
	xAtWall := g.current.X
	g.HandleInput("right")
	if g.current.X != xAtWall {
		t.Errorf("piece should stop at right wall, but x changed from %d to %d", xAtWall, g.current.X)
	}
}

// ---- rotate -----------------------------------------------------------------

func TestTetrisRotate(t *testing.T) {
	g := NewTetris(false, 0)
	for expectedRot := 0; expectedRot < 4; expectedRot++ {
		if g.current.Rotation != expectedRot%4 {
			t.Errorf("rotation %d: got %d", expectedRot, g.current.Rotation)
		}
		g.HandleInput("up")
	}
	// After 4 rotations should be back to 0
	if g.current.Rotation != 0 {
		t.Errorf("after 4 rotations: got %d, want 0", g.current.Rotation)
	}
}

func TestTetrisRotate_BlockedReverts(t *testing.T) {
	// Use an O piece (rotation doesn't change shape, always safe)
	// Instead test with an I piece pushed to left wall where rotation may be blocked
	g := NewTetris(false, 0)
	// Drive to left wall, then try to rotate — if blocked the rotation should revert
	for i := 0; i < BoardWidth; i++ {
		g.HandleInput("left")
	}
	beforeRot := g.current.Rotation
	g.HandleInput("up")
	// Either rotation succeeded (no collision) or reverted — piece must still be valid (not stuck in wall)
	if g.board.Collides(g.current) {
		t.Error("after rotate attempt, current piece should not collide with board")
	}
	_ = beforeRot
}

// ---- ghostY -----------------------------------------------------------------

func TestTetrisGhostY(t *testing.T) {
	g := NewTetris(true, 0)
	gy := g.ghostY()
	if gy < g.current.Y {
		t.Errorf("ghostY %d should be >= current.Y %d", gy, g.current.Y)
	}
	// ghostY+1 should collide
	test := &Piece{X: g.current.X, Y: gy + 1, Type: g.current.Type, Rotation: g.current.Rotation}
	if !g.board.Collides(test) {
		t.Errorf("piece at ghostY+1 (%d) should collide", gy+1)
	}
}

func TestTetrisGhostY_OnFloor(t *testing.T) {
	g := NewTetris(true, 0)
	// Move piece to bottom manually so it's sitting on the floor
	for !g.onGround {
		if !g.move(0, 1) {
			break
		}
	}
	gy := g.ghostY()
	if gy != g.current.Y {
		t.Errorf("piece already on floor: ghostY %d != current.Y %d", gy, g.current.Y)
	}
}

// ---- scoring ----------------------------------------------------------------

// fillRowExcept fills an entire row except for one column.
func fillRowExcept(b *Board, y, exceptX int, color byte) {
	for x := 0; x < BoardWidth; x++ {
		if x != exceptX {
			b.SetCell(x, y, color)
		}
	}
}

func TestTetrisLock_ScoringFourLines(t *testing.T) {
	g := NewTetris(false, 0)
	// Fill rows 16-19 except column 4
	for _, row := range []int{16, 17, 18, 19} {
		fillRowExcept(g.board, row, 4, 'X')
	}
	// I-piece rotation 1 cells: (x+1, y-1), (x+1, y), (x+1, y+1), (x+1, y+2)
	// At X=3, Y=17: fills (4,16),(4,17),(4,18),(4,19) — completes all 4 rows
	g.current = &Piece{Type: 'I', Color: 'I', X: 3, Y: 17, Rotation: 1}
	g.lock()

	// 4 lines: base=800, level=0 so multiplier=1
	if g.GetScore() != 800 {
		t.Errorf("4-line clear at level 0: got score %d, want 800", g.GetScore())
	}
	if g.GetLines() != 4 {
		t.Errorf("4-line clear: got lines %d, want 4", g.GetLines())
	}
}

func TestTetrisLock_LevelMultiplier(t *testing.T) {
	g := NewTetris(false, 2)
	fillRowExcept(g.board, BoardHeight-1, 4, 'X')
	g.current = &Piece{Type: 'I', Color: 'I', X: 4, Y: BoardHeight - 1, Rotation: 0}
	g.lock()
	// 1 line: base=100, level=2 so multiplier=3
	if g.GetScore() != 300 {
		t.Errorf("1-line clear at level 2: got %d, want 300", g.GetScore())
	}
}

func TestTetrisLock_LevelProgression(t *testing.T) {
	g := NewTetris(false, 0)
	// Clear 10 lines to advance level
	for i := 0; i < 10; i++ {
		fillRowExcept(g.board, BoardHeight-1, 4, 'X')
		g.current = &Piece{Type: 'I', Color: 'I', X: 4, Y: BoardHeight - 1, Rotation: 0}
		g.lock()
		if g.IsGameOver() {
			t.Fatal("game over during level progression test")
		}
	}
	if g.GetLevel() != 1 {
		t.Errorf("after 10 lines: got level %d, want 1", g.GetLevel())
	}
}

func TestTetrisLock_LevelCap(t *testing.T) {
	g := NewTetris(false, 0)
	g.lines = 210 // 210 lines → level 21, should be capped
	fillRowExcept(g.board, BoardHeight-1, 4, 'X')
	g.current = &Piece{Type: 'I', Color: 'I', X: 4, Y: BoardHeight - 1, Rotation: 0}
	g.lock()
	if g.GetLevel() > 20 {
		t.Errorf("level should be capped at 20, got %d", g.GetLevel())
	}
}

// ---- HandleInput ------------------------------------------------------------

func TestTetrisHandleInput_SoftDrop(t *testing.T) {
	g := NewTetris(false, 0)
	startY := g.current.Y
	g.HandleInput("down")
	if g.GetScore() != 1 {
		t.Errorf("soft drop 1 row: got score %d, want 1", g.GetScore())
	}
	if g.current.Y != startY+1 {
		t.Errorf("soft drop: expected y=%d, got %d", startY+1, g.current.Y)
	}
}

func TestTetrisHandleInput_HardDrop(t *testing.T) {
	g := NewTetris(false, 0)
	startY := g.current.Y
	gy := g.ghostY()
	g.HandleInput(" ")
	rows := gy - startY
	if rows > 0 && g.GetScore() != rows*2 {
		t.Errorf("hard drop %d rows: got score %d, want %d", rows, g.GetScore(), rows*2)
	}
}

func TestTetrisHandleInput_Pause(t *testing.T) {
	g := NewTetris(false, 0)
	g.HandleInput("p")
	if !g.IsPaused() {
		t.Error("game should be paused after 'p'")
	}
	g.HandleInput("p")
	if g.IsPaused() {
		t.Error("game should be unpaused after second 'p'")
	}
}

func TestTetrisHandleInput_Quit(t *testing.T) {
	g := NewTetris(false, 0)
	g.HandleInput("q")
	if !g.IsGameOver() {
		t.Error("'q' should set gameOver")
	}
}

func TestTetrisHandleInput_WhenGameOver(t *testing.T) {
	g := NewTetris(false, 0)
	g.HandleInput("q") // trigger game over
	scoreAfterQuit := g.GetScore()
	// Input should now be ignored
	g.HandleInput("left")
	g.HandleInput("right")
	g.HandleInput("down")
	if g.GetScore() != scoreAfterQuit {
		t.Error("input should be ignored when game over")
	}
}

// ---- Update -----------------------------------------------------------------

func TestTetrisUpdate_AutoDrop(t *testing.T) {
	g := NewTetris(false, 0)
	startY := g.current.Y
	// Force lastDrop to be far in the past so the next Update triggers a drop
	g.lastDrop = time.Now().Add(-2 * time.Second)
	if err := g.Update(16 * time.Millisecond); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if g.current.Y <= startY && !g.IsGameOver() {
		t.Errorf("piece should have dropped: startY=%d, currentY=%d", startY, g.current.Y)
	}
}

func TestTetrisUpdate_LockDelay(t *testing.T) {
	g := NewTetris(false, 0)
	// Move piece to floor
	for g.move(0, 1) {
	}
	g.onGround = true
	g.lockTimer = time.Now().Add(-2 * LockDelay) // past lock delay
	pieceBeforeUpdate := g.current
	if err := g.Update(16 * time.Millisecond); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	// Either the piece locked and a new one was spawned, or game ended
	if !g.IsGameOver() && g.current == pieceBeforeUpdate {
		t.Error("piece should have locked after lock delay elapsed")
	}
}

func TestTetrisUpdate_Paused(t *testing.T) {
	g := NewTetris(false, 0)
	g.HandleInput("p")
	startY := g.current.Y
	g.lastDrop = time.Now().Add(-2 * time.Second)
	_ = g.Update(time.Second)
	if g.current.Y != startY {
		t.Error("paused game should not drop pieces on Update")
	}
}

func TestTetrisUpdate_GameOver(t *testing.T) {
	g := NewTetris(false, 0)
	g.gameOver = true
	if err := g.Update(time.Second); err != nil {
		t.Errorf("Update on game over should return nil, got %v", err)
	}
}

// ---- Game Over on spawn -----------------------------------------------------

func TestTetrisGameOver_OnSpawn(t *testing.T) {
	g := NewTetris(false, 0)
	// Fill the top rows so the next spawn has no room
	for x := 0; x < BoardWidth; x++ {
		g.board.SetCell(x, 0, 'X')
		g.board.SetCell(x, 1, 'X')
	}
	g.spawnPiece()
	if !g.IsGameOver() {
		t.Error("spawning into a full board should trigger game over")
	}
}

// ---- Render -----------------------------------------------------------------

func TestTetrisRender_ContainsPiece(t *testing.T) {
	g := NewTetris(false, 0)
	out := g.Render()
	if !strings.Contains(out, "██") {
		t.Error("Render should contain '██' for the current piece")
	}
}

func TestTetrisRender_ContainsStats(t *testing.T) {
	g := NewTetris(false, 0)
	out := g.Render()
	for _, want := range []string{"SCORE", "LEVEL", "LINES"} {
		if !strings.Contains(out, want) {
			t.Errorf("Render should contain %q in stats", want)
		}
	}
}

func TestTetrisRender_ContainsNext(t *testing.T) {
	g := NewTetris(false, 0)
	out := g.Render()
	if !strings.Contains(out, "NEXT:") {
		t.Error("Render should contain 'NEXT:' label")
	}
}

func TestTetrisRender_GhostEnabled(t *testing.T) {
	// Ghost piece only shows when piece isn't already on the floor
	g := NewTetris(true, 0)
	// Ensure piece isn't at the floor yet
	out := g.Render()
	// Only check if the piece has room to show a ghost (ghostY != current.Y)
	if g.ghostY() != g.current.Y {
		if !strings.Contains(out, "░░") {
			t.Error("ghost enabled: render should contain '░░' when piece has space to fall")
		}
	}
}

func TestTetrisRender_GhostDisabled(t *testing.T) {
	g := NewTetris(false, 0)
	out := g.Render()
	if strings.Contains(out, "░░") {
		t.Error("ghost disabled: render should not contain '░░'")
	}
}

func TestTetrisRender_NilCurrent(t *testing.T) {
	g := NewTetris(false, 0)
	g.current = nil
	// Should not panic
	out := g.Render()
	if out == "" {
		t.Error("Render should return non-empty string even with nil current")
	}
}
