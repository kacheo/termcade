# Poker (Texas Hold'em) — Implementation Spec

This document is a self-contained specification for adding a Texas Hold'em poker game to the **tmvgs** terminal game suite. An implementer should be able to build the feature using only this document and the existing codebase as reference.

---

## 1. Overview

**Package path:** `games/poker/`

**Variant:** Texas Hold'em — community card poker with blinds, four betting rounds, and a showdown.

**Players:** 1 human + 2–4 AI opponents (configurable in the options screen, total seats 3–5).

**AI difficulty:** Easy / Medium / Hard (three distinct behavior tiers).

**Win/loss:** The game ends when the human runs out of chips (loss) or all AI opponents are eliminated (win). There is no persistent save state; chips reset each session.

---

## 1a. Shared Card Package — games/cards

Before writing any poker code, refactor `games/blackjack/cards.go` into a new shared package. This avoids duplicating ~150 lines of identical card/deck/rendering code.

### What to move: `games/blackjack/cards.go` → `games/cards/cards.go`

Types and functions that are identical across both games:

```go
// Types
type Suit int   // Spades=0, Hearts=1, Diamonds=2, Clubs=3
type Rank int   // Ace=1, Two=2, ..., King=13
type Card struct { Rank Rank; Suit Suit }
type Deck []Card

// Card methods
func (c Card) Symbol() string      // "A","T","J","Q","K" or digit
func (c Card) SuitSymbol() string  // "♠","♥","♦","♣"
func (c Card) IsRed() bool         // Hearts or Diamonds

// Deck functions
func NewDeck() Deck
func (d Deck) Shuffled(r *rand.Rand) Deck
func (d *Deck) Draw() Card         // panics on empty
```

Rendering helpers — rename (drop `bj` prefix) and export. Card color styles should be passed in as parameters so callers control the palette:

```go
func RenderCard(c Card, redSty, blackSty lipgloss.Style) string
func RenderHand(hand []Card, redSty, blackSty lipgloss.Style) string
func RenderHandBudget(hand []Card, maxWidth int, redSty, blackSty lipgloss.Style) string
func Pad(s string, width int) string    // right-pad to visual width
func Center(s string, width int) string // center in visual width
```

(`Pad` and `Center` use `runewidth` + `cxansi.Strip` — same logic as `bjPad`/`bjCenter` in blackjack.)

### What stays in `games/blackjack`

Blackjack-specific hand logic. Update `blackjack` to use `type Hand []cards.Card`:

```go
func (h Hand) BaseValue() int  // 10 for T/J/Q/K, else rank
func (h Hand) Value() int      // Ace = 1 or 11
func (h Hand) IsBust() bool
func (h Hand) IsBlackjack() bool
func (h Hand) IsSoft() bool
```

Update `blackjack`'s `bjRenderCard` / `bjRenderHand` / `bjRenderHandBudget` / `bjPad` / `bjCenter` calls to delegate to `games/cards`.

### Ace rank note

The shared `Rank` keeps `Ace = 1` (matching current blackjack). Poker's hand evaluator in `hand.go` treats Ace as 14 where it improves the hand — this is handled entirely inside `Evaluate()`, not in the shared type.

### Verify the refactor before writing poker

```bash
go build ./games/blackjack     # must still compile
go test ./games/blackjack/...  # all existing tests must still pass
```

---

## 2. Files to Create

```
games/cards/
  cards.go        — shared Card, Suit, Rank, Deck types + rendering helpers (see §1a)

games/poker/
  poker.go        — Poker struct, core.Game interface, Update(), Render(), HandleInput()
  hand.go         — Hand type, Evaluate(), Compare(), ranking constants
  ai.go           — Difficulty enum, MakeDecision()
  poker_test.go   — integration: full hand, pot attribution, AI smoke tests
  hand_test.go    — unit: one test per hand rank, ordering correctness
  README.md       — user-facing controls/rules (see Section 14)
```

`deck.go` is not needed — poker imports `games/cards` directly. Keep each file focused; do not combine game logic with rendering.

---

## 3. cmd/main.go Integration

Make these exact changes to `cmd/main.go`. Read the existing file before editing.

### 3a. menuState enum

Add `menuPokerOptions` immediately before `menuPlaying`:

```go
const (
    menuMain menuState = iota
    menuTetrisOptions
    menuSnakeOptions
    menuSudokuOptions
    menuBlackjackOptions
    menuPokerOptions     // ADD THIS
    menuPlaying
    menuPause
    menuGameOver
)
```

### 3b. gameKind enum

