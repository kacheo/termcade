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
	board       *Board
	rng         []byte
	rngIndex    int
	current     *Piece
	next        *Piece
	held        *Piece
	holdUsed    bool
	score       int
	level       int
	lines       int
	gameOver    bool
	paused      bool
	lastDrop    time.Time
	lockTimer   time.Time
	lockStart   time.Time
	onGround    bool
	ghost       bool
	startLevel  int
	lastRotate  bool
	combo       int
	backToBack  int
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
	if t.next == nil {
		// First spawn: draw current from bag
		pieceType := t.nextPieceType()
		t.current = &Piece{Type: pieceType, X: 4, Y: 0, Rotation: 0, Color: pieceType}
	} else {
		// Subsequent spawns: promote queued next to current
		t.current = t.next
		t.current.X = 4
		t.current.Y = 0
		t.current.Rotation = 0
	}
	// Always draw a fresh next from bag
	nextType := t.nextPieceType()
	t.next = &Piece{Type: nextType, X: 4, Y: 0, Rotation: 0, Color: nextType}

	if t.board.Collides(t.current) {
		t.gameOver = true
	}
	t.onGround = false
	t.lockTimer = time.Now()
	t.holdUsed = false
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
		if dx != 0 {
			t.lastRotate = false
		}
		return true
	}
	return false
}

func (t *Tetris) rotate() bool {
	if t.current == nil {
		return false
	}
	newRot := (t.current.Rotation + 1) % 4

	// Try rotation at current position first
	if !t.board.CollidesAt(t.current, t.current.X, t.current.Y, newRot) {
		t.current.Rotation = newRot
		t.lastRotate = true
		t.onGround = false
		return true
	}

	// Wall kicks: try -1, 1, -2, 2
	for _, dx := range []int{-1, 1, -2, 2} {
		if !t.board.CollidesAt(t.current, t.current.X+dx, t.current.Y, newRot) {
			t.current.X += dx
			t.current.Rotation = newRot
			t.lastRotate = true
			t.onGround = false
			return true
		}
	}

	return false
}

func (t *Tetris) lock() LockResult {
	if t.current == nil {
		return LockResult{}
	}

	result := LockResult{}
	result.TSpin = t.isTSpin()
	t.board.Lock(t.current)
	cleared, rows := t.board.ClearLines()
	result.Cleared = cleared
	result.ClearedRows = rows

	// Score calculation
	if result.TSpin {
		scoreTable := []int{400, 800, 1200, 1600}
		if result.Cleared >= 0 && result.Cleared < len(scoreTable) {
			result.ScoreDelta = scoreTable[result.Cleared] * (t.level + 1)
		}
	} else if result.Cleared > 0 {
		scoreTable := []int{0, 100, 300, 500, 800}
		result.ScoreDelta = scoreTable[result.Cleared] * (t.level + 1)
	}

	if result.ScoreDelta > 0 {
		t.score += result.ScoreDelta
	}

	if result.Cleared > 0 {
		t.lines += result.Cleared
		t.level = t.lines / 10
		if t.level > 20 {
			t.level = 20
		}
		t.combo++
		result.Combo = t.combo
		qualifiesB2B := (result.TSpin && result.Cleared > 0) || result.Cleared == 4
		if qualifiesB2B {
			t.backToBack++
		} else {
			t.backToBack = 0
		}
		result.BackToBack = t.backToBack
	} else {
		t.combo = 0
		t.backToBack = 0
	}

	t.spawnPiece()
	t.lastRotate = false
	return result
}

func (t *Tetris) doHold() {
	if t.holdUsed {
		return
	}
	if t.held == nil {
		// First hold: save current type, pull next piece into play
		t.held = &Piece{Type: t.current.Type, Color: t.current.Color}
		t.spawnPiece() // resets holdUsed to false internally
	} else {
		// Swap: held piece becomes new current; current goes to hold
		swapped := &Piece{Type: t.held.Type, Color: t.held.Color, X: 4, Y: 0, Rotation: 0}
		t.held = &Piece{Type: t.current.Type, Color: t.current.Color}
		t.current = swapped
		if t.board.Collides(t.current) {
			t.gameOver = true
			return
		}
		t.onGround = false
		t.lockTimer = time.Now()
	}
	t.holdUsed = true
}

