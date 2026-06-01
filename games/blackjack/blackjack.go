package blackjack

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phaseDealing    phase = iota
	phaseAITurn
	phasePlayerTurn
	phaseDealerTurn
	phaseResults
)

type playerStatus int

const (
	statusPlaying   playerStatus = iota
	statusStand
	statusBust
	statusBlackjack
)

type tablePlayer struct {
	name   string
	hand   Hand
	status playerStatus
	isAI   bool
	result string // "WIN", "LOSE", "PUSH", or ""
}

type Blackjack struct {
	rng      *rand.Rand
	deck     Deck
	dealer   Hand
	players  []*tablePlayer // [0]=human, [1..]=AI
	phase    phase
	aiIdx    int
	elapsed  time.Duration
	wins   int
	rounds int
}

const (
	dealDelay   = 600 * time.Millisecond
	aiStepDelay = 500 * time.Millisecond
	dealerDelay = 700 * time.Millisecond
)

func NewBlackjack(aiCount int) *Blackjack {
	if aiCount < 0 {
		aiCount = 0
	}
	if aiCount > 3 {
		aiCount = 3
	}
	b := &Blackjack{rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
	b.players = append(b.players, &tablePlayer{name: "YOU", isAI: false})
	for _, name := range []string{"AI-1", "AI-2", "AI-3"}[:aiCount] {
		b.players = append(b.players, &tablePlayer{name: name, isAI: true})
	}
	b.startRound()
	return b
}

func (b *Blackjack) startRound() {
	b.deck = NewDeck().Shuffled(b.rng)
	b.dealer = Hand{}
	for _, p := range b.players {
		p.hand = Hand{}
		p.status = statusPlaying
		p.result = ""
	}
	for i := 0; i < 2; i++ {
		for _, p := range b.players {
			p.hand = append(p.hand, b.deck.Draw())
		}
		b.dealer = append(b.dealer, b.deck.Draw())
	}
	for _, p := range b.players {
		if p.hand.IsBlackjack() {
			p.status = statusBlackjack
		}
	}
	b.phase = phaseDealing
	b.elapsed = 0
	b.aiIdx = 1
	b.rounds++
}

func (b *Blackjack) Name() string        { return "Blackjack" }
func (b *Blackjack) Description() string { return "Beat the dealer. Hit or stand." }
func (b *Blackjack) IsPaused() bool   { return false }
func (b *Blackjack) IsGameOver() bool { return false }
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
	case phaseAITurn:
		if b.elapsed >= aiStepDelay {
			b.elapsed = 0
			b.stepAI()
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
	if len(b.players) > 1 {
		b.aiIdx = 1
		b.phase = phaseAITurn
		b.skipDoneAIs()
	} else {
		b.phase = phasePlayerTurn
	}
}

func (b *Blackjack) skipDoneAIs() {
	for b.aiIdx < len(b.players) && b.players[b.aiIdx].status != statusPlaying {
		b.aiIdx++
	}
	if b.aiIdx >= len(b.players) {
		if b.players[0].status == statusPlaying {
			b.phase = phasePlayerTurn
		} else {
			b.phase = phaseDealerTurn
			b.elapsed = 0
		}
	}
}

func (b *Blackjack) stepAI() {
	if b.aiIdx >= len(b.players) {
		return
	}
	ai := b.players[b.aiIdx]
	if ShouldHit(ai.hand) {
		ai.hand = append(ai.hand, b.deck.Draw())
		if ai.hand.IsBust() {
			ai.status = statusBust
			b.aiIdx++
			b.skipDoneAIs()
			return
		}
		// Non-bust hit: AI gets another tick to decide again
	} else {
		ai.status = statusStand
		b.aiIdx++
		b.skipDoneAIs()
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
	for _, p := range b.players {
		switch p.status {
		case statusBust:
			p.result = "LOSE"
		case statusBlackjack:
			if dealerBJ {
				p.result = "PUSH"
			} else {
				p.result = "WIN"
				if !p.isAI {
					b.wins++
				}
			}
		default:
			pv := p.hand.Value()
			if dealerBust || pv > dv {
				p.result = "WIN"
				if !p.isAI {
					b.wins++
				}
			} else if pv == dv {
				p.result = "PUSH"
			} else {
				p.result = "LOSE"
			}
		}
	}
}

func (b *Blackjack) HandleInput(key string) {
	switch b.phase {
	case phasePlayerTurn:
		human := b.players[0]
		if human.status != statusPlaying {
			return
		}
		switch key {
		case "h", "left":
			human.hand = append(human.hand, b.deck.Draw())
			if human.hand.IsBust() {
				human.status = statusBust
				b.phase = phaseDealerTurn
				b.elapsed = 0
			}
		case "s", "right", "down":
			human.status = statusStand
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

func bjRenderCard(c Card) string {
	s := fmt.Sprintf("[%s%s]", c.Symbol(), c.SuitSymbol())
	if c.IsRed() {
		return bjRedCardSty.Render(s)
	}
	return bjWhiteCardSty.Render(s)
}

func bjRenderHidden() string { return bjHiddenSty.Render("[??]") }

func bjRenderHand(hand Hand) string {
	parts := make([]string, len(hand))
	for i, c := range hand {
		parts[i] = bjRenderCard(c)
	}
	return strings.Join(parts, " ")
}

// bjRenderHandBudget renders as many cards as fit in maxWidth visible chars,
// appending "…" if cards are truncated.
func bjRenderHandBudget(hand Hand, maxWidth int) string {
	ellipsis := bjLabelSty.Render("…")
	var parts []string
	used := 0
	for _, c := range hand {
		sep := 0
		if len(parts) > 0 {
			sep = 1
		}
		rendered := bjRenderCard(c)
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

func bjPad(s string, width int) string {
	vis := lipgloss.Width(s)
	if vis >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vis)
}

func bjCenter(s string, width int) string {
	vis := lipgloss.Width(s)
	if vis >= width {
		return s
	}
	l := (width - vis) / 2
	return strings.Repeat(" ", l) + s + strings.Repeat(" ", width-vis-l)
}

func (b *Blackjack) Render() string {
	revealed := b.phase == phaseDealerTurn || b.phase == phaseResults
	brd := func(s string) string { return bjBorderSty.Render(s) }
	row := func(content string) string {
		return brd("║") + bjPad(content, bjInnerWidth) + brd("║\n")
	}
	blank := func() string { return row("") }

	var sb strings.Builder
	sb.WriteString(brd("╔" + strings.Repeat("═", bjInnerWidth) + "╗\n"))
	sb.WriteString(row(bjCenter(bjTitleSty.Render("BLACKJACK"), bjInnerWidth)))
	sb.WriteString(brd("╠" + strings.Repeat("═", bjInnerWidth) + "╣\n"))

	// Dealer
	var dealerCards string
	if revealed {
		dealerCards = bjRenderHand(b.dealer)
	} else if len(b.dealer) > 0 {
		dealerCards = bjRenderCard(b.dealer[0])
		for range b.dealer[1:] {
			dealerCards += " " + bjRenderHidden()
		}
	}
	dealerVal := ""
	if len(b.dealer) > 0 {
		if revealed {
			dealerVal = bjLabelSty.Render(fmt.Sprintf(" (%d)", b.dealer.Value()))
		} else {
			showing := b.dealer[0].BaseValue()
			if b.dealer[0].Rank == Ace {
				showing = 11
			}
			dealerVal = bjLabelSty.Render(fmt.Sprintf(" (%d+?)", showing))
		}
	}
	sb.WriteString(row(" " + bjTextSty.Render("DEALER") + "  " + dealerCards + dealerVal))

	// AI players
	if len(b.players) > 1 {
		sb.WriteString(brd("╠" + strings.Repeat("═", bjInnerWidth) + "╣\n"))
		for i := 1; i < len(b.players); i++ {
			active := b.phase == phaseAITurn && b.aiIdx == i
			sb.WriteString(row(b.renderPlayerRow(b.players[i], active)))
		}
	}

	sb.WriteString(brd("╠" + strings.Repeat("═", bjInnerWidth) + "╣\n"))
	sb.WriteString(row(b.renderPlayerRow(b.players[0], b.phase == phasePlayerTurn)))
	sb.WriteString(blank())

	switch b.phase {
	case phasePlayerTurn:
		actions := "  " + bjHotSty.Render(" H-Hit ") + "   " + bjActionSty.Render(" S-Stand ")
		sb.WriteString(row(actions))
	case phaseResults:
		sb.WriteString(row("  " + bjPushSty.Render("Press ENTER / SPACE for next round")))
	case phaseAITurn:
		sb.WriteString(row("  " + bjLabelSty.Render("Waiting for AI...")))
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

func (b *Blackjack) renderPlayerRow(p *tablePlayer, active bool) string {
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

	// " " + 5-char name + "  " = 8 visible chars prefix; reserve space for val+status
	const prefixW = 8
	suffixW := lipgloss.Width(val) + lipgloss.Width(statusStr)
	cardBudget := bjInnerWidth - prefixW - suffixW
	cards := bjRenderHandBudget(p.hand, cardBudget)

	return " " + name + "  " + cards + val + statusStr
}
