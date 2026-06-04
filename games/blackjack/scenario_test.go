package blackjack

import (
	"testing"
	"time"

	cardpkg "tmvgs/games/cards"
)

// makeDeck builds a Deck from the given cards; the first card listed is drawn first.
func makeDeck(cards ...cardpkg.Card) cardpkg.Deck {
	return cardpkg.Deck(cards)
}

// advancePastDealerTurn advances time enough for the dealer to finish all steps.
// Each dealer action requires dealerDelay (700ms). We call Update several times
// to ensure the dealer finishes regardless of how many hits they need.
func advancePastDealerTurn(b *Blackjack) {
	for i := 0; i < 10; i++ {
		if b.phase == phaseResults {
			break
		}
		b.Update(time.Second) //nolint:errcheck
	}
}

// TestScenarioBlackjackHitToBust controls the deck so that hitting results in a bust,
// verifies player status becomes statusBust and phase moves to phaseDealerTurn.
func TestScenarioBlackjackHitToBust(t *testing.T) {
	b := NewBlackjack()

	// Give the player a hand of 19 (Ten + Nine) and set up the deck so that hitting
	// draws a Five (19 + 5 = 24, bust).
	// Dealer hand is set to a safe 17 so it doesn't interfere.
	b.player.hand = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades},
		cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Hearts},
	}
	b.player.status = statusPlaying
	b.dealer = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Diamonds},
	}
	// Put the bust card at the top of the deck.
	b.deck = makeDeck(
		cardpkg.Card{Rank: cardpkg.Five, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Spades}, // dealer spare
	)

	// Advance past deal animation so player input is accepted.
	b.phase = phaseTurn

	b.HandleInput("h") // draw the Five → 24, bust

	if b.player.status != statusBust {
		t.Errorf("player.status = %v, want statusBust", b.player.status)
	}
	if b.phase != phaseDealerTurn {
		t.Errorf("phase = %v, want phaseDealerTurn", b.phase)
	}
}

// TestScenarioBlackjackStand verifies that standing advances the game through the
// dealer's turn and lands in phaseResults without getting stuck.
func TestScenarioBlackjackStand(t *testing.T) {
	b := NewBlackjack()

	// Player holds 18; dealer holds 17 (hard) — dealer stands immediately.
	b.player.hand = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades},
		cardpkg.Card{Rank: cardpkg.Eight, Suit: cardpkg.Hearts},
	}
	b.player.status = statusPlaying
	b.dealer = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Diamonds},
	}
	// Spare cards in case dealer needs to draw (it won't here, but keep deck non-empty).
	b.deck = makeDeck(
		cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Spades},
		cardpkg.Card{Rank: cardpkg.Three, Suit: cardpkg.Clubs},
	)
	b.phase = phaseTurn

	b.HandleInput("s") // stand → phaseDealerTurn

	if b.phase != phaseDealerTurn {
		t.Errorf("after stand: phase = %v, want phaseDealerTurn", b.phase)
	}
	if b.player.status != statusStand {
		t.Errorf("player.status = %v, want statusStand", b.player.status)
	}

	// Advance time so the dealer finishes its turn.
	advancePastDealerTurn(b)

	if b.phase != phaseResults {
		t.Errorf("after dealer finishes: phase = %v, want phaseResults", b.phase)
	}
}

// TestScenarioBlackjackNewRound verifies that pressing "enter" in phaseResults
// increments b.rounds and returns the phase to phaseDealing.
func TestScenarioBlackjackNewRound(t *testing.T) {
	b := NewBlackjack()

	// Force the game straight to phaseResults.
	b.phase = phaseResults
	b.player.result = "WIN"
	roundsBefore := b.rounds

	b.HandleInput("enter")

	if b.rounds != roundsBefore+1 {
		t.Errorf("rounds = %d, want %d (incremented)", b.rounds, roundsBefore+1)
	}
	if b.phase != phaseDealing {
		t.Errorf("phase = %v, want phaseDealing after new round", b.phase)
	}
}

// TestScenarioBlackjackNaturalBlackjack directly assigns a natural blackjack hand
// (Ace + King) and verifies that statusBlackjack is set and the dealing phase
// transitions to phaseDealerTurn (skipping phaseTurn) when the timer fires.
func TestScenarioBlackjackNaturalBlackjack(t *testing.T) {
	b := NewBlackjack()

	// Directly set a natural blackjack hand (21 in two cards).
	b.player.hand = Hand{
		cardpkg.Card{Rank: cardpkg.Ace, Suit: cardpkg.Spades},
		cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Hearts},
	}
	b.player.status = statusBlackjack
	b.dealer = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Diamonds},
	}
	b.deck = makeDeck(
		cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Spades},
	)
	b.phase = phaseDealing
	b.elapsed = 0

	if !b.player.hand.IsBlackjack() {
		t.Fatalf("test setup: hand %v is not a natural blackjack", b.player.hand)
	}
	if b.player.status != statusBlackjack {
		t.Errorf("player.status = %v, want statusBlackjack", b.player.status)
	}

	// Advance past the deal animation; because status is statusBlackjack the game
	// should skip phaseTurn and go directly to phaseDealerTurn.
	b.Update(time.Second) //nolint:errcheck

	if b.phase != phaseDealerTurn {
		t.Errorf("phase after deal animation = %v, want phaseDealerTurn (blackjack skips player turn)", b.phase)
	}
}

// TestScenarioBlackjackWinCount plays two complete rounds where the player wins
// each time (player 20 vs dealer bust) and verifies b.wins increments correctly.
func TestScenarioBlackjackWinCount(t *testing.T) {
	b := NewBlackjack()

	for round := 1; round <= 2; round++ {
		wantWins := round - 1 // wins so far before this round

		// Set up the player with 20 and the dealer with a hand that will bust:
		// dealer has Ten+Six (16), and the next card in the deck is a King (10) → 26, bust.
		b.player.hand = Hand{
			cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Spades},
			cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Hearts},
		}
		b.player.status = statusPlaying
		b.dealer = Hand{
			cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Clubs},
			cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Diamonds},
		}
		// Dealer will draw one more card (16 < 17), so put a busting card on top.
		b.deck = makeDeck(
			cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Clubs},  // dealer draws → 26, bust
			cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Spades},  // spare
			cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Hearts},  // spare for next round dealing
			cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Clubs},   // spare
			cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Diamonds},// spare
		)
		b.phase = phaseTurn
		b.player.result = ""

		// Stand so dealer gets to play.
		b.HandleInput("s")

		if b.phase != phaseDealerTurn {
			t.Fatalf("round %d: expected phaseDealerTurn after stand, got %v", round, b.phase)
		}

		// Advance past dealer turn.
		advancePastDealerTurn(b)

		if b.phase != phaseResults {
			t.Fatalf("round %d: expected phaseResults, got %v", round, b.phase)
		}
		if b.wins != wantWins+1 {
			t.Errorf("round %d: wins = %d, want %d", round, b.wins, wantWins+1)
		}

		// Start next round (only needed if we are going to play again).
		if round < 2 {
			b.HandleInput("enter")
			// Reset the game state for the next iteration (startRound was called,
			// override what it dealt with our controlled scenario).
			b.phase = phaseTurn // skip deal animation
		}
	}

	if b.wins != 2 {
		t.Errorf("final wins = %d, want 2", b.wins)
	}
}
