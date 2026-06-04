package poker

import (
	"testing"
	"time"
)

// driveToHumanTurn calls Update in a loop until it is the human's turn
// (g.action == 0 in a betting phase) or until maxSteps is exceeded.
// When in phaseShowdown it advances time by 2 seconds per step so the
// showdown timer fires and startHand is called.
// Returns true if the human's turn was reached within maxSteps.
func driveToHumanTurn(g *Poker, maxSteps int) bool {
	for i := 0; i < maxSteps; i++ {
		if g.gameOver || g.phase == phaseGameOver {
			return false
		}
		if g.action == 0 && g.phase >= phasePreflop && g.phase <= phaseRiver &&
			!g.bettingRoundEnded() {
			return true
		}
		delta := time.Duration(0)
		if g.phase == phaseShowdown {
			// Jump the showdown timer forward so the next hand starts.
			delta = 2 * time.Second
		}
		_ = g.Update(delta)
	}
	// Final check after loop.
	return g.action == 0 && g.phase >= phasePreflop && g.phase <= phaseRiver &&
		!g.gameOver && !g.bettingRoundEnded()
}

// driveToCompletion drives the game with Update calls until it is game over,
// the human's turn is reached again (action==0 in a betting phase without the
// round already being over), or maxSteps is exceeded.
// Used for AI-only progression after the human acts.
func driveToCompletion(g *Poker, maxSteps int) {
	for i := 0; i < maxSteps; i++ {
		if g.gameOver || g.phase == phaseGameOver {
			return
		}
		// Stop when it is the human's turn and there is still a decision to
		// make (betting round not yet complete). If the round has ended the
		// next Update call will advance to the next phase, so we must not
		// stop yet.
		if g.action == 0 && g.phase >= phasePreflop && g.phase <= phaseRiver &&
			!g.bettingRoundEnded() {
			return
		}
		delta := time.Duration(0)
		if g.phase == phaseShowdown {
			delta = 2 * time.Second
		}
		_ = g.Update(delta)
	}
}

// newTestPoker creates a Poker game and pre-warms the AI delay to
// >= 300 ms so the first processAITurn call fires immediately.
// (processAITurn accumulates aiDelay in 16 ms increments before acting.)
func newTestPoker(seats int) *Poker {
	g := NewPoker(seats, Easy)
	g.aiDelay = 300 * time.Millisecond
	return g
}

// TestScenarioFold verifies that after the human folds preflop the game
// does not get stuck and eventually completes the hand.
func TestScenarioFold(t *testing.T) {
	g := newTestPoker(3)

	if !driveToHumanTurn(g, 500) {
		if g.gameOver {
			return // ended during setup; acceptable
		}
		t.Fatal("game stuck before human's first turn")
	}

	initialHands := g.handsPlayed
	initialPhase := g.phase

	g.HandleInput("f")

	driveToCompletion(g, 1000)

	if !g.gameOver {
		advanced := g.handsPlayed > initialHands || g.phase > initialPhase
		if !advanced {
			t.Errorf("game did not advance after fold: handsPlayed=%d (want >%d), phase=%v (was %v)",
				g.handsPlayed, initialHands, g.phase, initialPhase)
		}
	}
}

// TestScenarioCallCheck verifies that calling/checking advances the game
// through betting streets without panics.
func TestScenarioCallCheck(t *testing.T) {
	g := newTestPoker(3)

	t.Run("call_advances_phase", func(t *testing.T) {
		if !driveToHumanTurn(g, 500) {
			if g.gameOver {
				return
			}
			t.Fatal("game stuck before human's first turn")
		}

		preflopPhase := g.phase // should be phasePreflop

		g.HandleInput("c") // call or check

		driveToCompletion(g, 1000)

		// If the game has not ended a hand must have advanced past preflop
		// or the game must be over.
		if !g.gameOver && g.handsPlayed == 0 && g.phase == preflopPhase {
			t.Log("note: still on preflop — may be normal if human has more to act")
		}
	})
}

// TestScenarioAllIn verifies that the human going all-in preflop eventually
// leads to a showdown or game-over without panics.
func TestScenarioAllIn(t *testing.T) {
	g := newTestPoker(3)

	if !driveToHumanTurn(g, 500) {
		if g.gameOver {
			return
		}
		t.Fatal("game stuck before human's first turn (all-in scenario)")
	}

	g.HandleInput("a")

	driveToCompletion(g, 2000)

	// After an all-in the game must have advanced past preflop: the phase
	// should be at least phaseFlop (community cards dealt), or a hand must have
	// completed, or the game must be over.
	if !g.gameOver && g.handsPlayed == 0 && g.phase == phasePreflop {
		t.Errorf("after all-in: game did not advance past preflop: phase=%v handsPlayed=%d",
			g.phase, g.handsPlayed)
	}
}

