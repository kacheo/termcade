# Pong Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Pong game (Player vs AI) with configurable options following Tetris patterns.

**Architecture:** MVU pattern with Bubble Tea. Game state in struct, Update() for game logic, Render() for ASCII display, HandleInput() for keyboard.

**Tech Stack:** Go, Bubble Tea, Lipgloss

---

## File Structure

**New files:**
- `games/pong/pong.go` - Main game (~180 lines)
- `games/pong/pong_test.go` - Tests (~80 lines)

**Modified files:**
- `cmd/main.go` - Add menuPongOptions state, menu integration

---

## Task 1: Create games/pong/pong.go with Pong struct and NewPong

**Files:**
- Create: `games/pong/pong.go`

- [ ] **Step 1: Create games/pong directory**

```bash
mkdir -p games/pong
```

- [ ] **Step 2: Write Pong struct and NewPong function**

```go
package pong

import (
    "math/rand"
    "time"
)

const (
    FieldWidth  = 40
    FieldHeight = 20
    PaddleHeight = 4
    WinScore     = 5
)

type Pong struct {
    playerY      float64
    aiY         float64
    ballX       float64
    ballY       float64
    ballVX      float64
    ballVY      float64
    playerScore int
    aiScore     int
    paused      bool
    gameOver    bool
    winner      string
    speedIncrease bool
    aiDifficulty  int // 0=Easy, 1=Medium, 2=Hard
}

func NewPong(speedIncrease bool, aiDifficulty int) *Pong {
    p := &Pong{
        playerY:        0.5,
        aiY:            0.5,
        ballX:          0.5,
        ballY:          0.5,
        speedIncrease:  speedIncrease,
        aiDifficulty:  aiDifficulty,
    }
    p.resetBall(1) // 1 = right direction
    return p
}

func (p *Pong) resetBall(direction int) {
    p.ballX = 0.5
    p.ballY = 0.5 + (rand.Float64() - 0.5) * 0.3
    speed := 0.02
    p.ballVX = float64(direction) * speed
    p.ballVY = (rand.Float64() - 0.5) * speed
}
```

- [ ] **Step 3: Add interface methods**

```go
func (p *Pong) Name() string        { return "Pong" }
func (p *Pong) Description() string  { return "Classic paddle game" }
func (p *Pong) IsPaused() bool       { return p.paused }
func (p *Pong) IsGameOver() bool     { return p.gameOver }
func (p *Pong) GetScore() int        { return p.playerScore }
func (p *Pong) GetLevel() int       { return p.aiDifficulty }
func (p *Pong) GetLines() int        { return p.aiScore }
```

- [ ] **Step 4: Build to verify**

Run: `go build ./games/pong`
Expected: (no output = success)

- [ ] **Step 5: Commit**

```bash
git add games/pong/pong.go
git commit -m "feat: create Pong struct and NewPong constructor"
```

---

## Task 2: Implement Update() game logic

**Files:**
- Modify: `games/pong/pong.go`

- [ ] **Step 1: Add Update method with ball movement and collision**

```go
func (p *Pong) Update(delta time.Duration) error {
    if p.gameOver || p.paused {
        return nil
    }

    // Move ball
    p.ballX += p.ballVX
    p.ballY += p.ballVY

    // Wall collision (top/bottom)
    if p.ballY <= 0 {
        p.ballY = 0
        p.ballVY = -p.ballVY
    }
    if p.ballY >= 1 {
        p.ballY = 1
        p.ballVY = -p.ballVY
    }

    // Paddle collision (left = player)
    if p.ballX <= 0.05 && p.ballVX < 0 {
        if p.ballY >= p.playerY - 0.05 && p.ballY <= p.playerY + 0.05 {
            p.ballVX = -p.ballVX
            p.ballY += (p.ballY - p.playerY) * 0.5
            if p.speedIncrease {
                p.ballVX *= 1.1
                p.ballVY *= 1.1
            }
        }
    }

    // Paddle collision (right = AI)
    if p.ballX >= 0.95 && p.ballVX > 0 {
        if p.ballY >= p.aiY - 0.05 && p.ballY <= p.aiY + 0.05 {
            p.ballVX = -p.ballVX
            p.ballY += (p.ballY - p.aiY) * 0.5
        }
    }

    // Scoring
    if p.ballX < 0 {
        p.aiScore++
        if p.aiScore >= WinScore {
            p.gameOver = true
            p.winner = "AI"
        } else {
            p.resetBall(1)
        }
    }
    if p.ballX > 1 {
        p.playerScore++
        if p.playerScore >= WinScore {
            p.gameOver = true
            p.winner = "Player"
        } else {
            p.resetBall(-1)
        }
    }

    return nil
}
```

