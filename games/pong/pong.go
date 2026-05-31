package pong

import (
	"math/rand"
)

const (
	FieldWidth  = 40
	FieldHeight = 20
	PaddleHeight = 4
	WinScore     = 5
)

type Pong struct {
	playerY      float64
	aiY         float64
	ballX       float64
	ballY       float64
	ballVX      float64
	ballVY      float64
	playerScore int
	aiScore     int
	paused      bool
	gameOver    bool
	winner      string
	speedIncrease bool
	aiDifficulty  int // 0=Easy, 1=Medium, 2=Hard
}

func NewPong(speedIncrease bool, aiDifficulty int) *Pong {
	p := &Pong{
		playerY:        0.5,
		aiY:            0.5,
		ballX:          0.5,
		ballY:          0.5,
		speedIncrease:  speedIncrease,
		aiDifficulty:  aiDifficulty,
	}
	p.resetBall(1) // 1 = right direction
	return p
}

func (p *Pong) resetBall(direction int) {
	p.ballX = 0.5
	p.ballY = 0.5 + (rand.Float64() - 0.5) * 0.3
	speed := 0.02
	p.ballVX = float64(direction) * speed
	p.ballVY = (rand.Float64() - 0.5) * speed
}

func (p *Pong) Name() string        { return "Pong" }
func (p *Pong) Description() string  { return "Classic paddle game" }
func (p *Pong) IsPaused() bool       { return p.paused }
func (p *Pong) IsGameOver() bool     { return p.gameOver }
func (p *Pong) GetScore() int        { return p.playerScore }
func (p *Pong) GetLevel() int       { return p.aiDifficulty }
func (p *Pong) GetLines() int        { return p.aiScore }