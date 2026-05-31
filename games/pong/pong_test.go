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
	if p.PlayerY >= initialY {
		t.Error("PlayerY should have decreased after up key")
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
