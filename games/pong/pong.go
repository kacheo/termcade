package pong

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

const (
	FieldWidth          = 40
	FieldHeight         = 20
	PaddleHeight        = 4
	WinScore            = 5
	PaddleMargin        = 0.05
	InitialBallSpeed    = 0.016
	PaddleSpeed         = 0.016
	PaddleInputWindow   = 280 * time.Millisecond
	SpeedIncreaseRate   = 1.1
	MaxSpeedMultiplier  = 2.0
	MaxPaddleHits       = 10
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
	AiDifficulty int // 0=Easy, 1=Medium, 2=Hard
	ballHitCount int

	playerMoveDir     float64
	playerInputExpiry time.Time
}

func NewPong(speedIncrease bool, aiDifficulty int) *Pong {
	if aiDifficulty < 0 {
		aiDifficulty = 0
	}
	if aiDifficulty > 2 {
		aiDifficulty = 2
	}
	p := &Pong{
		PlayerY:       0.5,
		AiY:          0.5,
		BallX:        0.5,
		BallY:        0.5,
		SpeedIncrease: speedIncrease,
		AiDifficulty:  aiDifficulty,
		ballHitCount: 0,
	}
	p.resetBall(1)
	return p
}

func (p *Pong) resetBall(direction int) {
	p.BallX = 0.5
	p.BallY = 0.5 + (rand.Float64()-0.5)*0.3
	p.BallVX = float64(direction) * InitialBallSpeed
	p.BallVY = (rand.Float64() - 0.5) * InitialBallSpeed
	p.ballHitCount = 0
}

func (p *Pong) paddleHalf() float64 {
	return float64(PaddleHeight) / float64(FieldHeight) / 2
}

func (p *Pong) clampPaddleY(y float64) float64 {
	half := p.paddleHalf()
	if y < half {
		return half
	}
	if y > 1-half {
		return 1 - half
	}
	return y
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x float64) float64 {
	if x < 0 {
		return -1
	}
	return 1
}

func (p *Pong) Update(delta time.Duration) error {
	if p.GameOver || p.Paused {
		return nil
	}

	if time.Now().Before(p.playerInputExpiry) {
		p.PlayerY += p.playerMoveDir * PaddleSpeed
		p.PlayerY = p.clampPaddleY(p.PlayerY)
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

	if p.BallVX < 0 {
		t := (PaddleMargin - p.BallX) / p.BallVX
		if t >= 0 && t <= 1 {
			crossY := p.BallY + p.BallVY*t
			if crossY >= p.PlayerY-PaddleMargin && crossY <= p.PlayerY+PaddleMargin {
				p.BallX = PaddleMargin
				p.BallY = crossY
				p.BallVX = -p.BallVX
				p.BallVY = (crossY - p.PlayerY) * 0.5
				if p.SpeedIncrease && p.ballHitCount < MaxPaddleHits {
					maxSpeed := InitialBallSpeed * MaxSpeedMultiplier
					p.BallVX *= SpeedIncreaseRate
					p.BallVY *= SpeedIncreaseRate
					if abs(p.BallVX) > maxSpeed {
						p.BallVX = sign(p.BallVX) * maxSpeed
					}
					if abs(p.BallVY) > maxSpeed {
						p.BallVY = sign(p.BallVY) * maxSpeed
					}
					p.ballHitCount++
				}
			}
		}
	}

	if p.BallVX > 0 {
		t := (1 - PaddleMargin - p.BallX) / p.BallVX
		if t >= 0 && t <= 1 {
			crossY := p.BallY + p.BallVY*t
			if crossY >= p.AiY-PaddleMargin && crossY <= p.AiY+PaddleMargin {
				p.BallX = 1 - PaddleMargin
				p.BallY = crossY
				p.BallVX = -p.BallVX
				p.BallVY = (crossY - p.AiY) * 0.5
			}
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

	p.AiY = p.clampPaddleY(p.AiY)
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
		p.playerMoveDir = -1.0
		p.playerInputExpiry = time.Now().Add(PaddleInputWindow)
	case "down", "j":
		p.playerMoveDir = 1.0
		p.playerInputExpiry = time.Now().Add(PaddleInputWindow)
	case "p":
		p.Paused = !p.Paused
	case "q":
		p.GameOver = true
		p.Winner = "AI"
	}
}

func (p *Pong) Render() string {
	var sb strings.Builder
	sb.WriteString("\n")

	sb.WriteString("  ╔════════════════════════════════════════╗\n")
	sb.WriteString("  ║           PONG                          ║\n")
	sb.WriteString("  ╠════════════════════════════════════════╣\n")

	for y := 0; y < FieldHeight; y++ {
		rowY := float64(y) / float64(FieldHeight)
		sb.WriteString("  ║")

		for x := 0; x < FieldWidth; x++ {
			char := " "

			if x == 2 {
				paddleTop := p.PlayerY - p.paddleHalf()
				paddleBottom := p.PlayerY + p.paddleHalf()
				if rowY >= paddleTop && rowY <= paddleBottom {
					char = "█"
				}
			}

			if x == FieldWidth-3 {
				paddleTop := p.AiY - p.paddleHalf()
				paddleBottom := p.AiY + p.paddleHalf()
				if rowY >= paddleTop && rowY <= paddleBottom {
					char = "█"
				}
			}

			ballX := int(math.Round(p.BallX * float64(FieldWidth)))
			ballY := int(math.Round(p.BallY * float64(FieldHeight)))
			if ballX == x && ballY == y {
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
