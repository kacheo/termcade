package blackjack

import (
	"testing"
	"time"

	cardpkg "github.com/kacheo/termcade/games/cards"
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
	b := NewBlackjack(6)
	b.HandleInput("enter") // place bet, deal

	// Give the player a hand of 19 (Ten + Nine) and set up the shoe so that hitting
	// draws a Five (19 + 5 = 24, bust).
	// Dealer hand is set to a safe 17 so it doesn't interfere.
	b.hands[0].hand = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades},
		cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Hearts},
	}
	b.hands[0].status = statusPlaying
	b.dealer = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Diamonds},
	}
	// Put the bust card at the top of the shoe.
	b.shoe.cards = makeDeck(
		cardpkg.Card{Rank: cardpkg.Five, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Spades}, // dealer spare
	)

	// Advance past deal animation so player input is accepted.
	b.phase = phaseTurn

	b.HandleInput("h") // draw the Five → 24, bust

	if b.hands[0].status != statusBust {
		t.Errorf("hands[0].status = %v, want statusBust", b.hands[0].status)
	}
	if b.phase != phaseDealerTurn {
		t.Errorf("phase = %v, want phaseDealerTurn", b.phase)
	}
}

// TestScenarioBlackjackStand verifies that standing advances the game through the
// dealer's turn and lands in phaseResults without getting stuck.
func TestScenarioBlackjackStand(t *testing.T) {
	b := NewBlackjack(6)
	b.HandleInput("enter")

	// Player holds 18; dealer holds 17 (hard) — dealer stands immediately.
	b.hands[0].hand = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades},
		cardpkg.Card{Rank: cardpkg.Eight, Suit: cardpkg.Hearts},
	}
	b.hands[0].status = statusPlaying
	b.dealer = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Diamonds},
	}
	// Spare cards in case dealer needs to draw (it won't here, but keep shoe non-empty).
	b.shoe.cards = makeDeck(
		cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Spades},
		cardpkg.Card{Rank: cardpkg.Three, Suit: cardpkg.Clubs},
	)
	b.phase = phaseTurn

	b.HandleInput("s") // stand → phaseDealerTurn

	if b.phase != phaseDealerTurn {
		t.Errorf("after stand: phase = %v, want phaseDealerTurn", b.phase)
	}
	if b.hands[0].status != statusStand {
		t.Errorf("hands[0].status = %v, want statusStand", b.hands[0].status)
	}

	// Advance time so the dealer finishes its turn.
	advancePastDealerTurn(b)

	if b.phase != phaseResults {
		t.Errorf("after dealer finishes: phase = %v, want phaseResults", b.phase)
	}
}

// TestScenarioBlackjackNewRound verifies that pressing "enter" in phaseResults
// returns to phaseBetting, and that a subsequent "enter" there deals and
// increments b.rounds.
func TestScenarioBlackjackNewRound(t *testing.T) {
	b := NewBlackjack(6)

	// Force the game straight to phaseResults.
	b.phase = phaseResults
	b.hands = []playerHand{{result: "WIN"}}
	roundsBefore := b.rounds

	b.HandleInput("enter")
	if b.phase != phaseBetting {
		t.Errorf("phase = %v, want phaseBetting after results Enter", b.phase)
	}

	b.HandleInput("enter")
	if b.rounds != roundsBefore+1 {
		t.Errorf("rounds = %d, want %d (incremented)", b.rounds, roundsBefore+1)
	}
	if b.phase != phaseDealing {
		t.Errorf("phase = %v, want phaseDealing after new round dealt", b.phase)
	}
}

