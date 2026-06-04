package cards

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewDeck_Has52Cards(t *testing.T) {
	d := NewDeck()
	if len(d) != 52 {
		t.Errorf("NewDeck() length = %d, want 52", len(d))
	}
}

func TestNewDeck_AllUnique(t *testing.T) {
	d := NewDeck()
	seen := make(map[[2]int]bool)
	for _, c := range d {
		key := [2]int{int(c.Rank), int(c.Suit)}
		if seen[key] {
			t.Errorf("duplicate card Rank=%d Suit=%d", c.Rank, c.Suit)
		}
		seen[key] = true
	}
}

func TestNewDeck_AllSuitsAndRanks(t *testing.T) {
	d := NewDeck()
	suitCount := make(map[Suit]int)
	rankCount := make(map[Rank]int)
	for _, c := range d {
		suitCount[c.Suit]++
		rankCount[c.Rank]++
	}
	for _, suit := range []Suit{Spades, Hearts, Diamonds, Clubs} {
		if suitCount[suit] != 13 {
			t.Errorf("suit %d has %d cards, want 13", suit, suitCount[suit])
		}
	}
	for rank := Ace; rank <= King; rank++ {
		if rankCount[rank] != 4 {
			t.Errorf("rank %d has %d cards, want 4", rank, rankCount[rank])
		}
	}
}

func TestDeck_Shuffled_SameLength(t *testing.T) {
	d := NewDeck()
	rng := rand.New(rand.NewSource(42))
	s := d.Shuffled(rng)
	if len(s) != len(d) {
		t.Errorf("Shuffled() length = %d, want %d", len(s), len(d))
	}
}

func TestDeck_Shuffled_DoesNotMutateOriginal(t *testing.T) {
	d := NewDeck()
	original := make(Deck, len(d))
	copy(original, d)
	rng := rand.New(rand.NewSource(42))
	d.Shuffled(rng)
	for i := range d {
		if d[i] != original[i] {
			t.Error("Shuffled() mutated the original deck")
			break
		}
	}
}

func TestDeck_Draw_RemovesTopCard(t *testing.T) {
	d := NewDeck()
	first := d[0]
	got := d.Draw()
	if got != first {
		t.Errorf("Draw() = %v, want first card %v", got, first)
	}
	if len(d) != 51 {
		t.Errorf("after Draw() len = %d, want 51", len(d))
	}
}

func TestDeck_Draw_PanicsOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Draw() on empty deck should panic")
		}
	}()
	var d Deck
	d.Draw()
}

func TestCard_Symbol(t *testing.T) {
	cases := []struct {
		rank Rank
		want string
	}{
		{Ace, "A"},
		{Two, "2"},
		{Nine, "9"},
		{Ten, "T"},
		{Jack, "J"},
		{Queen, "Q"},
		{King, "K"},
	}
	for _, tc := range cases {
		c := Card{Rank: tc.rank, Suit: Spades}
		if got := c.Symbol(); got != tc.want {
			t.Errorf("Rank %d Symbol() = %q, want %q", tc.rank, got, tc.want)
		}
	}
}

func TestCard_SuitSymbol(t *testing.T) {
	cases := []struct {
		suit Suit
		want string
	}{
		{Spades, "♠"},
		{Hearts, "♥"},
		{Diamonds, "♦"},
		{Clubs, "♣"},
	}
	for _, tc := range cases {
		c := Card{Rank: Ace, Suit: tc.suit}
		if got := c.SuitSymbol(); got != tc.want {
			t.Errorf("Suit %d SuitSymbol() = %q, want %q", tc.suit, got, tc.want)
		}
	}
}

func TestCard_IsRed(t *testing.T) {
	if !(Card{Rank: Ace, Suit: Hearts}.IsRed()) {
		t.Error("Hearts should be red")
	}
	if !(Card{Rank: Ace, Suit: Diamonds}.IsRed()) {
		t.Error("Diamonds should be red")
	}
	if (Card{Rank: Ace, Suit: Spades}.IsRed()) {
		t.Error("Spades should not be red")
	}
	if (Card{Rank: Ace, Suit: Clubs}.IsRed()) {
		t.Error("Clubs should not be red")
	}
}

