# Blackjack Card-Counting Trainer Design

## Overview

**Project:** termcade Blackjack game
**Type:** Single-player casino card game / card-counting practice tool
**Status:** Implemented

The original implementation reshuffled a fresh 52-card deck every hand — equivalent to a continuous-shuffle machine, which real casinos use specifically to defeat counting. This design turns it into a genuine counting trainer: a persistent multi-deck shoe with penetration/reshuffle, a toggleable Hi-Lo count overlay, bankroll-based betting, insurance, and double-down/split.

## Architecture

- `games/blackjack/blackjack.go` — Game state machine (`core.Game` implementation), input routing, render orchestration, betting/insurance/double/split logic
- `games/blackjack/shoe.go` — Persistent multi-deck `Shoe` with Hi-Lo counting and penetration tracking
- `games/blackjack/cards.go` — `Hand` value/bust/blackjack/soft helpers (shared, unmodified)

`Shoe` lives in `games/blackjack`, not the shared `games/cards` package: penetration, cut cards, and Hi-Lo counting are blackjack-specific training concepts that poker (the other consumer of `games/cards`) doesn't need.

## Data Model

### Shoe (`shoe.go`)

```go
type Shoe struct {
    rng              *rand.Rand
    cards            cardpkg.Deck
    numDecks         int     // clamped [1,8]
    totalCards       int
    reshufflePending bool
    runningCount     int
}
```

- `shoePenetration = 0.75` — 75% of the shoe is dealt before a reshuffle is queued, matching typical real-world cut-card placement.
- `Draw()` never reshuffles mid-hand on its own; the emergency reshuffle inside it is a defensive fallback that should never trigger given the deck-count range and one-split cap, but exists so a pathological run of splits/hits can't panic.
- `CountCard(c)` must be called only when a card becomes visible to the player, not when it's physically drawn — see Counting below.

### Blackjack (`blackjack.go`)

```go
type playerHand struct {
    hand          Hand
    status        handStatus
    result        string // "WIN", "LOSE", "PUSH", or ""
    bet           int
    isDoubled     bool
    fromSplitAces bool
}

type Blackjack struct {
    shoe    *Shoe
    dealer  Hand
    hands   []playerHand // len 1, or 2 after a split
    active  int
    phase   phase

    bankroll int
    bet      int

    dealerHoleCounted bool
    insuranceOffered, insuranceTaken bool
    insuranceBet     int
    insuranceResult  string
    insuranceTrueCountAtDecision float64
    insuranceCorrectCount, insuranceTotalCount int

    showCount bool
    gameOver  bool
}
```

A slice of 0–2 `playerHand`s (instead of parallel `hand`/`hand2` fields) lets the render/evaluate loops handle the split case for free — single-hand play is just the `len(hands) == 1` case of the same code paths.

## Phase State Machine

```
phaseBetting → phaseDealing → phaseInsurance (if dealer shows Ace) → phaseTurn → phaseDealerTurn → phaseResults → phaseBetting
```

There's no separate "double-or-split-choice" phase — Double and Split are just two more keys available in `phaseTurn` on a hand's first decision, gated by `canDouble`/`canSplit` eligibility (shown as action hints only when true).

