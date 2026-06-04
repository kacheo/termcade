package poker

import (
	"math/rand"

	"tmvgs/games/cards"
)

type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
)

type Action int

const (
	ActionFold Action = iota
	ActionCheck
	ActionCall
	ActionRaise
	ActionAllIn
)

type Decision struct {
	Action Action
	Amount int
}

type bucket int

const (
	bucketTrash bucket = iota
	bucketWeak
	bucketMedium
	bucketStrong
	bucketMonster
)

func MakeDecision(
	difficulty Difficulty,
	rng *rand.Rand,
	holeCards [2]cards.Card,
	community []cards.Card,
	chips int,
	toCall int,
	pot int,
	minRaise int,
) Decision {
	switch difficulty {
	case Easy:
		return easyDecision(rng, toCall, minRaise)
	case Medium:
		b := preflopBucket(holeCards)
		if len(community) > 0 {
			eval := Evaluate(append(community[:], holeCards[:]...))
			b = handRankBucket(eval.Rank)
		}
		return mediumDecision(rng, b, toCall, pot, minRaise, chips)
	case Hard:
		b := preflopBucket(holeCards)
		if len(community) > 0 {
			eval := Evaluate(append(community[:], holeCards[:]...))
			b = handRankBucket(eval.Rank)
		}
		return hardDecision(rng, b, toCall, pot, minRaise, chips)
	}
	return Decision{Action: ActionFold}
}

func handRankBucket(rank HandRank) bucket {
	switch rank {
	case HighCard:
		return bucketTrash
	case OnePair:
		return bucketWeak
	case TwoPair:
		return bucketWeak
	case ThreeOfAKind:
		return bucketMedium
	case Straight:
		return bucketMedium
	case Flush:
		return bucketMedium
	case FullHouse:
		return bucketStrong
	case FourOfAKind:
		return bucketStrong
	case StraightFlush:
		return bucketMonster
	case RoyalFlush:
		return bucketMonster
	}
	return bucketTrash
}

func preflopBucket(holeCards [2]cards.Card) bucket {
	ranks := make([]int, 2)
	for i, c := range holeCards {
		if c.Rank == cards.Ace {
			ranks[i] = 14
		} else {
			ranks[i] = int(c.Rank)
		}
	}
	if ranks[0] == ranks[1] {
		if ranks[0] >= int(cards.Jack) {
			return bucketMedium
		}
		return bucketTrash
	}
	if holeCards[0].Suit == holeCards[1].Suit {
		diff := ranks[0] - ranks[1]
		if diff < 0 {
			diff = -diff
		}
		if diff <= 2 {
			return bucketWeak
		}
	}
	if ranks[0] == 14 || ranks[1] == 14 {
		high := ranks[0]
		if high == 14 {
			high = ranks[1]
		}
		if high >= int(cards.Ten) {
			return bucketWeak
		}
	}
	return bucketTrash
}

func easyDecision(rng *rand.Rand, toCall int, minRaise int) Decision {
	r := rng.Intn(100)
	if r < 40 {
		return Decision{Action: ActionFold}
	}
	if r < 90 {
		if toCall == 0 {
			return Decision{Action: ActionCheck}
		}
		return Decision{Action: ActionCall}
	}
	return Decision{Action: ActionRaise, Amount: minRaise}
}

func mediumDecision(rng *rand.Rand, b bucket, toCall int, pot int, minRaise int, chips int) Decision {
	switch b {
	case bucketTrash:
		if toCall == 0 {
			return Decision{Action: ActionCheck}
		}
		return Decision{Action: ActionFold}
	case bucketWeak:
		if toCall == 0 {
			return Decision{Action: ActionCheck}
		}
		if toCall < pot/3 {
			return Decision{Action: ActionCall}
		}
		return Decision{Action: ActionFold}
	case bucketMedium:
		if toCall == 0 {
			if rng.Intn(100) < 20 {
				return Decision{Action: ActionRaise, Amount: minRaise}
			}
			return Decision{Action: ActionCheck}
		}
		return Decision{Action: ActionCall}
	case bucketStrong:
		if toCall == 0 {
			return Decision{Action: ActionRaise, Amount: minRaise}
		}
		return Decision{Action: ActionRaise, Amount: minRaise}
	case bucketMonster:
		raiseAmount := minRaise * 2
		if raiseAmount > chips {
			raiseAmount = chips
		}
		return Decision{Action: ActionRaise, Amount: raiseAmount}
	}
	return Decision{Action: ActionFold}
}

func hardDecision(rng *rand.Rand, b bucket, toCall int, pot int, minRaise int, chips int) Decision {
	equity := bucketEquity(b)
	if toCall > 0 {
		odd := float64(toCall) / float64(pot+toCall)
		if odd >= equity {
			return Decision{Action: ActionFold}
		}
	}
	switch b {
	case bucketTrash:
		if toCall == 0 {
			return Decision{Action: ActionCheck}
		}
		return Decision{Action: ActionFold}
	case bucketWeak:
		if toCall == 0 {
			return Decision{Action: ActionCheck}
		}
		if toCall < pot/3 {
			return Decision{Action: ActionCall}
		}
		return Decision{Action: ActionFold}
	case bucketMedium:
		if toCall == 0 {
			if rng.Intn(100) < 20 {
				return Decision{Action: ActionRaise, Amount: minRaise}
			}
			return Decision{Action: ActionCheck}
		}
		return Decision{Action: ActionCall}
	case bucketStrong:
		if toCall == 0 {
			return Decision{Action: ActionRaise, Amount: minRaise}
		}
		return Decision{Action: ActionRaise, Amount: minRaise}
	case bucketMonster:
		raiseAmount := minRaise * 2
		if raiseAmount > chips {
			raiseAmount = chips
		}
		return Decision{Action: ActionRaise, Amount: raiseAmount}
	}
	return Decision{Action: ActionFold}
}

func bucketEquity(b bucket) float64 {
	switch b {
	case bucketTrash:
		return 0.15
	case bucketWeak:
		return 0.35
	case bucketMedium:
		return 0.55
	case bucketStrong:
		return 0.80
	case bucketMonster:
		return 0.95
	}
	return 0.15
}

