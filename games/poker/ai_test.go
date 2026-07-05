package poker

import (
	"math/rand"
	"testing"

	"github.com/kacheo/termcade/games/cards"
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

func TestHandRankBucketAllRanks(t *testing.T) {
	cases := []struct {
		rank HandRank
		want bucket
	}{
		{HighCard, bucketTrash},
		{OnePair, bucketWeak},
		{TwoPair, bucketWeak},
		{ThreeOfAKind, bucketMedium},
		{Straight, bucketMedium},
		{Flush, bucketMedium},
		{FullHouse, bucketStrong},
		{FourOfAKind, bucketStrong},
		{StraightFlush, bucketMonster},
		{RoyalFlush, bucketMonster},
	}
	for _, tc := range cases {
		got := handRankBucket(tc.rank)
		if got != tc.want {
			t.Errorf("handRankBucket(%v) = %v, want %v", tc.rank, got, tc.want)
		}
	}
	if handRankBucket(HandRank(99)) != bucketTrash {
		t.Error("handRankBucket(unknown) should return bucketTrash")
	}
}

func TestBucketEquityAllBuckets(t *testing.T) {
	cases := []struct {
		b    bucket
		want float64
	}{
		{bucketTrash, 0.15},
		{bucketWeak, 0.35},
		{bucketMedium, 0.55},
		{bucketStrong, 0.80},
		{bucketMonster, 0.95},
	}
	for _, tc := range cases {
		got := bucketEquity(tc.b)
		if got != tc.want {
			t.Errorf("bucketEquity(%v) = %v, want %v", tc.b, got, tc.want)
		}
	}
	if bucketEquity(bucket(99)) != 0.15 {
		t.Error("bucketEquity(unknown) should return 0.15")
	}
}

func TestMediumDecisionAllBuckets(t *testing.T) {
	rng := rand.New(rand.NewSource(0))

	d := mediumDecision(rng, bucketTrash, 0, 100, 10, 1000)
	if d.Action != ActionCheck {
		t.Errorf("trash+free: got %v, want Check", d.Action)
	}
	d = mediumDecision(rng, bucketTrash, 50, 100, 10, 1000)
	if d.Action != ActionFold {
		t.Errorf("trash+call: got %v, want Fold", d.Action)
	}
	d = mediumDecision(rng, bucketWeak, 0, 100, 10, 1000)
	if d.Action != ActionCheck {
		t.Errorf("weak+free: got %v, want Check", d.Action)
	}
	d = mediumDecision(rng, bucketWeak, 10, 100, 10, 1000)
	if d.Action != ActionCall {
		t.Errorf("weak+cheap: got %v, want Call", d.Action)
	}
	d = mediumDecision(rng, bucketWeak, 50, 100, 10, 1000)
	if d.Action != ActionFold {
		t.Errorf("weak+expensive: got %v, want Fold", d.Action)
	}
	for i := 0; i < 30; i++ {
		d = mediumDecision(rng, bucketMedium, 0, 100, 10, 1000)
		if d.Action != ActionCheck && d.Action != ActionRaise {
			t.Errorf("medium+free[%d]: got %v, want Check or Raise", i, d.Action)
		}
	}
	d = mediumDecision(rng, bucketMedium, 30, 100, 10, 1000)
	if d.Action != ActionCall {
		t.Errorf("medium+call: got %v, want Call", d.Action)
	}
	d = mediumDecision(rng, bucketStrong, 0, 100, 10, 1000)
	if d.Action != ActionRaise {
		t.Errorf("strong+free: got %v, want Raise", d.Action)
	}
	d = mediumDecision(rng, bucketStrong, 30, 100, 10, 1000)
	if d.Action != ActionRaise {
		t.Errorf("strong+call: got %v, want Raise", d.Action)
	}
	d = mediumDecision(rng, bucketMonster, 0, 100, 10, 1000)
	if d.Action != ActionRaise || d.Amount != 20 {
		t.Errorf("monster: got %v amount=%d, want Raise 20", d.Action, d.Amount)
	}
	d = mediumDecision(rng, bucketMonster, 0, 100, 10, 5)
	if d.Action != ActionRaise || d.Amount != 5 {
		t.Errorf("monster+low chips: got %v amount=%d, want Raise 5", d.Action, d.Amount)
	}
}

func TestPreflopBucketAllBranches(t *testing.T) {
	// Pair of Jacks or better → bucketMedium
	b := preflopBucket([2]cards.Card{
		{Rank: cards.Jack, Suit: cards.Spades},
		{Rank: cards.Jack, Suit: cards.Hearts},
	})
	if b != bucketMedium {
		t.Errorf("Jack pair: got %v, want bucketMedium", b)
	}
	// Pair of low cards → bucketTrash
	b = preflopBucket([2]cards.Card{
		{Rank: cards.Two, Suit: cards.Spades},
		{Rank: cards.Two, Suit: cards.Hearts},
	})
	if b != bucketTrash {
		t.Errorf("Low pair: got %v, want bucketTrash", b)
	}
	// Suited connectors (diff ≤ 2) → bucketWeak
	b = preflopBucket([2]cards.Card{
		{Rank: cards.Seven, Suit: cards.Hearts},
		{Rank: cards.Eight, Suit: cards.Hearts},
	})
	if b != bucketWeak {
		t.Errorf("Suited connector: got %v, want bucketWeak", b)
	}
	// Suited but wide spread (diff > 2) — falls through to trash
	b = preflopBucket([2]cards.Card{
		{Rank: cards.Two, Suit: cards.Hearts},
		{Rank: cards.Six, Suit: cards.Hearts},
	})
	if b != bucketTrash {
		t.Errorf("Suited spread: got %v, want bucketTrash", b)
	}
	// Ace + Ten off-suit → bucketWeak (high >= Ten)
	b = preflopBucket([2]cards.Card{
		{Rank: cards.Ace, Suit: cards.Spades},
		{Rank: cards.Ten, Suit: cards.Hearts},
	})
	if b != bucketWeak {
		t.Errorf("Ace+Ten offsuit: got %v, want bucketWeak", b)
	}
	// Ace + Nine off-suit → bucketTrash (high < Ten)
	b = preflopBucket([2]cards.Card{
		{Rank: cards.Ace, Suit: cards.Spades},
		{Rank: cards.Nine, Suit: cards.Hearts},
	})
	if b != bucketTrash {
		t.Errorf("Ace+Nine offsuit: got %v, want bucketTrash", b)
	}
}

func TestHardDecisionEquityCheck(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	d := hardDecision(rng, bucketTrash, 100, 100, 10, 1000)
	if d.Action != ActionFold {
		t.Errorf("hard trash bad odds: got %v, want Fold", d.Action)
	}
	d = hardDecision(rng, bucketMonster, 10, 1000, 10, 1000)
	if d.Action != ActionRaise {
		t.Errorf("hard monster good odds: got %v, want Raise", d.Action)
	}
	d = hardDecision(rng, bucketMedium, 100, 100, 10, 1000)
	if d.Action != ActionCall {
		t.Errorf("hard medium proceed: got %v, want Call", d.Action)
	}
	d = hardDecision(rng, bucketTrash, 0, 100, 10, 1000)
	if d.Action != ActionCheck {
		t.Errorf("hard trash free: got %v, want Check", d.Action)
	}
	d = hardDecision(rng, bucketWeak, 0, 100, 10, 1000)
	if d.Action != ActionCheck {
		t.Errorf("hard weak free: got %v, want Check", d.Action)
	}
	d = hardDecision(rng, bucketWeak, 10, 1000, 10, 1000)
	if d.Action != ActionCall {
		t.Errorf("hard weak cheap: got %v, want Call", d.Action)
	}
	d = hardDecision(rng, bucketStrong, 0, 100, 10, 1000)
	if d.Action != ActionRaise {
		t.Errorf("hard strong free: got %v, want Raise", d.Action)
	}
}