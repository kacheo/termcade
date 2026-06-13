package blackjack

import "github.com/kacheo/tmvgs/games/cards"

type Hand []cards.Card

func cardBaseValue(r cards.Rank) int {
	if r >= cards.Ten {
		return 10
	}
	return int(r)
}

func (h Hand) BaseValue() int {
	return cardBaseValue(h[0].Rank)
}

func (h Hand) Value() int {
	total, aces := 0, 0
	for _, c := range h {
		if c.Rank == cards.Ace {
			aces++
			total += 11
		} else {
			total += cardBaseValue(c.Rank)
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

func (h Hand) IsBust() bool { return h.Value() > 21 }

func (h Hand) IsBlackjack() bool { return len(h) == 2 && h.Value() == 21 }

func (h Hand) IsSoft() bool {
	total, aces := 0, 0
	for _, c := range h {
		if c.Rank == cards.Ace {
			aces++
			total += 11
		} else {
			total += cardBaseValue(c.Rank)
		}
	}
	reduced := 0
	for total > 21 && reduced < aces {
		total -= 10
		reduced++
	}
	return aces > reduced && total <= 21
}