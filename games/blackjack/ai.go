package blackjack

// ShouldHit returns true if the AI should draw. Hits hard <17 and soft 17.
func ShouldHit(hand Hand) bool {
	v := hand.Value()
	if v < 17 {
		return true
	}
	if v == 17 && hand.IsSoft() {
		return true
	}
	return false
}
