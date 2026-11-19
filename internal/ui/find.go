package ui

import (
	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/icons"
	"wtgo/internal/pane"
	"wtgo/locales"
)

// The find bar's size.
const (
	findWidth  float32 = 320
	findHeight float32 = 38
	findMargin float32 = 8
	// findNarrow is the width below which the bar stacks its buttons under its
	// field: the field and the three buttons need this much beside each other,
	// and a pane narrower than this gets a bar that says in a row of its own
	// which pane it searches instead of spending the room on the words.
	findNarrow float32 = 300
	// findScopeHeight is the height of that row of its own.
	findScopeHeight float32 = 18
)

// Find is the bar that asks the terminal in front for text, as Windows
// Terminal's does.
type Find struct {
	// Text is what the field holds.
	Text string
	// Status is what the terminal says of the matches, which the window
	// reads for each frame.
	Status terminal.SearchStatus
	// Pane is the terminal's pane, which the bar draws itself over: the search
	// is the terminal's, and a window of split panes says which one it asks by
	// putting the bar on that pane rather than on the window. nil draws the bar
	// over the window, as a search that named no pane does.
	Pane *pane.Node
}

// findBar draws the bar over the terminal it searches, and searches as the
// field is typed in.
func findBar(c *myui.Context, st State, act Actions) {
	f := st.Find
	if f == nil {
		return
	}
	t := c.Theme()
	room := Content(c)
	// The bar belongs to the pane it asks: it is drawn where the pane is, and
	// shrinks with it, so that the matches and the words that count them are
	// over the terminal whose matches they are.
	if r, ok := paneRoom(c, st, f.Pane); ok {
		room = r
	}
	wide := room.W >= findNarrow
	width := min(findWidth, max(room.W-2*findMargin, 0))
	myui.Overlay(c, func() {
		el := myui.Column(c).
			Absolute().
			Top(room.Y + findMargin).
			Width(width).
			Radius(8).
			Background(t.Surface).
			BorderWidth(1).
			BorderColor(t.Border)
		if wide {
			el.Left(max(room.X+room.W-width-findMargin, room.X)).Height(findHeight)
		} else {
			// The narrow bar keeps the margin on both sides of the pane, and
			// grows a row for what the wide one says beside the field.
			el.Left(room.X + findMargin).
				Width(max(room.W-2*findMargin, 0)).
				Height(findHeight + findScopeHeight)
		}
		el.Children(func() {
			myui.Row(c).Height(findHeight).Padding(0, 8).Gap(6).AlignItems(myui.Center).Children(func() {
				in := myui.TextInputBase(c.Key("find"), &f.Text).
					Grow(1).
					Placeholder(locales.Find())
				// The bar holds the keyboard while it is open: a click on the
				// pane or on a button would otherwise take the focus and leave
				// the field deaf to the next key.
				in.Focus()
				if in.Changed() {
					act.Find(f.Text)
				}
				if wide {
					myui.Text(c, status(f.Status)).FontSize(11).TextColor(t.TextMuted).SingleLine()
				}
				findButton(c, icons.Up, locales.FindPreviousHint(), act.FindPrevious)
				findButton(c, icons.Chevron, locales.FindNextHint(), act.FindNext)
				findButton(c, icons.Close, locales.Close(), act.EndFind)
				// Enter picks the next match, Shift+Enter the one before it. The
				// field takes Enter as its submit, and the bar cannot see the key
				// itself, so the modifiers held say which one it was.
				switch {
				case in.Submitted() && c.Modifiers()&myui.Shift != 0:
					act.FindPrevious()
				case in.Submitted():
					act.FindNext()
				}
				if el.OverlayShortcut(0, myui.KeyEscape) {
					act.EndFind()
				}
			})
			if wide {
				return
			}
			// The pane is named here rather than left to the bar's place: a bar
			// that fills the pane it searches has no place of its own to say
			// which pane that is.
			myui.Row(c).Height(findScopeHeight).Padding(0, 8).Gap(8).AlignItems(myui.Center).Children(func() {
				myui.Text(c, locales.FindInPane(paneName(f.Pane))).FontSize(11).
					TextColor(t.TextMuted).SingleLine().Grow(1).Ellipsis("…")
				myui.Text(c, status(f.Status)).FontSize(11).TextColor(t.TextMuted).SingleLine()
			})
		})
	})
}

// paneName is what the pane calls itself, which the narrow bar says it searches.
func paneName(n *pane.Node) string {
	if n == nil || n.Term == nil {
		return locales.Pane()
	}
	if title := n.Term.Title(); title != "" {
		return title
	}
	return locales.Pane()
}

// findButton is one of the bar's buttons: an icon, and what a click does. The
// buttons are the size of the strip's, so that the two rows of the window agree
// on what a button is, and the keys the two of them answer to are in their
// tooltips: a click is not the only way to go to the next match.
func findButton(c *myui.Context, ic *icons.Icon, tooltip string, onClick func()) {
	myui.Box(c).
		Size(28, 28).
		Radius(4).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Tooltip(tooltip).
		OnClick(onClick).
		Children(func() {
			myui.Icon(c, ic.SVG).Size(14, 14).TextColor(c.Theme().TextMuted)
		})
}

// status is what the bar says of the matches: where the selected one is among
// them, that there is none yet, or nothing while the field is empty.
func status(s terminal.SearchStatus) string {
	switch {
	case s.Text == "":
		return ""
	case s.Total == 0:
		return locales.FindNone()
	case s.Selected == 0:
		return locales.FindCount(s.Total)
	}
	return locales.FindPosition(s.Selected, s.Total)
}
