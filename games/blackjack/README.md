# Blackjack

Classic casino card game — you against the dealer. Get closer to 21 than the dealer without going over.

This is a card-counting trainer, not just a hand-by-hand game: cards are dealt from a persistent multi-deck shoe (not reshuffled every hand), and a toggleable Hi-Lo count overlay lets you check your running/true count against the game's own tracking.

---

## Controls

| Key | Action |
|-----|--------|
| `H` / `←` | Hit — draw another card |
| `S` / `→` / `↓` | Stand — end your turn |
| `D` | Double down (first decision on a hand only) |
| `X` | Split (first decision, matching pair only, one split per round) |
| `Y` / `I` | Take insurance (when offered) |
| `N` / `S` | Decline insurance (when offered) |
| `←` / `→` | Adjust bet (betting screen, in $10 steps) |
| `C` | Toggle the count overlay on/off |
| `Enter` / `Space` | Confirm bet / deal / next round |
| `P` | Pause |
| `Q` | Quit to menu (from pause menu) |

---

## Betting & bankroll

- You start each session with a $1000 bankroll.
- Before each round, adjust your bet (`$10` minimum, `$10` steps) and press `Enter` to deal.
- Running out of bankroll (below the $10 minimum bet) ends the game.

## The shoe

- The game deals from a persistent shoe of 1–8 decks (default 6, configurable from the options screen), not a fresh 52-card deck every hand — this matters for counting, since a continuous reshuffle would make counting meaningless.
- The shoe is cut at 75% penetration: once that much of it has been dealt, a reshuffle is queued and happens automatically before the *next* round — never mid-hand.

## Counting

- Press `C` any time to toggle a count overlay showing the Hi-Lo running count, the true count (running count ÷ decks remaining), decks remaining, a reshuffle-pending flag, and your insurance-decision accuracy.
- The overlay is off by default so you can practice counting in your head and check yourself, or leave it on as a learning aid.
- The dealer's hole card is not counted until it's actually revealed (dealer's turn, or an insurance-triggered peek at dealer blackjack) — the displayed count always matches what a real counter watching the table would see.

## Insurance

- Offered whenever the dealer's up-card is an Ace, for up to half your bet — you can only take it if your bankroll covers that amount.
- Standard counting strategy: insurance is a profitable bet only when the true count is +3 or higher. The count overlay tracks how often your insurance decisions matched that rule.
- If the dealer has blackjack, insurance pays 2:1 and the round settles immediately.
- The dealer only peeks for blackjack on an Ace up-card. A Ten-value up-card is never peeked, so a hidden dealer blackjack behind one isn't revealed until the dealer's turn — and a natural blackjack always beats any other 21, even a multi-card one, so it's a LOSE, not a PUSH, if that reveal turns out to be a dealer blackjack.

## Double down & split

- Double down: available on your first decision on a hand (any two cards), doubles your bet, draws exactly one more card, then auto-stands.
- Split: available on your first decision when your two cards share the same base rank and you can cover a second bet of the same size. Only one split per round (two hands total, no re-splitting). Split Aces each get exactly one card and auto-stand.
- A 21 made from a split hand pays even money, not the 3:2 blackjack bonus (standard casino rule).

---

## Rules

- Cards 2–10 are face value; J, Q, K = 10; Ace = 1 or 11 (whichever keeps you under 22).
- Dealer plays automatically after all your hands stand or bust.
- Dealer must hit on soft 17 or below, and stand on hard 17 or above.
- A natural blackjack (Ace + 10-value card on the initial deal) wins automatically (3:2 payout) unless the dealer also has blackjack.

---

## Outcomes

| Result | Condition |
|--------|-----------|
| WIN | Your hand beats the dealer, or the dealer busts |
| LOSE | You bust, or the dealer's hand is higher |
| PUSH | Equal hand values — no winner |
| BLACKJACK | Natural 21 on deal; wins 3:2 unless dealer also has blackjack (push) |

---

## Session Tracking

Bankroll, wins, and total rounds played are shown on screen. There is no persistent record between sessions.
