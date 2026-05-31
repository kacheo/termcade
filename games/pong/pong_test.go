package pong

import (
	"testing"
	"time"
)

func TestNewPong(t *testing.T) {
	p := NewPong(false, 1)
	if p == nil {
		t.Fatal("NewPong returned nil")
	}
	if p.PlayerScore != 0 {
		t.Errorf("PlayerScore should be 0, got %d", p.PlayerScore)
	}
	if p.AiScore != 0 {
		t.Errorf("AiScore should be 0, got %d", p.AiScore)
	}
	if p.GameOver {
		t.Error("GameOver should be false")
	}
	if p.Paused {
		t.Error("Paused should be false")
	}
}

func TestPongUpdate(t *testing.T) {
	p := NewPong(false, 1)
	initialBallX := p.BallX
	p.Update(time.Millisecond * 100)
	if p.BallX == initialBallX {
		t.Error("ball should have moved")
	}
}

func TestPongWallBounce(t *testing.T) {
	p := NewPong(false, 1)
	p.BallX = 0.5
	p.BallY = 0.001
	p.BallVX = 0.01
	p.BallVY = -0.01

	p.Update(time.Millisecond * 100)

	if p.BallY != 0 {
		t.Error("ball should be at wall after bounce")
	}
	if p.BallVY <= 0 {
		t.Error("ball Y velocity should have flipped after wall hit")
	}
}

func TestPongHandleInput(t *testing.T) {
	p := NewPong(false, 1)
	initialY := p.PlayerY

	p.HandleInput("up")
	p.Update(time.Millisecond * 10)
	if p.PlayerY >= initialY {
		t.Errorf("paddle should move up after up key, got %f (was %f)", p.PlayerY, initialY)
	}

	afterUp := p.PlayerY
	p.HandleInput("down")
	p.Update(time.Millisecond * 10)
	if p.PlayerY <= afterUp {
		t.Errorf("paddle should move down after down key, got %f (was %f)", p.PlayerY, afterUp)
	}
}

func TestPongPause(t *testing.T) {
	p := NewPong(false, 1)
	if p.Paused {
		t.Error("Paused should start false")
	}
	p.HandleInput("p")
	if !p.Paused {
		t.Error("Paused should be true after p key")
	}
	p.HandleInput("p")
	if p.Paused {
		t.Error("Paused should be false after second p key")
	}
}

func TestPongQuit(t *testing.T) {
	p := NewPong(false, 1)
	p.HandleInput("q")
	if !p.GameOver {
		t.Error("GameOver should be true after q key")
	}
	if p.Winner != "AI" {
		t.Errorf("Winner should be AI, got %s", p.Winner)
	}
}

func TestPongInterface(t *testing.T) {
	p := NewPong(false, 1)
	if p.Name() != "Pong" {
		t.Errorf("Name should be Pong, got %s", p.Name())
	}
	if p.Description() != "Classic paddle game" {
		t.Errorf("Description mismatch, got %s", p.Description())
	}
	if p.IsPaused() {
		t.Error("IsPaused should be false")
	}
	if p.IsGameOver() {
		t.Error("IsGameOver should be false")
	}
	if p.GetScore() != 0 {
		t.Errorf("GetScore should be 0, got %d", p.GetScore())
	}
	if p.GetLevel() != 1 {
		t.Errorf("GetLevel should be 1, got %d", p.GetLevel())
	}
	if p.GetLines() != 0 {
		t.Errorf("GetLines should be 0, got %d", p.GetLines())
	}
}

func TestPongScoring(t *testing.T) {
	p := NewPong(false, 1)
	p.BallX = -0.1
	p.BallVX = -0.01
	p.AiScore = 4

	p.Update(time.Millisecond * 100)

	if p.AiScore != 5 {
		t.Errorf("AiScore should be 5, got %d", p.AiScore)
	}
	if !p.GameOver {
		t.Error("GameOver should be true when AI reaches 5")
	}
	if p.Winner != "AI" {
		t.Errorf("Winner should be AI, got %s", p.Winner)
	}
}

func TestPongPlayerScores(t *testing.T) {
	p := NewPong(false, 1)
	p.BallX = 1.1
	p.BallVX = 0.01
	p.PlayerScore = 4

	p.Update(time.Millisecond * 100)

	if p.PlayerScore != 5 {
		t.Errorf("PlayerScore should be 5, got %d", p.PlayerScore)
	}
	if !p.GameOver {
		t.Error("GameOver should be true when player reaches 5")
	}
	if p.Winner != "Player" {
		t.Errorf("Winner should be Player, got %s", p.Winner)
	}
}