// doQueue swaps the held piece with the next piece in queue.
// The currently-falling piece is unaffected; nothing is lost.
func (t *Tetris) doQueue() {
	if t.held == nil || t.next == nil {
		return
	}
	t.held, t.next =
		&Piece{Type: t.next.Type, Color: t.next.Color},
		&Piece{Type: t.held.Type, Color: t.held.Color, X: 4, Y: 0}
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
	case "c":
		t.doHold()
	case "z":
		t.doQueue()
	case "p":
		t.paused = !t.paused
	case "q":
		t.gameOver = true
	}
}

func (t *Tetris) ghostY() int {
	if t.current == nil {
		return 0
	}
	gy := t.current.Y
	for {
		test := &Piece{X: t.current.X, Y: gy + 1, Type: t.current.Type, Rotation: t.current.Rotation, Color: t.current.Color}
		if t.board.Collides(test) {
			break
		}
		gy++
	}
	return gy
}

func (t *Tetris) isTSpin() bool {
	if t.current == nil || t.current.Type != 'T' {
		return false
	}
	if !t.lastRotate {
		return false
	}
	cx := t.current.X + 1
	cy := t.current.Y + 1
	corners := [][2]int{
		{cx - 1, cy - 1},
		{cx + 1, cy - 1},
		{cx - 1, cy + 1},
		{cx + 1, cy + 1},
	}
	filled := 0
	for _, c := range corners {
		x, y := c[0], c[1]
		if x < 0 || x >= BoardWidth || y < 0 || y >= BoardHeight {
			filled++
		} else if t.board.grid[y][x] != 0 {
			filled++
		}
	}
	return filled >= 3
}

// renderMiniPiece renders a piece as a 4×4 mini-grid, always at rotation 0.
// Returns 4 strings each representing one row (8 visible chars = 4 cells × 2).
func renderMiniPiece(p *Piece) [4]string {
	grayCell := lipgloss.NewStyle().Foreground(ui.ColorGray).Render("··")
	emptyRow := grayCell + grayCell + grayCell + grayCell

	if p == nil {
		return [4]string{emptyRow, emptyRow, emptyRow, emptyRow}
	}

	cells := getCells(&Piece{Type: p.Type, Color: p.Color, Rotation: 0})

	// Normalize: shift cells so minimum Y is 0
	minY := 0
	for _, c := range cells {
		if c.Y < minY {
			minY = c.Y
		}
	}
	offset := -minY

	var grid [4][4]bool
	for _, c := range cells {
		y := c.Y + offset
		if y >= 0 && y < 4 && c.X >= 0 && c.X < 4 {
			grid[y][c.X] = true
		}
	}

	color := ui.GetPieceColor(p.Color)
	var rows [4]string
	for y := 0; y < 4; y++ {
		var row strings.Builder
		for x := 0; x < 4; x++ {
			if grid[y][x] {
				row.WriteString(lipgloss.NewStyle().Foreground(color).Render("██"))
			} else {
				row.WriteString(grayCell)
			}
		}
		rows[y] = row.String()
	}
	return rows
}

// renderBoardLines returns the 22-line left panel (border + 20 board rows + border).
func (t *Tetris) renderBoardLines() []string {
	ghostCells := map[[2]int]bool{}
	currentCells := map[[2]int]bool{}
	if t.current != nil {
		for _, cell := range getCells(t.current) {
			currentCells[[2]int{t.current.X + cell.X, t.current.Y + cell.Y}] = true
		}
		if t.ghost {
			gy := t.ghostY()
			for _, cell := range getCells(t.current) {
				ghostCells[[2]int{t.current.X + cell.X, gy + cell.Y}] = true
			}
		}
	}

	lines := make([]string, 0, 22)
	lines = append(lines, "╔════════════════════════╗")

	for y := 0; y < BoardHeight; y++ {
		var row strings.Builder
		row.WriteString("║  ")
		for x := 0; x < BoardWidth; x++ {
			pos := [2]int{x, y}
			c := t.board.Cell(x, y)
			switch {
			case c != 0:
				row.WriteString(lipgloss.NewStyle().Foreground(ui.GetPieceColor(c)).Render("██"))
			case currentCells[pos]:
				row.WriteString(lipgloss.NewStyle().Foreground(ui.GetPieceColor(t.current.Color)).Render("██"))
			case ghostCells[pos]:
				row.WriteString(lipgloss.NewStyle().Foreground(ui.ColorGray).Render("░░"))
			default:
				row.WriteString(lipgloss.NewStyle().Foreground(ui.ColorGray).Render("··"))
			}
		}
		row.WriteString("  ║")
		lines = append(lines, row.String())
	}

	lines = append(lines, "╚════════════════════════╝")
	return lines
}

