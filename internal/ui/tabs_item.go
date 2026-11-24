package ui

import (
	myui "github.com/egoist/mygo/ui"
	"wtgo/internal/icons"
	"wtgo/locales"
)

// Boundary: one tab -- its mark, its name and its close button -- and the
// lines the strip draws between tabs. Where it stands is tabs.go; what a tab
// is is state.go.

// tab is one tab: the dot of the profile it runs, its name, and the button
// that closes it.
func tab(c *myui.Context, st State, act Actions, i int, share float32) myui.Element {
	t := c.Theme()
	tb := st.Tabs[i]
	active := i == st.Active
	menu := func(m *myui.Menu) { tabMenu(m, st, act, i) }
	el := myui.Row(c).
		Height(TitleBarHeight).
		AlignItems(myui.Center).
		Gap(8).
		Padding(0, 8, 0, 12).
		Cursor(myui.CursorPointer).
		OnClick(func() { act.FocusTab(i) }).
		Drag(tabDrag{from: i}).
		ContextMenu(menu)
	nameWeight := 0
	if active {
		// The tab in front is the one the eye looks for first: it names itself
		// in bold, as Windows Terminal's does.
		nameWeight = nameBold
	}
	name := measuredName(c, tb.Name, nameWeight)
	// The tab in front carries the pane's colors up into the strip; a tab that
	// draws a page of its own, the settings, has no pane of the terminal's, and
	// neither has a window that draws no pane below the strip. Those take the
	// theme's own colors.
	face := st.Pane
	if face == (Face{}) || (active && tb.Page != nil) {
		face = Face{Background: t.Background, Foreground: t.Text}
	}
	// A tab is as wide as the share of the strip it has, but not wider than
	// its name asks for -- unless the settings make every tab alike -- and not
	// narrower than the room its parts take: past that the strip scrolls
	// rather than squeezing the name and the close button away. What a name
	// asks for is its bold measure whichever tab is in front: the front tab
	// names itself in bold, a few DIPs wider than the rest, and the buttons
	// that open a tab stand beside the last tab -- so a width that followed
	// the weight of the moment would move them as the front tab changed.
	asked := max(name, measuredName(c, tb.Name, nameBold)) + tabChrome
	w := min(max(share, minTabWidth), maxTabWidth)
	if !st.EqualTabs {
		w = min(w, asked)
	}
	el = el.Width(w)
	name = w - tabChrome
	// A tab keeps its width when the strip is short of room: it is the strip
	// that scrolls, not the name and the close button that go.
	el = el.Shrink(0)
	if active {
		// The tab in front carries the pane's background up into the strip,
		// rounded at the top the way Windows Terminal does: the tab joins the
		// pane below it, and the strip is what tells the two apart. The color
		// is the pane's own, not a shade the theme derives: a window whose
		// theme is not the terminal's would otherwise seam at that edge.
		el = el.Background(face.Background).Radius(8, 8, 0, 0)
	} else if el.Hovered() {
		el = el.Background(t.SurfaceHover)
	}
	if active && st.RevealTab {
		// The strip shows the tab in front when it changes, once: asking
		// every frame would keep the user from scrolling it away.
		el = el.ScrollIntoView()
	}
	el = el.ObserveInput(func(ev myui.InputEvent) {
		if ev.Kind == myui.InputPointerDown && ev.Button == mouseMiddle {
			act.CloseTab(i)
		}
	})
	el.Children(func() {
		tabIcon(c, tb, active)
		nameColor := t.TextMuted
		if active {
			nameColor = face.Foreground
		}
		// The name is a label, not text to select: pressing it starts a drag
		// of the tab, the way pressing the rest of the tab does. It carries
		// the tab's menu itself: an unselectable element ends the search for
		// one above it, so a menu only on the row would not open over the
		// name, the widest part of the tab.
		text := myui.Text(c, tb.Name).
			SingleLine().MaxWidth(name).Ellipsis("…").TextColor(nameColor).
			Unselectable().ContextMenu(menu)
		// The name takes the room the tab gives it, not only what its letters
		// need: a short name would otherwise leave the close button a gap short
		// of the tab's edge, where nothing is drawn.
		text = text.Grow(1)
		if nameWeight != 0 {
			// Bold sets the weight on the element already drawing: the call
			// mutates it, so its result is not needed.
			text.Bold()
		}
		if active || el.Hovered() {
			closeButton(c, act, i, active, face)
		} else {
			// The button keeps its room, so that a tab does not change width
			// as the pointer comes over it.
			myui.Box(c).Size(16, 16)
		}
	})
	return el
}

// closeButton closes a tab. It takes the surface under the pointer, so that
// it reads as the button it is rather than as a mark on the tab, and it keeps
// the text color of the tab it sits on: Windows Terminal's close button is as
// bright on the tab behind as on the tab in front.
func closeButton(c *myui.Context, act Actions, i int, active bool, face Face) myui.Element {
	t := c.Theme()
	text := t.Text
	el := myui.Box(c).
		Size(16, 16).
		Radius(4).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Tooltip(locales.CloseTab()).
		OnClick(func() { act.CloseTab(i) })
	if active {
		// The button sits on the pane's background, which the theme knows
		// nothing about, so its hover takes that background rather than the
		// theme's surface.
		text = face.Foreground
		if el.Hovered() {
			el = el.Background(face.Foreground.Alpha(0.16))
		}
	} else if el.Hovered() {
		el = el.Background(t.SurfaceHover)
	}
	return el.Children(func() {
		myui.Icon(c, icons.Close.SVG).Size(14, 14).TextColor(text)
	})
}

// dividerSlot is one of the strip's lines -- the line itself when drawn is set,
// and the room it would take when it is not. The strip leaves a line out beside
// the tab in front, and lays itself out around it either way: a strip that moved
// as another tab came to the front would make the same press land on another
// button, or a tab change width under the pointer.
func dividerSlot(c *myui.Context, drawn bool, room float32) {
	if !drawn {
		myui.Box(c).Width(room).Height(16)
		return
	}
	stripDivider(c).Margin(0, (room-1)/2)
}

// stripDivider is the line the strip draws between two tabs, and beside the
// button that opens one: a DIP wide and as high as a tab's mark, centered in
// the strip, out of the border color between two surfaces. Windows Terminal
// draws the same line in both places.
func stripDivider(c *myui.Context) myui.Element {
	return myui.Box(c).Width(1).Height(16).Background(c.Theme().Border)
}
