package blackjack

import (
	"math/rand"

	cardpkg "github.com/kacheo/termcade/games/cards"
)

// shoePenetration is the fraction of the shoe dealt before a reshuffle is
// queued. 75% matches typical cut-card placement in real 6-deck shoe games:
// deep enough that true-count deviations are worth practicing, shallow
// enough to leave a safety margin against exhausting the shoe mid-hand.
const shoePenetration = 0.75

// Shoe is a persistent, multi-deck source of cards that is dealt down across
// many hands rather than reshuffled every round, plus a Hi-Lo running/true
// count so counting practice is actually meaningful. It intentionally lives
// in games/blackjack rather than the shared games/cards package: penetration,
// cut cards, and Hi-Lo counting are blackjack training concepts, not generic
// card-deck primitives.
type Shoe struct {
	rng              *rand.Rand
	cards            cardpkg.Deck
	numDecks         int
	totalCards       int
	reshufflePending bool
	runningCount     int
}

// NewShoe builds a shuffled shoe of numDecks decks (clamped to [1,8]).
func NewShoe(numDecks int, rng *rand.Rand) *Shoe {
	if numDecks < 1 {
		numDecks = 1
	}
	if numDecks > 8 {
		numDecks = 8
	}
	s := &Shoe{rng: rng, numDecks: numDecks, totalCards: numDecks * 52}
	s.Reshuffle()
	return s
}

// Reshuffle rebuilds the shoe from numDecks fresh decks, shuffles, and resets
// the running count and pending-reshuffle flag.
func (s *Shoe) Reshuffle() {
	full := make(cardpkg.Deck, 0, s.totalCards)
	for i := 0; i < s.numDecks; i++ {
		full = append(full, cardpkg.NewDeck()...)
	}
	s.cards = full.Shuffled(s.rng)
	s.runningCount = 0
	s.reshufflePending = false
}

// Draw pops the next card from the shoe. It never reshuffles mid-hand on its
// own — callers only reshuffle between hands via NeedsReshuffle/Reshuffle.
// The emergency reshuffle here is a defensive fallback that should never
// trigger given the deck-count range and split limits this game enforces;
// it exists only so a pathological run of splits/hits can't panic.
func (s *Shoe) Draw() cardpkg.Card {
	if len(s.cards) == 0 {
		s.Reshuffle()
	}
	return s.cards.Draw()
}

// CardsRemaining returns the number of undealt cards left in the shoe.
func (s *Shoe) CardsRemaining() int { return len(s.cards) }

// DecksRemaining estimates decks left, floored to avoid a near-zero
// denominator blowing up TrueCount late in the shoe.
func (s *Shoe) DecksRemaining() float64 {
	d := float64(len(s.cards)) / 52.0
	if d < 0.25 {
		d = 0.25
	}
	return d
}

// CountCard folds a card into the Hi-Lo running count. Call this only when
// the card becomes visible to the player (dealt player/up-cards immediately,
// but the dealer's hole card only once it's actually revealed) — counting it
// the instant it's physically drawn would make the on-screen count diverge
// from what a real counter watching the table would compute.
func (s *Shoe) CountCard(c cardpkg.Card) {
	s.runningCount += hiLoValue(c.Rank)
	s.checkPenetration()
}

// checkPenetration flags the shoe for reshuffle once enough of it has been
// physically dealt (CardsRemaining, not the Hi-Lo count). It never
// reshuffles immediately — startRound() checks the flag between hands so a
// reshuffle never happens mid-round.
func (s *Shoe) checkPenetration() {
	if float64(s.CardsRemaining()) <= float64(s.totalCards)*(1-shoePenetration) {
		s.reshufflePending = true
	}
}

// NeedsReshuffle reports whether the shoe has been dealt past its
// penetration threshold and should be reshuffled before the next hand.
func (s *Shoe) NeedsReshuffle() bool { return s.reshufflePending }

// RunningCount is the current Hi-Lo running count of all counted cards.
func (s *Shoe) RunningCount() int { return s.runningCount }

// TrueCount is the running count normalized by decks remaining.
func (s *Shoe) TrueCount() float64 {
	return float64(s.runningCount) / s.DecksRemaining()
}

// hiLoValue is the standard Hi-Lo tag for a rank: 2-6 = +1, 7-9 = 0,
// 10/face/Ace = -1.
func hiLoValue(r cardpkg.Rank) int {
	switch {
	case r == cardpkg.Ace || r >= cardpkg.Ten:
		return -1
	case r >= cardpkg.Two && r <= cardpkg.Six:
		return 1
	default: // Seven, Eight, Nine
		return 0
	}
}
