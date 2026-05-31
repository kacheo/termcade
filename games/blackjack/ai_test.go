package blackjack

import "testing"

func TestShouldHit(t *testing.T) {
	cases := []struct {
		name string
		hand Hand
		want bool
	}{
		{"hard 16 hits", Hand{{Ten, Spades}, {Six, Hearts}}, true},
		{"hard 17 stands", Hand{{Ten, Spades}, {Seven, Hearts}}, false},
		{"hard 18 stands", Hand{{Ten, Spades}, {Eight, Hearts}}, false},
		{"hard 12 hits", Hand{{Two, Spades}, {Ten, Hearts}}, true},
		{"soft 17 hits", Hand{{Ace, Spades}, {Six, Hearts}}, true},
		{"soft 18 stands", Hand{{Ace, Spades}, {Seven, Hearts}}, false},
		{"soft 16 hits", Hand{{Ace, Spades}, {Five, Hearts}}, true},
		{"blackjack stands", Hand{{Ace, Spades}, {King, Hearts}}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ShouldHit(c.hand); got != c.want {
				t.Errorf("ShouldHit = %v, want %v", got, c.want)
			}
		})
	}
}
