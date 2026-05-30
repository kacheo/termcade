package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestGetPieceColor(t *testing.T) {
	cases := []struct {
		piece byte
		want  lipgloss.Color
	}{
		{'I', ColorCyan},
		{'O', ColorYellow},
		{'T', ColorPurple},
		{'S', ColorGreen},
		{'Z', ColorRed},
		{'J', ColorBlue},
		{'L', ColorOrange},
		{'X', ColorGray},  // unknown → gray
		{0, ColorGray},    // zero byte → gray
		{'?', ColorGray},  // another unknown
	}

	for _, tc := range cases {
		got := GetPieceColor(tc.piece)
		if got != tc.want {
			t.Errorf("GetPieceColor(%c): got %v, want %v", tc.piece, got, tc.want)
		}
	}
}
