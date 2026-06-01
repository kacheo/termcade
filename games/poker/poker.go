package poker

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"tmvgs/games/cards"
)

type phase int

const (
	phasePreflop phase = iota
	phaseFlop
	phaseTurn
	phaseRiver
	phaseShowdown
	phaseGameOver
)

type player struct {
	name    string
	chips   int
	hole    [2]cards.Card
	bet     int
	folded  bool
	allIn   bool
	isHuman bool
}

type Poker struct {
	rng         *rand.Rand
	deck        cards.Deck
	players     []player
	community   []cards.Card
	pot         int
	phase       phase
	dealer      int
	action      int
	toCall      int
	minRaise    int
	handsPlayed int
	difficulty  Difficulty
	paused      bool
	gameOver    bool
	elapsed     time.Duration
	raiseMode   bool
	raiseAmount int
	message     string
	lastRaiser  int
	aiDelay     time.Duration
}

func NewPoker(seats int, difficulty Difficulty) *Poker {
	p := &Poker{
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
		difficulty: difficulty,
	}
	for i := 0; i < seats; i++ {
		isHuman := i == 0
		name := "YOU"
		if !isHuman {
			name = fmt.Sprintf("AI-%d", i)
		}
		p.players = append(p.players, player{
			name:    name,
			chips:   1000,
			isHuman: isHuman,
		})
	}
	p.dealer = p.rng.Intn(len(p.players))
	p.startHand()
	return p
}

func (p *Poker) startHand() {
	p.players = filterPlayers(p.players)
	if len(p.players) == 0 {
		p.gameOver = true
		p.phase = phaseGameOver
		return
	}
	if p.players[0].chips == 0 {
		p.gameOver = true
		p.phase = phaseGameOver
		return
	}
	activeHuman := false
	for _, pl := range p.players {
		if pl.isHuman {
			activeHuman = true
			break
		}
	}
	if !activeHuman {
		p.gameOver = true
		p.phase = phaseGameOver
		return
	}
	if len(p.players) == 1 {
		p.gameOver = true
		p.phase = phaseGameOver
		return
	}
	p.dealer = (p.dealer + 1) % len(p.players)
	p.deck = cards.NewDeck().Shuffled(p.rng)
	for i := range p.players {
		p.players[i].hole = [2]cards.Card{p.deck.Draw(), p.deck.Draw()}
		p.players[i].bet = 0
		p.players[i].folded = false
		p.players[i].allIn = false
	}
	p.pot = 0
	smallBlindIdx := (p.dealer + 1) % len(p.players)
	bigBlindIdx := (p.dealer + 2) % len(p.players)
	p.players[smallBlindIdx].chips -= 10
	p.players[smallBlindIdx].bet = 10
	p.pot += 10
	if p.players[smallBlindIdx].chips == 0 {
		p.players[smallBlindIdx].allIn = true
	}
	p.players[bigBlindIdx].chips -= 20
	p.players[bigBlindIdx].bet = 20
	p.pot += 20
	if p.players[bigBlindIdx].chips == 0 {
		p.players[bigBlindIdx].allIn = true
	}
	p.toCall = 20
	p.minRaise = 20
	p.action = (p.dealer + 3) % len(p.players)
	p.phase = phasePreflop
	p.community = nil
	p.message = ""
	p.raiseMode = false
	p.raiseAmount = 0
	p.lastRaiser = -1
	p.aiDelay = 0
	if p.players[p.action].folded || p.players[p.action].allIn {
		p.advanceToNextPlayer()
	}
}

func filterPlayers(players []player) []player {
	var result []player
	for _, pl := range players {
		if pl.chips > 0 || pl.isHuman {
			result = append(result, pl)
		}
	}
	return result
}

func (p *Poker) activePlayerCount() int {
	count := 0
	for _, pl := range p.players {
		if !pl.folded && !pl.allIn {
			count++
		}
	}
	return count
}

