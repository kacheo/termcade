package blackjack

import (
	"strings"
	"testing"
	"time"

	cardpkg "github.com/kacheo/termcade/games/cards"
)

// dealNow moves a fresh game from phaseBetting straight into a dealt round.
func dealNow(g *Blackjack) {
	g.HandleInput("enter")
}

func TestNewBlackjack_Metadata(t *testing.T) {
	g := NewBlackjack(6)
	if g.Name() != "Blackjack" {
		t.Errorf("Name() = %q, want Blackjack", g.Name())
	}
	if g.IsGameOver() {
		t.Error("should not be game over at start")
	}
	if g.IsPaused() {
		t.Error("should not be paused at start")
	}
	if g.GetScore() != startingBankroll {
		t.Errorf("GetScore() = %d, want %d", g.GetScore(), startingBankroll)
	}
	if g.GetLevel() != 6 {
		t.Errorf("GetLevel() = %d, want 6", g.GetLevel())
	}
	if g.GetLines() != 0 {
		t.Errorf("GetLines() = %d, want 0 (no round dealt yet)", g.GetLines())
	}
	if g.phase != phaseBetting {
		t.Errorf("initial phase = %v, want phaseBetting", g.phase)
	}
}

func TestNewBlackjack_DefaultsDeckCount(t *testing.T) {
	g := NewBlackjack(0)
	if g.GetLevel() != 6 {
		t.Errorf("GetLevel() = %d, want 6 (default)", g.GetLevel())
	}
}

func TestBetting_AdjustClamped(t *testing.T) {
	g := NewBlackjack(6)
	for i := 0; i < 200; i++ {
		g.HandleInput("left")
	}
	if g.bet != minBet {
		t.Errorf("bet = %d, want clamped to minBet %d", g.bet, minBet)
	}
	for i := 0; i < 200; i++ {
		g.HandleInput("right")
	}
	if g.bet != g.bankroll {
		t.Errorf("bet = %d, want clamped to bankroll %d", g.bet, g.bankroll)
	}
}

func TestBetting_EnterDeals(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	if g.phase != phaseDealing && g.phase != phaseInsurance && g.phase != phaseTurn && g.phase != phaseDealerTurn {
		t.Errorf("phase after Enter in phaseBetting = %v, want a dealt-round phase", g.phase)
	}
	if len(g.hands[0].hand) != 2 {
		t.Errorf("player cards = %d, want 2", len(g.hands[0].hand))
	}
	if len(g.dealer) != 2 {
		t.Errorf("dealer cards = %d, want 2", len(g.dealer))
	}
	if g.bankroll != startingBankroll-minBet {
		t.Errorf("bankroll = %d, want %d after placing bet", g.bankroll, startingBankroll-minBet)
	}
}

func TestPhase_DealingToTurn(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.dealer = Hand{cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Clubs}, cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Diamonds}} // no ace, no insurance
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}}
	g.hands[0].status = statusPlaying
	if err := g.Update(dealDelay + time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if g.phase != phaseTurn {
		t.Errorf("expected phaseTurn, got %v", g.phase)
	}
}

func TestPhase_DealingToInsurance_OnDealerAce(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.dealer = Hand{cardpkg.Card{Rank: cardpkg.Ace, Suit: cardpkg.Clubs}, cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Diamonds}}
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}}
	g.hands[0].status = statusPlaying
	if err := g.Update(dealDelay + time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if g.phase != phaseInsurance {
		t.Errorf("expected phaseInsurance when dealer shows Ace, got %v", g.phase)
	}
}

func TestPhase_DealingToDealer_OnBlackjack(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.dealer = Hand{cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Clubs}, cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Diamonds}}
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Ace, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Hearts}} // blackjack
	g.hands[0].status = statusBlackjack
	if err := g.Update(dealDelay + time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if g.phase != phaseDealerTurn {
		t.Errorf("expected phaseDealerTurn on blackjack, got %v", g.phase)
	}
}

func TestEvaluate_PlayerWinsVsBustedDealer(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.dealer = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}, cardpkg.Card{Rank: cardpkg.Seven, Suit: cardpkg.Clubs}} // 23
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Eight, Suit: cardpkg.Hearts}}                                            // 18
	g.hands[0].status = statusStand
	bankrollBefore := g.bankroll
	g.evaluateResults()
	if g.hands[0].result != "WIN" {
		t.Errorf("result = %q, want WIN", g.hands[0].result)
	}
	if g.wins != 1 {
		t.Errorf("wins = %d, want 1", g.wins)
	}
	if want := bankrollBefore + 2*g.hands[0].bet; g.bankroll != want {
		t.Errorf("bankroll = %d, want %d (bet returned + winnings)", g.bankroll, want)
	}
}

