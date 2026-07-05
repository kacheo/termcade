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
	phaseBetting phase = iota
	phaseDealing
	phaseInsurance
	phaseTurn
	phaseDealerTurn
	phaseResults
)

type handStatus int

const (
	statusPlaying handStatus = iota
	statusStand
	statusBust
	statusBlackjack
)

// playerHand is one hand the player is playing this round. There are two of
// these only after a split; every other round has exactly one.
type playerHand struct {
	hand          Hand
	status        handStatus
	result        string // "WIN", "LOSE", "PUSH", or ""
	bet           int
	isDoubled     bool
	fromSplitAces bool // one card only, forced stand — standard split-Aces rule
}

const (
	startingBankroll = 1000
	minBet           = 10
	betStep          = 10

	// insuranceOptimalTrueCount is the standard Hi-Lo threshold above which
	// taking even-money insurance is the correct (positive-EV) play.
	insuranceOptimalTrueCount = 3.0
)

type Blackjack struct {
	rng     *rand.Rand
	shoe    *Shoe
	dealer  Hand
	hands   []playerHand
	active  int
	phase   phase
	elapsed time.Duration
	wins    int
	rounds  int

	bankroll int
	bet      int

	dealerHoleCounted bool

	insuranceOffered             bool
	insuranceTaken               bool
	insuranceBet                 int
	insuranceResult              string
	insuranceTrueCountAtDecision float64
	insuranceCorrectCount        int
	insuranceTotalCount          int

	showCount bool
	gameOver  bool
}

const (
	dealDelay   = 600 * time.Millisecond
	dealerDelay = 700 * time.Millisecond
)

// NewBlackjack creates a new game with a shoe of numDecks decks. A numDecks
// of 0 or less defaults to 6 (a standard Vegas shoe game).
func NewBlackjack(numDecks int) *Blackjack {
	if numDecks <= 0 {
		numDecks = 6
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := &Blackjack{
		rng:      rng,
		shoe:     NewShoe(numDecks, rng),
		bankroll: startingBankroll,
		bet:      minBet,
	}
	b.enterBetting()
	return b
}

func (b *Blackjack) enterBetting() {
	if b.bet > b.bankroll {
		b.bet = b.bankroll
	}
	if b.bet < minBet {
		b.bet = minBet
	}
	b.phase = phaseBetting
	b.elapsed = 0
}

func (b *Blackjack) startRound() {
	if b.shoe.NeedsReshuffle() {
		b.shoe.Reshuffle()
	}
	b.dealerHoleCounted = false
	b.insuranceOffered = false
	b.insuranceTaken = false
	b.insuranceBet = 0
	b.insuranceResult = ""

	b.dealer = Hand{}
	b.hands = []playerHand{{bet: b.bet}}
	b.active = 0

	for i := 0; i < 2; i++ {
		pc := b.shoe.Draw()
		b.hands[0].hand = append(b.hands[0].hand, pc)
		b.shoe.CountCard(pc)
		b.dealer = append(b.dealer, b.shoe.Draw())
	}
	// The dealer's up-card (first card dealt) is visible immediately; the
	// hole card (second) is deferred until it's actually revealed.
	b.shoe.CountCard(b.dealer[0])

	if b.hands[0].hand.IsBlackjack() {
		b.hands[0].status = statusBlackjack
	}
	b.phase = phaseDealing
	b.elapsed = 0
	b.rounds++
}

func (b *Blackjack) Name() string { return "Blackjack" }
func (b *Blackjack) Description() string {
	return "Beat the dealer. Bet, hit, stand, double, or split — closest to 21 wins."
}
func (b *Blackjack) IsPaused() bool   { return false }
func (b *Blackjack) IsGameOver() bool { return b.gameOver }
func (b *Blackjack) GetScore() int    { return b.bankroll }
func (b *Blackjack) GetLevel() int    { return b.shoe.numDecks }
func (b *Blackjack) GetLines() int    { return b.rounds }

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
	if b.dealer[0].Rank == cardpkg.Ace {
		b.insuranceOffered = true
		b.insuranceTrueCountAtDecision = b.shoe.TrueCount()
		b.phase = phaseInsurance
		return
	}
	b.afterInsuranceDecision()
}

