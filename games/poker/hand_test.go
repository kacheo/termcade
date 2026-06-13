package poker

import (
	"testing"
	"github.com/kacheo/tmvgs/games/cards"
)

func makeCard(rank cards.Rank, suit cards.Suit) cards.Card {
	return cards.Card{Rank: rank, Suit: suit}
}

func highCardValue(r cards.Rank) int {
	if r == cards.Ace {
		return 14
	}
	return int(r)
}

func TestRoyalFlush(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Ace, cards.Spades),
		makeCard(cards.King, cards.Spades),
		makeCard(cards.Queen, cards.Spades),
		makeCard(cards.Jack, cards.Spades),
		makeCard(cards.Ten, cards.Spades),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != RoyalFlush {
		t.Errorf("expected RoyalFlush, got %v", evaluated.Rank)
	}
}

func TestStraightFlush(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Nine, cards.Spades),
		makeCard(cards.King, cards.Spades),
		makeCard(cards.Queen, cards.Spades),
		makeCard(cards.Jack, cards.Spades),
		makeCard(cards.Ten, cards.Spades),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != StraightFlush {
		t.Errorf("expected StraightFlush, got %v", evaluated.Rank)
	}
	if len(evaluated.Tiebreakers) == 0 || evaluated.Tiebreakers[0] != highCardValue(cards.King) {
		t.Errorf("expected King high, got %v", evaluated.Tiebreakers)
	}
}

func TestFourOfAKind(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Ace, cards.Spades),
		makeCard(cards.Ace, cards.Hearts),
		makeCard(cards.Ace, cards.Diamonds),
		makeCard(cards.Ace, cards.Clubs),
		makeCard(cards.King, cards.Spades),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != FourOfAKind {
		t.Errorf("expected FourOfAKind, got %v", evaluated.Rank)
	}
	if evaluated.Tiebreakers[0] != highCardValue(cards.Ace) {
		t.Errorf("expected four Aces, got %v", evaluated.Tiebreakers[0])
	}
}

func TestFullHouse(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.King, cards.Spades),
		makeCard(cards.King, cards.Hearts),
		makeCard(cards.King, cards.Diamonds),
		makeCard(cards.Queen, cards.Spades),
		makeCard(cards.Queen, cards.Hearts),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != FullHouse {
		t.Errorf("expected FullHouse, got %v", evaluated.Rank)
	}
	if evaluated.Tiebreakers[0] != highCardValue(cards.King) {
		t.Errorf("expected Kings full, got %v", evaluated.Tiebreakers[0])
	}
	if evaluated.Tiebreakers[1] != highCardValue(cards.Queen) {
		t.Errorf("expected Queens full, got %v", evaluated.Tiebreakers[1])
	}
}

func TestFlush(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Ace, cards.Spades),
		makeCard(cards.King, cards.Spades),
		makeCard(cards.Seven, cards.Spades),
		makeCard(cards.Jack, cards.Spades),
		makeCard(cards.Three, cards.Spades),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != Flush {
		t.Errorf("expected Flush, got %v", evaluated.Rank)
	}
	if evaluated.Tiebreakers[0] != highCardValue(cards.Ace) {
		t.Errorf("expected Ace high flush, got %v", evaluated.Tiebreakers[0])
	}
}

func TestStraight(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Nine, cards.Spades),
		makeCard(cards.Eight, cards.Hearts),
		makeCard(cards.Seven, cards.Diamonds),
		makeCard(cards.Six, cards.Clubs),
		makeCard(cards.Five, cards.Spades),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != Straight {
		t.Errorf("expected Straight, got %v", evaluated.Rank)
	}
	if evaluated.Tiebreakers[0] != highCardValue(cards.Nine) {
		t.Errorf("expected Nine high, got %v", evaluated.Tiebreakers[0])
	}
}

func TestAceLowStraight(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Ace, cards.Spades),
		makeCard(cards.Two, cards.Hearts),
		makeCard(cards.Three, cards.Diamonds),
		makeCard(cards.Four, cards.Clubs),
		makeCard(cards.Five, cards.Spades),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != Straight {
		t.Errorf("expected Straight (ace-low), got %v", evaluated.Rank)
	}
	if evaluated.Tiebreakers[0] != 5 {
		t.Errorf("expected 5-high straight, got %v", evaluated.Tiebreakers[0])
	}
}

