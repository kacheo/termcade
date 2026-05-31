# NOVA — Game Design Document
*Working title: **NOVA** | Platform: tmvgs (terminal) | Genre: vertical shooter*

---

## Concept

A vertical-scrolling shoot-'em-up for the terminal. Pure shooter — move, fire, survive waves, beat the boss. Uses **Unicode braille-character vector rendering** (the terminal equivalent of bladedancer's Bresenham PixelCanvas) for genuine vector-art sprites — not block pixel art, not single-glyph substitutes. Shapes look like they were drawn with a CRT vector beam.

---

## Platform Fit

| Concern | Solution |
|---------|----------|
| `Game` interface | Implement all 9 methods; `GetLines()` returns 0 |
| Rendering | `Render() string` builds the playfield + sidebar as a Bubble Tea view string |
| Input | `HandleInput(key)` receives "left"/"right"/"up"/"down"/" "/"q"/"p" |
| Timing | `Update(delta time.Duration)` advances all entity positions and animations at 60 Hz |
| Registration | Add menu option in `cmd/main.go` alongside Tetris |

**New package:** `games/nova/`

---

## Vector Rendering: BrailleCanvas

This is the core technical insight from bladedancer — ported to terminal.

### How it works

Unicode braille characters (U+2800–U+28FF) pack a **2-column × 4-row dot grid** into a single terminal character. This gives sub-character pixel resolution:

```
Terminal char at (tx, ty)
represents braille sub-pixels at:
  (tx*2,   ty*4)   (tx*2+1, ty*4)
  (tx*2,   ty*4+1) (tx*2+1, ty*4+1)
  (tx*2,   ty*4+2) (tx*2+1, ty*4+2)
  (tx*2,   ty*4+3) (tx*2+1, ty*4+3)

Dot-to-bit mapping (Braille standard):
  dot 1 (bit 0) = col 0, row 0
  dot 2 (bit 1) = col 0, row 1
  dot 3 (bit 2) = col 0, row 2
  dot 4 (bit 3) = col 1, row 0
  dot 5 (bit 4) = col 1, row 1
  dot 6 (bit 5) = col 1, row 2
  dot 7 (bit 6) = col 0, row 3
  dot 8 (bit 7) = col 1, row 3

Char = '⠀' + bit_sum (U+2800 + flags)
```

A 32×40 terminal playfield becomes a **64×160 sub-pixel canvas** — enough for recognizable vector shapes.

### BrailleCanvas type (canvas.go)

```go
type BrailleCanvas struct {
    W, H   int              // terminal dimensions (chars)
    pixels [160][64]bool    // sub-pixels: [row][col], H*4 × W*2
    colors [40][32]lipgloss.Color  // one color per terminal char
}

func (c *BrailleCanvas) Clear()
func (c *BrailleCanvas) SetPixel(x, y int, color lipgloss.Color)
func (c *BrailleCanvas) DrawLine(x0, y0, x1, y1 int, color lipgloss.Color)  // Bresenham
func (c *BrailleCanvas) DrawCircle(cx, cy, r int, color lipgloss.Color)     // Midpoint circle
func (c *BrailleCanvas) DrawFilledCircle(cx, cy, r int, color lipgloss.Color)
func (c *BrailleCanvas) Row(ty int) string  // render one terminal row to string
```

Caller draws sprites onto the canvas by position. Last write to a sub-pixel wins color of that terminal char.

### Algorithms (ported directly from bladedancer)

**Bresenham line** (pixel_canvas.dart → Go):
```go
func bresenhamLine(x0, y0, x1, y1 int) [][2]int {
    pts := [][2]int{}
    dx, dy := abs(x1-x0), -abs(y1-y0)
    sx, sy := sign(x0, x1), sign(y0, y1)
    err := dx + dy
    cx, cy := x0, y0
    for {
        pts = append(pts, [2]int{cx, cy})
        if cx == x1 && cy == y1 { break }
        e2 := 2 * err
        if e2 >= dy { err += dy; cx += sx }
        if e2 <= dx { err += dx; cy += sy }
    }
    return pts
}
```

**Midpoint circle** (8-way symmetry, pixel_canvas.dart → Go):
```go
func midpointCircle(r int) [][2]int {
    pts := [][2]int{}
    x, y, d := r, 0, 1-r
    for x >= y {
        for _, p := range [][2]int{{x,y},{-x,y},{x,-y},{-x,-y},{y,x},{-y,x},{y,-x},{-y,-x}} {
            pts = append(pts, p)
        }
        y++
        if d > 0 { x--; d += 2*(y-x) + 1 } else { d += 2*y + 1 }
    }
    return pts
}
```

