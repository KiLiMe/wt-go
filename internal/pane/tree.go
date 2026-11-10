package pane

import (
	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/layout"
)

// Node is a pane, or the split that holds two of them.
type Node struct {
	// Term is the terminal of a pane; a split has none.
	Term *terminal.Terminal
	// Dir and Share are how a split divides its room, the first pane of it
	// taking Share of the room. A pane has no use for them.
	Dir   layout.Dir
	Share float32
	// A and B are the panes of a split.
	A, B *Node
}

// Pane reports whether n is a pane rather than a split.
func (n *Node) Pane() bool {
	return n != nil && n.Term != nil
}

// Tab is one tab: the panes it holds, the one with the keyboard, and the
// profile it runs. It owns nothing: the caller closes the terminals it hands
// back.
type Tab struct {
	// Root is the pane the tab holds, or the split of all of them.
	Root *Node
	// Focus is the pane with the keyboard.
	Focus *Node
	// ProfileID, ProfileName and Color are the profile the tab runs and the
	// color of its dot.
	ProfileID   string
	ProfileName string
	Color       string
	// Name is the name the user gave the tab.
	Name string
}

// New is a tab that holds term.
func New(term *terminal.Terminal) *Tab {
	root := &Node{Term: term}
	return &Tab{Root: root, Focus: root}
}

// Panes are the panes of the tab, in the order they take room.
func (t *Tab) Panes() []*Node {
	out := make([]*Node, 0, 4)
	t.Root.each(func(n *Node) { out = append(out, n) })
	return out
}

// Count is how many panes the tab holds.
func (t *Tab) Count() int {
	n := 0
	t.Root.each(func(*Node) { n++ })
	return n
}

// Terminals are the terminals of the tab's panes, in the order they take
// room.
func (t *Tab) Terminals() []*terminal.Terminal {
	out := make([]*terminal.Terminal, 0, 4)
	t.Root.each(func(n *Node) { out = append(out, n.Term) })
	return out
}

// Running reports whether a pane still runs its program. A terminal that has
// ended answers at once on Done, so a tab of ended shells does not ask before
// it closes.
func (t *Tab) Running() bool {
	for _, term := range t.Terminals() {
		select {
		case <-term.Done():
		default:
			return true
		}
	}
	return false
}

// each calls fn for each pane of the tree, in the order they take room.
func (n *Node) each(fn func(*Node)) {
	if n == nil {
		return
	}
	if n.Pane() {
		fn(n)
		return
	}
	n.A.each(fn)
	n.B.each(fn)
}

// Split divides the pane with the keyboard in two, the pane it held first and
// a pane that runs term second, and gives the keyboard to the new pane. How
// many panes a tab may hold is the caller's to decide.
func (t *Tab) Split(dir layout.Dir, term *terminal.Terminal) bool {
	leaf := t.Focus
	if !leaf.Pane() {
		return false
	}
	// The pane keeps its room: it becomes the split, holding what it held
	// as its first pane.
	held := leaf.Term
	leaf.Term, leaf.Dir, leaf.Share = nil, dir, 0.5
	leaf.A, leaf.B = &Node{Term: held}, &Node{Term: term}
	t.Focus = leaf.B
	return true
}

// Close drops the pane with the keyboard: the pane it shared a split with
// takes the place of that split, and the keyboard when it is a pane. Close
// hands back the terminal it dropped, and reports whether the tab has panes
// left.
func (t *Tab) Close() (*terminal.Terminal, bool) {
	return t.CloseAt(t.Focus)
}

// CloseAt drops pane n as Close drops the pane with the keyboard, and leaves
// the keyboard where it is when n is not the pane that has it.
func (t *Tab) CloseAt(n *Node) (*terminal.Terminal, bool) {
	if !n.Pane() {
		return nil, t.Root.Pane()
	}
	term := n.Term
	if n == t.Root {
		t.Root, t.Focus = nil, nil
		return term, false
	}
	split, ok := t.Root.parent(n)
	if !ok {
		return term, false
	}
	sibling := split.A
	if sibling == n {
		sibling = split.B
	}
	if split == t.Root {
		t.Root = sibling
	} else if grand, ok := t.Root.parent(split); ok {
		if grand.A == split {
			grand.A = sibling
		} else {
			grand.B = sibling
		}
	}
	if t.Focus == n {
		t.Focus = first(sibling)
	}
	return term, true
}

// parent is the split that holds n, if any.
func (n *Node) parent(target *Node) (*Node, bool) {
	if n == nil || n.Pane() {
		return nil, false
	}
	if n.A == target || n.B == target {
		return n, true
	}
	if p, ok := n.A.parent(target); ok {
		return p, true
	}
	return n.B.parent(target)
}

// first is the pane that takes room first in the tree of n.
// First is the first pane of the tab: the one a split leaves where it was,
// and the one whose title names the tab when the window does not follow the
// pane with the keyboard.
func (t *Tab) First() *Node {
	return first(t.Root)
}

func first(n *Node) *Node {
	for n != nil && !n.Pane() {
		n = n.A
	}
	return n
}
