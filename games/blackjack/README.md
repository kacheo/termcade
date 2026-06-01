# Blackjack

Classic casino card game — you against the dealer. Get closer to 21 than the dealer without going over.

---

## Controls

| Key | Action |
|-----|--------|
| `H` / `←` | Hit — draw another card |
| `S` / `→` / `↓` | Stand — end your turn |
| `Enter` / `Space` | Next round (shown after results) |
| `P` | Pause |
| `Q` | Quit to menu (from pause menu) |

---

## Rules

- Cards 2–10 are face value; J, Q, K = 10; Ace = 1 or 11 (whichever keeps you under 22).
- Dealer plays automatically after you stand or bust.
- Dealer must hit on soft 17 or below, and stand on hard 17 or above.
- A natural blackjack (Ace + 10-value card on the initial deal) wins automatically unless the dealer also has blackjack.

---

## Outcomes

| Result | Condition |
|--------|-----------|
| WIN | Your hand beats the dealer, or the dealer busts |
| LOSE | You bust, or the dealer's hand is higher |
| PUSH | Equal hand values — no winner |
| BLACKJACK | Natural 21 on deal; wins unless dealer also has blackjack (push) |

---

## Session Tracking

Wins and total rounds played are shown on screen. There is no persistent high score between sessions.
