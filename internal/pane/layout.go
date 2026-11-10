package pane

import "wtgo/internal/layout"

// Place is where a pane goes.
type Place struct {
	Node *Node
	Rect layout.Rect
}

// Divide is a split as the window draws it: the line between its panes, and
// the room the split shares, which the line moves in.
type Divide struct {
	Node *Node
	Dir  layout.Dir
	Line layout.Rect
	Room layout.Rect
}

// Geometry is where the panes of a tab go, and where the lines between them
// are.
type Geometry struct {
	Panes   []Place
	Divides []Divide
}

// Direction is a way the keyboard moves between the panes of a tab.
type Direction int

const (
	// Left and Right move between panes beside each other.
	Left Direction = iota
	Right
	// Up and Down move between panes above each other.
	Up
	Down
)

// Layout is where the panes of the tab go in r, and where the lines between
// them are.
func (t *Tab) Layout(r layout.Rect) Geometry {
	g := Geometry{}
	t.Root.place(r, &g)
	return g
}

// place puts n in r, and the panes it holds.
func (n *Node) place(r layout.Rect, g *Geometry) {
	if n == nil {
		return
	}
	if n.Pane() {
		g.Panes = append(g.Panes, Place{Node: n, Rect: r})
		return
	}
	d := layout.Split(r, n.Dir, n.Share)
	g.Divides = append(g.Divides, Divide{Node: n, Dir: n.Dir, Line: d.Line, Room: r})
	n.A.place(d.First, g)
	n.B.place(d.Second, g)
}

// Resize moves the line of a split by delta DIPs along the way it divides its
// room, within what the room allows.
func (t *Tab) Resize(d Divide, delta float32) {
	if d.Node == nil || d.Node.Pane() {
		return
	}
	d.Node.Share = layout.Share(d.Room, d.Dir, d.Node.Share, delta)
}

// Auto is the way a split of the pane with the keyboard divides it: the way
// that pane has more room, side by side when it is as wide as it is tall.
func (t *Tab) Auto(r layout.Rect) layout.Dir {
	if p, ok := t.room(r); ok && p.H > p.W {
		return layout.Stacked
	}
	return layout.SideBySide
}

// Move gives the keyboard to the pane next to the one with it in dir, if
// there is one: of the panes that way, the one sharing the most room with it
// across the line between them, and then the nearest.
func (t *Tab) Move(dir Direction, r layout.Rect) bool {
	from, ok := t.room(r)
	if !ok {
		return false
	}
	var best *Node
	var bestOverlap, bestGap float32
	for _, p := range t.Layout(r).Panes {
		if p.Node == t.Focus {
			continue
		}
		gap, overlap := reach(from, p.Rect, dir)
		if gap < 0 || overlap <= 0 {
			continue
		}
		if best == nil || overlap > bestOverlap || (overlap == bestOverlap && gap < bestGap) {
			best, bestOverlap, bestGap = p.Node, overlap, gap
		}
	}
	if best == nil {
		return false
	}
	t.Focus = best
	return true
}

// room is where the pane with the keyboard goes in r.
func (t *Tab) room(r layout.Rect) (layout.Rect, bool) {
	if !t.Focus.Pane() {
		return layout.Rect{}, false
	}
	for _, p := range t.Layout(r).Panes {
		if p.Node == t.Focus {
			return p.Rect, true
		}
	}
	return layout.Rect{}, false
}

// reach is how far to is from from in dir, and how much of the side across
// dir the two share. to is that way only when it is clear of from: a gap of
// no less than nothing.
func reach(from, to layout.Rect, dir Direction) (gap, overlap float32) {
	switch dir {
	case Left:
		return from.X - (to.X + to.W), span(from.Y, from.H, to.Y, to.H)
	case Right:
		return to.X - (from.X + from.W), span(from.Y, from.H, to.Y, to.H)
	case Up:
		return from.Y - (to.Y + to.H), span(from.X, from.W, to.X, to.W)
	default:
		return to.Y - (from.Y + from.H), span(from.X, from.W, to.X, to.W)
	}
}

// span is how much of [a, a+n) and [b, b+m) share.
func span(a, n, b, m float32) float32 {
	return min(a+n, b+m) - max(a, b)
}