func (b *Blackjack) afterInsuranceDecision() {
	if b.hands[0].status == statusBlackjack {
		b.enterDealerTurn()
		return
	}
	b.phase = phaseTurn
	b.active = 0
}

// enterDealerTurn moves into the dealer's turn, revealing (and, exactly
// once, counting) the hole card at the moment it actually becomes visible.
func (b *Blackjack) enterDealerTurn() {
	if !b.dealerHoleCounted {
		b.shoe.CountCard(b.dealer[1])
		b.dealerHoleCounted = true
	}
	b.phase = phaseDealerTurn
	b.elapsed = 0
}

func (b *Blackjack) stepDealer() {
	dv := b.dealer.Value()
	if dv < 17 || (dv == 17 && b.dealer.IsSoft()) {
		c := b.shoe.Draw()
		b.shoe.CountCard(c)
		b.dealer = append(b.dealer, c)
	} else {
		b.evaluateResults()
		b.phase = phaseResults
	}
}

func (b *Blackjack) evaluateResults() {
	dv := b.dealer.Value()
	dealerBust := b.dealer.IsBust()
	dealerBJ := b.dealer.IsBlackjack()

	for i := range b.hands {
		h := &b.hands[i]
		switch h.status {
		case statusBust:
			h.result = "LOSE"
		case statusBlackjack:
			if dealerBJ {
				h.result = "PUSH"
			} else {
				h.result = "WIN"
			}
		default:
			pv := h.hand.Value()
			switch {
			case dealerBust || pv > dv:
				h.result = "WIN"
			case pv == dv:
				h.result = "PUSH"
			default:
				h.result = "LOSE"
			}
		}
		b.settleHand(h)
	}

	if b.insuranceOffered {
		b.settleInsurance(dealerBJ)
	}

	if b.bankroll < minBet {
		b.gameOver = true
	}
}

func (b *Blackjack) settleHand(h *playerHand) {
	switch h.result {
	case "WIN":
		b.wins++
		payout := h.bet
		if h.status == statusBlackjack {
			payout = h.bet * 3 / 2
		}
		b.bankroll += h.bet + payout
	case "PUSH":
		b.bankroll += h.bet
	}
}

func (b *Blackjack) settleInsurance(dealerBJ bool) {
	if !b.insuranceTaken {
		return
	}
	if dealerBJ {
		b.insuranceResult = "WIN"
		b.bankroll += b.insuranceBet * 3 // original insurance bet + 2:1 payout
	} else {
		b.insuranceResult = "LOSE"
	}
}

func (b *Blackjack) recordInsuranceDecision() {
	optimal := b.insuranceTrueCountAtDecision >= insuranceOptimalTrueCount
	b.insuranceTotalCount++
	if b.insuranceTaken == optimal {
		b.insuranceCorrectCount++
	}
}

func (b *Blackjack) resolveInsuranceReveal() {
	if b.dealer.IsBlackjack() {
		b.enterDealerTurn()
		b.evaluateResults()
		b.phase = phaseResults
		return
	}
	b.afterInsuranceDecision()
}

func (b *Blackjack) canDouble(h *playerHand) bool {
	return len(h.hand) == 2 && !h.isDoubled && !h.fromSplitAces && b.bankroll >= h.bet
}

func (b *Blackjack) canSplit(h *playerHand) bool {
	if len(b.hands) != 1 || len(h.hand) != 2 {
		return false
	}
	if cardBaseValue(h.hand[0].Rank) != cardBaseValue(h.hand[1].Rank) {
		return false
	}
	return b.bankroll >= h.bet
}

func (b *Blackjack) performDouble(h *playerHand) {
	b.bankroll -= h.bet
	h.bet *= 2
	h.isDoubled = true
	c := b.shoe.Draw()
	b.shoe.CountCard(c)
	h.hand = append(h.hand, c)
	if h.hand.IsBust() {
		h.status = statusBust
	} else {
		h.status = statusStand
	}
	b.advanceOrDealer()
}

