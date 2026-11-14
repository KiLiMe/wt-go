package ui

import (
	"testing"
)

// TestNameWidth covers the room a tab's name takes: a short name keeps a
// tab clickable, a long one stops growing.
func TestNameWidth(t *testing.T) {
	tests := []struct{ measured, want float32 }{
		{0, minTabName},
		{12, minTabName},
		{minTabName, minTabName},
		{120, 120},
		{maxTabName, maxTabName},
		{900, maxTabName},
	}
	for _, tt := range tests {
		if got := nameWidth(tt.measured); got != tt.want {
			t.Errorf("nameWidth(%v) = %v, want %v", tt.measured, got, tt.want)
		}
	}
}

// TestTabNameFillsItsRoom covers the room a tab gives its name: the whole of
// it, so that the button that closes the tab sits on the tab's edge rather
// than a gap short of it, where a short name leaves nothing drawn.
func TestTabNameFillsItsRoom(t *testing.T) {
	tt, _, _ := stripOfNamed(t, 1, 0, 900, "pwsh")
	name, ok := tt.Find("pwsh")
	if !ok {
		t.Fatalf("the tab draws no name: %q", tt.Texts())
	}
	if name.W != minTabName {
		t.Errorf("the name takes %v DIPs of the %v a tab gives it: the close button stands %v short of the edge",
			name.W, minTabName, minTabName-name.W)
	}
}
