package ui

import (
	"testing"

	myui "github.com/egoist/mygo/ui"
)

// TestInkOn covers the color a mark takes on the face of a profile: the pale
// shells get the dark ink, the saturated ones keep the white.
func TestInkOn(t *testing.T) {
	for _, tc := range []struct{ color, want string }{
		{"#3b78ff", "#ffffff"}, // PowerShell 7
		{"#2671be", "#ffffff"}, // Windows PowerShell
		{"#e95420", "#ffffff"}, // the distributions of wsl
		{"#7c6cf5", "#ffffff"}, // Nushell
		{"#cccccc", "#101010"}, // Command Prompt
		{"#e5c07b", "#101010"}, // Zsh
		{"#b0b0b0", "#101010"}, // sh
		{"#4eaa25", "#101010"}, // Bash
	} {
		got, want := inkOn(myui.Hex(tc.color)), myui.Hex(tc.want)
		if got.R != want.R || got.G != want.G || got.B != want.B {
			t.Errorf("inkOn(%s) = #%02x%02x%02x, want %s", tc.color, got.R, got.G, got.B, tc.want)
		}
	}
}

// TestColorOf covers the colors settings.json writes, and the ones it does
// not: a color it cannot read is no color, never a panic.
func TestColorOf(t *testing.T) {
	for _, s := range []string{"#3b78ff", "#fff", "#3b78ffcc", "#ffff"} {
		if c := colorOf(s); c.A == 0 {
			t.Errorf("colorOf(%q) = %v", s, c)
		}
	}
	for _, s := range []string{"", "3b78ff", "#3b78f", "#3b78fff", "#gggggg", "blue"} {
		if c := colorOf(s); c.A != 0 || c.R != 0 || c.G != 0 || c.B != 0 {
			t.Errorf("colorOf(%q) = %v", s, c)
		}
	}
}