Both can be memoized (the results only depend on endpoints / radius, not position).

---

## Sprite Definitions

All sprites are expressed as draw-call sequences on the BrailleCanvas, scaled to braille sub-pixel coordinates. Entity float position `(ex, ey)` maps to braille pixel origin at `(int(ex)*2, int(ey)*4)`. Each sprite is centered at `(0,0)` in its local grid and translated to the entity's braille origin before drawing.

**Scale**: 1 bladedancer grid unit ≈ 1 braille sub-pixel. Sprites are naturally compact.

### Player Ship — NOVA-I

*Inspired by DART's delta-wing hull (5 lines):*

| Draw call | Local coords | Effect |
|-----------|-------------|--------|
| `DrawLine` | (0,-7)→(5,5) | right wing leading edge |
| `DrawLine` | (5,5)→(3,6) | right wing trailing edge |
| `DrawLine` | (3,6)→(-3,6) | tail base |
| `DrawLine` | (-3,6)→(-5,5) | left wing trailing edge |
| `DrawLine` | (-5,5)→(0,-7) | left wing leading edge |
| `DrawLine` | (0,-5)→(0,4) | fuselage spine (accent color) |

**Color**: hull = `#00FFFF`, spine = `#FFFFFF`
**Size in terminal chars**: ~5 wide × 4 tall

Invincibility blink: alternate drawing hull in `#004444` dim cyan at 8 Hz.

### Enemy: Drifter

*Port of bladedancer's Drift diamond (4 lines + animated accent):*

| Draw call | Local coords |
|-----------|-------------|
| `DrawLine` | (0,-8)→(6,0) |
| `DrawLine` | (6,0)→(0,8) |
| `DrawLine` | (0,8)→(-6,0) |
| `DrawLine` | (-6,0)→(0,-8) |
| `DrawLine` (accent) | (-5, ±1)→(5, ±1) — toggles y at 4 Hz |

**Color**: body = `#FF4444`, accent = `#FF8888`

### Enemy: Weaver

*Port of bladedancer's Swirl triangle with spikes:*

| Draw call | Local coords |
|-----------|-------------|
| `DrawLine` | (0,-8)→(6,6) |
| `DrawLine` | (6,6)→(-6,6) |
| `DrawLine` | (-6,6)→(0,-8) |
| `DrawLine` (spike) | (0,-8)→(0,-12) — extends by 1 at 3 Hz |
| `DrawLine` (spike) | (-6,6)→(-9,9) — same |
| `DrawLine` (spike) | (6,6)→(9,9) — same |

**Color**: body = `#FF8800`, spikes = `#FFCC44`

### Enemy: Anchor

*Port of bladedancer's Fairy compact diamond + rotating accent:*

| Draw call | Local coords |
|-----------|-------------|
| `DrawLine` | (0,-5)→(4,0) |
| `DrawLine` | (4,0)→(0,5) |
| `DrawLine` | (0,5)→(-4,0) |
| `DrawLine` | (-4,0)→(0,-5) |
| `DrawFilledCircle` | center (0,0) r=2 (hub) |
| `DrawLine` (accent diagonal) | alternates orientation at 8 Hz |

**Color**: body = `#AA44FF`, hub = `#CC88FF`

### Bullet Sprites

| Entity | Draw calls | Color |
|--------|-----------|-------|
| Player bullet | `DrawLine` (0,0)→(0,-5) — 5px tall vertical line | `#FFFFFF` |
| Enemy aimed | `SetPixel` (0,0) — single braille dot | `#FF6666` |
| Enemy spread | `DrawCircle` r=2 — small ring | `#FFAA44` |

### Boss: APEX

*Multi-line shape, larger scale (×1.5 vs enemies):*

```
Outer hexagonal frame (6 DrawLine calls):
  (0,-12)→(8,-6)→(8,6)→(0,12)→(-8,6)→(-8,-6)→(0,-12)
Center cross:
  DrawLine (-6,0)→(6,0)  [horizontal]
  DrawLine (0,-6)→(0,6)  [vertical]
Corner diamonds (4× DrawFilledCircle r=2 at corners):
  at (8,0), (-8,0), (0,-12), (0,12)
```