// renderSidebarLines returns the 22-line right panel with NEXT, HOLD, stats, and hints.
func (t *Tetris) renderSidebarLines() []string {
	border := lipgloss.NewStyle().Foreground(ui.ColorBorder)
	text := lipgloss.NewStyle().Foreground(ui.ColorText)

	sideW := 14 // inner visible width

	pad := func(s string, w int) string {
		// Pad a plain string to visible width w (no ANSI in s assumed here)
		for len(s) < w {
			s += " "
		}
		return s
	}

	fullRow := func(content string) string {
		return border.Render("║") + content + border.Render("║")
	}
	sepRow := border.Render("╠══════════════╣")

	nextGrid := renderMiniPiece(t.next)
	holdGrid := renderMiniPiece(t.held)

	// Each mini-grid row: 3 spaces + 8 visible grid chars + 3 spaces = 14 inner
	miniRow := func(gridRow string) string {
		return fullRow("   " + gridRow + "   ")
	}

	holdLabel := "    HOLD      "
	if t.held == nil {
		holdLabel = "    HOLD  ----"
	}

	lines := []string{
		border.Render("╔══════════════╗"),
		fullRow(text.Render("    NEXT      ")),
		fullRow(pad("", sideW)),
		miniRow(nextGrid[0]),
		miniRow(nextGrid[1]),
		miniRow(nextGrid[2]),
		miniRow(nextGrid[3]),
		fullRow(pad("", sideW)),
		sepRow,
		fullRow(text.Render(holdLabel)),
		fullRow(pad("", sideW)),
		miniRow(holdGrid[0]),
		miniRow(holdGrid[1]),
		miniRow(holdGrid[2]),
		miniRow(holdGrid[3]),
		fullRow(pad("", sideW)),
		sepRow,
		fullRow(text.Render(fmt.Sprintf(" SCORE: %-6d", t.score))),
		fullRow(text.Render(fmt.Sprintf(" LEVEL: %-6d", t.level))),
		fullRow(text.Render(fmt.Sprintf(" LINES: %-6d", t.lines))),
		fullRow(pad("", sideW)),
		border.Render("╚══════════════╝"),
	}
	return lines
}

func renderControlsLines() []string {
	border := lipgloss.NewStyle().Foreground(ui.ColorBorder)
	hint := lipgloss.NewStyle().Foreground(ui.ColorBorder)

	inner := 42
	row := func(s string) string {
		runes := []rune(s)
		for len(runes) < inner {
			runes = append(runes, ' ')
		}
		return border.Render("║") + hint.Render(string(runes)) + border.Render("║")
	}

	return []string{
		border.Render("╔══════════════════════════════════════════╗"),
		row(" [←→] Move  [↑] Rotate  [↓] Soft drop"),
		row(" [Spc] Hard drop  [C] Hold  [Z] Swap"),
		row(" [P] Pause  [Q] Quit"),
		border.Render("╚══════════════════════════════════════════╝"),
	}
}

func (t *Tetris) Render() string {
	board := t.renderBoardLines()
	side := t.renderSidebarLines()
	controls := renderControlsLines()

	var sb strings.Builder
	sb.WriteString("\n")
	for i := range board {
		sb.WriteString("  ")
		sb.WriteString(board[i])
		sb.WriteString("  ")
		sb.WriteString(side[i])
		sb.WriteString("\n")
	}
	for _, line := range controls {
		sb.WriteString("  ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}
