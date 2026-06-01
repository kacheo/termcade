package poker

import (
	"testing"

	"tmvgs/core"
	"tmvgs/games/cards"
)

func TestNewPokerCreatesCorrectNumberOfPlayers(t *testing.T) {
	testCases := []struct {
		seats      int
		wantHumans int
		wantAIs    int
	}{
		{3, 1, 2},
		{4, 1, 3},
		{5, 1, 4},
	}
	for _, tc := range testCases {
		p := NewPoker(tc.seats, Easy)
		if len(p.players) != tc.seats {
			t.Errorf("NewPoker(%d) created %d players, want %d", tc.seats, len(p.players), tc.seats)
		}
		humanCount := 0
		for _, pl := range p.players {
			if pl.isHuman {
				humanCount++
			}
		}
		if humanCount != tc.wantHumans {
			t.Errorf("NewPoker(%d) has %d humans, want %d", tc.seats, humanCount, tc.wantHumans)
		}
		for _, pl := range p.players {
			if pl.chips > 1000 || pl.chips < 970 {
				t.Errorf("player %s has %d chips, expected 970-1000 (after blinds)", pl.name, pl.chips)
			}
		}
	}
}

func TestStartHandDealsHoleCardsAndPostsBlinds(t *testing.T) {
	p := NewPoker(3, Easy)
	initialHandsPlayed := p.handsPlayed
	p.startHand()
	if p.handsPlayed != initialHandsPlayed {
		t.Errorf("startHand should not increment handsPlayed")
	}
	for i, pl := range p.players {
		if pl.hole[0].Rank == 0 && pl.hole[1].Rank == 0 {
			t.Errorf("player %d has empty hole cards", i)
		}
	}
	if p.pot != 30 {
		t.Errorf("pot is %d, want 30 (10 SB + 20 BB)", p.pot)
	}
}

func TestFoldingAllAILeavesPotToHuman(t *testing.T) {
	p := NewPoker(3, Easy)
	// Reset chips to a known baseline — NewPoker randomly assigns blinds so
	// the human may have paid 0, 10, or 20 chips before we set up the scenario.
	for i := range p.players {
		p.players[i].chips = 1000
	}
	p.pot = 100
	p.community = make([]cards.Card, 5)
	p.players[0].folded = false
	p.players[1].folded = true
	p.players[2].folded = true
	p.players[0].hole = [2]cards.Card{
		{Rank: cards.Ace, Suit: cards.Spades},
		{Rank: cards.King, Suit: cards.Spades},
	}
	p.showdown()
	if p.players[0].chips != 1100 {
		t.Errorf("human chips after showdown is %d, want 1100", p.players[0].chips)
	}
	if p.pot != 0 {
		t.Errorf("pot after showdown is %d, want 0", p.pot)
	}
}

func TestPokerImplementsCoreGameInterface(t *testing.T) {
	p := NewPoker(3, Easy)
	var _ core.Game = p
}