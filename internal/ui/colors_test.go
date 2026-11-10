package ui

import "testing"

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