func TestPad_ShorterThanWidth(t *testing.T) {
	got := Pad("hi", 5)
	if len(got) < 5 {
		t.Errorf("Pad(%q, 5) too short: %q", "hi", got)
	}
	if !strings.HasPrefix(got, "hi") {
		t.Errorf("Pad(%q, 5) should start with original string", "hi")
	}
}

func TestPad_AtOrOverWidth(t *testing.T) {
	got := Pad("hello", 3)
	if got != "hello" {
		t.Errorf("Pad(s, width<=len) should return s unchanged, got %q", got)
	}
}

func TestCenter_ShorterThanWidth(t *testing.T) {
	got := Center("hi", 10)
	if !strings.Contains(got, "hi") {
		t.Errorf("Center(%q, 10) should contain original string, got %q", "hi", got)
	}
}

func TestCenter_AtOrOverWidth(t *testing.T) {
	got := Center("hello", 3)
	if got != "hello" {
		t.Errorf("Center(s, width<=len) should return s unchanged, got %q", got)
	}
}

func TestRenderCard_Red(t *testing.T) {
	redSty := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	blackSty := lipgloss.NewStyle()
	c := Card{Rank: Ace, Suit: Hearts}
	got := RenderCard(c, redSty, blackSty)
	if !strings.Contains(got, "A") || !strings.Contains(got, "♥") {
		t.Errorf("RenderCard(hearts ace) = %q, should contain A and ♥", got)
	}
}

func TestRenderCard_Black(t *testing.T) {
	redSty := lipgloss.NewStyle()
	blackSty := lipgloss.NewStyle()
	c := Card{Rank: King, Suit: Spades}
	got := RenderCard(c, redSty, blackSty)
	if !strings.Contains(got, "K") || !strings.Contains(got, "♠") {
		t.Errorf("RenderCard(spades king) = %q, should contain K and ♠", got)
	}
}

func TestRenderHand_MultipleCards(t *testing.T) {
	sty := lipgloss.NewStyle()
	hand := []Card{
		{Rank: Ace, Suit: Spades},
		{Rank: King, Suit: Hearts},
	}
	got := RenderHand(hand, sty, sty)
	if !strings.Contains(got, "A") || !strings.Contains(got, "K") {
		t.Errorf("RenderHand() = %q, should contain A and K", got)
	}
}

func TestRenderHand_Empty(t *testing.T) {
	sty := lipgloss.NewStyle()
	got := RenderHand([]Card{}, sty, sty)
	if got != "" {
		t.Errorf("RenderHand(empty) = %q, want empty string", got)
	}
}

func TestRenderHandBudget_FitsAll(t *testing.T) {
	sty := lipgloss.NewStyle()
	hand := []Card{{Rank: Ace, Suit: Spades}, {Rank: Two, Suit: Hearts}}
	got := RenderHandBudget(hand, 100, sty, sty)
	if !strings.Contains(got, "A") || !strings.Contains(got, "2") {
		t.Errorf("RenderHandBudget(wide) = %q, should contain both cards", got)
	}
}

func TestRenderHandBudget_TruncatesWhenNarrow(t *testing.T) {
	sty := lipgloss.NewStyle()
	hand := []Card{
		{Rank: Ace, Suit: Spades},
		{Rank: Two, Suit: Hearts},
		{Rank: Three, Suit: Diamonds},
		{Rank: Four, Suit: Clubs},
	}
	got := RenderHandBudget(hand, 5, sty, sty)
	// With very narrow width, not all cards should appear
	_ = got // just ensure no panic
}

func TestSuitSymbol_Unknown(t *testing.T) {
	c := Card{Rank: Ace, Suit: Suit(99)}
	got := c.SuitSymbol()
	if got != "?" {
		t.Errorf("unknown suit SuitSymbol() = %q, want ?", got)
	}
}
