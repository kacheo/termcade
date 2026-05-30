package tetris

import (
	"time"
)

const (
	BoardWidth  = 10
	BoardHeight = 20
	LockDelay   = 500 * time.Millisecond
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
			{{0, 0}, {1, 0}, {1, 1}, {1, 2}},
		},
	}
	S pieceData = pieceData{
		Rotations: [][]Position{
			{{1, -1}, {2, -1}, {0, 0}, {1, 0}},
			{{1, -1}, {1, 0}, {2, 0}, {2, 1}},
			{{0, 0}, {1, 0}, {1, 1}, {2, 1}},
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
	for _, c := range getCells(p) {
		x := p.X + c.X
		y := p.Y + c.Y
		if x < 0 || x >= BoardWidth || y >= BoardHeight {
			return true
		}
		if y >= 0 && b.grid[y][x] != 0 {
			return true
		}
	}
	return false
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

func (b *Board) ClearLines() int {
	cleared := 0
	for y := BoardHeight - 1; y >= 0; y-- {
		full := true
		for x := 0; x < BoardWidth; x++ {
			if b.grid[y][x] == 0 {
				full = false
				break
			}
		}
		if full {
			for vy := y; vy > 0; vy-- {
				for x := 0; x < BoardWidth; x++ {
					b.grid[vy][x] = b.grid[vy-1][x]
				}
			}
			for x := 0; x < BoardWidth; x++ {
				b.grid[0][x] = 0
			}
			cleared++
			y++
		}
	}
	return cleared
}