```go
const (
    gameKindTetris gameKind = iota
    gameKindSnake
    gameKindSudoku
    gameKindBlackjack
    gameKindPoker        // ADD THIS
)
```

### 3c. model struct — add options fields

```go
type model struct {
    // ... existing fields ...
    pokerOpts struct {
        seats      int  // total seats including human; valid: 3, 4, 5; default: 4
        difficulty poker.Difficulty  // Easy=0, Medium=1, Hard=2; default: Medium
    }
}
```

Initialize defaults in the zero value (seats=0 means use default). In `updatePokerOptions` treat `seats == 0` as `seats = 4`.

### 3d. Main menu items

Change the items slice in `updateMainMenu`:

```go
items := []string{"Play Tetris", "Play Snake", "Play Sudoku", "Play Blackjack", "Play Poker", "", "Quit"}
```

Route `case 4` to `menuPokerOptions`. Shift the existing Quit case to `case 6`.

### 3e. Update dispatcher

In `Update()`, add:

```go
case menuPokerOptions:
    return m.updatePokerOptions(msg)
```

### 3f. View dispatcher

In `View()`, add:

```go
case menuPokerOptions:
    return m.renderPokerOptions()
```

### 3g. Options handler

```go
func (m *model) updatePokerOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // 4 items: 0=Seats, 1=Difficulty, 2=Start Game, 3=Back
    if m.pokerOpts.seats == 0 {
        m.pokerOpts.seats = 4
    }
    switch msg.String() {
    case "up", "k":
        if m.selected > 0 {
            m.selected--
        }
    case "down", "j":
        if m.selected < 3 {
            m.selected++
        }
    case "left", "h":
        if m.selected == 0 && m.pokerOpts.seats > 3 {
            m.pokerOpts.seats--
        }
        if m.selected == 1 && m.pokerOpts.difficulty > 0 {
            m.pokerOpts.difficulty--
        }
    case "right", "l":
        if m.selected == 0 && m.pokerOpts.seats < 5 {
            m.pokerOpts.seats++
        }
        if m.selected == 1 && m.pokerOpts.difficulty < 2 {
            m.pokerOpts.difficulty++
        }
    case "enter", " ":
        switch m.selected {
        case 2:
            m.game = poker.NewPoker(m.pokerOpts.seats, m.pokerOpts.difficulty)
            m.activeGame = gameKindPoker
            m.currentMenu = menuPlaying
            m.gameOver = false
            m.selected = 0
        case 3:
            m.currentMenu = menuMain
            m.selected = 0
        }
    case "q":
        m.currentMenu = menuMain
        m.selected = 0
    }
    return m, nil
}
```

### 3h. Restart

In `restartGame()`:

```go
case gameKindPoker:
    m.game = poker.NewPoker(m.pokerOpts.seats, m.pokerOpts.difficulty)
```

### 3i. Import

Add `"tmvgs/games/poker"` to the import block in `main.go`.

---

## 4. core.Game Interface

The `core.Game` interface (defined in `core/game.go`) requires these nine methods:

```go
func (p *Poker) Name() string        { return "Poker" }
func (p *Poker) Description() string { return "Texas Hold'em — bet, raise, or fold." }
func (p *Poker) IsPaused() bool      { return p.paused }
func (p *Poker) IsGameOver() bool    { return p.gameOver }
func (p *Poker) GetScore() int       { return p.humanChips() }  // human's current chip count
func (p *Poker) GetLevel() int       { return int(p.difficulty) } // 0=Easy 1=Med 2=Hard
func (p *Poker) GetLines() int       { return p.handsPlayed }
func (p *Poker) Update(delta time.Duration) error { ... }
func (p *Poker) Render() string { ... }
func (p *Poker) HandleInput(key string) { ... }
```

---

## 5. Card & Deck Types

Import `games/cards`. All card/deck types and rendering helpers come from there — do not redefine them in `games/poker`.

```go
import "tmvgs/games/cards"

// Use directly:
// cards.Card, cards.Deck, cards.NewDeck(), cards.Suit, cards.Rank
// cards.RenderCard(c, redSty, blackSty)
// cards.RenderHand(hand, redSty, blackSty)
// cards.RenderHandBudget(hand, maxWidth, redSty, blackSty)
// cards.Pad(s, width), cards.Center(s, width)
```

Define the poker `Hand` type as a slice alias in `hand.go`:

```go
type Hand []cards.Card
```

**Ace handling:** `cards.Rank` uses `Ace = 1`. Inside `Evaluate()`, treat Ace as rank 14 when computing straights and high-card tiebreakers. Ace-low straights (A-2-3-4-5) use Ace = 1. No changes needed to the shared type.