func TestThreeOfAKind(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Queen, cards.Spades),
		makeCard(cards.Queen, cards.Hearts),
		makeCard(cards.Queen, cards.Diamonds),
		makeCard(cards.King, cards.Spades),
		makeCard(cards.Ace, cards.Hearts),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != ThreeOfAKind {
		t.Errorf("expected ThreeOfAKind, got %v", evaluated.Rank)
	}
	if evaluated.Tiebreakers[0] != highCardValue(cards.Queen) {
		t.Errorf("expected Queens trip, got %v", evaluated.Tiebreakers[0])
	}
}

func TestTwoPair(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Ace, cards.Spades),
		makeCard(cards.Ace, cards.Hearts),
		makeCard(cards.King, cards.Diamonds),
		makeCard(cards.King, cards.Clubs),
		makeCard(cards.Queen, cards.Spades),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != TwoPair {
		t.Errorf("expected TwoPair, got %v", evaluated.Rank)
	}
	if evaluated.Tiebreakers[0] != highCardValue(cards.Ace) || evaluated.Tiebreakers[1] != highCardValue(cards.King) {
		t.Errorf("expected Aces over Kings, got %v %v", evaluated.Tiebreakers[0], evaluated.Tiebreakers[1])
	}
}

func TestOnePair(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Jack, cards.Spades),
		makeCard(cards.Jack, cards.Hearts),
		makeCard(cards.Ten, cards.Diamonds),
		makeCard(cards.Seven, cards.Clubs),
		makeCard(cards.Three, cards.Spades),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != OnePair {
		t.Errorf("expected OnePair, got %v", evaluated.Rank)
	}
	if evaluated.Tiebreakers[0] != highCardValue(cards.Jack) {
		t.Errorf("expected Jacks pair, got %v", evaluated.Tiebreakers[0])
	}
}

func TestHighCard(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Ace, cards.Spades),
		makeCard(cards.King, cards.Hearts),
		makeCard(cards.Nine, cards.Diamonds),
		makeCard(cards.Seven, cards.Clubs),
		makeCard(cards.Three, cards.Spades),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != HighCard {
		t.Errorf("expected HighCard, got %v", evaluated.Rank)
	}
	if evaluated.Tiebreakers[0] != highCardValue(cards.Ace) {
		t.Errorf("expected Ace high, got %v", evaluated.Tiebreakers[0])
	}
}

func TestCompareDifferentRanks(t *testing.T) {
	straight := EvaluatedHand{Rank: Straight, Tiebreakers: []int{9}}
	flush := EvaluatedHand{Rank: Flush, Tiebreakers: []int{9, 7, 5, 3, 1}}
	if Compare(straight, flush) >= 0 {
		t.Error("Straight should lose to Flush")
	}
	if Compare(flush, straight) <= 0 {
		t.Error("Flush should beat Straight")
	}
}

func TestCompareSameRank(t *testing.T) {
	straight1 := EvaluatedHand{Rank: Straight, Tiebreakers: []int{9}}
	straight2 := EvaluatedHand{Rank: Straight, Tiebreakers: []int{8}}
	if Compare(straight1, straight2) != 1 {
		t.Error("Nine-high straight should beat Eight-high straight")
	}
	if Compare(straight2, straight1) != -1 {
		t.Error("Eight-high straight should lose to Nine-high straight")
	}
}

func TestCompareEqual(t *testing.T) {
	straight1 := EvaluatedHand{Rank: Straight, Tiebreakers: []int{9, 7, 5, 3, 1}}
	straight2 := EvaluatedHand{Rank: Straight, Tiebreakers: []int{9, 7, 5, 3, 1}}
	if Compare(straight1, straight2) != 0 {
		t.Error("Identical hands should be equal")
	}
}

func TestSevenCardBestHand(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Ace, cards.Spades),
		makeCard(cards.King, cards.Spades),
		makeCard(cards.Queen, cards.Spades),
		makeCard(cards.Jack, cards.Spades),
		makeCard(cards.Ten, cards.Spades),
		makeCard(cards.Nine, cards.Hearts),
		makeCard(cards.Eight, cards.Diamonds),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != RoyalFlush {
		t.Errorf("expected RoyalFlush from 7 cards, got %v", evaluated.Rank)
	}
}

func TestSevenCardSelectsBest(t *testing.T) {
	h := []cards.Card{
		makeCard(cards.Ace, cards.Spades),
		makeCard(cards.King, cards.Spades),
		makeCard(cards.Queen, cards.Spades),
		makeCard(cards.Jack, cards.Spades),
		makeCard(cards.Eight, cards.Spades),
		makeCard(cards.Nine, cards.Hearts),
		makeCard(cards.Eight, cards.Diamonds),
	}
	evaluated := Evaluate(h)
	if evaluated.Rank != Flush {
		t.Errorf("expected best Flush, got %v", evaluated.Rank)
	}
}