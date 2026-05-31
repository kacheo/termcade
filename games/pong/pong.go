package pong

import (
	"math/rand"
	"time"
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

func (p *Pong) Update(delta time.Duration) error {
	if p.gameOver || p.paused {
		return nil
	}

	p.ballX += p.ballVX
	p.ballY += p.ballVY

	if p.ballY <= 0 {
		p.ballY = 0
		p.ballVY = -p.ballVY
	}
	if p.ballY >= 1 {
		p.ballY = 1
		p.ballVY = -p.ballVY
	}

	if p.ballX <= 0.05 && p.ballVX < 0 {
		if p.ballY >= p.playerY-0.05 && p.ballY <= p.playerY+0.05 {
			p.ballVX = -p.ballVX
			p.ballY += (p.ballY - p.playerY) * 0.5
			if p.speedIncrease {
				p.ballVX *= 1.1
				p.ballVY *= 1.1
			}
		}
	}

	if p.ballX >= 0.95 && p.ballVX > 0 {
		if p.ballY >= p.aiY-0.05 && p.ballY <= p.aiY+0.05 {
			p.ballVX = -p.ballVX
			p.ballY += (p.ballY - p.aiY) * 0.5
		}
	}

	if p.ballX < 0 {
		p.aiScore++
		if p.aiScore >= WinScore {
			p.gameOver = true
			p.winner = "AI"
		} else {
			p.resetBall(1)
		}
	}
	if p.ballX > 1 {
		p.playerScore++
		if p.playerScore >= WinScore {
			p.gameOver = true
			p.winner = "Player"
		} else {
			p.resetBall(-1)
		}
	}

	p.updateAI()

	return nil
}

func (p *Pong) updateAI() {
	if p.gameOver || p.paused {
		return
	}

	var reactionSpeed float64
	var accuracy float64

	switch p.aiDifficulty {
	case 0:
		reactionSpeed = 0.01
		accuracy = 0.6
	case 1:
		reactionSpeed = 0.02
		accuracy = 0.8
	case 2:
		reactionSpeed = 0.04
		accuracy = 0.95
	}

	targetY := p.ballY
	if p.ballVX < 0 {
		targetY = 0.5
	}

	diff := targetY - p.aiY
	if diff > reactionSpeed {
		p.aiY += reactionSpeed
	} else if diff < -reactionSpeed {
		p.aiY -= reactionSpeed
	}

	if rand.Float64() > accuracy {
		p.aiY += (rand.Float64() - 0.5) * 0.02
	}

	halfPaddle := float64(PaddleHeight) / float64(FieldHeight) / 2
	if p.aiY < halfPaddle {
		p.aiY = halfPaddle
	}
	if p.aiY > 1-halfPaddle {
		p.aiY = 1 - halfPaddle
	}
}

func (p *Pong) Name() string        { return "Pong" }
func (p *Pong) Description() string  { return "Classic paddle game" }
func (p *Pong) IsPaused() bool       { return p.paused }
func (p *Pong) IsGameOver() bool     { return p.gameOver }
func (p *Pong) GetScore() int        { return p.playerScore }
func (p *Pong) GetLevel() int       { return p.aiDifficulty }
func (p *Pong) GetLines() int        { return p.aiScore }