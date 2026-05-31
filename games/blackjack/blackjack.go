package blackjack

import (
	"math/rand"
	"time"
)

type phase int

const (
	phaseDealing    phase = iota
	phaseAITurn
	phasePlayerTurn
	phaseDealerTurn
	phaseResults
)

type playerStatus int

const (
	statusPlaying   playerStatus = iota
	statusStand
	statusBust
	statusBlackjack
)

type tablePlayer struct {
	name   string
	hand   Hand
	status playerStatus
	isAI   bool
	result string // "WIN", "LOSE", "PUSH", or ""
}

type Blackjack struct {
	rng      *rand.Rand
	deck     Deck
	dealer   Hand
	players  []*tablePlayer // [0]=human, [1..]=AI
	phase    phase
	aiIdx    int
	elapsed  time.Duration
	wins     int
	rounds   int
	paused   bool
	gameOver bool
}

const (
	dealDelay   = 600 * time.Millisecond
	aiStepDelay = 500 * time.Millisecond
	dealerDelay = 700 * time.Millisecond
)

func NewBlackjack(aiCount int) *Blackjack {
	if aiCount < 0 {
		aiCount = 0
	}
	if aiCount > 3 {
		aiCount = 3
	}
	b := &Blackjack{rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
	b.players = append(b.players, &tablePlayer{name: "YOU", isAI: false})
	for _, name := range []string{"AI-1", "AI-2", "AI-3"}[:aiCount] {
		b.players = append(b.players, &tablePlayer{name: name, isAI: true})
	}
	b.startRound()
	return b
}

func (b *Blackjack) startRound() {
	b.deck = NewDeck().Shuffled(b.rng)
	b.dealer = Hand{}
	for _, p := range b.players {
		p.hand = Hand{}
		p.status = statusPlaying
		p.result = ""
	}
	for i := 0; i < 2; i++ {
		for _, p := range b.players {
			p.hand = append(p.hand, b.deck.Draw())
		}
		b.dealer = append(b.dealer, b.deck.Draw())
	}
	for _, p := range b.players {
		if p.hand.IsBlackjack() {
			p.status = statusBlackjack
		}
	}
	b.phase = phaseDealing
	b.elapsed = 0
	b.aiIdx = 1
	b.rounds++
}

func (b *Blackjack) Name() string        { return "Blackjack" }
func (b *Blackjack) Description() string { return "Beat the dealer. Hit or stand." }
func (b *Blackjack) IsPaused() bool      { return b.paused }
func (b *Blackjack) IsGameOver() bool    { return b.gameOver }
func (b *Blackjack) GetScore() int       { return b.wins }
func (b *Blackjack) GetLevel() int       { return 0 }
func (b *Blackjack) GetLines() int       { return b.rounds }

func (b *Blackjack) Update(delta time.Duration) error { return nil }
func (b *Blackjack) HandleInput(key string)           {}
func (b *Blackjack) Render() string                   { return "BLACKJACK\nloading..." }