func (p *Poker) allActed() bool {
	for _, pl := range p.players {
		if !pl.folded && !pl.allIn && pl.bet < p.toCall {
			return false
		}
	}
	return true
}

func (p *Poker) bettingRoundEnded() bool {
	if p.lastRaiser == -1 {
		return p.allActed()
	}
	if p.action != p.lastRaiser {
		return false
	}
	for _, pl := range p.players {
		if !pl.folded && !pl.allIn && pl.bet < p.toCall {
			return false
		}
	}
	return true
}

func (p *Poker) advanceToNextPlayer() {
	start := p.action
	for {
		p.action = (p.action + 1) % len(p.players)
		if p.action == start {
			break
		}
		if !p.players[p.action].folded && !p.players[p.action].allIn {
			break
		}
	}
}

func (p *Poker) processAITurn() {
	if p.phase < phasePreflop || p.phase > phaseRiver {
		return
	}
	if p.action == 0 {
		return
	}
	if p.players[p.action].folded || p.players[p.action].allIn {
		p.advanceToNextPlayer()
		return
	}
	if p.aiDelay < 300*time.Millisecond {
		p.aiDelay += 16 * time.Millisecond
		return
	}
	p.aiDelay = 0
	decision := MakeDecision(
		p.difficulty,
		p.rng,
		p.players[p.action].hole,
		p.community,
		p.players[p.action].chips,
		p.toCall,
		p.pot,
		p.minRaise,
	)
	p.applyAction(decision)
}

func (p *Poker) applyAction(d Decision) {
	switch d.Action {
	case ActionFold:
		p.players[p.action].folded = true
		p.message = fmt.Sprintf("%s folds", p.players[p.action].name)
	case ActionCheck:
		p.players[p.action].bet = 0
		p.message = fmt.Sprintf("%s checks", p.players[p.action].name)
	case ActionCall:
		toCall := p.toCall - p.players[p.action].bet
		if toCall > p.players[p.action].chips {
			toCall = p.players[p.action].chips
		}
		p.players[p.action].chips -= toCall
		p.players[p.action].bet += toCall
		p.pot += toCall
		if p.players[p.action].chips == 0 {
			p.players[p.action].allIn = true
		}
		p.message = fmt.Sprintf("%s calls %d", p.players[p.action].name, toCall)
	case ActionRaise:
		raiseAmount := d.Amount
		totalToCall := p.toCall - p.players[p.action].bet + raiseAmount
		playerChips := p.players[p.action].chips
		if totalToCall > playerChips {
			allInAmount := playerChips
			p.players[p.action].chips = 0
			p.players[p.action].allIn = true
			p.players[p.action].bet += allInAmount
			p.pot += allInAmount
			if p.toCall < p.players[p.action].bet {
				p.toCall = p.players[p.action].bet
			}
			p.lastRaiser = p.action
			p.message = fmt.Sprintf("%s is all-in!", p.players[p.action].name)
			p.advanceToNextPlayer()
			p.aiDelay = 0
			return
		}
		additional := raiseAmount
		p.players[p.action].chips -= additional
		p.players[p.action].bet += additional
		p.pot += additional
		p.toCall = p.players[p.action].bet
		p.minRaise = raiseAmount
		if raiseAmount > 0 {
			p.minRaise = raiseAmount
		} else {
			p.minRaise = 20
		}
		if p.minRaise < 20 {
			p.minRaise = 20
		}
		p.lastRaiser = p.action
		if p.players[p.action].chips == 0 {
			p.players[p.action].allIn = true
		}
		p.message = fmt.Sprintf("%s raises to %d", p.players[p.action].name, p.toCall)
	case ActionAllIn:
		allInAmount := p.players[p.action].chips
		p.players[p.action].chips = 0
		p.players[p.action].allIn = true
		p.players[p.action].bet += allInAmount
		p.pot += allInAmount
		if p.players[p.action].bet > p.toCall {
			p.toCall = p.players[p.action].bet
			p.lastRaiser = p.action
		}
		p.message = fmt.Sprintf("%s is all-in!", p.players[p.action].name)
	}
	p.advanceToNextPlayer()
	p.aiDelay = 0
}