func TestEvaluate_DealerWins(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.dealer = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Hearts}}
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Eight, Suit: cardpkg.Hearts}} // 18 vs 19
	g.hands[0].status = statusStand
	bankrollBefore := g.bankroll
	g.evaluateResults()
	if g.hands[0].result != "LOSE" {
		t.Errorf("result = %q, want LOSE", g.hands[0].result)
	}
	if g.bankroll != bankrollBefore {
		t.Errorf("bankroll = %d, want unchanged %d on loss", g.bankroll, bankrollBefore)
	}
}

func TestEvaluate_Push(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.dealer = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Hearts}}
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Clubs}} // 19 vs 19
	g.hands[0].status = statusStand
	bankrollBefore := g.bankroll
	g.evaluateResults()
	if g.hands[0].result != "PUSH" {
		t.Errorf("result = %q, want PUSH", g.hands[0].result)
	}
	if want := bankrollBefore + g.hands[0].bet; g.bankroll != want {
		t.Errorf("bankroll = %d, want %d (bet returned on push)", g.bankroll, want)
	}
}

func TestEvaluate_BustLoses(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}, cardpkg.Card{Rank: cardpkg.Eight, Suit: cardpkg.Clubs}} // 24
	g.hands[0].status = statusBust
	g.evaluateResults()
	if g.hands[0].result != "LOSE" {
		t.Errorf("result = %q, want LOSE", g.hands[0].result)
	}
}

func TestEvaluate_GameOverOnBankrollBelowMinBet(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.bankroll = minBet - 1
	g.hands[0].status = statusBust
	g.evaluateResults()
	if !g.IsGameOver() {
		t.Error("expected IsGameOver() true when bankroll falls below minBet")
	}
}

func TestHandleInput_HitDrawsCard(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}} // 16
	g.hands[0].status = statusPlaying
	g.phase = phaseTurn
	before := len(g.hands[0].hand)
	g.HandleInput("h")
	if len(g.hands[0].hand) != before+1 {
		t.Errorf("hand size = %d, want %d", len(g.hands[0].hand), before+1)
	}
}

func TestHandleInput_HitBust_GoesToDealer(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Hearts}} // 19
	g.hands[0].status = statusPlaying
	g.phase = phaseTurn
	g.shoe.cards = append(cardpkg.Deck{cardpkg.Card{Rank: cardpkg.Five, Suit: cardpkg.Clubs}}, g.shoe.cards...) // 19+5=24, bust
	g.HandleInput("h")
	if g.hands[0].status != statusBust {
		t.Errorf("status = %v, want statusBust", g.hands[0].status)
	}
	if g.phase != phaseDealerTurn {
		t.Errorf("phase = %v, want phaseDealerTurn", g.phase)
	}
}

func TestHandleInput_Stand_GoesToDealer(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}} // 16
	g.hands[0].status = statusPlaying
	g.phase = phaseTurn
	g.HandleInput("s")
	if g.hands[0].status != statusStand {
		t.Errorf("status = %v, want statusStand", g.hands[0].status)
	}
	if g.phase != phaseDealerTurn {
		t.Errorf("phase = %v, want phaseDealerTurn", g.phase)
	}
}

func TestHandleInput_Double_DrawsOneCardAndStands(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Five, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}} // 11
	g.hands[0].status = statusPlaying
	g.hands[0].bet = 50
	g.phase = phaseTurn
	betBefore := g.hands[0].bet
	bankrollBefore := g.bankroll
	g.shoe.cards = append(cardpkg.Deck{cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Clubs}}, g.shoe.cards...)

	g.HandleInput("d")

	if len(g.hands[0].hand) != 3 {
		t.Errorf("hand size = %d, want 3 after double", len(g.hands[0].hand))
	}
	if !g.hands[0].isDoubled {
		t.Error("expected isDoubled true")
	}
	if g.hands[0].bet != betBefore*2 {
		t.Errorf("bet = %d, want %d (doubled)", g.hands[0].bet, betBefore*2)
	}
	if g.bankroll != bankrollBefore-betBefore {
		t.Errorf("bankroll = %d, want %d (debited additional bet)", g.bankroll, bankrollBefore-betBefore)
	}
	if g.hands[0].status != statusStand {
		t.Errorf("status = %v, want statusStand after double", g.hands[0].status)
	}
	if g.phase != phaseDealerTurn {
		t.Errorf("phase = %v, want phaseDealerTurn after double", g.phase)
	}
}

