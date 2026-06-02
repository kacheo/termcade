package cards

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type Suit int

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

type Rank int

const (
	Ace Rank = iota + 1
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

type Card struct {
	Rank Rank
	Suit Suit
}

func (c Card) Symbol() string {
	switch c.Rank {
	case Ace:
		return "A"
	case Ten:
		return "T"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	default:
		return fmt.Sprintf("%d", int(c.Rank))
	}
}

func (c Card) SuitSymbol() string {
	switch c.Suit {
	case Spades:
		return "♠"
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	}
	return "?"
}

func (c Card) IsRed() bool {
	return c.Suit == Hearts || c.Suit == Diamonds
}

type Deck []Card

func NewDeck() Deck {
	var d Deck
	for suit := Spades; suit <= Clubs; suit++ {
		for rank := Ace; rank <= King; rank++ {
			d = append(d, Card{Rank: rank, Suit: suit})
		}
	}
	return d
}

func (d Deck) Shuffled(r *rand.Rand) Deck {
	out := make(Deck, len(d))
	copy(out, d)
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func (d *Deck) Draw() Card {
	if len(*d) == 0 {
		panic("cards: draw from empty deck")
	}
	c := (*d)[0]
	*d = (*d)[1:]
	return c
}

func Pad(s string, width int) string {
	vis := runewidth.StringWidth(ansi.Strip(s))
	if vis >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vis)
}

func Center(s string, width int) string {
	vis := runewidth.StringWidth(ansi.Strip(s))
	if vis >= width {
		return s
	}
	l := (width - vis) / 2
	return strings.Repeat(" ", l) + s + strings.Repeat(" ", width-vis-l)
}

func RenderCard(c Card, redSty, blackSty lipgloss.Style) string {
	s := fmt.Sprintf("[%s%s]", c.Symbol(), c.SuitSymbol())
	if c.IsRed() {
		return redSty.Render(s)
	}
	return blackSty.Render(s)
}

func RenderHand(hand []Card, redSty, blackSty lipgloss.Style) string {
	parts := make([]string, len(hand))
	for i, c := range hand {
		parts[i] = RenderCard(c, redSty, blackSty)
	}
	return strings.Join(parts, " ")
}

func RenderHandBudget(hand []Card, maxWidth int, redSty, blackSty lipgloss.Style) string {
	ellipsis := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("…")
	var parts []string
	used := 0
	for _, c := range hand {
		sep := 0
		if len(parts) > 0 {
			sep = 1
		}
		rendered := RenderCard(c, redSty, blackSty)
		w := lipgloss.Width(rendered)
		if used+sep+w > maxWidth {
			if len(parts) > 0 {
				parts = append(parts, ellipsis)
			}
			break
		}
		used += sep + w
		parts = append(parts, rendered)
	}
	return strings.Join(parts, " ")
}