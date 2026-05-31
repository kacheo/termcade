package tetris

import (
	"testing"
	"time"
)

// ---- getCells ---------------------------------------------------------------

func TestGetCells_AllPiecesAllRotations(t *testing.T) {
	cases := []struct {
		pieceType byte
		rotation  int
		want      []Position
	}{
		// I
		{'I', 0, []Position{{0, 0}, {1, 0}, {2, 0}, {3, 0}}},
		{'I', 1, []Position{{1, -1}, {1, 0}, {1, 1}, {1, 2}}},
		{'I', 2, []Position{{0, 1}, {1, 1}, {2, 1}, {3, 1}}},
		{'I', 3, []Position{{2, -1}, {2, 0}, {2, 1}, {2, 2}}},
		// O (all rotations identical)
		{'O', 0, []Position{{0, 0}, {1, 0}, {0, 1}, {1, 1}}},
		{'O', 1, []Position{{0, 0}, {1, 0}, {0, 1}, {1, 1}}},
		{'O', 2, []Position{{0, 0}, {1, 0}, {0, 1}, {1, 1}}},
		{'O', 3, []Position{{0, 0}, {1, 0}, {0, 1}, {1, 1}}},
		// T
		{'T', 0, []Position{{0, 0}, {1, 0}, {2, 0}, {1, 1}}},
		{'T', 1, []Position{{1, -1}, {1, 0}, {1, 1}, {2, 0}}},
		{'T', 2, []Position{{1, -1}, {0, 0}, {1, 0}, {2, 0}}},
		{'T', 3, []Position{{1, -1}, {0, 0}, {1, 0}, {1, 1}}},
		// S
		{'S', 0, []Position{{1, -1}, {2, -1}, {0, 0}, {1, 0}}},
		{'S', 1, []Position{{1, -1}, {1, 0}, {2, 0}, {2, 1}}},
		{'S', 2, []Position{{1, -1}, {2, -1}, {0, 0}, {1, 0}}},
		{'S', 3, []Position{{0, -1}, {0, 0}, {1, 0}, {1, 1}}},
		// Z
		{'Z', 0, []Position{{0, -1}, {1, -1}, {1, 0}, {2, 0}}},
		{'Z', 1, []Position{{2, -1}, {1, 0}, {2, 0}, {1, 1}}},
		{'Z', 2, []Position{{0, 0}, {1, 0}, {1, 1}, {2, 1}}},
		{'Z', 3, []Position{{1, -1}, {0, 0}, {1, 0}, {0, 1}}},
		// J
		{'J', 0, []Position{{0, -1}, {0, 0}, {1, 0}, {2, 0}}},
		{'J', 1, []Position{{1, -1}, {2, -1}, {1, 0}, {1, 1}}},
		{'J', 2, []Position{{0, 0}, {1, 0}, {2, 0}, {2, 1}}},
		{'J', 3, []Position{{1, -1}, {1, 0}, {0, 1}, {1, 1}}},
		// L
		{'L', 0, []Position{{2, -1}, {0, 0}, {1, 0}, {2, 0}}},
		{'L', 1, []Position{{1, -1}, {1, 0}, {1, 1}, {2, 1}}},
		{'L', 2, []Position{{0, 0}, {1, 0}, {2, 0}, {0, 1}}},
		{'L', 3, []Position{{1, -1}, {2, -1}, {1, 0}, {1, 1}}},
	}

	for _, tc := range cases {
		p := &Piece{Type: tc.pieceType, Rotation: tc.rotation}
		got := getCells(p)
		if len(got) != len(tc.want) {
			t.Errorf("piece %c rot %d: got %d cells, want %d", tc.pieceType, tc.rotation, len(got), len(tc.want))
			continue
		}
		for i, g := range got {
			if g != tc.want[i] {
				t.Errorf("piece %c rot %d cell[%d]: got %v, want %v", tc.pieceType, tc.rotation, i, g, tc.want[i])
			}
		}
	}
}