func TestHandleInput_Split_CreatesTwoHands(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Eight, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Eight, Suit: cardpkg.Hearts}}
	g.hands[0].status = statusPlaying
	g.hands[0].bet = 50
	g.phase = phaseTurn
	bankrollBefore := g.bankroll
	g.shoe.cards = append(cardpkg.Deck{
		cardpkg.Card{Rank: cardpkg.Two, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Three, Suit: cardpkg.Diamonds},
	}, g.shoe.cards...)

	g.HandleInput("x")

	if len(g.hands) != 2 {
		t.Fatalf("hands = %d, want 2 after split", len(g.hands))
	}
	if g.hands[0].bet != 50 || g.hands[1].bet != 50 {
		t.Errorf("split hand bets = %d, %d, want 50, 50", g.hands[0].bet, g.hands[1].bet)
	}
	if g.bankroll != bankrollBefore-50 {
		t.Errorf("bankroll = %d, want %d (debited second bet)", g.bankroll, bankrollBefore-50)
	}
	if g.canSplit(&g.hands[0]) {
		t.Error("should not be able to split again after a split (max one split)")
	}
}

func TestHandleInput_SplitAces_OneCardEachAutoStand(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Ace, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Ace, Suit: cardpkg.Hearts}}
	g.hands[0].status = statusPlaying
	g.hands[0].bet = 50
	g.phase = phaseTurn
	g.shoe.cards = append(cardpkg.Deck{
		cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Clubs},
		cardpkg.Card{Rank: cardpkg.Queen, Suit: cardpkg.Diamonds},
	}, g.shoe.cards...)

	g.HandleInput("x")

	if len(g.hands) != 2 {
		t.Fatalf("hands = %d, want 2 after split", len(g.hands))
	}
	for i, h := range g.hands {
		if len(h.hand) != 2 {
			t.Errorf("hand %d has %d cards, want 2 (one Ace + one draw)", i, len(h.hand))
		}
		if h.status != statusStand {
			t.Errorf("hand %d status = %v, want statusStand (split Aces auto-stand)", i, h.status)
		}
		if !h.fromSplitAces {
			t.Errorf("hand %d fromSplitAces = false, want true", i)
		}
	}
	// Both hands auto-stand, so the round should move straight to the dealer's turn.
	if g.phase != phaseDealerTurn {
		t.Errorf("phase = %v, want phaseDealerTurn after split-Aces auto-stand", g.phase)
	}
}

func TestHandleInput_ResultsEnter_GoesToBetting(t *testing.T) {
	g := NewBlackjack(6)
	g.phase = phaseResults
	g.HandleInput("enter")
	if g.phase != phaseBetting {
		t.Errorf("phase = %v, want phaseBetting", g.phase)
	}
}

func TestHandleInput_IgnoredDuringDealing(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.phase = phaseDealing
	before := len(g.hands[0].hand)
	g.HandleInput("h")
	if len(g.hands[0].hand) != before {
		t.Error("hit during phaseDealing should be ignored")
	}
}

func TestHandleInput_ToggleCount_WorksInAnyPhase(t *testing.T) {
	g := NewBlackjack(6)
	if g.showCount {
		t.Fatal("showCount should default to false")
	}
	g.HandleInput("c")
	if !g.showCount {
		t.Error("expected showCount true after toggling")
	}
	g.HandleInput("c")
	if g.showCount {
		t.Error("expected showCount false after toggling again")
	}
}