// TestScenarioBlackjackNaturalBlackjack directly assigns a natural blackjack hand
// (Ace + King) and verifies that statusBlackjack is set and the dealing phase
// transitions to phaseDealerTurn (skipping phaseTurn) when the timer fires.
func TestScenarioBlackjackNaturalBlackjack(t *testing.T) {
	b := NewBlackjack(6)
	b.HandleInput("enter")

	// Directly set a natural blackjack hand (21 in two cards).
	b.hands[0].hand = Hand{
		cardpkg.Card{Rank: cardpkg.Ace, Suit: cardpkg.Spades},
		cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Hearts},
	}
	b.hands[0].status = statusBlackjack
	b.dealer = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Diamonds},
	}
	b.shoe.cards = makeDeck(
		cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Spades},
	)
	b.phase = phaseDealing
	b.elapsed = 0

	if !b.hands[0].hand.IsBlackjack() {
		t.Fatalf("test setup: hand %v is not a natural blackjack", b.hands[0].hand)
	}
	if b.hands[0].status != statusBlackjack {
		t.Errorf("hands[0].status = %v, want statusBlackjack", b.hands[0].status)
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
	b := NewBlackjack(6)

	for round := 1; round <= 2; round++ {
		wantWins := round - 1 // wins so far before this round

		if b.phase == phaseBetting {
			b.HandleInput("enter")
		}

		// Set up the player with 20 and the dealer with a hand that will bust:
		// dealer has Ten+Six (16), and the next card in the shoe is a King (10) → 26, bust.
		b.hands[0].hand = Hand{
			cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Spades},
			cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Hearts},
		}
		b.hands[0].status = statusPlaying
		b.dealer = Hand{
			cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Clubs},
			cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Diamonds},
		}
		// Dealer will draw one more card (16 < 17), so put a busting card on top.
		b.shoe.cards = makeDeck(
			cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Clubs},   // dealer draws → 26, bust
			cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Spades},   // spare
			cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Hearts},   // spare for next round dealing
			cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Clubs},    // spare
			cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Diamonds}, // spare
		)
		b.phase = phaseTurn
		b.hands[0].result = ""

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
			b.HandleInput("enter") // results -> betting
			b.HandleInput("enter") // betting -> deal
			b.phase = phaseTurn    // skip deal animation for the controlled scenario
		}
	}

	if b.wins != 2 {
		t.Errorf("final wins = %d, want 2", b.wins)
	}
}

// TestScenarioReshuffleNeverHappensMidHand is the single most important
// counting-integrity regression: once the shoe is flagged for reshuffle
// mid-round, it must not actually reshuffle until the *next* startRound(),
// never in the middle of the current hand (which would leak information and
// defeat counting).
func TestScenarioReshuffleNeverHappensMidHand(t *testing.T) {
	b := NewBlackjack(1) // single deck shoe, easiest to drive to penetration
	b.HandleInput("enter")

	b.hands[0].hand = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades},
		cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts},
	}
	b.hands[0].status = statusPlaying
	b.dealer = Hand{
		cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Diamonds},
	}
	// Force the shoe down near the bottom, and flag it as already pending a
	// reshuffle (as if penetration had just been crossed).
	b.shoe.cards = makeDeck(
		cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Spades},
		cardpkg.Card{Rank: cardpkg.Three, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Four, Suit: cardpkg.Hearts},
	)
	b.shoe.reshufflePending = true
	cardsBefore := b.shoe.CardsRemaining()

	b.phase = phaseTurn
	b.HandleInput("h") // draw mid-hand

	if b.shoe.CardsRemaining() >= cardsBefore {
		t.Fatalf("expected a card to be drawn mid-hand (remaining should decrease from %d)", cardsBefore)
	}
	if b.shoe.CardsRemaining() == 52 {
		t.Error("shoe reshuffled mid-hand — this must only happen between hands")
	}
	if !b.shoe.NeedsReshuffle() {
		t.Error("reshufflePending should still be true mid-hand, deferred to next startRound()")
	}

	// Now finish the hand and start the next round — the reshuffle should
	// happen there, and only there.
	b.HandleInput("s")
	advancePastDealerTurn(b)
	b.HandleInput("enter") // results -> betting
	b.HandleInput("enter") // betting -> deal, reshuffle happens here

	if b.shoe.NeedsReshuffle() {
		t.Error("shoe should have reshuffled (and cleared reshufflePending) at the start of the new round")
	}
}

// TestScenarioHoleCardNotCountedUntilRevealed proves the running count does
// not reflect the dealer's hole card until it's actually revealed (dealer's
// turn) — otherwise the on-screen count would diverge from what a real
// counter watching the table computes.
func TestScenarioHoleCardNotCountedUntilRevealed(t *testing.T) {
	b := NewBlackjack(6)
	b.HandleInput("enter")

	b.hands[0].hand = Hand{
		cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Spades},
		cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Hearts},
	}
	b.hands[0].status = statusPlaying
	b.dealer = Hand{
		cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Clubs},     // up-card, neutral count value
		cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Diamonds}, // hole card, would count -1 if visible
	}
	b.shoe.runningCount = 0
	b.shoe.CountCard(b.dealer[0]) // simulate what startRound would have done for the up-card
	countBeforeReveal := b.shoe.RunningCount()

	b.phase = phaseTurn
	b.dealerHoleCounted = false
	b.shoe.cards = makeDeck(cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Spades})

	b.HandleInput("s") // stand -> enters dealer turn, hole card should be revealed+counted now

	wantAfter := countBeforeReveal + hiLoValue(cardpkg.King)
	if b.shoe.RunningCount() != wantAfter {
		t.Errorf("RunningCount() after dealer-turn reveal = %d, want %d (hole card now counted)", b.shoe.RunningCount(), wantAfter)
	}
	if !b.dealerHoleCounted {
		t.Error("dealerHoleCounted should be true after entering dealer's turn")
	}
}