func (b *Blackjack) performSplit() {
	h := b.hands[0]
	b.bankroll -= h.bet
	isAces := h.hand[0].Rank == cardpkg.Ace

	hand1 := playerHand{hand: Hand{h.hand[0]}, bet: h.bet}
	hand2 := playerHand{hand: Hand{h.hand[1]}, bet: h.bet}

	d1 := b.shoe.Draw()
	b.shoe.CountCard(d1)
	hand1.hand = append(hand1.hand, d1)

	d2 := b.shoe.Draw()
	b.shoe.CountCard(d2)
	hand2.hand = append(hand2.hand, d2)

	if isAces {
		// Standard rule: split Aces get exactly one card each and cannot be
		// hit further, regardless of the resulting total.
		hand1.status = statusStand
		hand1.fromSplitAces = true
		hand2.status = statusStand
		hand2.fromSplitAces = true
	}
	// Note: a split hand reaching 21 is a plain 21, not a "natural" blackjack
	// — standard rule, no 3:2 bonus — so status stays statusPlaying/Stand,
	// never statusBlackjack, for split hands.

	b.hands = []playerHand{hand1, hand2}
	b.active = 0
	if b.hands[0].status != statusPlaying {
		b.advanceOrDealer()
	}
}

// advanceOrDealer moves to the next hand still in play, or to the dealer's
// turn once every hand is resolved.
func (b *Blackjack) advanceOrDealer() {
	for i := b.active + 1; i < len(b.hands); i++ {
		if b.hands[i].status == statusPlaying {
			b.active = i
			return
		}
	}
	b.enterDealerTurn()
}

func (b *Blackjack) HandleInput(key string) {
	if key == "c" {
		b.showCount = !b.showCount
		return
	}
	switch b.phase {
	case phaseBetting:
		switch key {
		case "left", "h":
			if b.bet-betStep >= minBet {
				b.bet -= betStep
			}
		case "right", "l":
			if b.bet+betStep <= b.bankroll {
				b.bet += betStep
			}
		case "enter", " ":
			b.bankroll -= b.bet
			b.startRound()
		}
	case phaseInsurance:
		switch key {
		case "y", "i":
			b.insuranceTaken = true
			b.insuranceBet = b.hands[0].bet / 2
			b.bankroll -= b.insuranceBet
		case "n", "s":
			b.insuranceTaken = false
		default:
			return
		}
		b.recordInsuranceDecision()
		b.resolveInsuranceReveal()
	case phaseTurn:
		h := &b.hands[b.active]
		if h.status != statusPlaying {
			return
		}
		switch key {
		case "h", "left":
			c := b.shoe.Draw()
			b.shoe.CountCard(c)
			h.hand = append(h.hand, c)
			if h.hand.IsBust() {
				h.status = statusBust
				b.advanceOrDealer()
			}
		case "s", "right", "down":
			h.status = statusStand
			b.advanceOrDealer()
		case "d":
			if b.canDouble(h) {
				b.performDouble(h)
			}
		case "x":
			if b.canSplit(h) {
				b.performSplit()
			}
		}
	case phaseResults:
		if key == "enter" || key == " " {
			b.enterBetting()
		}
	}
}

var (
	bjBorderSty        = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	bjTitleSty         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
	bjRedCardSty       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	bjWhiteCardSty     = lipgloss.NewStyle().Foreground(lipgloss.Color("#EEEEEE"))
	bjHiddenSty        = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	bjLabelSty         = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	bjActiveSty        = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")).Bold(true)
	bjTextSty          = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	bjWinSty           = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	bjLoseSty          = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
	bjPushSty          = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	bjBustSty          = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
	bjBJSty            = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
	bjActionSty        = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Background(lipgloss.Color("#333333"))
	bjHotSty           = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#00FF88")).Bold(true)
	bjBankrollSty      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	bjCountPositiveSty = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	bjCountNegativeSty = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
	bjInsuranceSty     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
)

