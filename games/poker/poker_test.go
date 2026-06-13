package poker

import (
	"testing"

	"github.com/kacheo/tmvgs/core"
	"github.com/kacheo/tmvgs/games/cards"
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

func TestBettingRoundNotEndedUntilBBActsPreflop(t *testing.T) {
	p := NewPoker(3, Easy)
	for i := range p.players {
		p.players[i].chips = 980
		p.players[i].bet = 20
		p.players[i].folded = false
		p.players[i].allIn = false
	}
	p.players[0].acted = true
	p.players[1].acted = true
	p.players[2].acted = false
	p.toCall = 20
	p.action = 2
	if p.bettingRoundEnded() {
		t.Error("bettingRoundEnded should return false when BB has not acted yet")
	}
	p.players[2].acted = true
	if !p.bettingRoundEnded() {
		t.Error("bettingRoundEnded should return true after BB acts")
	}
}

func TestEndBettingRoundOnRiverDistributesPot(t *testing.T) {
	p := NewPoker(3, Easy)
	for i := range p.players {
		p.players[i].chips = 1000
		p.players[i].acted = true
		p.players[i].folded = false
		p.players[i].allIn = false
		p.players[i].bet = 0
	}
	p.pot = 150
	p.phase = phaseRiver
	p.community = make([]cards.Card, 5)
	p.endBettingRound()
	if p.pot != 0 {
		t.Errorf("pot should be 0 after river endBettingRound, got %d", p.pot)
	}
	total := 0
	for _, pl := range p.players {
		total += pl.chips
	}
	if total != 3150 {
		t.Errorf("total chips should equal 3150 after pot distribution, got %d", total)
	}
}

func TestEndBettingRoundWithOnlyOneRemainingGoesToShowdown(t *testing.T) {
	p := NewPoker(3, Easy)
	for i := range p.players {
		p.players[i].chips = 1000
	}
	p.pot = 80
	p.phase = phasePreflop
	p.community = make([]cards.Card, 5)
	p.players[0].folded = false
	p.players[0].acted = true
	p.players[0].bet = 0
	p.players[1].folded = true
	p.players[2].folded = true
	p.endBettingRound()
	if p.phase != phaseShowdown {
		t.Errorf("phase should be phaseShowdown when 1 player remains, got %v", p.phase)
	}
	if p.pot != 0 {
		t.Errorf("pot should be 0 after showdown, got %d", p.pot)
	}
	if p.players[0].chips != 1080 {
		t.Errorf("last standing player should win pot (1080), got %d chips", p.players[0].chips)
	}
}

func TestShowdown_SidePotAllIn(t *testing.T) {
	p := NewPoker(3, Easy)
	// Community: three Kings + two low cards → full houses determined by hole pair rank.
	p.community = []cards.Card{
		{Rank: cards.King, Suit: cards.Spades},
		{Rank: cards.King, Suit: cards.Hearts},
		{Rank: cards.King, Suit: cards.Diamonds},
		{Rank: cards.Two, Suit: cards.Clubs},
		{Rank: cards.Three, Suit: cards.Spades},
	}
	// Player 0 (human): all-in at $50, best hand (KKK+AA full house).
	p.players[0].chips = 0
	p.players[0].contributed = 50
	p.players[0].allIn = true
	p.players[0].folded = false
	p.players[0].hole = [2]cards.Card{
		{Rank: cards.Ace, Suit: cards.Hearts},
		{Rank: cards.Ace, Suit: cards.Diamonds},
	}
	// Player 1 (AI): contributed $100, second-best hand (KKK+QQ full house).
	p.players[1].chips = 0
	p.players[1].contributed = 100
	p.players[1].allIn = false
	p.players[1].folded = false
	p.players[1].hole = [2]cards.Card{
		{Rank: cards.Queen, Suit: cards.Hearts},
		{Rank: cards.Queen, Suit: cards.Diamonds},
	}
	// Player 2 (AI): contributed $100, worst hand (KKK+JJ full house).
	p.players[2].chips = 0
	p.players[2].contributed = 100
	p.players[2].allIn = false
	p.players[2].folded = false
	p.players[2].hole = [2]cards.Card{
		{Rank: cards.Jack, Suit: cards.Hearts},
		{Rank: cards.Jack, Suit: cards.Diamonds},
	}
	// pot must equal sum of contributions to activate the side-pot path.
	p.pot = 250

	p.showdown()

	// Main pot (level $50): $150 total, all 3 eligible → player 0 wins.
	// Side pot (level $100): $100 total, players 1 and 2 eligible → player 1 wins.
	if p.players[0].chips != 150 {
		t.Errorf("all-in player chips = %d, want 150 (main pot)", p.players[0].chips)
	}
	if p.players[1].chips != 100 {
		t.Errorf("player 1 chips = %d, want 100 (side pot)", p.players[1].chips)
	}
	if p.players[2].chips != 0 {
		t.Errorf("player 2 chips = %d, want 0", p.players[2].chips)
	}
	if p.pot != 0 {
		t.Errorf("pot = %d, want 0 after showdown", p.pot)
	}
}