**Damaged variant** (below 33% HP): outer hexagon drawn in `#441111`, center cross in `#FF2222`

**Color**: full health = `#FF2222` frame + `#FF6666` cross; taking damage flash = `#FFFFFF` for 3 frames

---

## Play Area

```
  ╔══════════════════════════════╗  ╔══════════════╗
  ║  [BrailleCanvas 32×40]       ║  ║ SCORE        ║
  ║  entities rendered as        ║  ║ 000000        ║
  ║  braille vector sprites      ║  ╠══════════════╣
  ║  on near-black background    ║  ║ LEVEL   01    ║
  ║                              ║  ║ LIVES  ▲▲▲    ║
  ╚══════════════════════════════╝  ║ BOMBS   ●●●   ║
  ╔════════════════════════════════╝  ╚══════════════╝
  ╔════════════════════════════════════════════════╗
  ║ [←→↑↓] Move  [Spc] Bomb  [P] Pause  [Q] Quit  ║
  ╚════════════════════════════════════════════════╝
```

- **Playfield**: 32 cols × 40 rows terminal chars = 64×160 braille sub-pixels
- **Sidebar**: 16 chars inner (matches Tetris sidebar width)
- Entity game positions are `float64` in `[0,32) × [0,40)`, snapped to braille coords for drawing

**Background**: `BrailleCanvas.Clear()` fills all chars with `⠂` (single-dot) in `#0D0D0D` — faint star-texture rather than pure black.

---

## Core Mechanics

### Player
- Position: `float64` (x, y), initialized at `(16, 36)`
- Speed: 18 cells/sec; clamped to playfield bounds
- Hitbox: 0.4-cell radius (applies at float64 level, not braille)
- Auto-fire: bullet every 0.12s; bullet speed 30 cells/sec upward
- Lives: 3 starting; hit = respawn at center-bottom, 1.5s invincibility (ship blinks at 8 Hz)
- Bombs: 3 starting; clears all enemy bullets, deals 80 dmg to all enemies; 2s recharge

### Enemies

| Type | HP | Move | Fire |
|------|----|------|------|
| Drifter | 1 | Straight down, 8 cells/sec | 1 aimed bullet every 2.0s |
| Weaver | 1 | Sine-wave descent, amplitude 4, freq 1 Hz | None |
| Anchor | 3 | Enter from top, stop at y=10 | 3-way spread every 1.5s |

Enemy bullets: 12 cells/sec. All enemies cleared on player death.

### Scoring
- Drifter kill: 100 pts
- Weaver kill: 150 pts
- Anchor kill: 300 pts
- Boss damage per 10% HP: 500 pts
- **Level multiplier**: × level
- **Combo**: consecutive kills without being hit = ×1.1 per kill, max ×3.0, resets on hit

### Lives & Progression
- 3 lives; last life lost → game over with final score
- 5 waves cleared → level up, enemies scale ×1.15 speed and HP
- 5 levels → endless loop with continued scaling

---

## Boss Design

**APEX** — enters from top-center at wave 5.

**HP**: 600 × level

**Phases** (at 66% and 33% HP):
- **Phase 1**: 4-way radial burst every 2.0s + 2 Drifter spawns every 4s
- **Phase 2**: 8-way radial burst every 1.5s + aimed shot every 1.0s
- **Phase 3**: 12-way radial every 1.2s + aimed every 0.8s + ±30° sweeping aimed

Boss glyph switches to damaged variant below 33% HP.

---

## Wave System

```go
type WaveEntry struct {
    Delay     float64
    EnemyType EnemyType
    X         float64  // spawn x
    Y         float64  // spawn y (usually -2)
}
```

**Level 1 waves**:
| Wave | Enemies | Pattern |
|------|---------|---------|
| 1 | 4× Drifter | Spread across columns |
| 2 | 6× Weaver | Three pairs, staggered 0.5s |
| 3 | 4× Drifter + 2× Anchor | Anchors enter first |
| 4 | 8× Weaver + 2× Anchor | Crossing sine paths |
| 5 | APEX | Descends from top-center |

Wave advances when all enemies and enemy bullets are gone.

---

## State Machine

```
idle → playing → paused → waveClearing → bossEntry → gameOver
```

---

## Collision Detection

