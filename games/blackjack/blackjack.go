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

func (b *Blackjack) Update(delta time.Duration) error {
	if b.paused || b.gameOver {
		return nil
	}
	b.elapsed += delta
	switch b.phase {
	case phaseDealing:
		if b.elapsed >= dealDelay {
			b.elapsed = 0
			b.transitionFromDealing()
		}
	case phaseAITurn:
		if b.elapsed >= aiStepDelay {
			b.elapsed = 0
			b.stepAI()
		}
	case phaseDealerTurn:
		if b.elapsed >= dealerDelay {
			b.elapsed = 0
			b.stepDealer()
		}
	}
	return nil
}

func (b *Blackjack) transitionFromDealing() {
	if len(b.players) > 1 {
		b.aiIdx = 1
		b.phase = phaseAITurn
		b.skipDoneAIs()
	} else {
		b.phase = phasePlayerTurn
	}
}

func (b *Blackjack) skipDoneAIs() {
	for b.aiIdx < len(b.players) && b.players[b.aiIdx].status != statusPlaying {
		b.aiIdx++
	}
	if b.aiIdx >= len(b.players) {
		if b.players[0].status == statusPlaying {
			b.phase = phasePlayerTurn
		} else {
			b.phase = phaseDealerTurn
			b.elapsed = 0
		}
	}
}

func (b *Blackjack) stepAI() {
	if b.aiIdx >= len(b.players) {
		return
	}
	ai := b.players[b.aiIdx]
	if ShouldHit(ai.hand) {
		ai.hand = append(ai.hand, b.deck.Draw())
		if ai.hand.IsBust() {
			ai.status = statusBust
			b.aiIdx++
			b.skipDoneAIs()
			return
		}
		// Non-bust hit: AI gets another tick to decide again
	} else {
		ai.status = statusStand
		b.aiIdx++
		b.skipDoneAIs()
	}
}

func (b *Blackjack) stepDealer() {
	dv := b.dealer.Value()
	if dv < 17 || (dv == 17 && b.dealer.IsSoft()) {
		b.dealer = append(b.dealer, b.deck.Draw())
	} else {
		b.evaluateResults()
		b.phase = phaseResults
	}
}

func (b *Blackjack) evaluateResults() {
	dv := b.dealer.Value()
	dealerBust := b.dealer.IsBust()
	dealerBJ := b.dealer.IsBlackjack()
	for _, p := range b.players {
		switch p.status {
		case statusBust:
			p.result = "LOSE"
		case statusBlackjack:
			if dealerBJ {
				p.result = "PUSH"
			} else {
				p.result = "WIN"
				if !p.isAI {
					b.wins++
				}
			}
		default:
			pv := p.hand.Value()
			if dealerBust || pv > dv {
				p.result = "WIN"
				if !p.isAI {
					b.wins++
				}
			} else if pv == dv {
				p.result = "PUSH"
			} else {
				p.result = "LOSE"
			}
		}
	}
}

func (b *Blackjack) HandleInput(key string) {
	switch b.phase {
	case phasePlayerTurn:
		human := b.players[0]
		if human.status != statusPlaying {
			return
		}
		switch key {
		case "h", "left":
			human.hand = append(human.hand, b.deck.Draw())
			if human.hand.IsBust() {
				human.status = statusBust
				b.phase = phaseDealerTurn
				b.elapsed = 0
			}
		case "s", "right", "down":
			human.status = statusStand
			b.phase = phaseDealerTurn
			b.elapsed = 0
		}
	case phaseResults:
		if key == "enter" || key == " " {
			b.startRound()
		}
	}
}
func (b *Blackjack) Render() string         { return "BLACKJACK\nloading..." }
