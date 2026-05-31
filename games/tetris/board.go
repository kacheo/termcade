package tetris

import (
	"time"
)

const (
	BoardWidth  = 10
	BoardHeight = 20
	LockDelay   = 250 * time.Millisecond
)

type Piece struct {
	Type     byte
	X, Y     int
	Rotation int
	Color    byte
}

type Position struct {
	X, Y int
}

var (
	I pieceData = pieceData{
		Rotations: [][]Position{
			{{0, 0}, {1, 0}, {2, 0}, {3, 0}},
			{{1, -1}, {1, 0}, {1, 1}, {1, 2}},
			{{0, 1}, {1, 1}, {2, 1}, {3, 1}},
			{{2, -1}, {2, 0}, {2, 1}, {2, 2}},
		},
	}
	O pieceData = pieceData{
		Rotations: [][]Position{
			{{0, 0}, {1, 0}, {0, 1}, {1, 1}},
			{{0, 0}, {1, 0}, {0, 1}, {1, 1}},
			{{0, 0}, {1, 0}, {0, 1}, {1, 1}},
			{{0, 0}, {1, 0}, {0, 1}, {1, 1}},
		},
	}
	T pieceData = pieceData{
		Rotations: [][]Position{
			{{0, 0}, {1, 0}, {2, 0}, {1, 1}},
			{{1, -1}, {1, 0}, {1, 1}, {2, 0}},
			{{1, -1}, {0, 0}, {1, 0}, {2, 0}},
			{{1, -1}, {0, 0}, {1, 0}, {1, 1}},
		},
	}
	S pieceData = pieceData{
		Rotations: [][]Position{
			{{1, -1}, {2, -1}, {0, 0}, {1, 0}},
			{{1, -1}, {1, 0}, {2, 0}, {2, 1}},
			{{1, -1}, {2, -1}, {0, 0}, {1, 0}},
			{{0, -1}, {0, 0}, {1, 0}, {1, 1}},
		},
	}
	Z pieceData = pieceData{
		Rotations: [][]Position{
			{{0, -1}, {1, -1}, {1, 0}, {2, 0}},
			{{2, -1}, {1, 0}, {2, 0}, {1, 1}},
			{{0, 0}, {1, 0}, {1, 1}, {2, 1}},
			{{1, -1}, {0, 0}, {1, 0}, {0, 1}},
		},
	}
	J pieceData = pieceData{
		Rotations: [][]Position{
			{{0, -1}, {0, 0}, {1, 0}, {2, 0}},
			{{1, -1}, {2, -1}, {1, 0}, {1, 1}},
			{{0, 0}, {1, 0}, {2, 0}, {2, 1}},
			{{1, -1}, {1, 0}, {0, 1}, {1, 1}},
		},
	}
	L pieceData = pieceData{
		Rotations: [][]Position{
			{{2, -1}, {0, 0}, {1, 0}, {2, 0}},
			{{1, -1}, {1, 0}, {1, 1}, {2, 1}},
			{{0, 0}, {1, 0}, {2, 0}, {0, 1}},
			{{1, -1}, {2, -1}, {1, 0}, {1, 1}},
		},
	}
)

type pieceData struct {
	Rotations [][]Position
}

func getCells(p *Piece) []Position {
	switch p.Type {
	case 'I':
		return I.Rotations[p.Rotation]
	case 'O':
		return O.Rotations[p.Rotation]
	case 'T':
		return T.Rotations[p.Rotation]
	case 'S':
		return S.Rotations[p.Rotation]
	case 'Z':
		return Z.Rotations[p.Rotation]
	case 'J':
		return J.Rotations[p.Rotation]
	case 'L':
		return L.Rotations[p.Rotation]
	}
	return nil
}

func getDropInterval(level int) time.Duration {
	intervals := []int{800, 720, 630, 550, 480, 450, 380, 330, 280, 230,
		200, 180, 160, 140, 120, 100, 80, 60, 50, 50}
	if level >= len(intervals) {
		return 50 * time.Millisecond
	}
	return time.Duration(intervals[level]) * time.Millisecond
}

type Board struct {
	grid [BoardHeight][BoardWidth]byte
}

type LockResult struct {
	TSpin       bool
	Cleared     int
	ClearedRows []int
	Combo       int
	BackToBack  int
	ScoreDelta  int
}

func NewBoard() *Board {
	return &Board{}
}

func (b *Board) Cell(x, y int) byte {
	if x < 0 || x >= BoardWidth || y < 0 || y >= BoardHeight {
		return 0
	}
	return b.grid[y][x]
}

func (b *Board) SetCell(x, y int, v byte) {
	if x >= 0 && x < BoardWidth && y >= 0 && y < BoardHeight {
		b.grid[y][x] = v
	}
}

func (b *Board) Collides(p *Piece) bool {
	return b.CollidesAt(p, p.X, p.Y, p.Rotation)
}

func (b *Board) CollidesAt(p *Piece, x, y, rotation int) bool {
	for _, c := range getCellsAt(p.Type, rotation) {
		cx := x + c.X
		cy := y + c.Y
		if cx < 0 || cx >= BoardWidth || cy >= BoardHeight {
			return true
		}
		if cy >= 0 && b.grid[cy][cx] != 0 {
			return true
		}
	}
	return false
}

func getCellsAt(pieceType byte, rotation int) []Position {
	switch pieceType {
	case 'I':
		return I.Rotations[rotation]
	case 'O':
		return O.Rotations[rotation]
	case 'T':
		return T.Rotations[rotation]
	case 'S':
		return S.Rotations[rotation]
	case 'Z':
		return Z.Rotations[rotation]
	case 'J':
		return J.Rotations[rotation]
	case 'L':
		return L.Rotations[rotation]
	}
	return nil
}

func (b *Board) Lock(p *Piece) {
	for _, c := range getCells(p) {
		x := p.X + c.X
		y := p.Y + c.Y
		if y >= 0 {
			b.SetCell(x, y, p.Color)
		}
	}
}

func (b *Board) ClearLines() (int, []int) {
	cleared := 0
	rows := []int{}
	for y := BoardHeight - 1; y >= 0; y-- {
		full := true
		for x := 0; x < BoardWidth; x++ {
			if b.grid[y][x] == 0 {
				full = false
				break
			}
		}
		if full {
			cleared++
			rows = append(rows, y)
		}
	}
	if cleared == 0 {
		return 0, nil
	}
	b.removeRows(rows)
	return cleared, rows
}

func (b *Board) FullRows() []int {
	rows := []int{}
	for y := BoardHeight - 1; y >= 0; y-- {
		full := true
		for x := 0; x < BoardWidth; x++ {
			if b.grid[y][x] == 0 {
				full = false
				break
			}
		}
		if full {
			rows = append(rows, y)
		}
	}
	return rows
}

func (b *Board) removeRows(rows []int) {
	if len(rows) == 0 {
		return
	}
	rowsMap := make(map[int]struct{}, len(rows))
	for _, row := range rows {
		if row >= 0 && row < BoardHeight {
			rowsMap[row] = struct{}{}
		}
	}
	if len(rowsMap) == 0 {
		return
	}
	dst := BoardHeight - 1
	for src := BoardHeight - 1; src >= 0; src-- {
		if _, remove := rowsMap[src]; remove {
			continue
		}
		if dst != src {
			copy(b.grid[dst][:], b.grid[src][:])
		}
		dst--
	}
	for ; dst >= 0; dst-- {
		for x := 0; x < BoardWidth; x++ {
			b.grid[dst][x] = 0
		}
	}
}