- [ ] **Step 2: Add AI logic**

```go
func (p *Pong) updateAI() {
    if p.gameOver || p.paused {
        return
    }

    var reactionSpeed float64
    var accuracy float64

    switch p.aiDifficulty {
    case 0: // Easy
        reactionSpeed = 0.01
        accuracy = 0.6
    case 1: // Medium
        reactionSpeed = 0.02
        accuracy = 0.8
    case 2: // Hard
        reactionSpeed = 0.04
        accuracy = 0.95
    }

    targetY := p.ballY
    if p.ballVX < 0 {
        targetY = 0.5 // Return to center when ball going away
    }

    diff := targetY - p.aiY
    if diff > reactionSpeed {
        p.aiY += reactionSpeed
    } else if diff < -reactionSpeed {
        p.aiY -= reactionSpeed
    }

    // Add imperfection
    if rand.Float64() > accuracy {
        p.aiY += (rand.Float64() - 0.5) * 0.02
    }

    // Clamp
    if p.aiY < PaddleHeight/20 {
        p.aiY = PaddleHeight / 20
    }
    if p.aiY > 1 - PaddleHeight/20 {
        p.aiY = 1 - PaddleHeight/20
    }
}
```

- [ ] **Step 3: Update Update() to call updateAI() and add time import**

Update the Update method to call updateAI() at the end, and add "time" to imports.

- [ ] **Step 4: Build and test**

Run: `go build ./games/pong && go test ./games/pong`
Expected: Success

- [ ] **Step 5: Commit**

```bash
git add games/pong/pong.go
git commit -m "feat: add Update() with ball physics and AI logic"
```

---

## Task 3: Implement Render() and HandleInput()

**Files:**
- Modify: `games/pong/pong.go`

- [ ] **Step 1: Add HandleInput method**

```go
func (p *Pong) HandleInput(key string) {
    if p.gameOver {
        return
    }
    switch key {
    case "up", "k":
        p.playerY -= 0.05
    case "down", "j":
        p.playerY += 0.05
    case "p":
        p.paused = !p.paused
    case "q":
        p.gameOver = true
        p.winner = "AI"
    }

    // Clamp player paddle
    halfPaddle := float64(PaddleHeight) / float64(FieldHeight) / 2
    if p.playerY < halfPaddle {
        p.playerY = halfPaddle
    }
    if p.playerY > 1-halfPaddle {
        p.playerY = 1 - halfPaddle
    }
}
```

- [ ] **Step 2: Add Render method**

```go
func (p *Pong) Render() string {
    var sb strings.Builder
    sb.WriteString("\n")

    // Header
    sb.WriteString("  ╔════════════════════════════════════════╗\n")
    sb.WriteString("║           PONG                          ║\n")
    sb.WriteString("  ╠════════════════════════════════════════╣\n")

    // Draw field
    for y := 0; y < FieldHeight; y++ {
        rowY := float64(y) / float64(FieldHeight)
        sb.WriteString("  ║")

        for x := 0; x < FieldWidth; x++ {
            char := " "

            // Left paddle
            if x == 2 {
                paddleTop := p.playerY - float64(PaddleHeight)/float64(FieldHeight)/2
                paddleBottom := p.playerY + float64(PaddleHeight)/float64(FieldHeight)/2
                if rowY >= paddleTop && rowY <= paddleBottom {
                    char = "█"
                }
            }

            // Right paddle (AI)
            if x == FieldWidth-3 {
                paddleTop := p.aiY - float64(PaddleHeight)/float64(FieldHeight)/2
                paddleBottom := p.aiY + float64(PaddleHeight)/float64(FieldHeight)/2
                if rowY >= paddleTop && rowY <= paddleBottom {
                    char = "█"
                }
            }

            // Ball
            ballX := int(p.ballX * float64(FieldWidth))
            if ballX == x && int(p.ballY*float64(FieldHeight)) == y {
                char = "●"
            }

            // Center line
            if x == FieldWidth/2 {
                char = "│"
            }

            sb.WriteString(char)
        }
        sb.WriteString("║\n")
    }

    // Scores
    sb.WriteString("  ╠════════════════════════════════════════╣\n")
    sb.WriteString(fmt.Sprintf("║  Player: %d          AI: %d              ║\n", p.playerScore, p.aiScore))
    sb.WriteString("  ╚════════════════════════════════════════╝\n")

    // Controls
    sb.WriteString("    [↑/↓] Move   [P] Pause   [Q] Quit\n")

    return sb.String()
}
```

