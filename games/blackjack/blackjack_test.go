package blackjack

import (
	"strings"
	"testing"
	"time"

	cardpkg "tmvgs/games/cards"
)

func TestNewBlackjack_Metadata(t *testing.T) {
	g := NewBlackjack()
	if g.Name() != "Blackjack" {
		t.Errorf("Name() = %q, want Blackjack", g.Name())
	}
	if g.IsGameOver() {
		t.Error("should not be game over at start")
	}
	if g.IsPaused() {
		t.Error("should not be paused at start")
	}
	if g.GetScore() != 0 {
		t.Errorf("GetScore() = %d, want 0", g.GetScore())
	}
	if g.GetLevel() != 0 {
		t.Errorf("GetLevel() = %d, want 0", g.GetLevel())
	}
	if g.GetLines() != 1 {
		t.Errorf("GetLines() = %d, want 1 (first round)", g.GetLines())
	}
}

func TestInitialDeal(t *testing.T) {
	g := NewBlackjack()
	if len(g.dealer) != 2 {
		t.Errorf("dealer cards = %d, want 2", len(g.dealer))
	}
	if len(g.player.hand) != 2 {
		t.Errorf("player cards = %d, want 2", len(g.player.hand))
	}
	if g.phase != phaseDealing {
		t.Errorf("initial phase = %v, want phaseDealing", g.phase)
	}
}

func TestPhase_DealingToTurn(t *testing.T) {
	g := NewBlackjack()
	g.player.hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}} // 16, not BJ
	g.player.status = statusPlaying
	if err := g.Update(dealDelay + time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if g.phase != phaseTurn {
		t.Errorf("expected phaseTurn, got %v", g.phase)
	}
}

func TestPhase_DealingToDealer_OnBlackjack(t *testing.T) {
	g := NewBlackjack()
	g.player.hand = Hand{cardpkg.Card{Rank: cardpkg.Ace, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Hearts}} // blackjack
	g.player.status = statusBlackjack
	if err := g.Update(dealDelay + time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if g.phase != phaseDealerTurn {
		t.Errorf("expected phaseDealerTurn on blackjack, got %v", g.phase)
	}
}

func TestEvaluate_PlayerWinsVsBustedDealer(t *testing.T) {
	g := NewBlackjack()
	g.dealer = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}, cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Clubs}} // 23
	g.player.hand = Hand{cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Eight, Suit: cardpkg.Hearts}}                                                  // 18
	g.player.status = statusStand
	g.evaluateResults()
	if g.player.result != "WIN" {
		t.Errorf("result = %q, want WIN", g.player.result)
	}
	if g.wins != 1 {
		t.Errorf("wins = %d, want 1", g.wins)
	}
}

func TestEvaluate_DealerWins(t *testing.T) {
	g := NewBlackjack()
	g.dealer = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Hearts}}
	g.player.hand = Hand{cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Eight, Suit: cardpkg.Hearts}} // 18 vs 19
	g.player.status = statusStand
	g.evaluateResults()
	if g.player.result != "LOSE" {
		t.Errorf("result = %q, want LOSE", g.player.result)
	}
}

func TestEvaluate_Push(t *testing.T) {
	g := NewBlackjack()
	g.dealer = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Hearts}}
	g.player.hand = Hand{cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Clubs}} // 19 vs 19
	g.player.status = statusStand
	g.evaluateResults()
	if g.player.result != "PUSH" {
		t.Errorf("result = %q, want PUSH", g.player.result)
	}
}

func TestEvaluate_BustLoses(t *testing.T) {
	g := NewBlackjack()
	g.player.hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}, cardpkg.Card{Rank: cardpkg.Eight, Suit: cardpkg.Clubs}} // 24
	g.player.status = statusBust
	g.evaluateResults()
	if g.player.result != "LOSE" {
		t.Errorf("result = %q, want LOSE", g.player.result)
	}
}

func TestHandleInput_HitDrawsCard(t *testing.T) {
	g := NewBlackjack()
	g.player.hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}} // 16
	g.player.status = statusPlaying
	g.phase = phaseTurn
	before := len(g.player.hand)
	g.HandleInput("h")
	if len(g.player.hand) != before+1 {
		t.Errorf("hand size = %d, want %d", len(g.player.hand), before+1)
	}
}

func TestHandleInput_HitBust_GoesToDealer(t *testing.T) {
	g := NewBlackjack()
	g.player.hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Hearts}} // 19
	g.player.status = statusPlaying
	g.phase = phaseTurn
	g.deck = append(cardpkg.Deck{cardpkg.Card{Rank: cardpkg.Five, Suit: cardpkg.Clubs}}, g.deck...) // 19+5=24, bust
	g.HandleInput("h")
	if g.player.status != statusBust {
		t.Errorf("status = %v, want statusBust", g.player.status)
	}
	if g.phase != phaseDealerTurn {
		t.Errorf("phase = %v, want phaseDealerTurn", g.phase)
	}
}

func TestHandleInput_Stand_GoesToDealer(t *testing.T) {
	g := NewBlackjack()
	g.player.hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}} // 16
	g.player.status = statusPlaying
	g.phase = phaseTurn
	g.HandleInput("s")
	if g.player.status != statusStand {
		t.Errorf("status = %v, want statusStand", g.player.status)
	}
	if g.phase != phaseDealerTurn {
		t.Errorf("phase = %v, want phaseDealerTurn", g.phase)
	}
}

func TestHandleInput_Enter_StartsNextRound(t *testing.T) {
	g := NewBlackjack()
	g.phase = phaseResults
	roundsBefore := g.rounds
	g.HandleInput("enter")
	if g.rounds != roundsBefore+1 {
		t.Errorf("rounds = %d, want %d", g.rounds, roundsBefore+1)
	}
	if g.phase != phaseDealing {
		t.Errorf("phase = %v, want phaseDealing", g.phase)
	}
}

func TestHandleInput_IgnoredDuringDealing(t *testing.T) {
	g := NewBlackjack()
	before := len(g.player.hand)
	g.HandleInput("h")
	if len(g.player.hand) != before {
		t.Error("hit during phaseDealing should be ignored")
	}
}

func TestRender_ContainsLabels(t *testing.T) {
	g := NewBlackjack()
	out := g.Render()
	for _, want := range []string{"BLACKJACK", "DEALER", "YOU"} {
		if !strings.Contains(out, want) {
			t.Errorf("Render() missing %q", want)
		}
	}
}

func TestRender_ResultsPhaseShowsPromptAndResult(t *testing.T) {
	g := NewBlackjack()
	g.phase = phaseResults
	g.player.result = "WIN"
	out := g.Render()
	if !strings.Contains(out, "WIN") {
		t.Error("Render() in phaseResults should show WIN")
	}
	if !strings.Contains(out, "ENTER") {
		t.Error("Render() in phaseResults should show ENTER prompt")
	}
}

func TestRender_PlayerTurnShowsActions(t *testing.T) {
	g := NewBlackjack()
	g.player.hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}}
	g.player.status = statusPlaying
	g.phase = phaseTurn
	out := g.Render()
	if !strings.Contains(out, "Hit") && !strings.Contains(out, "H-Hit") {
		t.Error("player turn should show hit action")
	}
	if !strings.Contains(out, "Stand") && !strings.Contains(out, "S-Stand") {
		t.Error("player turn should show stand action")
	}
}
