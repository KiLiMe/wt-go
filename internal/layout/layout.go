// Package layout computes the room the panes of a tab take: the rectangles
// the two panes of a split take, the line between them, and how dragging that
// line moves it. It knows nothing of terminals, tabs, or the window.
package layout

// Rect is a rectangle in the window, in DIPs.
type Rect struct {
	X, Y, W, H float32
}

// Dir is how a split shares its room between its two panes.
type Dir int

const (
	// SideBySide puts the panes one beside the other, sharing the width.
	SideBySide Dir = iota
	// Stacked puts one above the other, sharing the height.
	Stacked
)

const (
	// Divider is the room the line between two panes takes.
	Divider float32 = 4
	// Min is the least room a pane takes along the way its split shares,
	// which keeps a pane big enough to read.
	Min float32 = 40
)

// Divide is where a split puts its two panes, and the line between them.
type Divide struct {
	First, Line, Second Rect
}

// Split divides r between the two panes of a split: the first takes share of
// the room, the second the rest, and Divider of the room goes to the line
// between them. A share that would leave a pane narrower than Min is brought
// back to what leaves it Min; when the room cannot hold two panes of Min, the
// panes take half of it each.
func Split(r Rect, dir Dir, share float32) Divide {
	length := r.W
	if dir == Stacked {
		length = r.H
	}
	room := max(length-Divider, 0)
	first := room / 2
	if room >= 2*Min {
		first = min(max(share*room, Min), room-Min)
	}
	rest := room - first
	if dir == Stacked {
		return Divide{
			First:  Rect{X: r.X, Y: r.Y, W: r.W, H: first},
			Line:   Rect{X: r.X, Y: r.Y + first, W: r.W, H: Divider},
			Second: Rect{X: r.X, Y: r.Y + first + Divider, W: r.W, H: rest},
		}
	}
	return Divide{
		First:  Rect{X: r.X, Y: r.Y, W: first, H: r.H},
		Line:   Rect{X: r.X + first, Y: r.Y, W: Divider, H: r.H},
		Second: Rect{X: r.X + first + Divider, Y: r.Y, W: rest, H: r.H},
	}
}

// Room is the room a split of r shares between its panes along dir: what the
// two panes take, without the line.
func Room(r Rect, dir Dir) float32 {
	if dir == Stacked {
		return max(r.H-Divider, 0)
	}
	return max(r.W-Divider, 0)
}

// Share is share moved by delta DIPs along dir: the share that puts the line
// delta further along, within what Split allows. When the room cannot hold
// two panes of Min, the share stays what it was, as it says nothing there.
func Share(r Rect, dir Dir, share, delta float32) float32 {
	room := Room(r, dir)
	if room < 2*Min {
		return share
	}
	first := share*room + delta
	return min(max(first, Min), room-Min) / room
}