- [ ] **Step 3: Build and test**

Run: `go build ./games/pong`
Expected: Success

- [ ] **Step 4: Commit**

```bash
git add games/pong/pong.go
git commit -m "feat: add Render() and HandleInput() for Pong"
```

---

## Task 4: Integrate into cmd/main.go menu system

**Files:**
- Modify: `cmd/main.go`

- [ ] **Step 1: Add menuPongOptions state and pongOpts struct**

Add to the menuState const block:
```go
menuPongOptions menuState = iota
```

Add pongOpts to the model struct:
```go
pongOpts struct {
    speedIncrease bool
    aiDifficulty  int
}
```

- [ ] **Step 2: Add pong options to main menu**

In updateMainMenu, change items to include Pong:
```go
items := []string{"Play Tetris", "Play Pong", "", "Quit"}
```

And case 0 goes to Tetris options, case 1 goes to Pong options.

- [ ] **Step 3: Add updatePongOptions and renderPongOptions functions**

Following the pattern of updateTetrisOptions and renderTetrisOptions:
- Option 0: Speed Increase (ON/OFF)
- Option 1: AI Difficulty (Easy/Medium/Hard)
- Option 2: Start Game
- Option 3: Back

- [ ] **Step 4: Wire up the menu state machine**

In Update(), add:
```go
case menuPongOptions:
    return m.updatePongOptions(msg)
```

In updateTetrisOptions, change "Play Tetris" case to go to menuTetrisOptions, add "Play Pong" case to go to menuPongOptions.

- [ ] **Step 5: Build and test**

Run: `go build ./... && go test ./...`
Expected: All pass

- [ ] **Step 6: Commit**

```bash
git add cmd/main.go
git commit -m "feat: integrate Pong into menu system"
```

---

## Task 5: Add basic tests

**Files:**
- Create: `games/pong/pong_test.go`

- [ ] **Step 1: Write basic Pong tests**

```go
package pong

import (
    "testing"
    "time"
)

func TestNewPong(t *testing.T) {
    p := NewPong(false, 1)
    if p == nil {
        t.Fatal("NewPong returned nil")
    }
    if p.playerScore != 0 {
        t.Errorf("playerScore should be 0, got %d", p.playerScore)
    }
    if p.aiScore != 0 {
        t.Errorf("aiScore should be 0, got %d", p.aiScore)
    }
    if p.gameOver {
        t.Error("gameOver should be false")
    }
}

func TestPongInterface(t *testing.T) {
    p := NewPong(false, 1)
    var _ Game = p // Verify Pong implements Game interface
}

func TestPongUpdate(t *testing.T) {
    p := NewPong(false, 1)
    p.ballX = 0.5
    p.ballY = 0.5
    p.ballVX = 0.01
    p.ballVY = 0

    initialBallX := p.ballX
    p.Update(time.Millisecond * 100)

    if p.ballX == initialBallX {
        t.Error("ball should have moved")
    }
}

func TestPongWallBounce(t *testing.T) {
    p := NewPong(false, 1)
    p.ballX = 0.5
    p.ballY = 0
    p.ballVX = 0.01
    p.ballVY = -0.01

    p.Update(time.Millisecond * 100)

    if p.ballVY >= 0 {
        t.Error("ball Y velocity should have flipped after wall hit")
    }
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./games/pong -v`
Expected: All pass

- [ ] **Step 3: Commit**

```bash
git add games/pong/pong_test.go
git commit -m "test: add Pong tests"
```

---

## Task 6: Update roadmap

**Files:**
- Modify: `docs/roadmap.md`

- [ ] **Step 1: Mark Pong as in progress**

Change "Pong (Low Priority)" to "Pong (IN PROGRESS)"

- [ ] **Step 2: Commit**

```bash
git add docs/roadmap.md
git commit -m "docs: mark Pong as in progress"
```

---

## Task 7: Final verification

- [ ] **Step 1: Run all tests**

Run: `go test ./...`
Expected: All pass

- [ ] **Step 2: Run coverage check**

Run: `make test-coverage`
Expected: Pass (80%+)

- [ ] **Step 3: Final build**

Run: `go build ./...`
Expected: Success

- [ ] **Step 4: Push**

```bash
git push -u origin feature/pong
```

---

## Verification

After all tasks complete:
1. Run `go run ./cmd` - should see "Play Pong" in menu
2. Select Pong - should see options screen
3. Start game - should see paddle and ball
4. Arrow keys should move player paddle
5. AI should respond to ball
6. Scores should increment when ball passes paddle
7. Game over at 5 points with winner message
