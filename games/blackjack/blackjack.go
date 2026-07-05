package blackjack

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	cardpkg "github.com/kacheo/termcade/games/cards"
)

type phase int

const (
	phaseDealing phase = iota
	phaseTurn
	phaseDealerTurn
	phaseResults
)

type playerStatus int

const (
	statusPlaying playerStatus = iota
	statusStand
	statusBust
	statusBlackjack
)

type tablePlayer struct {
	name   string
	hand   Hand
	status playerStatus
	result string // "WIN", "LOSE", "PUSH", or ""
}

type Blackjack struct {
	rng     *rand.Rand
	deck    cardpkg.Deck
	dealer  Hand
	player  tablePlayer
	phase   phase
	elapsed time.Duration
	wins    int
	rounds  int
}

const (
	dealDelay   = 600 * time.Millisecond
	dealerDelay = 700 * time.Millisecond
)

func NewBlackjack() *Blackjack {
	b := &Blackjack{rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
	b.startRound()
	return b
}

func (b *Blackjack) startRound() {
	b.deck = cardpkg.NewDeck().Shuffled(b.rng)
	b.dealer = Hand{}
	b.player = tablePlayer{name: "YOU"}
	for i := 0; i < 2; i++ {
		b.player.hand = append(b.player.hand, b.deck.Draw())
		b.dealer = append(b.dealer, b.deck.Draw())
	}
	if b.player.hand.IsBlackjack() {
		b.player.status = statusBlackjack
	}
	b.phase = phaseDealing
	b.elapsed = 0
	b.rounds++
}

func (b *Blackjack) Name() string        { return "Blackjack" }
func (b *Blackjack) Description() string { return "Beat the dealer. Hit or stand." }
func (b *Blackjack) IsPaused() bool      { return false }
func (b *Blackjack) IsGameOver() bool    { return false }
func (b *Blackjack) GetScore() int       { return b.wins }
func (b *Blackjack) GetLevel() int       { return 0 }
func (b *Blackjack) GetLines() int       { return b.rounds }

func (b *Blackjack) Update(delta time.Duration) error {
	b.elapsed += delta
	switch b.phase {
	case phaseDealing:
		if b.elapsed >= dealDelay {
			b.elapsed = 0
			b.transitionFromDealing()
		}
	case phaseDealerTurn:
		if b.elapsed >= dealerDelay {
			b.elapsed = 0
			b.stepDealer()
		}
	}
	return nil
}

func (b *Blackjack) transitionFromDealing() {
	if b.player.status == statusPlaying {
		b.phase = phaseTurn
	} else {
		b.phase = phaseDealerTurn
		b.elapsed = 0
	}
}

func (b *Blackjack) stepDealer() {
	dv := b.dealer.Value()
	if dv < 17 || (dv == 17 && b.dealer.IsSoft()) {
		b.dealer = append(b.dealer, b.deck.Draw())
	} else {
		b.evaluateResults()
		b.phase = phaseResults
	}
}

func (b *Blackjack) evaluateResults() {
	dv := b.dealer.Value()
	dealerBust := b.dealer.IsBust()
	dealerBJ := b.dealer.IsBlackjack()
	p := &b.player
	switch p.status {
	case statusBust:
		p.result = "LOSE"
	case statusBlackjack:
		if dealerBJ {
			p.result = "PUSH"
		} else {
			p.result = "WIN"
			b.wins++
		}
	default:
		pv := p.hand.Value()
		if dealerBust || pv > dv {
			p.result = "WIN"
			b.wins++
		} else if pv == dv {
			p.result = "PUSH"
		} else {
			p.result = "LOSE"
		}
	}
}

func (b *Blackjack) HandleInput(key string) {
	switch b.phase {
	case phaseTurn:
		if b.player.status != statusPlaying {
			return
		}
		switch key {
		case "h", "left":
			b.player.hand = append(b.player.hand, b.deck.Draw())
			if b.player.hand.IsBust() {
				b.player.status = statusBust
			}
			b.phase = phaseDealerTurn
			b.elapsed = 0
		case "s", "right", "down":
			b.player.status = statusStand
			b.phase = phaseDealerTurn
			b.elapsed = 0
		}
	case phaseResults:
		if key == "enter" || key == " " {
			b.startRound()
		}
	}
}

var (
	bjBorderSty    = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	bjTitleSty     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
	bjRedCardSty   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	bjWhiteCardSty = lipgloss.NewStyle().Foreground(lipgloss.Color("#EEEEEE"))
	bjHiddenSty    = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	bjLabelSty     = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	bjActiveSty    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")).Bold(true)
	bjTextSty      = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	bjWinSty       = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	bjLoseSty      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
	bjPushSty      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	bjBustSty      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
	bjBJSty        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
	bjActionSty    = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Background(lipgloss.Color("#333333"))
	bjHotSty       = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#00FF88")).Bold(true)
)

const bjInnerWidth = 46

func bjRenderCard(c cardpkg.Card) string {
	return cardpkg.RenderCard(c, bjRedCardSty, bjWhiteCardSty)
}

func bjRenderHidden() string { return bjHiddenSty.Render("[??]") }

// bjRenderHandMasked returns (cards, valueLabel).
// revealed=false: first card shown, rest as [??], value shown as "(X+?)".
// revealed=true:  all cards shown, value as "(N)".
func bjRenderHandMasked(hand Hand, revealed bool) (string, string) {
	if revealed || len(hand) == 0 {
		return bjRenderHand(hand), bjLabelSty.Render(fmt.Sprintf(" (%d)", hand.Value()))
	}
	cards := bjRenderCard(hand[0])
	for range hand[1:] {
		cards += " " + bjRenderHidden()
	}
	showing := int(hand[0].Rank)
	if hand[0].Rank >= cardpkg.Ten {
		showing = 10
	} else if hand[0].Rank == cardpkg.Ace {
		showing = 11
	}
	return cards, bjLabelSty.Render(fmt.Sprintf(" (%d+?)", showing))
}

func bjRenderHand(hand Hand) string {
	return cardpkg.RenderHand(hand[:], bjRedCardSty, bjWhiteCardSty)
}

// bjRenderHandBudget renders as many cards as fit in maxWidth visible chars,
// appending "…" if cards are truncated.
func bjRenderHandBudget(hand Hand, maxWidth int) string {
	return cardpkg.RenderHandBudget(hand[:], maxWidth, bjRedCardSty, bjWhiteCardSty)
}

func bjPad(s string, width int) string {
	return cardpkg.Pad(s, width)
}

func bjCenter(s string, width int) string {
	return cardpkg.Center(s, width)
}

func (b *Blackjack) Render() string {
	revealed := b.phase == phaseDealerTurn || b.phase == phaseResults
	brd := func(s string) string { return bjBorderSty.Render(s) }
	row := func(content string) string {
		return brd("║") + bjPad(content, bjInnerWidth) + brd("║") + "\n"
	}
	blank := func() string { return row("") }

	var sb strings.Builder
	sb.WriteString(brd("╔"+strings.Repeat("═", bjInnerWidth)+"╗") + "\n")
	sb.WriteString(row(bjCenter(bjTitleSty.Render("BLACKJACK"), bjInnerWidth)))
	sb.WriteString(brd("╠"+strings.Repeat("═", bjInnerWidth)+"╣") + "\n")

	// Dealer
	dealerCards, dealerVal := bjRenderHandMasked(b.dealer, revealed)
	sb.WriteString(row(" " + bjTextSty.Render("DEALER") + "  " + dealerCards + dealerVal))

	sb.WriteString(brd("╠"+strings.Repeat("═", bjInnerWidth)+"╣") + "\n")
	sb.WriteString(row(b.renderPlayerRow(b.phase == phaseTurn)))
	sb.WriteString(blank())

	switch b.phase {
	case phaseTurn:
		actions := "  " + bjHotSty.Render(" H-Hit ") + "   " + bjActionSty.Render(" S-Stand ")
		sb.WriteString(row(actions))
	case phaseResults:
		sb.WriteString(row("  " + bjPushSty.Render("Press ENTER / SPACE for next round")))
	case phaseDealerTurn:
		sb.WriteString(row("  " + bjLabelSty.Render("Dealer's turn...")))
	default:
		sb.WriteString(blank())
	}

	sb.WriteString(blank())
	sb.WriteString(row("  " + bjLabelSty.Render(fmt.Sprintf("Wins: %d   Rounds: %d", b.wins, b.rounds))))
	sb.WriteString(brd("╚" + strings.Repeat("═", bjInnerWidth) + "╝"))
	return sb.String()
}

func (b *Blackjack) renderPlayerRow(active bool) string {
	p := &b.player
	nameSty := bjTextSty
	if active {
		nameSty = bjActiveSty
	}
	name := nameSty.Render(fmt.Sprintf("%-5s", p.name))

	val := bjLabelSty.Render(fmt.Sprintf(" (%d)", p.hand.Value()))

	statusStr := ""
	switch p.status {
	case statusBust:
		statusStr = "  " + bjBustSty.Render("BUST")
	case statusStand:
		statusStr = "  " + bjLabelSty.Render("STAND")
	case statusBlackjack:
		statusStr = "  " + bjBJSty.Render("BLACKJACK!")
	}
	switch p.result {
	case "WIN":
		statusStr += "  " + bjWinSty.Render("WIN")
	case "LOSE":
		statusStr += "  " + bjLoseSty.Render("LOSE")
	case "PUSH":
		statusStr += "  " + bjPushSty.Render("PUSH")
	}

	const prefixW = 8
	suffixW := lipgloss.Width(val) + lipgloss.Width(statusStr)
	cardBudget := bjInnerWidth - prefixW - suffixW
	cards := bjRenderHandBudget(p.hand, cardBudget)

	return " " + name + "  " + cards + val + statusStr
}
