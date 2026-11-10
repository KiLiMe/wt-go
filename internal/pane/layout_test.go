package pane

import (
	"testing"

	"wtgo/internal/layout"
)

// sideBySide is a tab of two panes beside each other in a room of 204 by
// 100: the first takes the left 100 DIPs, the second the right ones.
func sideBySide() (*Tab, layout.Rect) {
	tb := newTab()
	tb.Split(layout.SideBySide, term())
	return tb, layout.Rect{W: 204, H: 100}
}

// TestLayout covers where the panes of a split go, and where the line
// between them is.
func TestLayout(t *testing.T) {
	tb, r := sideBySide()
	g := tb.Layout(r)
	if len(g.Panes) != 2 || len(g.Divides) != 1 {
		t.Fatalf("panes = %d, divides = %d", len(g.Panes), len(g.Divides))
	}
	if g.Panes[0].Rect != (layout.Rect{W: 100, H: 100}) {
		t.Errorf("first pane = %+v", g.Panes[0].Rect)
	}
	if g.Panes[1].Rect != (layout.Rect{X: 104, W: 100, H: 100}) {
		t.Errorf("second pane = %+v", g.Panes[1].Rect)
	}
	if g.Divides[0].Line != (layout.Rect{X: 100, W: layout.Divider, H: 100}) {
		t.Errorf("line = %+v", g.Divides[0].Line)
	}
	if g.Divides[0].Room != r {
		t.Errorf("room = %+v, want %+v", g.Divides[0].Room, r)
	}
}

// TestLayoutNested covers a split of a pane: the panes inside it take the
// room of the pane, less the line between them.
func TestLayoutNested(t *testing.T) {
	tb, r := sideBySide()
	tb.Focus = tb.Root.B
	tb.Split(layout.Stacked, term())
	g := tb.Layout(r)
	if len(g.Panes) != 3 || len(g.Divides) != 2 {
		t.Fatalf("panes = %d, divides = %d", len(g.Panes), len(g.Divides))
	}
	want := []layout.Rect{
		{X: 0, Y: 0, W: 100, H: 100},
		{X: 104, Y: 0, W: 100, H: 48},
		{X: 104, Y: 52, W: 100, H: 48},
	}
	for i, p := range g.Panes {
		if p.Rect != want[i] {
			t.Errorf("pane %d = %+v, want %+v", i, p.Rect, want[i])
		}
	}
}

// TestLayoutOfNoPanes covers a tab that holds none: nothing to place.
func TestLayoutOfNoPanes(t *testing.T) {
	g := (&Tab{}).Layout(layout.Rect{W: 100, H: 100})
	if len(g.Panes) != 0 || len(g.Divides) != 0 {
		t.Errorf("panes = %d, divides = %d", len(g.Panes), len(g.Divides))
	}
}

// TestResize covers dragging a line: the split takes more room for its first
// pane, and no more than leaves the second pane room to read.
func TestResize(t *testing.T) {
	tb, r := sideBySide()
	d := tb.Layout(r).Divides[0]
	tb.Resize(d, 50)
	if got := tb.Root.Share; got != 0.75 {
		t.Errorf("share = %v, want 0.75", got)
	}
	tb.Resize(d, 500)
	if got := tb.Root.Share; got != 0.8 {
		t.Errorf("share = %v, want 0.8", got)
	}
}

// TestAuto covers the way an automatic split divides a pane: the way it has
// more room.
func TestAuto(t *testing.T) {
	tb := newTab()
	if got := tb.Auto(layout.Rect{W: 204, H: 100}); got != layout.SideBySide {
		t.Errorf("wide pane splits %v", got)
	}
	if got := tb.Auto(layout.Rect{W: 100, H: 204}); got != layout.Stacked {
		t.Errorf("tall pane splits %v", got)
	}
	if got := tb.Auto(layout.Rect{W: 100, H: 100}); got != layout.SideBySide {
		t.Errorf("square pane splits %v", got)
	}
}

// TestMove covers the way the keyboard moves between panes: to the pane
// beside the one with the keyboard, if there is one.
func TestMoveBetweenPanes(t *testing.T) {
	tb, r := sideBySide()
	tb.Focus = tb.Root.A
	if !tb.Move(Right, r) || tb.Focus != tb.Root.B {
		t.Error("right did not reach the pane beside")
	}
	if !tb.Move(Left, r) || tb.Focus != tb.Root.A {
		t.Error("left did not reach the pane beside")
	}
	for _, dir := range []Direction{Left, Up, Down} {
		if tb.Move(dir, r) {
			t.Errorf("%v moved past the edge", dir)
		}
	}
}

// TestMoveAcrossSplits covers the way the keyboard moves between the panes
// of nested splits: to the pane that shares the most room across the line.
func TestMoveAcrossSplits(t *testing.T) {
	tb, r := sideBySide()
	tb.Focus = tb.Root.B
	tb.Split(layout.Stacked, term()) // A | (B1 / B2)
	top, bottom := tb.Root.B.A, tb.Root.B.B
	tb.Focus = top
	if !tb.Move(Left, r) || tb.Focus != tb.Root.A {
		t.Error("left did not reach the pane beside")
	}
	tb.Focus = tb.Root.A
	if !tb.Move(Right, r) {
		t.Fatal("right did not reach the split beside")
	}
	if tb.Focus != top && tb.Focus != bottom {
		t.Errorf("focus = %p, want one of the panes beside", tb.Focus)
	}
	tb.Focus = top
	if !tb.Move(Down, r) || tb.Focus != bottom {
		t.Error("down did not reach the pane below")
	}
	if !tb.Move(Up, r) || tb.Focus != top {
		t.Error("up did not reach the pane above")
	}
}