func TestGetCells_UnknownPiece(t *testing.T) {
	p := &Piece{Type: 'X', Rotation: 0}
	if getCells(p) != nil {
		t.Error("unknown piece type should return nil")
	}
}

// ---- getDropInterval --------------------------------------------------------

func TestGetDropInterval(t *testing.T) {
	// Levels 0-17 strictly decrease
	prev := time.Duration(1<<62 - 1)
	for level := 0; level <= 17; level++ {
		d := getDropInterval(level)
		if d >= prev {
			t.Errorf("level %d: interval %v not less than level %d interval %v", level, d, level-1, prev)
		}
		prev = d
	}
	// Levels 18 and 19 both plateau at 50ms
	if getDropInterval(18) != 50*time.Millisecond {
		t.Errorf("level 18: want 50ms, got %v", getDropInterval(18))
	}
	if getDropInterval(19) != 50*time.Millisecond {
		t.Errorf("level 19: want 50ms, got %v", getDropInterval(19))
	}
	// Beyond array: capped at 50ms
	if getDropInterval(25) != 50*time.Millisecond {
		t.Errorf("level 25: want 50ms, got %v", getDropInterval(25))
	}
}

// ---- Board.Cell / SetCell ---------------------------------------------------

func TestBoardCell_Bounds(t *testing.T) {
	b := NewBoard()
	// Out-of-bounds reads return 0
	for _, pos := range [][2]int{{-1, 0}, {0, -1}, {BoardWidth, 0}, {0, BoardHeight}, {-99, -99}} {
		if v := b.Cell(pos[0], pos[1]); v != 0 {
			t.Errorf("Cell(%d,%d) = %d, want 0", pos[0], pos[1], v)
		}
	}
	// In-bounds read after SetCell
	b.SetCell(5, 10, 'I')
	if v := b.Cell(5, 10); v != 'I' {
		t.Errorf("Cell(5,10) = %c, want I", v)
	}
}

func TestBoardSetCell_Bounds(t *testing.T) {
	b := NewBoard()
	// Out-of-bounds writes are silently ignored
	b.SetCell(-1, 0, 'T')
	b.SetCell(BoardWidth, 0, 'T')
	b.SetCell(0, -1, 'T')
	b.SetCell(0, BoardHeight, 'T')
	// Board should still be empty
	for y := 0; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			if b.Cell(x, y) != 0 {
				t.Errorf("board dirty after out-of-bounds SetCell at (%d,%d)", x, y)
			}
		}
	}
}

// ---- Board.Collides ---------------------------------------------------------

func TestBoardCollides_LeftWall(t *testing.T) {
	b := NewBoard()
	p := &Piece{Type: 'O', X: -1, Y: 5, Rotation: 0}
	if !b.Collides(p) {
		t.Error("O piece at x=-1 should collide with left wall")
	}
}

func TestBoardCollides_RightWall(t *testing.T) {
	b := NewBoard()
	// O piece occupies x and x+1; at x=9 the right cell is x=10 which is out of bounds
	p := &Piece{Type: 'O', X: 9, Y: 5, Rotation: 0}
	if !b.Collides(p) {
		t.Error("O piece at x=9 should collide with right wall")
	}
}

func TestBoardCollides_Floor(t *testing.T) {
	b := NewBoard()
	// O piece at y=BoardHeight-1: lower cells land at y=BoardHeight which is out of bounds
	p := &Piece{Type: 'O', X: 4, Y: BoardHeight - 1, Rotation: 0}
	if !b.Collides(p) {
		t.Error("O piece at bottom row should collide with floor")
	}
}

func TestBoardCollides_OccupiedCell(t *testing.T) {
	b := NewBoard()
	b.SetCell(4, 5, 'I')
	p := &Piece{Type: 'O', X: 4, Y: 5, Rotation: 0} // occupies (4,5),(5,5),(4,6),(5,6)
	if !b.Collides(p) {
		t.Error("piece should collide with occupied cell")
	}
}