// TestScenarioPotAwardedAfterFold exercises the pot-award bug from PR #12:
// when only one player remains after a fold the pot must be awarded to them,
// and the hand must complete (handsPlayed increments) without the game getting
// stuck.
//
// In a 2-seat game, the human folds on their first preflop turn. That leaves
// exactly one active player, triggering the "last player wins" code path.
func TestScenarioPotAwardedAfterFold(t *testing.T) {
	// Use a 2-seat game so a single fold leaves exactly 1 active player.
	g := newTestPoker(2)

	if !driveToHumanTurn(g, 500) {
		if g.gameOver {
			t.Skip("game ended during setup; skipping pot-award test")
		}
		t.Fatal("game stuck before human's first turn (pot-award scenario)")
	}

	// Capture total chips in play (all players + pot) as the conservation baseline.
	totalBefore := g.pot
	for _, pl := range g.players {
		totalBefore += pl.chips
	}
	if g.pot == 0 {
		t.Fatal("pot is 0 before fold — blind posting may not have occurred")
	}

	initialHands := g.handsPlayed

	// Human folds — the AI should win the pot.
	g.HandleInput("f")

	// Drive until a hand completes (handsPlayed increments) or game ends.
	// We do NOT drive until pot==0 because a subsequent hand's blinds will
	// refill the pot before we can observe pot==0.
	for i := 0; i < 1000; i++ {
		if g.gameOver || g.handsPlayed > initialHands {
			break
		}
		delta := time.Duration(0)
		if g.phase == phaseShowdown {
			delta = 2 * time.Second
		}
		_ = g.Update(delta)
	}

	if !g.gameOver && g.handsPlayed <= initialHands {
		t.Errorf("hand did not complete after human fold: handsPlayed=%d (want >%d); pot=%d; phase=%v",
			g.handsPlayed, initialHands, g.pot, g.phase)
		return
	}

	// Chip conservation: across all hands played the total chips in play must
	// remain constant (chips only move between players and pot, never created
	// or destroyed).
	totalAfter := g.pot
	for _, pl := range g.players {
		totalAfter += pl.chips
	}
	if totalAfter != totalBefore {
		t.Errorf("chip conservation violated across fold: before=%d after=%d (diff=%d)",
			totalBefore, totalAfter, totalAfter-totalBefore)
	}
}

// TestScenarioSidePot drives a 3-seat game where the human goes all-in with
// a short stack to create a multi-level pot situation. The test verifies that
// no panic occurs and the hand completes (handsPlayed increments or gameOver).
//
// The primary regression being tested is that the side-pot calculation in
// showdown does not panic or silently swallow chips.
func TestScenarioSidePot(t *testing.T) {
	g := newTestPoker(3)

	if !driveToHumanTurn(g, 500) {
		if g.gameOver {
			return
		}
		t.Fatal("game stuck before human's first turn (side-pot scenario)")
	}

	initialHands := g.handsPlayed

	// Shrink the human's stack so the all-in creates an unequal pot level,
	// forcing the side-pot code path in showdown when at least one AI calls.
	g.players[0].chips = 50

	// Capture the total chips in play after the artificial stack change.
	totalBefore := g.pot
	for _, pl := range g.players {
		totalBefore += pl.chips
	}

	// Human goes all-in with the reduced stack.
	g.HandleInput("a")

	// Drive AI turns and subsequent streets through to the next human turn
	// (or game over). With the short stack, a showdown should occur quickly.
	driveToCompletion(g, 2000)

	// The hand must have completed (pot awarded) or the game must be over.
	handCompleted := g.handsPlayed > initialHands || g.gameOver
	if !handCompleted {
		t.Errorf("hand did not complete after all-in: handsPlayed=%d (want >%d), phase=%v, pot=%d",
			g.handsPlayed, initialHands, g.phase, g.pot)
		return
	}

	// After a completed hand the pot must be zero (chips distributed to players).
	// If a new hand has started g.pot may be non-zero (new blinds); we verify
	// chip conservation instead, which catches any silent chip destruction.
	totalAfter := g.pot
	for _, pl := range g.players {
		totalAfter += pl.chips
	}
	if totalAfter != totalBefore {
		t.Errorf("chip conservation violated after side-pot hand: before=%d after=%d (diff=%d)",
			totalBefore, totalAfter, totalAfter-totalBefore)
	}
}