- **phaseBetting** — adjust `bet` in `$10` steps (clamped to `[minBet, bankroll]`); `Enter` debits the bankroll and calls `startRound()`.
- **phaseInsurance** — only entered if the dealer's up-card is an Ace. `insuranceTrueCountAtDecision` snapshots the true count *before* the answer, computed only from currently-visible cards (correctly excluding the still-hidden hole card). The decision is graded against the standard "true count ≥ +3 → take" rule and tallied into `insuranceCorrectCount`/`insuranceTotalCount`. Dealer blackjack settles immediately; otherwise the hole card's reveal (and count contribution) is deferred to `phaseDealerTurn`. Insurance is gated by `canTakeInsurance()` (`bankroll >= bet/2`); the hint is hidden and the key is a no-op when unaffordable, so bankroll can never go negative from a side bet it can't cover.
- **No-peek on a Ten up-card** — the dealer is only ever peeked for blackjack when the up-card is an Ace (standard American-style peek). A Ten-value up-card is deliberately **not** peeked: a hidden dealer blackjack behind one stays unresolved until the dealer's turn naturally reveals the hole card (`dealerBlackjackUnresolved()`). This means a player's natural blackjack against a Ten up-card isn't settled until that reveal, and — per real rules — a non-natural 21 must still `LOSE` (not `PUSH`) if that reveal turns out to be a dealer blackjack (`evaluateResults()`'s `dealerBJ` case). Conversely, whenever the dealer is *known* not to have blackjack (Ace peeked negative, or a 2-9 up-card that structurally can't make 21), a player's natural blackjack settles immediately in `afterInsuranceDecision()` without dealing the dealer any further cards a real counter would never see.
- **phaseTurn** — loops over `hands`, `active` tracks which is live. Hit/Stand/Double/Split; advances `active` or moves to `phaseDealerTurn` once every hand is resolved (`advanceOrDealer`).
- **phaseResults** — per-hand results, insurance result, bankroll deltas; `Enter` returns to `phaseBetting` (betting happens before the next deal, not automatically).

`canDouble(h)`: `len(h.hand)==2 && !h.isDoubled && !h.fromSplitAces && bankroll >= h.bet`.
`canSplit(h)`: exactly one hand in play, both cards share the same base rank value, and bankroll covers a second bet of the same size. Max one split (no re-splitting).

## Counting

Hi-Lo tags: 2–6 = +1, 7–9 = 0, 10/face/Ace = −1. `Shoe.CountCard` is called exactly at the moment a card becomes visible to the player:

- Player cards and the dealer's up-card: counted immediately at deal time.
- The dealer's **hole card**: counted only inside `enterDealerTurn()` (guarded by `dealerHoleCounted` to avoid double-counting), or inside the insurance-reveal path if a dealer blackjack is peeked and settled immediately.

Counting the hole card the instant it's physically drawn would make the on-screen count diverge from what a real counter watching the table computes — this would silently defeat the entire point of the trainer, so it's treated as an invariant (see `TestScenarioHoleCardNotCountedUntilRevealed` and `TestScenarioReshuffleNeverHappensMidHand` in `scenario_test.go`).

The count overlay (toggled with `C`, off by default) shows: running count, true count (color-coded green/red), decks remaining, a `[reshuffle pending]` tag once penetration is crossed, and the insurance-decision accuracy tally.

## Betting, Insurance, Double, Split settlement

- Starting bankroll: $1000. Min bet / bet step: $10.
- Bet is debited when placed (phaseBetting `Enter`), and again when doubling or splitting incurs an additional bet.
- At settlement: WIN credits back `bet + payout` (payout is `bet*3/2` for a natural blackjack, else `bet`); PUSH credits back just `bet`; LOSE credits nothing (already debited).
- Insurance is capped at half the original bet, auto-computed (no adjustable UI); it pays 2:1 on a dealer blackjack.
- Split Aces get exactly one card each and auto-stand; a 21 made from a split hand is a plain 21 (even money), never a blackjack — standard casino rule, called out explicitly in code so a future edit doesn't silently misapply the 3:2 bonus to split hands.
- Bankroll dropping below `minBet` after a round resolves sets `gameOver = true`.

## `core.Game` interface mapping

| Method | Behavior | Rationale |
|---|---|---|
| `GetScore()` | `bankroll` | The number that actually measures counting/betting success |
| `GetLevel()` | `shoe.numDecks` | Deck count is the real difficulty knob |
| `GetLines()` | `rounds` | Existing "rounds played" convention |
| `IsGameOver()` | `gameOver` | Set once bankroll < minBet after a round resolves |

## Options screen (`cmd/main.go`)

Mirrors poker's seats/difficulty adjuster: a `Decks` row (`←`/`→`, clamped `[1,8]`, default 6) above Start/Back, feeding `blackjack.NewBlackjack(numDecks)`.

## Controls

| Key | Action |
|-----|--------|
| `H` / `←` | Hit |
| `S` / `→` / `↓` | Stand |
| `D` | Double down |
| `X` | Split |
| `Y` / `I` | Take insurance |
| `N` / `S` | Decline insurance |
| `←` / `→` | Adjust bet (betting screen) |
| `C` | Toggle count overlay |
| `Enter` / `Space` | Confirm bet / deal / next round |

Split is bound to `X`, not the more conventional `P` — `cmd/main.go`'s top-level key router intercepts `P` unconditionally to open the pause menu before any key reaches the active game's `HandleInput`, so `P` is not available to any game for an in-play action.

## Testing

- `shoe_test.go` — Hi-Lo bucket correctness, true-count arithmetic, penetration boundary, reshuffle state reset, empty-shoe safety net.
- `scenario_test.go` — reshuffle-pending never triggers mid-hand (only deferred to the next `startRound()`), hole-card counting timing, plus the pre-existing hit/stand/blackjack/win-count scenarios updated for the shoe/hands API.
- `blackjack_test.go` — unit tests for betting clamps, insurance offer/settlement/accuracy (including the bankroll-affordability gate), double, split (including split Aces), count toggle, bankroll math on win/push/lose, game-over on bankroll depletion, the no-peek-on-Ten unresolved-blackjack path, and a non-natural 21 losing (not pushing) to an unpeeked dealer blackjack.
- `render_test.go` — golden snapshots for the betting screen, initial dealt state, and the count overlay.