---

## 6. Hand Evaluation (`hand.go`)

### Types

```go
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
    Rank    HandRank
    Tiebreakers []int  // descending rank values for tie-breaking
    Cards   [5]Card   // the five cards making this hand
}
```

### Evaluate(cards []Card) EvaluatedHand

Given 5–7 cards, return the best possible 5-card hand.

**Algorithm:**
1. If `len(cards) == 5`, evaluate directly.
2. If `len(cards) > 5`, generate all C(n,5) combinations (for 7 cards: 21 combos), evaluate each, return the highest.

**Evaluation order** (check from strongest to weakest):
1. **Royal Flush**: A-K-Q-J-T all same suit → RoyalFlush, tiebreaker: [14]
2. **Straight Flush**: 5 consecutive same suit → StraightFlush, tiebreaker: [high card rank]
3. **Four of a Kind**: four same rank → FourOfAKind, tiebreaker: [quad rank, kicker]
4. **Full House**: three + pair → FullHouse, tiebreaker: [trips rank, pair rank]
5. **Flush**: 5 same suit → Flush, tiebreaker: [r1,r2,r3,r4,r5] descending
6. **Straight**: 5 consecutive (Ace can be low: A-2-3-4-5) → Straight, tiebreaker: [high card rank]
7. **Three of a Kind**: → ThreeOfAKind, tiebreaker: [trips rank, k1, k2]
8. **Two Pair**: → TwoPair, tiebreaker: [high pair, low pair, kicker]
9. **One Pair**: → OnePair, tiebreaker: [pair rank, k1, k2, k3]
10. **High Card**: → HighCard, tiebreaker: [r1,r2,r3,r4,r5] descending

### Compare(a, b EvaluatedHand) int

Returns -1 (a < b), 0 (tie), or 1 (a > b). Compare HandRank first, then Tiebreakers lexicographically.

---

## 7. Game State & Phases

### Phases

```go
type phase int

const (
    phasePreflop  phase = iota  // 2 hole cards dealt; first betting round
    phaseFlop                   // 3 community cards; second betting round
    phaseTurn                   // 4th community card; third betting round
    phaseRiver                  // 5th community card; final betting round
    phaseShowdown               // evaluate hands; award pot; 2s pause before next hand
    phaseGameOver
)
```

### Poker struct (poker.go)

```go
type Poker struct {
    rng        *rand.Rand
    deck       Deck
    players    []player       // index 0 = human
    community  []Card         // 0–5 cards
    pot        int
    sidePots   []sidePot      // for all-in situations (can be omitted in v1 — see note)
    phase      phase
    dealer     int            // index into players of current dealer button
    action     int            // index of player whose turn it is
    toCall     int            // chips needed to call current bet
    minRaise   int            // minimum raise increment
    handsPlayed int
    difficulty Difficulty
    paused     bool
    gameOver   bool
    elapsed    time.Duration  // used to auto-advance phaseShowdown after 2s
    raiseMode  bool           // true when human has pressed R and is picking amount
    raiseAmount int
    message    string         // one-line status message shown in TUI
}

type player struct {
    name    string
    chips   int
    hole    [2]Card
    bet     int    // chips committed this betting round
    folded  bool
    allIn   bool
    isHuman bool
}
```

**Side pots note:** In v1, side pots are optional. If a player goes all-in, you may simplify by awarding the pot only up to their all-in contribution; award excess to the next eligible player. Document this limitation in code comments.

---

## 8. Game Flow

### NewPoker(seats int, difficulty Difficulty) *Poker

- Create `seats` players. `players[0]` is human. AI players named "AI-1", "AI-2", etc.
- Each player starts with **1000 chips**.
- Set dealer button to random seat.
- Call `startHand()`.

### startHand()

1. Remove eliminated players (chips == 0) except human (if human == 0, set `gameOver = true, phase = phaseGameOver`).
2. If only human remains, set `gameOver = true, phase = phaseGameOver` (win).
3. Advance dealer button to next active player.
4. Shuffle fresh deck.
5. Deal 2 hole cards to each active player.
6. Post blinds: small blind (dealer+1) = 10 chips, big blind (dealer+2) = 20 chips.
7. Set `toCall = 20`, `minRaise = 20`, `pot = 30`.
8. Set `action` to player after big blind.
9. Set `phase = phasePreflop`.

### Betting round

A betting round proceeds until every active (non-folded, non-all-in) player has acted and `toCall` is met by all.