func (p *Poker) endBettingRound() {
	for i := range p.players {
		p.players[i].bet = 0
	}
	p.toCall = 0
	p.minRaise = 20
	p.lastRaiser = -1
	switch p.phase {
	case phasePreflop:
		p.phase = phaseFlop
		p.community = append(p.community, p.deck.Draw(), p.deck.Draw(), p.deck.Draw())
	case phaseFlop:
		p.phase = phaseTurn
		p.community = append(p.community, p.deck.Draw())
	case phaseTurn:
		p.phase = phaseRiver
		p.community = append(p.community, p.deck.Draw())
	case phaseRiver:
		p.phase = phaseShowdown
		p.elapsed = 0
	}
	if p.action >= len(p.players) {
		p.action = 0
	}
	for p.players[p.action].folded || p.players[p.action].allIn {
		p.action = (p.action + 1) % len(p.players)
		if p.action == 0 {
			break
		}
	}
}

func (p *Poker) showdown() {
	activePlayers := make([]int, 0)
	for i, pl := range p.players {
		if !pl.folded {
			activePlayers = append(activePlayers, i)
		}
	}
	if len(activePlayers) == 0 {
		return
	}

	bestIdx := activePlayers[0]
	bestHand := Evaluate(append(p.players[bestIdx].hole[:], p.community...))

	winners := []int{bestIdx}
	for _, idx := range activePlayers[1:] {
		hand := Evaluate(append(p.players[idx].hole[:], p.community...))
		cmp := Compare(hand, bestHand)
		if cmp > 0 {
			bestHand = hand
			winners = []int{idx}
		} else if cmp == 0 {
			winners = append(winners, idx)
		}
	}

	splitAmount := p.pot / len(winners)
	for _, w := range winners {
		p.players[w].chips += splitAmount
	}
	remainder := p.pot % len(winners)
	if remainder > 0 {
		p.players[winners[0]].chips += remainder
	}
	p.pot = 0

	rankName := handRankName(bestHand.Rank)
	if len(winners) == 1 {
		p.message = fmt.Sprintf("%s wins with %s — $%d", p.players[winners[0]].name, rankName, splitAmount)
	} else {
		p.message = fmt.Sprintf("Split pot! %d players tie with %s — $%d each", len(winners), rankName, splitAmount)
	}
}

func handRankName(rank HandRank) string {
	switch rank {
	case HighCard:
		return "High Card"
	case OnePair:
		return "One Pair"
	case TwoPair:
		return "Two Pair"
	case ThreeOfAKind:
		return "Three of a Kind"
	case Straight:
		return "Straight"
	case Flush:
		return "Flush"
	case FullHouse:
		return "Full House"
	case FourOfAKind:
		return "Four of a Kind"
	case StraightFlush:
		return "Straight Flush"
	case RoyalFlush:
		return "Royal Flush"
	}
	return "Unknown"
}

func (p *Poker) Update(delta time.Duration) error {
	p.elapsed += delta
	switch p.phase {
	case phasePreflop, phaseFlop, phaseTurn, phaseRiver:
		if p.action != 0 && !p.players[p.action].folded && !p.players[p.action].allIn {
			p.processAITurn()
		}
		if p.bettingRoundEnded() {
			p.endBettingRound()
		}
	case phaseShowdown:
		if p.elapsed >= 2*time.Second {
			p.handsPlayed++
			p.startHand()
		}
	}
	return nil
}

