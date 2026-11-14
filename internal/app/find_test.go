package app

import (
	"testing"

	"wtgo/internal/ui"
)

// TestFindNeedsATerminal covers the bar the window opens: it opens on a tab
// that holds a terminal, and not on one that does not, as the settings. It
// also covers an overlay opened before the bar: the bar takes the keyboard, so
// what held it goes.
func TestFindNeedsATerminal(t *testing.T) {
	a := &App{}
	a.OpenFind()
	if a.find != nil {
		t.Error("the find bar opened on a window with no terminal")
	}
	if a.findState() != nil {
		t.Error("findState showed a bar the window does not have")
	}

	// Whatever holds the keys goes when the bar opens: an overlay opened
	// before the bar closes instead of the two showing at once.
	a.confirm = &ask{Confirm: ui.Confirm{Open: true}}
	a.openOverlay(overlayFind)
	if a.top() != overlayFind || a.find == nil {
		t.Fatalf("the bar did not open over the question: top %v, bar %v", a.top(), a.find)
	}
	if a.confirm != nil {
		t.Error("the question stayed open under the bar")
	}
	a.EndFind()
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
