package blackjack

import (
	"strings"
	"testing"
	"time"
)

func TestNewBlackjack_Metadata(t *testing.T) {
	g := NewBlackjack(2)
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

func TestNewBlackjack_Players(t *testing.T) {
	cases := []struct{ ai, total int }{{0, 1}, {1, 2}, {3, 4}, {4, 4}}
	for _, c := range cases {
		g := NewBlackjack(c.ai)
		if len(g.players) != c.total {
			t.Errorf("aiCount=%d: players=%d, want %d", c.ai, len(g.players), c.total)
		}
	}
	g := NewBlackjack(3)
	if g.players[0].isAI {
		t.Error("players[0] should be human")
	}
	for i := 1; i <= 3; i++ {
		if !g.players[i].isAI {
			t.Errorf("players[%d] should be AI", i)
		}
	}
}

func TestInitialDeal(t *testing.T) {
	g := NewBlackjack(2)
	if len(g.dealer) != 2 {
		t.Errorf("dealer cards = %d, want 2", len(g.dealer))
	}
	for i, p := range g.players {
		if len(p.hand) != 2 {
			t.Errorf("player[%d] cards = %d, want 2", i, len(p.hand))
		}
	}
	if g.phase != phaseDealing {
		t.Errorf("initial phase = %v, want phaseDealing", g.phase)
	}
}

func TestPhase_DealingToAITurn(t *testing.T) {
	g := NewBlackjack(2)
	// Pre-assign non-blackjack hands to ensure AIs are still statusPlaying
	g.players[1].hand = Hand{{Ten, Spades}, {Six, Hearts}}   // 16, not BJ
	g.players[1].status = statusPlaying
	g.players[2].hand = Hand{{Nine, Clubs}, {Seven, Diamonds}} // 16, not BJ
	g.players[2].status = statusPlaying
	g.Update(dealDelay + time.Millisecond)
	if g.phase != phaseAITurn {
		t.Errorf("expected phaseAITurn, got %v", g.phase)
	}
}

func TestPhase_DealingToPlayerTurn_NoAI(t *testing.T) {
	g := NewBlackjack(0)
	g.Update(dealDelay + time.Millisecond)
	if g.phase != phasePlayerTurn {
		t.Errorf("expected phasePlayerTurn (no AI), got %v", g.phase)
	}
}

func TestPhase_AIStandsAdvancesToPlayer(t *testing.T) {
	g := NewBlackjack(1)
	g.players[1].hand = Hand{{Ten, Spades}, {Seven, Hearts}} // hard 17 — AI will stand
	g.players[1].status = statusPlaying                       // Ensure statusPlaying
	g.Update(dealDelay + time.Millisecond)                    // → phaseAITurn
	g.Update(aiStepDelay + time.Millisecond)                  // AI acts
	if g.phase != phasePlayerTurn {
		t.Errorf("after AI stands expected phasePlayerTurn, got %v", g.phase)
	}
	if g.players[1].status != statusStand {
		t.Errorf("AI status = %v, want statusStand", g.players[1].status)
	}
}

func TestEvaluate_PlayerWinsVsBustedDealer(t *testing.T) {
	g := NewBlackjack(0)
	g.dealer = Hand{{Ten, Spades}, {Six, Hearts}, {Seven, Clubs}} // 23
	g.players[0].hand = Hand{{King, Spades}, {Eight, Hearts}}     // 18
	g.players[0].status = statusStand
	g.evaluateResults()
	if g.players[0].result != "WIN" {
		t.Errorf("result = %q, want WIN", g.players[0].result)
	}
	if g.wins != 1 {
		t.Errorf("wins = %d, want 1", g.wins)
	}
}

func TestEvaluate_DealerWins(t *testing.T) {
	g := NewBlackjack(0)
	g.dealer = Hand{{Ten, Spades}, {Nine, Hearts}}            // 19
	g.players[0].hand = Hand{{King, Spades}, {Eight, Hearts}} // 18
	g.players[0].status = statusStand
	g.evaluateResults()
	if g.players[0].result != "LOSE" {
		t.Errorf("result = %q, want LOSE", g.players[0].result)
	}
}

func TestEvaluate_Push(t *testing.T) {
	g := NewBlackjack(0)
	g.dealer = Hand{{Ten, Spades}, {Nine, Hearts}}           // 19
	g.players[0].hand = Hand{{King, Spades}, {Nine, Clubs}}  // 19
	g.players[0].status = statusStand
	g.evaluateResults()
	if g.players[0].result != "PUSH" {
		t.Errorf("result = %q, want PUSH", g.players[0].result)
	}
}

func TestEvaluate_BustLoses(t *testing.T) {
	g := NewBlackjack(0)
	g.players[0].hand = Hand{{Ten, Spades}, {Six, Hearts}, {Eight, Clubs}} // 24
	g.players[0].status = statusBust
	g.evaluateResults()
	if g.players[0].result != "LOSE" {
		t.Errorf("result = %q, want LOSE", g.players[0].result)
	}
}

func TestHandleInput_HitDrawsCard(t *testing.T) {
	g := NewBlackjack(0)
	g.Update(dealDelay + time.Millisecond) // → phasePlayerTurn
	before := len(g.players[0].hand)
	g.HandleInput("h")
	if len(g.players[0].hand) != before+1 {
		t.Errorf("hand size = %d, want %d", len(g.players[0].hand), before+1)
	}
}

func TestHandleInput_HitBust_GoesToDealer(t *testing.T) {
	g := NewBlackjack(0)
	g.Update(dealDelay + time.Millisecond)
	g.players[0].hand = Hand{{Ten, Spades}, {Nine, Hearts}} // 19
	g.players[0].status = statusPlaying                      // Ensure statusPlaying
	g.deck = append(Deck{Card{Five, Clubs}}, g.deck...)     // 19+5=24, bust
	g.HandleInput("h")
	if g.players[0].status != statusBust {
		t.Errorf("status = %v, want statusBust", g.players[0].status)
	}
	if g.phase != phaseDealerTurn {
		t.Errorf("phase = %v, want phaseDealerTurn", g.phase)
	}
}

func TestHandleInput_Stand_GoesToDealer(t *testing.T) {
	g := NewBlackjack(0)
	g.Update(dealDelay + time.Millisecond)
	// Ensure player is in statusPlaying (not blackjack)
	g.players[0].hand = Hand{{Ten, Spades}, {Six, Hearts}} // 16, not a blackjack
	g.players[0].status = statusPlaying
	g.HandleInput("s")
	if g.players[0].status != statusStand {
		t.Errorf("status = %v, want statusStand", g.players[0].status)
	}
	if g.phase != phaseDealerTurn {
		t.Errorf("phase = %v, want phaseDealerTurn", g.phase)
	}
}

func TestHandleInput_Enter_StartsNextRound(t *testing.T) {
	g := NewBlackjack(0)
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
	g := NewBlackjack(0) // starts in phaseDealing
	before := len(g.players[0].hand)
	g.HandleInput("h")
	if len(g.players[0].hand) != before {
		t.Error("hit during phaseDealing should be ignored")
	}
}

func TestRender_ContainsLabels(t *testing.T) {
	g := NewBlackjack(2)
	out := g.Render()
	for _, want := range []string{"BLACKJACK", "DEALER", "YOU", "AI-1", "AI-2"} {
		if !strings.Contains(out, want) {
			t.Errorf("Render() missing %q", want)
		}
	}
}

func TestRender_ResultsPhaseShowsPromptAndResult(t *testing.T) {
	g := NewBlackjack(0)
	g.phase = phaseResults
	g.players[0].result = "WIN"
	out := g.Render()
	if !strings.Contains(out, "WIN") {
		t.Error("Render() in phaseResults should show WIN")
	}
	if !strings.Contains(out, "ENTER") {
		t.Error("Render() in phaseResults should show ENTER prompt")
	}
}

func TestRender_PlayerTurnShowsActions(t *testing.T) {
	g := NewBlackjack(0)
	g.Update(dealDelay + time.Millisecond) // → phasePlayerTurn
	out := g.Render()
	if !strings.Contains(out, "Hit") && !strings.Contains(out, "H-Hit") {
		t.Error("player turn should show hit action")
	}
	if !strings.Contains(out, "Stand") && !strings.Contains(out, "S-Stand") {
		t.Error("player turn should show stand action")
	}
}
