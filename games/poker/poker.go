package poker

import (
	"time"
)

type Poker struct{}

func NewPoker(seats, difficulty int) *Poker {
	return &Poker{}
}

func (p *Poker) Update(delta time.Duration) error {
	return nil
}

func (p *Poker) Render() string {
	return "Poker"
}

func (p *Poker) HandleInput(key string) {}

func (p *Poker) Name() string {
	return "Poker"
}

func (p *Poker) Description() string {
	return "Texas Hold'em Poker"
}

func (p *Poker) IsPaused() bool {
	return false
}

func (p *Poker) IsGameOver() bool {
	return false
}

func (p *Poker) GetScore() int {
	return 0
}

func (p *Poker) GetLevel() int {
	return 0
}

func (p *Poker) GetLines() int {
	return 0
}