func (p *Poker) HandleInput(key string) {
	if p.phase < phasePreflop || p.phase > phaseRiver {
		return
	}
	if p.action != 0 {
		return
	}
	if p.raiseMode {
		switch key {
		case "up":
			p.raiseAmount += p.minRaise
			if p.raiseAmount > p.players[0].chips {
				p.raiseAmount = p.players[0].chips
			}
		case "down":
			p.raiseAmount -= p.minRaise
			if p.raiseAmount < p.minRaise {
				p.raiseAmount = p.minRaise
			}
		case "enter":
			if p.raiseAmount > 0 {
				d := Decision{Action: ActionRaise, Amount: p.raiseAmount - p.toCall}
				if d.Amount < 0 {
					d.Amount = 0
				}
				p.applyAction(d)
			}
			p.raiseMode = false
		case "esc":
			p.raiseMode = false
		}
		return
	}
	switch key {
	case "f":
		p.applyAction(Decision{Action: ActionFold})
	case "c":
		if p.toCall == 0 {
			p.applyAction(Decision{Action: ActionCheck})
		} else {
			d := Decision{Action: ActionCall}
			p.applyAction(d)
		}
	case "r":
		if p.players[0].chips > p.toCall {
			p.raiseMode = true
			p.raiseAmount = p.toCall + p.minRaise
			if p.raiseAmount > p.players[0].chips {
				p.raiseAmount = p.players[0].chips
			}
		}
	case "a":
		if p.players[0].chips > 0 {
			d := Decision{Action: ActionAllIn}
			p.applyAction(d)
		}
	}
}

func (p *Poker) Render() string {
	return p.render()
}