func TestBoardCollides_AboveBoardOK(t *testing.T) {
	b := NewBoard()
	// I piece with negative Y cells — spawn zone, should not collide
	p := &Piece{Type: 'I', X: 3, Y: -1, Rotation: 0} // cells at y=-1
	if b.Collides(p) {
		t.Error("piece above board top (negative Y) should not collide")
	}
}

func TestBoardCollides_FreeCells(t *testing.T) {
	b := NewBoard()
	p := &Piece{Type: 'O', X: 4, Y: 5, Rotation: 0}
	if b.Collides(p) {
		t.Error("piece on empty board should not collide")
	}
}

// ---- Board.Lock -------------------------------------------------------------

func TestBoardLock(t *testing.T) {
	b := NewBoard()
	p := &Piece{Type: 'O', Color: 'O', X: 4, Y: 5, Rotation: 0}
	b.Lock(p)
	// O piece occupies (4,5),(5,5),(4,6),(5,6)
	for _, expected := range [][2]int{{4, 5}, {5, 5}, {4, 6}, {5, 6}} {
		if b.Cell(expected[0], expected[1]) != 'O' {
			t.Errorf("expected 'O' at (%d,%d) after Lock", expected[0], expected[1])
		}
	}
}

func TestBoardLock_NegativeYIgnored(t *testing.T) {
	b := NewBoard()
	// S piece at y=0 has some cells at y=-1; those should not be written
	p := &Piece{Type: 'S', Color: 'S', X: 3, Y: 0, Rotation: 0}
	b.Lock(p)
	// Row -1 cells should not appear in board (they'd be out of grid)
	// Just verify board is not corrupted — any in-bounds cells should be 'S'
	found := false
	for y := 0; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			if b.Cell(x, y) == 'S' {
				found = true
			}
		}
	}
	if !found {
		t.Error("Lock with y=0 S piece should write at least some cells in-bounds")
	}
}

// ---- Board.ClearLines -------------------------------------------------------

func fillRow(b *Board, y int) {
	for x := 0; x < BoardWidth; x++ {
		b.SetCell(x, y, 'X')
	}
}

func TestBoardClearLines_None(t *testing.T) {
	b := NewBoard()
	b.SetCell(0, 19, 'I') // partial row
	if n, _ := b.ClearLines(); n != 0 {
		t.Errorf("ClearLines with no full row: got %d, want 0", n)
	}
	if b.Cell(0, 19) != 'I' {
		t.Error("partial row should not be cleared")
	}
}

func TestBoardClearLines_Single(t *testing.T) {
	b := NewBoard()
	b.SetCell(0, 18, 'T') // marker above the full row
	fillRow(b, 19)
	if n, _ := b.ClearLines(); n != 1 {
		t.Errorf("ClearLines single: got %d, want 1", n)
	}
	// Row 18 should have shifted down to row 19
	if b.Cell(0, 19) != 'T' {
		t.Error("cell above cleared row should shift down")
	}
	// Row 18 should now be empty
	if b.Cell(0, 18) != 0 {
		t.Error("row 18 should be empty after gravity")
	}
}

func TestBoardClearLines_Double(t *testing.T) {
	b := NewBoard()
	fillRow(b, 18)
	fillRow(b, 19)
	if n, _ := b.ClearLines(); n != 2 {
		t.Errorf("ClearLines double: got %d, want 2", n)
	}
}

func TestBoardClearLines_Tetris(t *testing.T) {
	b := NewBoard()
	for _, row := range []int{16, 17, 18, 19} {
		fillRow(b, row)
	}
	if n, _ := b.ClearLines(); n != 4 {
		t.Errorf("ClearLines tetris: got %d, want 4", n)
	}
}

func TestBoardClearLines_NonContiguous(t *testing.T) {
	b := NewBoard()
	fillRow(b, 0)
	fillRow(b, 19)
	if n, _ := b.ClearLines(); n != 2 {
		t.Errorf("ClearLines non-contiguous: got %d, want 2", n)
	}
}
