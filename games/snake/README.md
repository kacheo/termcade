# Snake

Guide your snake around a 20×20 grid to eat food. Each meal grows your snake and increases your score. Hit a wall or yourself and it's game over.

---

## Controls

| Key | Action |
|-----|--------|
| `↑ ↓ ← →` | Change direction |
| `P` | Pause / Resume |
| `Q` | Quit to menu |

---

## Levels & Speed

The game has 10 levels. You advance one level every 5 food eaten.

| Level | Tick interval |
|-------|--------------|
| 1 | 200 ms |
| 2 | 185 ms |
| … | … |
| 10 | 65 ms (floor: 50 ms) |

Speed increases by 15 ms per level.

---

## Scoring

**10 × current level** per food eaten.

Eat food at higher levels to score faster.