func TestInsurance_OfferedOnlyOnDealerAce(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.dealer = Hand{cardpkg.Card{Rank: cardpkg.Ace, Suit: cardpkg.Clubs}, cardpkg.Card{Rank: cardpkg.Nine, Suit: cardpkg.Diamonds}}
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}}
	g.hands[0].status = statusPlaying
	g.phase = phaseDealing
	g.elapsed = 0
	if err := g.Update(dealDelay + time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if g.phase != phaseInsurance {
		t.Fatalf("phase = %v, want phaseInsurance", g.phase)
	}

	g.hands[0].bet = 50
	bankrollBefore := g.bankroll
	g.HandleInput("y")
	if !g.insuranceTaken {
		t.Error("expected insuranceTaken true")
	}
	if g.insuranceBet != 25 {
		t.Errorf("insuranceBet = %d, want 25 (half of 50)", g.insuranceBet)
	}
	if g.bankroll != bankrollBefore-25 {
		t.Errorf("bankroll = %d, want %d after insurance debit", g.bankroll, bankrollBefore-25)
	}
	if g.insuranceTotalCount != 1 {
		t.Errorf("insuranceTotalCount = %d, want 1", g.insuranceTotalCount)
	}
}

func TestInsurance_DealerBlackjackSettlesImmediately(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.dealer = Hand{cardpkg.Card{Rank: cardpkg.Ace, Suit: cardpkg.Clubs}, cardpkg.Card{Rank: cardpkg.King, Suit: cardpkg.Diamonds}}
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}}
	g.hands[0].status = statusPlaying
	g.hands[0].bet = 50
	g.phase = phaseInsurance
	g.insuranceOffered = true

	bankrollBefore := g.bankroll
	g.HandleInput("y")

	if g.phase != phaseResults {
		t.Fatalf("phase = %v, want phaseResults (dealer blackjack settles immediately)", g.phase)
	}
	if g.insuranceResult != "WIN" {
		t.Errorf("insuranceResult = %q, want WIN", g.insuranceResult)
	}
	if want := bankrollBefore - 25 + 75; g.bankroll != want { // -insurance bet, +3x insurance bet payout
		t.Errorf("bankroll = %d, want %d", g.bankroll, want)
	}
	if g.hands[0].result != "LOSE" {
		t.Errorf("main hand result = %q, want LOSE (non-blackjack hand loses to dealer blackjack)", g.hands[0].result)
	}
}

func TestRender_ContainsLabels(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	out := g.Render()
	for _, want := range []string{"BLACKJACK", "DEALER", "YOU"} {
		if !strings.Contains(out, want) {
			t.Errorf("Render() missing %q", want)
		}
	}
}

func TestRender_ResultsPhaseShowsPromptAndResult(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.phase = phaseResults
	g.hands[0].result = "WIN"
	out := g.Render()
	if !strings.Contains(out, "WIN") {
		t.Error("Render() in phaseResults should show WIN")
	}
	if !strings.Contains(out, "ENTER") {
		t.Error("Render() in phaseResults should show ENTER prompt")
	}
}

func TestRender_PlayerTurnShowsActions(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	g.hands[0].hand = Hand{cardpkg.Card{Rank: cardpkg.Ten, Suit: cardpkg.Spades}, cardpkg.Card{Rank: cardpkg.Six, Suit: cardpkg.Hearts}}
	g.hands[0].status = statusPlaying
	g.phase = phaseTurn
	out := g.Render()
	if !strings.Contains(out, "Hit") && !strings.Contains(out, "H-Hit") {
		t.Error("player turn should show hit action")
	}
	if !strings.Contains(out, "Stand") && !strings.Contains(out, "S-Stand") {
		t.Error("player turn should show stand action")
	}
}

func TestRender_BettingPhaseShowsBankrollAndBet(t *testing.T) {
	g := NewBlackjack(6)
	out := g.Render()
	if !strings.Contains(out, "Bankroll") {
		t.Error("betting phase should show bankroll")
	}
	if !strings.Contains(out, "Bet:") {
		t.Error("betting phase should show current bet")
	}
}

func TestRender_CountOverlayOnlyWhenToggled(t *testing.T) {
	g := NewBlackjack(6)
	dealNow(g)
	out := g.Render()
	if strings.Contains(out, "True:") {
		t.Error("count overlay should be hidden by default")
	}
	g.HandleInput("c")
	out = g.Render()
	if !strings.Contains(out, "True:") {
		t.Error("count overlay should show once toggled on")
	}
}

func TestBlackjackDescription(t *testing.T) {
	b := NewBlackjack(6)
	if b.Description() != "Beat the dealer. Bet, hit, stand, double, or split — closest to 21 wins." {
		t.Errorf("Description() = %q", b.Description())
	}
}
