package pong

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	FieldWidth  = 40
	FieldHeight = 20
	PaddleHeight = 4
	WinScore     = 5
)

type Pong struct {
	PlayerY      float64
	AiY          float64
	BallX        float64
	BallY        float64
	BallVX       float64
	BallVY       float64
	PlayerScore  int
	AiScore      int
	Paused       bool
	GameOver     bool
	Winner       string
	SpeedIncrease bool
	AiDifficulty  int // 0=Easy, 1=Medium, 2=Hard
}

func NewPong(speedIncrease bool, aiDifficulty int) *Pong {
	p := &Pong{
		PlayerY:        0.5,
		AiY:            0.5,
		BallX:          0.5,
		BallY:          0.5,
		SpeedIncrease:  speedIncrease,
		AiDifficulty:   aiDifficulty,
	}
	p.resetBall(1) // 1 = right direction
	return p
}

func (p *Pong) resetBall(direction int) {
	p.BallX = 0.5
	p.BallY = 0.5 + (rand.Float64() - 0.5) * 0.3
	speed := 0.02
	p.BallVX = float64(direction) * speed
	p.BallVY = (rand.Float64() - 0.5) * speed
}

func (p *Pong) Update(delta time.Duration) error {
	if p.GameOver || p.Paused {
		return nil
	}

	p.BallX += p.BallVX
	p.BallY += p.BallVY

	if p.BallY <= 0 {
		p.BallY = 0
		p.BallVY = -p.BallVY
	}
	if p.BallY >= 1 {
		p.BallY = 1
		p.BallVY = -p.BallVY
	}

	if p.BallX <= 0.05 && p.BallVX < 0 {
		if p.BallY >= p.PlayerY-0.05 && p.BallY <= p.PlayerY+0.05 {
			p.BallVX = -p.BallVX
			p.BallY += (p.BallY - p.PlayerY) * 0.5
			if p.SpeedIncrease {
				p.BallVX *= 1.1
				p.BallVY *= 1.1
			}
		}
	}

	if p.BallX >= 0.95 && p.BallVX > 0 {
		if p.BallY >= p.AiY-0.05 && p.BallY <= p.AiY+0.05 {
			p.BallVX = -p.BallVX
			p.BallY += (p.BallY - p.AiY) * 0.5
		}
	}

	if p.BallX < 0 {
		p.AiScore++
		if p.AiScore >= WinScore {
			p.GameOver = true
			p.Winner = "AI"
		} else {
			p.resetBall(1)
		}
	}
	if p.BallX > 1 {
		p.PlayerScore++
		if p.PlayerScore >= WinScore {
			p.GameOver = true
			p.Winner = "Player"
		} else {
			p.resetBall(-1)
		}
	}

	p.updateAI()

	return nil
}

func (p *Pong) updateAI() {
	if p.GameOver || p.Paused {
		return
	}

	var reactionSpeed float64
	var accuracy float64

	switch p.AiDifficulty {
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

	targetY := p.BallY
	if p.BallVX < 0 {
		targetY = 0.5
	}

	diff := targetY - p.AiY
	if diff > reactionSpeed {
		p.AiY += reactionSpeed
	} else if diff < -reactionSpeed {
		p.AiY -= reactionSpeed
	}

	if rand.Float64() > accuracy {
		p.AiY += (rand.Float64() - 0.5) * 0.02
	}

	halfPaddle := float64(PaddleHeight) / float64(FieldHeight) / 2
	if p.AiY < halfPaddle {
		p.AiY = halfPaddle
	}
	if p.AiY > 1-halfPaddle {
		p.AiY = 1 - halfPaddle
	}
}

func (p *Pong) Name() string        { return "Pong" }
func (p *Pong) Description() string  { return "Classic paddle game" }
func (p *Pong) IsPaused() bool       { return p.Paused }
func (p *Pong) IsGameOver() bool     { return p.GameOver }
func (p *Pong) GetScore() int        { return p.PlayerScore }
func (p *Pong) GetLevel() int       { return p.AiDifficulty }
func (p *Pong) GetLines() int        { return p.AiScore }

func (p *Pong) HandleInput(key string) {
	if p.GameOver {
		return
	}
	switch key {
	case "up", "k":
		p.PlayerY -= 0.05
	case "down", "j":
		p.PlayerY += 0.05
	case "p":
		p.Paused = !p.Paused
	case "q":
		p.GameOver = true
		p.Winner = "AI"
	}

	halfPaddle := float64(PaddleHeight) / float64(FieldHeight) / 2
	if p.PlayerY < halfPaddle {
		p.PlayerY = halfPaddle
	}
	if p.PlayerY > 1-halfPaddle {
		p.PlayerY = 1 - halfPaddle
	}
}

func (p *Pong) Render() string {
	var sb strings.Builder
	sb.WriteString("\n")

	sb.WriteString("  ╔════════════════════════════════════════╗\n")
	sb.WriteString("║           PONG                          ║\n")
	sb.WriteString("  ╠════════════════════════════════════════╣\n")

	for y := 0; y < FieldHeight; y++ {
		rowY := float64(y) / float64(FieldHeight)
		sb.WriteString("  ║")

		for x := 0; x < FieldWidth; x++ {
			char := " "

			if x == 2 {
				paddleTop := p.PlayerY - float64(PaddleHeight)/float64(FieldHeight)/2
				paddleBottom := p.PlayerY + float64(PaddleHeight)/float64(FieldHeight)/2
				if rowY >= paddleTop && rowY <= paddleBottom {
					char = "█"
				}
			}

			if x == FieldWidth-3 {
				paddleTop := p.AiY - float64(PaddleHeight)/float64(FieldHeight)/2
				paddleBottom := p.AiY + float64(PaddleHeight)/float64(FieldHeight)/2
				if rowY >= paddleTop && rowY <= paddleBottom {
					char = "█"
				}
			}

			ballX := int(p.BallX * float64(FieldWidth))
			if ballX == x && int(p.BallY*float64(FieldHeight)) == y {
				char = "●"
			}

			if x == FieldWidth/2 {
				char = "│"
			}

			sb.WriteString(char)
		}
		sb.WriteString("║\n")
	}

	sb.WriteString("  ╠════════════════════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("║  Player: %d          AI: %d              ║\n", p.PlayerScore, p.AiScore))
	sb.WriteString("  ╚════════════════════════════════════════╝\n")

	sb.WriteString("    [↑/↓] Move   [P] Pause   [Q] Quit\n")

	return sb.String()
}