Track `lastRaiser` — when action returns to `lastRaiser` with all others having called/folded, end the round.

**Human turn:** `Update()` waits for input. Do nothing until `HandleInput` is called.

**AI turn:** In `Update()`, call `MakeDecision()` and apply immediately (no delay needed, but a 300ms pause between AI actions improves readability — use `elapsed` accumulator).

**Advancing phase after betting:**
- Preflop → deal 3 community cards → phaseFlop
- Flop → deal 1 → phaseTurn
- Turn → deal 1 → phaseRiver
- River → phaseShowdown

Reset `toCall = 0`, `minRaise = 20`, each player's `bet = 0` between phases. First action starts left of dealer.

### Showdown

1. Reveal all non-folded players' hole cards.
2. Build each player's 7-card pool (2 hole + 5 community).
3. Evaluate best hand per player.
4. Award pot to winner (or split on tie).
5. Set `message` to result string, e.g. "YOU win with Full House — $340".
6. Set `elapsed = 0`; wait 2 seconds in `Update()`, then call `startHand()`.

---

## 9. AI Behavior (`ai.go`)

```go
type Difficulty int
const (Easy Difficulty = iota; Medium; Hard)

type Action int
const (ActionFold Action = iota; ActionCheck; ActionCall; ActionRaise; ActionAllIn)

type Decision struct {
    Action Action
    Amount int  // only used for ActionRaise
}

func MakeDecision(
    difficulty Difficulty,
    rng *rand.Rand,
    holeCards [2]Card,
    community []Card,
    chips int,
    toCall int,
    pot int,
    minRaise int,
) Decision
```

### Hand strength bucket

Used by Medium and Hard. Evaluate the best hand from available cards (hole + community); map HandRank to bucket:

| HandRank | Bucket |
|----------|--------|
| HighCard, OnePair (low pair) | trash |
| OnePair (high pair), TwoPair | weak |
| ThreeOfAKind, Straight, Flush | medium |
| FullHouse, FourOfAKind | strong |
| StraightFlush, RoyalFlush | monster |

Pre-flop (no community cards): use hole card ranks only.
- Pocket pair ≥ J: medium
- Suited connectors (same suit, consecutive): weak
- Ace + high card: weak
- Otherwise: trash

### Easy AI

Pure weighted random — ignore hand strength:

- 40% → fold (if check is free, check instead)
- 50% → call (or check if toCall == 0)
- 10% → raise `minRaise`

### Medium AI

Heuristic by bucket:

| Bucket | toCall == 0 | toCall > 0 |
|--------|-------------|------------|
| trash | check | fold |
| weak | check | call if pot odds > 3:1 (toCall < pot/3), else fold |
| medium | check or raise 20% | call |
| strong | raise minRaise | raise minRaise |
| monster | check (slow-play pre-flop), raise on flop+ | raise minRaise*2 |

Pot odds: `toCall < pot / 3`.

### Hard AI

Add position and bluffing to Medium logic:

- **Position bonus**: if acting last (closest to dealer in remaining players), treat bucket one tier higher.
- **Bluffing**: on phaseRiver with a missed draw (trash bucket but had medium+ on turn), bluff raise with 10% probability.
- **Pot odds (precise)**: call if `toCall / (pot + toCall) < equity`. Use simple equity estimate: trash=15%, weak=35%, medium=55%, strong=80%, monster=95%.

---

## 10. Input Handling (`HandleInput`)

Keys passed in are already normalized by `cmd/main.go`'s `convertKey()`. The following normalized strings are used:

| Normalized key | Meaning |
|---------------|---------|
| `"f"` | Fold |
| `"c"` | Check or Call |
| `"r"` | Open raise amount selector |
| `"a"` | All-in |
| `"up"` | Increase raise amount (when raiseMode) |
| `"down"` | Decrease raise amount (when raiseMode) |
| `"enter"` | Confirm raise (when raiseMode) |
| `"esc"` | Cancel raise mode |

Only process input when `phase == phasePreflop/Flop/Turn/River` and it is the human player's turn (`action == 0`).

