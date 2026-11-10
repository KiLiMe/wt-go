package pane

import (
	"testing"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/layout"
)

// term is a terminal the tree holds without running anything: these tests
// only care where the panes are.
func term() *terminal.Terminal { return &terminal.Terminal{} }

// newTab is a tab of one pane.
func newTab() *Tab { return New(term()) }

// TestNew covers a tab of one pane: it holds the pane, and the pane has the
// keyboard.
func TestNew(t *testing.T) {
	tb := newTab()
	if tb.Count() != 1 || tb.Root != tb.Focus || !tb.Focus.Pane() {
		t.Fatalf("count = %d, root = %p, focus = %p", tb.Count(), tb.Root, tb.Focus)
	}
}

// TestSplit covers dividing a pane: the pane keeps its room as the first of
// the split, the new pane is the second, and the new pane has the keyboard.
func TestSplit(t *testing.T) {
	tb := newTab()
	held := tb.Root.Term
	if !tb.Split(layout.SideBySide, term()) {
		t.Fatal("split did nothing")
	}
	if tb.Count() != 2 {
		t.Fatalf("count = %d, want 2", tb.Count())
	}
	if tb.Root.Pane() {
		t.Error("the split is a pane")
	}
	if tb.Root.Dir != layout.SideBySide || tb.Root.Share != 0.5 {
		t.Errorf("dir = %v, share = %v", tb.Root.Dir, tb.Root.Share)
	}
	if tb.Root.A.Term != held {
		t.Error("the first pane does not hold what the pane held")
	}
	if tb.Focus != tb.Root.B {
		t.Error("the new pane has no keyboard")
	}
	if got := len(tb.Terminals()); got != 2 {
		t.Errorf("terminals = %d, want 2", got)
	}
}

// TestSplitNested covers dividing a pane of a split: the split holds the new
// split, and the other pane keeps its room.
func TestSplitNested(t *testing.T) {
	tb := newTab()
	tb.Split(layout.SideBySide, term())
	right := tb.Root.B
	tb.Split(layout.Stacked, term())
	if tb.Root.B != right {
		t.Fatal("the split took the place of the pane")
	}
	if right.Pane() || right.Dir != layout.Stacked {
		t.Errorf("the pane did not become a split: %+v", right)
	}
	if got := tb.Count(); got != 3 {
		t.Errorf("count = %d, want 3", got)
	}
}

// TestClose covers dropping a pane: the pane it shared a split with takes the
// room of the split, and the keyboard.
func TestClose(t *testing.T) {
	tb := newTab()
	tb.Split(layout.SideBySide, term())
	left := tb.Root.A
	closed := tb.Focus.Term
	dropped, ok := tb.Close()
	if !ok {
		t.Fatal("tab reported empty")
	}
	if dropped != closed {
		t.Error("close dropped another terminal")
	}
	if tb.Count() != 1 || tb.Root != left || tb.Focus != left {
		t.Errorf("root = %p, focus = %p, want the pane left: %p", tb.Root, tb.Focus, left)
	}
}

// TestCloseOfSplit covers dropping a pane whose sibling is a split: the split
// takes the room, and the keyboard goes to the pane that takes room first.
func TestCloseOfSplit(t *testing.T) {
	tb := newTab()
	tb.Split(layout.SideBySide, term())
	tb.Split(layout.Stacked, term())
	tb.Focus = tb.Root.A
	if _, ok := tb.Close(); !ok {
		t.Fatal("tab reported empty")
	}
	if tb.Count() != 2 || tb.Root.Dir != layout.Stacked {
		t.Fatalf("count = %d, dir = %v", tb.Count(), tb.Root.Dir)
	}
	if tb.Focus != tb.Root.A || !tb.Focus.Pane() {
		t.Error("the keyboard is not on the first pane")
	}
}

// TestCloseAt covers dropping a pane the keyboard is not on: the keyboard
// stays where it is.
func TestCloseAt(t *testing.T) {
	tb := newTab()
	tb.Split(layout.SideBySide, term())
	tb.Split(layout.Stacked, term()) // A | (B1 / B2), the keyboard on B2
	top := tb.Root.B.A
	dropped, ok := tb.CloseAt(top)
	if !ok || dropped == nil {
		t.Fatal("close dropped nothing")
	}
	if tb.Count() != 2 || !tb.Root.B.Pane() {
		t.Fatalf("count = %d, the split left = %v", tb.Count(), !tb.Root.B.Pane())
	}
	if tb.Focus == top {
		t.Error("the keyboard is on the pane that is gone")
	}
	if !tb.Focus.Pane() {
		t.Error("the keyboard is not on a pane")
	}
}

// TestCloseLast covers dropping the only pane: the tab holds none, and says
// so.
func TestCloseLast(t *testing.T) {
	tb := newTab()
	if _, ok := tb.Close(); ok {
		t.Error("tab kept a pane")
	}
	if tb.Root != nil || tb.Focus != nil || tb.Count() != 0 {
		t.Errorf("root = %p, focus = %p, count = %d", tb.Root, tb.Focus, tb.Count())
	}
}

// TestPanes covers the order of the panes: the order they take room.
func TestPanes(t *testing.T) {
	tb := newTab()
	tb.Split(layout.SideBySide, term()) // A | (B)
	tb.Focus = tb.Root.A
	tb.Split(layout.Stacked, term()) // (A1 / A2) | B
	panes := tb.Panes()
	if len(panes) != 3 {
		t.Fatalf("panes = %d, want 3", len(panes))
	}
	if panes[0] != tb.Root.A.A || panes[1] != tb.Root.A.B || panes[2] != tb.Root.B {
		t.Error("panes are out of the order they take room")
	}
}