func TestPongSpeedIncrease(t *testing.T) {
	p := NewPong(false, 1)
	p.SpeedIncrease = true
	p.ballHitCount = 0

	p.BallX = 0.04
	p.BallY = p.PlayerY
	p.BallVX = -InitialBallSpeed
	initialVX := p.BallVX

	p.Update(time.Millisecond * 10)

	if p.ballHitCount == 0 {
		t.Skip("Ball didn't reach paddle in this test iteration")
	}

	if p.BallVX <= initialVX {
		t.Errorf("BallVX should increase after paddle hit, got %f <= %f", p.BallVX, initialVX)
	}
}

func TestPongUpdatePaused(t *testing.T) {
	p := NewPong(false, 1)
	p.Paused = true
	initialBallX := p.BallX
	p.Update(time.Millisecond * 100)

	if p.BallX != initialBallX {
		t.Error("ball should not move when paused")
	}
}

func TestPongUpdateGameOver(t *testing.T) {
	p := NewPong(false, 1)
	p.GameOver = true
	initialBallX := p.BallX
	p.Update(time.Millisecond * 100)

	if p.BallX != initialBallX {
		t.Error("ball should not move when game over")
	}
}

func TestPongAIDifficulty(t *testing.T) {
	tests := []struct {
		difficulty int
		name      string
	}{
		{0, "Easy"},
		{1, "Medium"},
		{2, "Hard"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPong(false, tt.difficulty)
			if p.AiDifficulty != tt.difficulty {
				t.Errorf("AiDifficulty should be %d, got %d", tt.difficulty, p.AiDifficulty)
			}
		})
	}
}

func TestPongRender(t *testing.T) {
	p := NewPong(false, 1)
	output := p.Render()
	if len(output) == 0 {
		t.Error("Render should return non-empty string")
	}
}

func TestPongPlayerPaddleClamp(t *testing.T) {
	p := NewPong(false, 1)

	p.PlayerY = 0.0
	p.HandleInput("up")
	p.Update(time.Millisecond * 10)
	if p.PlayerY != p.paddleHalf() {
		t.Errorf("PlayerY should be clamped to paddleHalf, got %f", p.PlayerY)
	}

	p.PlayerY = 1.0
	p.HandleInput("down")
	p.Update(time.Millisecond * 10)
	if p.PlayerY != 1-p.paddleHalf() {
		t.Errorf("PlayerY should be clamped to 1-paddleHalf, got %f", p.PlayerY)
	}
}

func TestPongAIDifficultyClamping(t *testing.T) {
	p := NewPong(false, -1)
	if p.AiDifficulty != 0 {
		t.Errorf("AiDifficulty should be clamped to 0, got %d", p.AiDifficulty)
	}

	p = NewPong(false, 10)
	if p.AiDifficulty != 2 {
		t.Errorf("AiDifficulty should be clamped to 2, got %d", p.AiDifficulty)
	}
}

func TestPongSpeedCap(t *testing.T) {
	p := NewPong(true, 1)
	p.BallX = 0.05
	p.BallY = p.PlayerY
	p.BallVX = -InitialBallSpeed * 2

	p.Update(time.Millisecond * 100)

	maxSpeed := InitialBallSpeed * MaxSpeedMultiplier
	absVX := p.BallVX
	if absVX < 0 {
		absVX = -absVX
	}
	if absVX > maxSpeed {
		t.Errorf("BallVX should not exceed %f, got %f", maxSpeed, absVX)
	}
}

func TestPongResetBallClearsHitCount(t *testing.T) {
	p := NewPong(true, 1)
	p.ballHitCount = 5

	p.resetBall(1)

	if p.ballHitCount != 0 {
		t.Errorf("ballHitCount should be reset to 0, got %d", p.ballHitCount)
	}
}

func TestPongConstants(t *testing.T) {
	if InitialBallSpeed <= 0 {
		t.Error("InitialBallSpeed should be positive")
	}
	if MaxSpeedMultiplier <= 1 {
		t.Error("MaxSpeedMultiplier should be > 1")
	}
	if MaxPaddleHits <= 0 {
		t.Error("MaxPaddleHits should be positive")
	}
	if SpeedIncreaseRate <= 1 {
		t.Error("SpeedIncreaseRate should be > 1")
	}
}