**Raise mode:** When `R` is pressed, set `raiseMode = true` and `raiseAmount = minRaise`. Up/Down adjust by `minRaise` increments (capped at human's chip count). Enter confirms, applying `raiseAmount` as the raise.

---

## 11. TUI Layout (Render)

Target: fits comfortably in an 80-column, 24-row terminal. Use `lipgloss` for color; use `strings.Builder` for assembly. Inner width: **45 characters**.

```
╔═════════════════ POKER ═════════════════╗
║  Pot: 340     Hand #12       FLOP       ║
╠═════════════════════════════════════════╣
║  AI-1  [●●]  $480   —                  ║
║  AI-2  [●●]  $210   Folded             ║
║  AI-3  [●●]  $960   Raised 80          ║
╠═════════════════════════════════════════╣
║  Board:  [A♠][K♦][7♣]  __    __       ║
╠═════════════════════════════════════════╣
║  YOU:  [Q♠][J♠]  $650   to call: 80   ║
║                                         ║
║  [F]old   [C]all 80   [R]aise   [A]ll-in║
╚═════════════════════════════════════════╝
```

**Rendering rules:**

- **Header row**: pot size, hand number, current phase name (PREFLOP / FLOP / TURN / RIVER / SHOWDOWN).
- **AI rows**: one row per AI. Show `[●●]` for hole cards (hidden). After Showdown, reveal hole cards as `[Q♠][J♥]`. Dim folded players. Highlight the active AI player in yellow.
- **Board row**: community cards as `[R S]` where R=rank, S=suit symbol. Undealt slots shown as `__`.
- **Human row**: hole cards face-up. Show current chip count. Show `to call: N` if N > 0, `check` if N == 0.
- **Action row**: show `[F]old`, `[C]heck` or `[C]all N`, `[R]aise`, `[A]ll-in`. Only show when it is the human's turn.
- **Raise mode**: replace action row with raise selector: `Raise amount: $N  [↑↓ adjust] [Enter confirm] [Esc cancel]`.
- **Message bar**: below actions, show `p.message` (result from last hand, errors, etc.) — dim gray.
- **Showdown**: reveal all non-folded hands; show hand rank name next to each player.

**Color palette** (use lipgloss directly; use `cards.RenderCard` / `cards.RenderHandBudget` from `games/cards` for card rendering — don't rely on `core/ui`):

| Element | Color |
|---------|-------|
| Border | `#666666` |
| Title | `#FFD700` bold |
| Red cards (♥ ♦) | `#FF5555` |
| Black cards (♣ ♠) | `#EEEEEE` |
| Hidden cards `[●●]` | `#555555` |
| Active player | `#00FF88` bold |
| Folded player | `#444444` |
| Win message | `#00FF00` bold |
| Lose message | `#FF4444` |
| Action keys | `#CCCCCC` on `#333333` |
| Hot action (selected) | `#000000` on `#00FF88` bold |

---

## 12. Chip Management

- All chip transfers go through a helper `awardPot(winnerIdx int)` that clears `p.pot`.
- On split (equal hands), divide evenly; leftover chips (odd amounts) go to the player closest to the dealer's left.
- Eliminations: a player with 0 chips after a hand is removed from `p.players` before the next `startHand()`.
- Human elimination (chips == 0): set `p.gameOver = true`, `p.phase = phaseGameOver`. The game-over screen (rendered by `cmd/main.go`) shows final chip count via `GetScore()`.

---

## 13. Testing Requirements

### hand_test.go

One test per hand rank covering:
1. Correct rank identified
2. Correct tiebreakers
3. Higher rank beats lower rank (`Compare`)
4. Equal hands return 0
5. Ace-low straight (A-2-3-4-5) recognized
6. 7-card evaluation picks the best 5

### poker_test.go

1. `NewPoker` creates correct number of players with 1000 chips each.
2. After `startHand`, each player has 2 hole cards; blinds posted correctly.
3. Folding all AI leaves pot to human (award test).
4. Easy AI never raises (statistical: run 100 decisions, assert raise count < 20).
5. Hard AI folds `trash` bucket when pot odds don't justify calling.

Run with: `go test ./games/poker/... ./games/blackjack/... ./games/cards/...`

---

## 14. games/poker/README.md

Create this file after implementation, following the same structure as other game READMEs in this repo. Include:

- One-sentence description
- **Options** table (Seats, Difficulty)
- **Controls** table (all keys from Section 10)
- **Hand Rankings** table (all 10 ranks, highest to lowest)
- **Rules** summary (blinds, phases, win condition)

---

## 15. Definition of Done

```bash
# Step 1 — shared package refactor
go build ./games/blackjack         # blackjack still compiles after refactor
go test ./games/blackjack/...      # all existing blackjack tests pass

# Step 2 — poker implementation
go build ./cmd                     # compiles without errors
go test ./games/poker/... ./games/cards/...  # all tests pass
go run ./cmd                       # "Play Poker" appears in main menu
                                   # options screen shows Seats and Difficulty
                                   # game deals cards, AI acts, human can fold/call/raise
                                   # game ends when human busts or wins
```