func (p *Poker) render() string {
	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	redStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555"))

	blackStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EEEEEE"))

	hiddenStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))

	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF88")).
		Bold(true)

	foldedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444"))
	_ = foldedStyle

	winStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	loseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF4444"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	var sb strings.Builder

	sb.WriteString(borderStyle.Render(fmt.Sprintf("╔%s POKER %s╗", strings.Repeat("═", 17), strings.Repeat("═", 17))))
	sb.WriteString("\n")

	phaseName := ""
	switch p.phase {
	case phasePreflop:
		phaseName = "PREFLOP"
	case phaseFlop:
		phaseName = "FLOP"
	case phaseTurn:
		phaseName = "TURN"
	case phaseRiver:
		phaseName = "RIVER"
	case phaseShowdown:
		phaseName = "SHOWDOWN"
	case phaseGameOver:
		phaseName = "GAME OVER"
	}

	handNum := p.handsPlayed + 1
	header := fmt.Sprintf("║  Pot: %-5d  Hand #%-3d  %-9s║", p.pot, handNum, phaseName)
	sb.WriteString(cards.Pad(header, 47))
	sb.WriteString("\n")

	sb.WriteString(borderStyle.Render(fmt.Sprintf("╠%s╣", strings.Repeat("═", 45))))
	sb.WriteString("\n")

	for i, pl := range p.players {
		if pl.isHuman {
			continue
		}
		holeStr := cards.RenderCard(pl.hole[0], redStyle, blackStyle) + cards.RenderCard(pl.hole[1], redStyle, blackStyle)
		if p.phase < phaseShowdown {
			holeStr = hiddenStyle.Render("[●●]") + hiddenStyle.Render("[●●]")
		}

		status := ""
		if pl.folded {
			status = "Folded"
		} else if pl.allIn {
			status = "All-in"
		} else if pl.bet > 0 {
			status = fmt.Sprintf("Bet %d", pl.bet)
		}

		row := fmt.Sprintf("║  %s  %s  $%-5d  %-10s║", cards.Pad(pl.name, 5), holeStr, pl.chips, status)
		if i == p.action {
			row = activeStyle.Render(row)
		}
		sb.WriteString(cards.Pad(row, 47))
		sb.WriteString("\n")

		if p.phase == phaseShowdown && !pl.folded {
			eval := Evaluate(append(pl.hole[:], p.community...))
			rankStr := handRankName(eval.Rank)
			showRow := fmt.Sprintf("║       └─ %s", rankStr)
			sb.WriteString(cards.Pad(showRow, 47))
			sb.WriteString("\n")
		}
	}

	sb.WriteString(borderStyle.Render(fmt.Sprintf("╠%s╣", strings.Repeat("═", 45))))
	sb.WriteString("\n")

	boardStr := "Board: "
	for i := 0; i < 5; i++ {
		if i < len(p.community) {
			boardStr += cards.RenderCard(p.community[i], redStyle, blackStyle) + " "
		} else {
			boardStr += " __  "
		}
	}
	sb.WriteString(cards.Pad(boardStr, 47))
	sb.WriteString("\n")

	sb.WriteString(borderStyle.Render(fmt.Sprintf("╠%s╣", strings.Repeat("═", 45))))
	sb.WriteString("\n")

	humanHole := cards.RenderCard(p.players[0].hole[0], redStyle, blackStyle) + cards.RenderCard(p.players[0].hole[1], redStyle, blackStyle)
	toCallStr := ""
	if p.toCall > 0 {
		toCallStr = fmt.Sprintf("to call: %d", p.toCall-p.players[0].bet)
	} else {
		toCallStr = "check"
	}
	humanRow := fmt.Sprintf("║  %s  %s  $%-5d  %-12s║", cards.Pad("YOU", 5), humanHole, p.players[0].chips, toCallStr)
	sb.WriteString(cards.Pad(humanRow, 47))
	sb.WriteString("\n")

	if p.phase == phaseShowdown && !p.players[0].folded {
		eval := Evaluate(append(p.players[0].hole[:], p.community...))
		rankStr := handRankName(eval.Rank)
		showRow := fmt.Sprintf("║       └─ %s", rankStr)
		sb.WriteString(cards.Pad(showRow, 47))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	if p.action == 0 && !p.raiseMode && p.phase >= phasePreflop && p.phase <= phaseRiver {
		if p.toCall > 0 {
			callStr := fmt.Sprintf("[C]all %d", p.toCall-p.players[0].bet)
			sb.WriteString(fmt.Sprintf("  %-12s  ", callStr))
		} else {
			sb.WriteString("  [C]heck       ")
		}
		sb.WriteString("  [F]old  [R]aise  [A]ll-in")
		sb.WriteString("\n")
	}

	if p.raiseMode {
		raiseStr := fmt.Sprintf("Raise amount: $%d  [↑↓ adjust]  [Enter confirm]  [Esc cancel]", p.raiseAmount)
		sb.WriteString("  " + raiseStr + "\n")
	}

	sb.WriteString("\n")

	if p.message != "" {
		isWin := false
		isLose := false

		if strings.Contains(p.message, "YOU") && strings.Contains(p.message, "wins") {
			isWin = true
		}
		if strings.Contains(p.message, "wins") && strings.Contains(p.message, "AI") {
			isLose = true
		}
		if p.phase == phaseGameOver {
			isLose = true
		}
		msgStyle := dimStyle
		if isWin {
			msgStyle = winStyle
		} else if isLose {
			msgStyle = loseStyle
		}
		sb.WriteString("  " + msgStyle.Render(p.message))
		sb.WriteString("\n")
	}

	sb.WriteString(borderStyle.Render(fmt.Sprintf("╚%s╝", strings.Repeat("═", 45))))

	return sb.String()
}

func (p *Poker) Name() string {
	return "Poker"
}

func (p *Poker) Description() string {
	return "Texas Hold'em — bet, raise, or fold."
}

func (p *Poker) IsPaused() bool {
	return p.paused
}

func (p *Poker) IsGameOver() bool {
	return p.gameOver
}

func (p *Poker) GetScore() int {
	if len(p.players) == 0 {
		return 0
	}
	return p.players[0].chips
}

func (p *Poker) GetLevel() int {
	return int(p.difficulty)
}

func (p *Poker) GetLines() int {
	return p.handsPlayed
}