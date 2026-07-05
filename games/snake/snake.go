package snake

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/kacheo/termcade/core"
)

const (
	BoardWidth   = 20
	BoardHeight  = 20
	baseTickMS   = 200
	speedupMS    = 15
	foodPerLevel = 5
	maxLevel     = 10
)

type direction int

const (
	dirRight direction = iota
	dirDown
	dirLeft
	dirUp
)

var dirDelta = map[direction]core.Position{
	dirRight: {X: 1, Y: 0},
	dirDown:  {X: 0, Y: 1},
	dirLeft:  {X: -1, Y: 0},
	dirUp:    {X: 0, Y: -1},
}

var opposite = map[direction]direction{
	dirRight: dirLeft,
	dirLeft:  dirRight,
	dirUp:    dirDown,
	dirDown:  dirUp,
}

type Snake struct {
	body         []core.Position
	dir          direction
	nextDir      direction
	food         core.Position
	score        int
	foodEaten    int
	gameOver     bool
	paused       bool
	elapsed      time.Duration
	tickInterval time.Duration
	rng          *rand.Rand
}

func NewSnake() *Snake {
	s := &Snake{
		dir:          dirRight,
		nextDir:      dirRight,
		tickInterval: time.Duration(baseTickMS) * time.Millisecond,
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	cx, cy := BoardWidth/2, BoardHeight/2
	s.body = []core.Position{
		{X: cx, Y: cy},
		{X: cx - 1, Y: cy},
		{X: cx - 2, Y: cy},
	}
	s.spawnFood()
	return s
}

func (s *Snake) spawnFood() {
	occupied := make(map[[2]int]bool, len(s.body))
	for _, p := range s.body {
		occupied[[2]int{p.X, p.Y}] = true
	}
	// Fast path: random sampling
	for range 100 {
		p := core.Position{
			X: s.rng.Intn(BoardWidth),
			Y: s.rng.Intn(BoardHeight),
		}
		if !occupied[[2]int{p.X, p.Y}] {
			s.food = p
			return
		}
	}
	// Slow path: exhaustive scan when board is nearly full
	for y := range BoardHeight {
		for x := range BoardWidth {
			if !occupied[[2]int{x, y}] {
				s.food = core.Position{X: x, Y: y}
				return
			}
		}
	}
	// Board completely full — player wins
	s.gameOver = true
}

func (s *Snake) tickInterval_for(level int) time.Duration {
	ms := baseTickMS - (level-1)*speedupMS
	if ms < 50 {
		ms = 50
	}
	return time.Duration(ms) * time.Millisecond
}

func (s *Snake) Update(delta time.Duration) error {
	if s.gameOver || s.paused {
		return nil
	}
	s.elapsed += delta
	for s.elapsed >= s.tickInterval {
		s.elapsed -= s.tickInterval
		s.step()
		if s.gameOver {
			return nil
		}
	}
	return nil
}

func (s *Snake) step() {
	if opposite[s.nextDir] != s.dir {
		s.dir = s.nextDir
	}

	d := dirDelta[s.dir]
	head := s.body[0]
	newHead := core.Position{X: head.X + d.X, Y: head.Y + d.Y}

	if newHead.X < 0 || newHead.X >= BoardWidth || newHead.Y < 0 || newHead.Y >= BoardHeight {
		s.gameOver = true
		return
	}

	eating := newHead == s.food

	// When not eating, the tail vacates its cell this step — exclude it from collision.
	checkBody := s.body
	if !eating {
		checkBody = s.body[:len(s.body)-1]
	}
	for _, p := range checkBody {
		if p == newHead {
			s.gameOver = true
			return
		}
	}

	if eating {
		s.body = append([]core.Position{newHead}, s.body...)
		s.foodEaten++
		s.score += 10 * s.GetLevel()
		s.tickInterval = s.tickInterval_for(s.GetLevel())
		s.spawnFood()
	} else {
		s.body = append([]core.Position{newHead}, s.body[:len(s.body)-1]...)
	}
}

func (s *Snake) HandleInput(key string) {
	switch key {
	case "up":
		s.nextDir = dirUp
	case "down":
		s.nextDir = dirDown
	case "left":
		s.nextDir = dirLeft
	case "right":
		s.nextDir = dirRight
	case "q":
		s.gameOver = true
	}
}

func (s *Snake) Name() string        { return "Snake" }
func (s *Snake) IsPaused() bool      { return s.paused }
func (s *Snake) IsGameOver() bool    { return s.gameOver }
func (s *Snake) GetScore() int       { return s.score }
func (s *Snake) GetLines() int       { return s.foodEaten }
func (s *Snake) Description() string { return "Classic snake — eat food, grow longer, don't hit walls or yourself" }

func (s *Snake) GetLevel() int {
	level := s.foodEaten/foodPerLevel + 1
	if level > maxLevel {
		level = maxLevel
	}
	return level
}

var (
	borderSty = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	headSty   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	bodySty   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AA00"))
	foodSty   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
	emptySty  = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
	textSty   = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	labelSty  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)

func (s *Snake) Render() string {
	bodySet := make(map[[2]int]bool, len(s.body))
	for _, p := range s.body {
		bodySet[[2]int{p.X, p.Y}] = true
	}
	head := s.body[0]

	boardLines := make([]string, 0, BoardHeight+2)
	topBorder := borderSty.Render("╔" + strings.Repeat("═", BoardWidth*2) + "╗")
	boardLines = append(boardLines, topBorder)
	for y := range BoardHeight {
		var row strings.Builder
		row.WriteString(borderSty.Render("║"))
		for x := range BoardWidth {
			p := core.Position{X: x, Y: y}
			switch {
			case p == head:
				row.WriteString(headSty.Render("██"))
			case bodySet[[2]int{x, y}]:
				row.WriteString(bodySty.Render("██"))
			case p == s.food:
				row.WriteString(foodSty.Render("◆◆"))
			default:
				row.WriteString(emptySty.Render("··"))
			}
		}
		row.WriteString(borderSty.Render("║"))
		boardLines = append(boardLines, row.String())
	}
	botBorder := borderSty.Render("╚" + strings.Repeat("═", BoardWidth*2) + "╝")
	boardLines = append(boardLines, botBorder)

	sideLines := make([]string, 0, BoardHeight+2)
	sideLines = append(sideLines, "")
	sideLines = append(sideLines, "")
	sideLines = append(sideLines, labelSty.Render(" SCORE"))
	sideLines = append(sideLines, textSty.Render(fmt.Sprintf(" %d", s.score)))
	sideLines = append(sideLines, "")
	sideLines = append(sideLines, labelSty.Render(" LEVEL"))
	sideLines = append(sideLines, textSty.Render(fmt.Sprintf(" %d", s.GetLevel())))
	sideLines = append(sideLines, "")
	sideLines = append(sideLines, labelSty.Render(" LENGTH"))
	sideLines = append(sideLines, textSty.Render(fmt.Sprintf(" %d", len(s.body))))
	for len(sideLines) < len(boardLines) {
		sideLines = append(sideLines, "")
	}

	var sb strings.Builder
	sb.WriteString("\n")
	for i, bl := range boardLines {
		sl := ""
		if i < len(sideLines) {
			sl = sideLines[i]
		}
		sb.WriteString(bl + "  " + sl + "\n")
	}
	sb.WriteString("\n")
	sb.WriteString(labelSty.Render("  [←→↑↓] Move   [P] Pause   [Q] Quit") + "\n")
	return sb.String()
}
