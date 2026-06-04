package poker

import (
	"sort"
	"tmvgs/games/cards"
)

type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

type EvaluatedHand struct {
	Rank        HandRank
	Tiebreakers []int
	Cards       [5]cards.Card
}

type Hand []cards.Card

func Evaluate(cardList []cards.Card) EvaluatedHand {
	if len(cardList) < 5 {
		panic("Evaluate requires at least 5 cards")
	}

	if len(cardList) == 5 {
		return evaluateFive(cardList)
	}

	best := evaluateFive([]cards.Card{cardList[0], cardList[1], cardList[2], cardList[3], cardList[4]})
	for i := 0; i < len(cardList); i++ {
		for j := i + 1; j < len(cardList); j++ {
			for k := j + 1; k < len(cardList); k++ {
				for l := k + 1; l < len(cardList); l++ {
					for m := l + 1; m < len(cardList); m++ {
						if i == 0 && j == 1 && k == 2 && l == 3 && m == 4 {
							continue
						}
						combo := []cards.Card{cardList[i], cardList[j], cardList[k], cardList[l], cardList[m]}
						evaluated := evaluateFive(combo)
						if Compare(evaluated, best) > 0 {
							best = evaluated
						}
					}
				}
			}
		}
	}
	return best
}

func Compare(a, b EvaluatedHand) int {
	if a.Rank != b.Rank {
		if a.Rank > b.Rank {
			return 1
		}
		return -1
	}
	for i := 0; i < len(a.Tiebreakers) && i < len(b.Tiebreakers); i++ {
		if a.Tiebreakers[i] != b.Tiebreakers[i] {
			if a.Tiebreakers[i] > b.Tiebreakers[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}

func evaluateFive(cardList []cards.Card) EvaluatedHand {
	sorted := make([]cards.Card, len(cardList))
	copy(sorted, cardList)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Rank != sorted[j].Rank {
			return highCardRank(sorted[i].Rank) > highCardRank(sorted[j].Rank)
		}
		return sorted[i].Suit > sorted[j].Suit
	})

	isFlush := true
	for i := 0; i < len(sorted)-1; i++ {
		if sorted[i].Suit != sorted[i+1].Suit {
			isFlush = false
			break
		}
	}

	isStraight, highCard := isStraight(sorted)
	isRoyal := isFlush && isStraight && highCard == 14

	if isRoyal {
		return EvaluatedHand{
			Rank:  RoyalFlush,
			Cards: toCardArray(sorted),
		}
	}
	if isFlush && isStraight {
		return EvaluatedHand{
			Rank:        StraightFlush,
			Tiebreakers: []int{highCard},
			Cards:       toCardArray(sorted),
		}
	}

	counts := rankCountsHighCard(sorted)
	for rank, count := range counts {
		if count == 4 {
			kicker := findKickerHighCard(sorted, []int{rank})
			return EvaluatedHand{
				Rank:        FourOfAKind,
				Tiebreakers: []int{rank, kicker},
				Cards:       toCardArray(sorted),
			}
		}
	}

	var three, pair1, pair2 int
	threes := 0
	pairs := 0
	for rank, count := range counts {
		switch count {
		case 3:
			three = rank
			threes++
		case 2:
			if pairs == 0 {
				pair1 = rank
			} else {
				pair2 = rank
			}
			pairs++
		}
	}

	if threes == 1 && pairs == 1 {
		return EvaluatedHand{
			Rank:        FullHouse,
			Tiebreakers: []int{three, pair1},
			Cards:       toCardArray(sorted),
		}
	}

	if isFlush {
		tiebreakers := toHighCardRanks(sorted)
		return EvaluatedHand{
			Rank:        Flush,
			Tiebreakers: tiebreakers,
			Cards:       toCardArray(sorted),
		}
	}

	if isStraight {
		return EvaluatedHand{
			Rank:        Straight,
			Tiebreakers: []int{highCard},
			Cards:       toCardArray(sorted),
		}
	}

	if threes == 1 {
		kickers := findKickersHighCard(sorted, []int{three}, 2)
		return EvaluatedHand{
			Rank:        ThreeOfAKind,
			Tiebreakers: []int{three, kickers[0], kickers[1]},
			Cards:       toCardArray(sorted),
		}
	}

	if pairs == 2 {
		var highPair, lowPair, kicker int
		if pair1 > pair2 {
			highPair, lowPair = pair1, pair2
		} else {
			highPair, lowPair = pair2, pair1
		}
		kicker = findKickerHighCard(sorted, []int{highPair, lowPair})
		return EvaluatedHand{
			Rank:        TwoPair,
			Tiebreakers: []int{highPair, lowPair, kicker},
			Cards:       toCardArray(sorted),
		}
	}

	if pairs == 1 {
		kickers := findKickersHighCard(sorted, []int{pair1}, 3)
		return EvaluatedHand{
			Rank:        OnePair,
			Tiebreakers: []int{pair1, kickers[0], kickers[1], kickers[2]},
			Cards:       toCardArray(sorted),
		}
	}

	tiebreakers := toHighCardRanks(sorted)
	return EvaluatedHand{
		Rank:        HighCard,
		Tiebreakers: tiebreakers,
		Cards:       toCardArray(sorted),
	}
}

func highCardRank(r cards.Rank) int {
	if r == cards.Ace {
		return 14
	}
	return int(r)
}

func toHighCardRanks(sorted []cards.Card) []int {
	ranks := make([]int, len(sorted))
	for i, c := range sorted {
		ranks[i] = highCardRank(c.Rank)
	}
	return ranks
}

func isStraight(sorted []cards.Card) (bool, int) {
	ranks := toHighCardRanks(sorted)

	if ranks[0] == 14 && ranks[1] == 5 && ranks[2] == 4 && ranks[3] == 3 && ranks[4] == 2 {
		return true, 5
	}

	for i := 0; i < len(ranks)-1; i++ {
		if ranks[i] != ranks[i+1]+1 {
			return false, 0
		}
	}
	return true, ranks[0]
}

func rankCountsHighCard(sorted []cards.Card) map[int]int {
	counts := make(map[int]int)
	for _, c := range sorted {
		counts[highCardRank(c.Rank)]++
	}
	return counts
}

func findKickerHighCard(sorted []cards.Card, used []int) int {
	for _, c := range sorted {
		rank := highCardRank(c.Rank)
		found := false
		for _, u := range used {
			if rank == u {
				found = true
				break
			}
		}
		if !found {
			return rank
		}
	}
	return 0
}

func findKickersHighCard(sorted []cards.Card, used []int, n int) []int {
	var kickers []int
	for _, c := range sorted {
		rank := highCardRank(c.Rank)
		found := false
		for _, u := range used {
			if rank == u {
				found = true
				break
			}
		}
		if !found {
			kickers = append(kickers, rank)
			if len(kickers) == n {
				break
			}
		}
	}
	return kickers
}

func toCardArray(sorted []cards.Card) [5]cards.Card {
	var arr [5]cards.Card
	for i := 0; i < 5; i++ {
		arr[i] = sorted[i]
	}
	return arr
}