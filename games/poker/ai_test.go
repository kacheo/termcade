package poker

import (
	"math/rand"
	"testing"

	"github.com/kacheo/tmvgs/games/cards"
)

func TestEasyAIRaisesLessThan20Percent(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	holeCards := [2]cards.Card{
		{Rank: cards.Ace, Suit: cards.Spades},
		{Rank: cards.King, Suit: cards.Spades},
	}
	community := []cards.Card{
		{Rank: cards.Queen, Suit: cards.Hearts},
		{Rank: cards.Jack, Suit: cards.Hearts},
		{Rank: cards.Ten, Suit: cards.Hearts},
	}

	raises := 0
	for i := 0; i < 100; i++ {
		decision := MakeDecision(Easy, rng, holeCards, community, 1000, 50, 300, 20)
		if decision.Action == ActionRaise {
			raises++
		}
	}
	if raises >= 20 {
		t.Errorf("Easy AI raised %d times out of 100, expected < 20", raises)
	}
}

func TestHardAIFoldsTrashWhenPotOddsBad(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	holeCards := [2]cards.Card{
		{Rank: cards.Two, Suit: cards.Spades},
		{Rank: cards.Five, Suit: cards.Hearts},
	}
	community := []cards.Card{
		{Rank: cards.Seven, Suit: cards.Clubs},
		{Rank: cards.Nine, Suit: cards.Diamonds},
		{Rank: cards.King, Suit: cards.Spades},
	}

	folds := 0
	calls := 0
	for i := 0; i < 100; i++ {
		decision := MakeDecision(Hard, rng, holeCards, community, 1000, 100, 100, 20)
		if decision.Action == ActionFold {
			folds++
		}
		if decision.Action == ActionCall {
			calls++
		}
	}
	if folds == 0 {
		t.Errorf("Hard AI with trash hand never folded when pot odds were bad")
	}
}

func TestMediumAIDoesNotFoldStrongHands(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	holeCards := [2]cards.Card{
		{Rank: cards.Ace, Suit: cards.Spades},
		{Rank: cards.Ace, Suit: cards.Hearts},
	}
	community := []cards.Card{
		{Rank: cards.King, Suit: cards.Clubs},
		{Rank: cards.Queen, Suit: cards.Diamonds},
		{Rank: cards.Jack, Suit: cards.Spades},
	}

	folds := 0
	for i := 0; i < 50; i++ {
		decision := MakeDecision(Medium, rng, holeCards, community, 1000, 50, 300, 20)
		if decision.Action == ActionFold {
			folds++
		}
	}
	if folds > 5 {
		t.Errorf("Medium AI with strong hand (four of a kind possible) folded %d times out of 50", folds)
	}
}

func TestEasyAIHandlesFreeCheck(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	holeCards := [2]cards.Card{
		{Rank: cards.Seven, Suit: cards.Spades},
		{Rank: cards.Two, Suit: cards.Hearts},
	}

	checks := 0
	for i := 0; i < 50; i++ {
		decision := MakeDecision(Easy, rng, holeCards, nil, 1000, 0, 100, 20)
		if decision.Action == ActionCheck {
			checks++
		}
	}
	if checks < 10 {
		t.Errorf("Easy AI checked only %d times out of 50 when check was free", checks)
	}
}

func TestHardAIBluffingOnRiver(t *testing.T) {
	rng := rand.New(rand.NewSource(777))
	holeCards := [2]cards.Card{
		{Rank: cards.Ace, Suit: cards.Spades},
		{Rank: cards.King, Suit: cards.Spades},
	}
	community := []cards.Card{
		{Rank: cards.Queen, Suit: cards.Hearts},
		{Rank: cards.Jack, Suit: cards.Hearts},
		{Rank: cards.Ten, Suit: cards.Hearts},
		{Rank: cards.Nine, Suit: cards.Clubs},
	}

	raises := 0
	for i := 0; i < 100; i++ {
		decision := MakeDecision(Hard, rng, holeCards, community, 1000, 0, 500, 50)
		if decision.Action == ActionRaise {
			raises++
		}
	}
	if raises == 0 {
		t.Errorf("Hard AI never raised in 100 iterations with medium bucket")
	}
}