Per frame (float64 space, not braille space):
1. Player bullets × enemies: circle r = 0.4 + enemy_r
2. Enemy bullets × player: circle r = 0.4 + 0.4
3. Enemy body × player: circle r = enemy_r + 0.4

Enemy radii: Drifter=0.4, Weaver=0.3, Anchor=0.5, Boss=1.5

Skip player collision during invincibility window.

---

## File Layout

```
games/nova/
  nova.go         — Nova struct, Game interface, state machine, Update/Render
  canvas.go       — BrailleCanvas type, Bresenham line, midpoint circle, Draw* methods
  sprites.go      — drawPlayer(), drawDrifter(), drawWeaver(), drawAnchor(), drawBoss(),
                    drawPlayerBullet(), drawEnemyBullet(), drawExplosion()
  entities.go     — Player, Enemy, Bullet, Explosion structs + constructors
  wave.go         — WaveEntry, WaveScript, per-level wave tables
  render.go       — renderPlayfield(), renderSidebar(), renderControls()
  collision.go    — circleCollide(ax, ay, ar, bx, by, br float64) bool
```

---

## Implementation Phases

### Phase 1 — BrailleCanvas + Core Engine
- Implement `BrailleCanvas` with Bresenham line, midpoint circle, `Row()` renderer
- `Nova` struct: player, `[]Bullet`, `[]Enemy`, lives, score, elapsed `float64`
- 4-directional player movement (float64), auto-fire, lives counter
- Player sprite: delta-wing 5-line hull via `drawPlayer(canvas, x, y, elapsed)`
- Single enemy type: Drifter, diamond sprite, straight descent, no fire
- Collision: player bullets kill Drifters; Drifter body kills player (float64 circles)
- `Render()`: BrailleCanvas playfield rows + sidebar + controls
- `HandleInput()`: arrows, space (bomb stub), p, q
- Clear 10 Drifters → "WAVE CLEAR" → game over
- **Deliverable**: playable prototype with vector sprites

### Phase 2 — Wave System + Enemy Variety
- Weaver (triangle+spikes) and Anchor (compact diamond) sprites and AI
- `WaveScript` spawn system; level 1 all 5 waves
- Enemy bullets (aimed dot, 3-way spread ring) with collision
- Bomb mechanic
- Sprite animations: Drifter accent toggle, Weaver spike pulse, Anchor diagonal toggle
- Combo multiplier
- Level advance: wave 5 cleared → level up ×1.15
- **Deliverable**: full level 1 end-to-end

### Phase 3 — Boss Fight
- APEX sprite (hexagonal frame + cross + corner diamonds)
- 3-phase attack AI: radial burst helper, aimed shot, sweep
- Boss HP bar in sidebar
- Damaged glyph at 33% HP (dim frame + bright cross)
- 3-frame explosion death (flash → ring → fade)
- Victory screen
- **Deliverable**: complete game loop

### Phase 4 — Polish
- 5-level scaling
- Endless mode after level 5
- Player invincibility blink (dim hull at 8 Hz during i-frames)
- Expanding circle explosion effect (DrawCircle at growing radius, fading color)
- Difficulty option: Normal / Hard (Hard: +25% bullet speed, -1 starting life)
- **Deliverable**: shippable game

---

## Integration (cmd/main.go)

1. Add `menuNovaOptions` to `menuState` enum
2. Add `"NOVA"` to the main menu
3. Wire `nova.NewNova(difficulty)` into the game instantiation block
4. `GetLines()` returns 0

---

## What We're Stealing from Bladedancer

| Concept | Bladedancer source | How we port it |
|---------|-------------------|----------------|
| Bresenham line algorithm | `pixel_canvas.dart:46-78` | Direct Go port |
| Midpoint circle (8-way symmetry) | `pixel_canvas.dart:17-44` | Direct Go port |
| Delta-wing ship hull (5 lines) | `dart_ship_painter.dart:58-79` | Same coords on BrailleCanvas |
| Drifter diamond (4 lines + accent) | `stage1_visuals.dart:9-34` | Same draw calls |
| Swirl triangle + spikes | `stage1_visuals.dart:36-64` | Scaled to braille |
| Fairy compact diamond | `stage1_visuals.dart:121-147` | Scaled to braille |
| Binary-toggle animation (4/3/8 Hz) | `animation_math.dart` | `(elapsed * hz) % 2 < 1` |
| Result memoization for line/circle | `_lineCache`, `_circleCache` | `sync.Map` in canvas.go |