const bjInnerWidth = 54

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

	if b.phase == phaseBetting {
		sb.WriteString(row(" " + bjBankrollSty.Render(fmt.Sprintf("Bankroll: $%d", b.bankroll))))
		sb.WriteString(blank())
		sb.WriteString(row(fmt.Sprintf("  Bet: $%-4d   ◀ ▶ Adjust   ENTER Deal", b.bet)))
	} else {
		// Dealer
		dealerCards, dealerVal := bjRenderHandMasked(b.dealer, revealed)
		sb.WriteString(row(" " + bjTextSty.Render("DEALER") + "  " + dealerCards + dealerVal))

		if b.phase == phaseInsurance {
			sb.WriteString(row("  " + bjInsuranceSty.Render(fmt.Sprintf("Dealer shows Ace — Insurance? ($%d)", b.hands[0].bet/2))))
		}

		sb.WriteString(brd("╠"+strings.Repeat("═", bjInnerWidth)+"╣") + "\n")
		for i := range b.hands {
			sb.WriteString(row(b.renderPlayerRow(i, b.phase == phaseTurn && i == b.active)))
		}
	}
	sb.WriteString(blank())

	switch b.phase {
	case phaseBetting:
		sb.WriteString(blank())
	case phaseInsurance:
		actions := "  " + bjHotSty.Render(" Y-Insurance ") + "   " + bjActionSty.Render(" N-Decline ")
		sb.WriteString(row(actions))
	case phaseTurn:
		h := &b.hands[b.active]
		actions := "  " + bjHotSty.Render(" H-Hit ") + "   " + bjActionSty.Render(" S-Stand ")
		if b.canDouble(h) {
			actions += "   " + bjActionSty.Render(" D-Double ")
		}
		if b.canSplit(h) {
			actions += "   " + bjActionSty.Render(" X-Split ")
		}
		sb.WriteString(row(actions))
	case phaseResults:
		sb.WriteString(row("  " + bjPushSty.Render("Press ENTER / SPACE for next round")))
	case phaseDealerTurn:
		sb.WriteString(row("  " + bjLabelSty.Render("Dealer's turn...")))
	default:
		sb.WriteString(blank())
	}

	if b.showCount {
		sb.WriteString(blank())
		sb.WriteString(row("  " + b.renderCountOverlay()))
	}

	sb.WriteString(blank())
	sb.WriteString(row("  " + bjLabelSty.Render(fmt.Sprintf("Bankroll: $%d   Wins: %d   Rounds: %d", b.bankroll, b.wins, b.rounds))))
	sb.WriteString(brd("╚" + strings.Repeat("═", bjInnerWidth) + "╝"))
	return sb.String()
}

func (b *Blackjack) renderCountOverlay() string {
	tc := b.shoe.TrueCount()
	tcSty := bjCountPositiveSty
	if tc < 0 {
		tcSty = bjCountNegativeSty
	}
	s := fmt.Sprintf("Count: %+d   True: %s   Decks left: %.1f",
		b.shoe.RunningCount(), tcSty.Render(fmt.Sprintf("%+.1f", tc)), b.shoe.DecksRemaining())
	if b.shoe.NeedsReshuffle() {
		s += "   " + bjLabelSty.Render("[reshuffle pending]")
	}
	if b.insuranceTotalCount > 0 {
		s += fmt.Sprintf("   Insurance: %d/%d correct", b.insuranceCorrectCount, b.insuranceTotalCount)
	}
	return s
}

func (b *Blackjack) renderPlayerRow(idx int, active bool) string {
	h := &b.hands[idx]
	name := "YOU"
	if len(b.hands) > 1 {
		name = fmt.Sprintf("YOU-%d", idx+1)
	}
	nameSty := bjTextSty
	if active {
		nameSty = bjActiveSty
	}
	nameStr := nameSty.Render(fmt.Sprintf("%-6s", name))

	val := bjLabelSty.Render(fmt.Sprintf(" (%d)", h.hand.Value()))

	statusStr := ""
	switch h.status {
	case statusBust:
		statusStr = "  " + bjBustSty.Render("BUST")
	case statusStand:
		statusStr = "  " + bjLabelSty.Render("STAND")
	case statusBlackjack:
		statusStr = "  " + bjBJSty.Render("BLACKJACK!")
	}
	if h.isDoubled {
		statusStr += "  " + bjLabelSty.Render("(DOUBLED)")
	}
	switch h.result {
	case "WIN":
		statusStr += "  " + bjWinSty.Render("WIN")
	case "LOSE":
		statusStr += "  " + bjLoseSty.Render("LOSE")
	case "PUSH":
		statusStr += "  " + bjPushSty.Render("PUSH")
	}
	betStr := "  " + bjLabelSty.Render(fmt.Sprintf("$%d", h.bet))

	const prefixW = 9
	suffixW := lipgloss.Width(val) + lipgloss.Width(statusStr) + lipgloss.Width(betStr)
	cardBudget := bjInnerWidth - prefixW - suffixW
	cards := bjRenderHandBudget(h.hand, cardBudget)

	return " " + nameStr + "  " + cards + val + betStr + statusStr
}
