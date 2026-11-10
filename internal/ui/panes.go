package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/i18n"
	"wtgo/internal/layout"
	"wtgo/internal/pane"
)

// grip is how wide the pointer may take the line between two panes: wider
// than the line, so that it takes no care.
const grip float32 = 8

// content is the panes of the tab in front, or, with no tabs, what says there
// are none.
func content(c *myui.Context, st State, act Actions, r layout.Rect) {
	if st.Active < 0 || st.Active >= len(st.Tabs) {
		emptyState(c, act)
		return
	}
	tb := st.Tabs[st.Active]
	if tb.Page != nil {
		// A tab that is not a terminal, the settings, takes the room a
		// terminal would.
		myui.Box(c).
			Absolute().
			Left(r.X).
			Top(r.Y).
			Width(r.W).
			Height(r.H).
			Clip().
			Children(func() { tb.Page(c) })
		return
	}
	if tb.Panes == nil {
		emptyState(c, act)
		return
	}
	g := tb.Panes.Layout(r)
	if len(g.Panes) == 0 {
		emptyState(c, act)
		return
	}
	for _, p := range g.Panes {
		paneAt(c, st, act, p)
	}
	for _, d := range g.Divides {
		divider(c, act, st.Active, d)
	}
}

// paneAt is one pane: its terminal, and the keyboard it takes when the
// pointer presses it.
func paneAt(c *myui.Context, st State, act Actions, p pane.Place) {
	tb := st.Tabs[st.Active]
	box := myui.Box(c).
		Absolute().
		Left(p.Rect.X).
		Top(p.Rect.Y).
		Width(p.Rect.W).
		Height(p.Rect.H).
		Clip()
	box.ObserveInput(func(ev myui.InputEvent) {
		if ev.Kind == myui.InputPointerDown && ev.Button == 0 {
			act.FocusPane(st.Active, p.Node)
		}
	})
	box.Children(func() {
		if tb.Mount == nil {
			return
		}
		el := tb.Mount(c, p.Node).Fill()
		if p.Node == tb.Panes.Focus && st.Rename == nil {
			// The pane with the keyboard keeps it, so that the typing goes
			// where the window says it does, unless a dialog takes it.
			el.Focus()
		}
	})
}

// divider is the line between the two panes of a split: it shows the line,
// and takes the pointer that drags it.
func divider(c *myui.Context, act Actions, tab int, d pane.Divide) {
	myui.Box(c).
		Absolute().
		Left(d.Line.X).
		Top(d.Line.Y).
		Width(d.Line.W).
		Height(d.Line.H).
		Background(c.Theme().Border)
	h := gripRect(d)
	handle := myui.Box(c).
		Absolute().
		Left(h.X).
		Top(h.Y).
		Width(h.W).
		Height(h.H)
	delta, stacked := float32(0), d.Dir == layout.Stacked
	if stacked {
		handle = handle.Cursor(myui.CursorResizeNS)
	} else {
		handle = handle.Cursor(myui.CursorResizeEW)
	}
	if dx, dy, held := handle.Dragged(); held {
		delta = dx
		if stacked {
			delta = dy
		}
	}
	if delta != 0 {
		act.ResizePane(tab, d, delta)
	}
}

// gripRect is where the pointer takes a line: wider than the line across it,
// so that the pointer need not aim, and as long as the line.
func gripRect(d pane.Divide) layout.Rect {
	r := d.Line
	if d.Dir == layout.Stacked {
		r.Y -= (grip - r.H) / 2
		r.H = grip
		return r
	}
	r.X -= (grip - r.W) / 2
	r.W = grip
	return r
}

// emptyState is what the window shows with no panes: a button that opens a
// tab.
func emptyState(c *myui.Context, act Actions) {
	t := c.Theme()
	myui.Column(c).
		Absolute().
		Left(0).
		Right(0).
		Top(TitleBarHeight).
		Bottom(0).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Gap(12).
		Children(func() {
			myui.Text(c, i18n.T("No tabs open")).TextColor(t.TextMuted)
			myui.Button(c, i18n.T("Open a tab")).OnClick(func() { act.NewTab("") })
		})
}
