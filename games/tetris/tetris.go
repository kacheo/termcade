package tetris

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"tmvgs/core/ui"
)

type Tetris struct {
	board      *Board
	rng        []byte
	rngIndex   int
	current    *Piece
	next       *Piece
	score      int
	level      int
	lines      int
	gameOver   bool
	paused     bool
	lastDrop   time.Time
	lockTimer  time.Time
	onGround   bool
	ghost      bool
	startLevel int
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewTetris(ghost bool, startLevel int) *Tetris {
	t := &Tetris{
		board:      NewBoard(),
		rng:        make([]byte, 7),
		rngIndex:   0,
		level:      startLevel,
		lastDrop:   time.Now(),
		lockTimer:  time.Now(),
		ghost:      ghost,
		startLevel: startLevel,
	}
	t.shuffleRNG()
	t.spawnPiece()
	return t
}

func (t *Tetris) shuffleRNG() {
	t.rng = []byte{'I', 'O', 'T', 'S', 'Z', 'J', 'L'}
	rand.Shuffle(len(t.rng), func(i, j int) {
		t.rng[i], t.rng[j] = t.rng[j], t.rng[i]
	})
	t.rngIndex = 0
}

func (t *Tetris) nextPieceType() byte {
	if t.rngIndex >= len(t.rng) {
		t.shuffleRNG()
	}
	pt := t.rng[t.rngIndex]
	t.rngIndex++
	return pt
}

func (t *Tetris) spawnPiece() {
	pieceType := t.nextPieceType()
	t.current = &Piece{
		Type:     pieceType,
		X:        4,
		Y:        0,
		Rotation: 0,
		Color:    pieceType,
	}
	nextPieceType := t.nextPieceType()
	if t.next == nil {
		t.next = &Piece{}
	}
	t.next.Type = nextPieceType
	t.next.X = 4
	t.next.Y = 0
	t.next.Rotation = 0
	t.next.Color = nextPieceType

	if t.board.Collides(t.current) {
		t.gameOver = true
	}
	t.onGround = false
	t.lockTimer = time.Now()
}

func (t *Tetris) Name() string        { return "Tetris" }
func (t *Tetris) Description() string  { return "Classic block-stacking puzzle" }
func (t *Tetris) IsPaused() bool       { return t.paused }
func (t *Tetris) IsGameOver() bool     { return t.gameOver }
func (t *Tetris) GetScore() int        { return t.score }
func (t *Tetris) GetLevel() int        { return t.level }
func (t *Tetris) GetLines() int        { return t.lines }

func (t *Tetris) Update(delta time.Duration) error {
	if t.gameOver || t.paused {
		return nil
	}

	// Auto drop
	interval := getDropInterval(t.level)
	if time.Since(t.lastDrop) >= interval {
		if !t.move(0, 1) {
			t.onGround = true
			t.lockTimer = time.Now()
		} else {
			t.onGround = false
		}
		t.lastDrop = time.Now()
	}

	// Lock delay
	if t.onGround && time.Since(t.lockTimer) >= LockDelay {
		t.lock()
	}

	return nil
}

func (t *Tetris) move(dx, dy int) bool {
	if t.current == nil {
		return false
	}
	np := &Piece{X: t.current.X + dx, Y: t.current.Y + dy, Type: t.current.Type, Rotation: t.current.Rotation, Color: t.current.Color}
	if !t.board.Collides(np) {
		t.current.X += dx
		t.current.Y += dy
		if dy > 0 {
			t.onGround = false
		}
		return true
	}
	return false
}

func (t *Tetris) rotate() bool {
	if t.current == nil {
		return false
	}
	oldRot := t.current.Rotation
	t.current.Rotation = (t.current.Rotation + 1) % 4
	if t.board.Collides(t.current) {
		t.current.Rotation = oldRot
		return false
	}
	t.onGround = false
	return true
}

func (t *Tetris) lock() {
	if t.current == nil {
		return
	}
	t.board.Lock(t.current)
	cleared := t.board.ClearLines()
	t.lines += cleared
	if cleared > 0 {
		t.score += []int{0, 100, 300, 500, 800}[cleared] * (t.level + 1)
	}
	t.level = t.lines / 10
	if t.level > 20 {
		t.level = 20
	}
	t.current = t.next
	t.spawnPiece()
}

func (t *Tetris) HandleInput(key string) {
	if t.gameOver {
		return
	}
	switch key {
	case "left":
		t.move(-1, 0)
	case "right":
		t.move(1, 0)
	case "down":
		if t.move(0, 1) {
			t.score++
		}
	case "up":
		t.rotate()
	case " ":
		for t.move(0, 1) {
			t.score += 2
		}
		t.lock()
	case "p":
		t.paused = !t.paused
	case "q":
		t.gameOver = true
	}
}

func (t *Tetris) Render() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("  ╔════════════════════════════════════╗\n")
	sb.WriteString("  ║                                    ║\n")

	for y := 0; y < BoardHeight; y++ {
		sb.WriteString("  ║  ")
		for x := 0; x < BoardWidth; x++ {
			c := t.board.Cell(x, y)
			if c != 0 {
				color := ui.GetPieceColor(c)
				sb.WriteString(lipgloss.NewStyle().Foreground(color).Render("██"))
			} else if t.current != nil {
				pieceCell := false
				for _, cell := range getCells(t.current) {
					if t.current.X+cell.X == x && t.current.Y+cell.Y == y {
						pieceCell = true
						break
					}
				}
				if pieceCell {
					color := ui.GetPieceColor(t.current.Color)
					sb.WriteString(lipgloss.NewStyle().Foreground(color).Render("██"))
				} else {
					sb.WriteString(lipgloss.NewStyle().Foreground(ui.ColorGray).Render("··"))
				}
			} else {
				sb.WriteString(lipgloss.NewStyle().Foreground(ui.ColorGray).Render("··"))
			}
		}
		sb.WriteString("  ║\n")
	}

	sb.WriteString("  ║                                    ║\n")
	sb.WriteString("  ╠════════════════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("  ║  SCORE: %-5d  LEVEL: %-2d  LINES: %-2d  ║\n", t.score, t.level, t.lines))
	sb.WriteString("  ║  [P] Pause                [Q] Quit ║\n")
	sb.WriteString("  ╚════════════════════════════════════╝\n")

	return sb.String()
}