package app

import (
	"testing"

	"wtgo/internal/ui"
)

// TestFindNeedsATerminal covers the bar the window opens: it opens on a tab
// that holds a terminal, and not on one that does not, as the settings.
func TestFindNeedsATerminal(t *testing.T) {
	a := &App{}
	a.OpenFind()
	if a.find != nil {
		t.Error("the find bar opened on a window with no terminal")
	}
	if a.findState() != nil {
		t.Error("findState showed a bar the window does not have")
	}

	// A bar left open as the tab in front stops being a terminal goes with
	// it.
	a.find = &ui.Find{Text: "x"}
	if got := a.findState(); got != nil {
		t.Errorf("findState = %+v, want none", got)
	}
	if a.find != nil {
		t.Error("the find bar stayed open")
	}
	a.EndFind()
}
