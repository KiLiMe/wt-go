package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
	"wtgo/internal/i18n"
)

// TitleBarHeight is the height of the tab strip, which is the title bar: the
// window keeps it clear of its window controls, and the strip runs under
// them.
const TitleBarHeight = 36

// tabWidth is how wide a tab is while the settings make them all alike.
const tabWidth float32 = 176

// The room a tab's name may take: as little as a name that is a single
// letter and a close button, as much as stays readable in a strip of many.
const (
	minTabName float32 = 48
	maxTabName float32 = 240
)

// mouseMiddle is the button that closes the tab it points at, as it does in
// Windows Terminal.
const mouseMiddle = 2

// The room the strip keeps whatever the tabs do: the buttons that move them,
// and the part of the title bar that moves the window. A strip of many tabs
// never takes the last of either.
const (
	stepWidth float32 = 24
	dragRoom  float32 = 56
)

// The room a window button takes, the gap between two of them, and the room
// the buttons keep from the corner of the window.
const (
	windowButtonWidth float32 = 34
	windowButtonGap   float32 = 2
	windowButtonRoom  float32 = 6
)

// tabStep is how far the buttons beside the tabs move what the strip shows:
// about one tab, so that a press brings the next tabs into view.
const tabStep float32 = 176

// tabDrag is a tab on its way to another place in the strip.
type tabDrag struct {
	from int
}

// tabMove is a tab dropped on another place: the tab at from belongs at to.
type tabMove struct {
	from, to int
}

// strip is the title bar: the window's buttons where the window draws them,
// the tabs, the menu that opens tabs, the button that opens one, and the room
// that moves the window.
func strip(c *myui.Context, st State, act Actions) {
	t := c.Theme()
	bar := c.TitleBar()
	myui.Row(c).
		Height(TitleBarHeight).
		Background(t.Surface).
		Padding(0, bar.Right, 0, bar.Left).
		DragWindow().
		Children(func() {
			if b := st.WindowButtons; b != nil && len(b.Start) > 0 {
				windowButtonGroup(c, b, act, b.Start, false)
			}
			stripTabs(c, st, act)
			profilesMenu(c, st, act)
			newTabButton(c, act)
			myui.Box(c).Width(dragRoom).Height(TitleBarHeight)
			if b := st.WindowButtons; b != nil && len(b.End) > 0 {
				windowButtonGroup(c, b, act, b.End, true)
			}
		})
}

// windowButtonGroup draws the window's buttons the desktop puts on a side of
// the title bar, at the end of it when end is set, and nothing where it puts
// none: a desktop may name some, and a tiling window manager names none.
func windowButtonGroup(c *myui.Context, b *WindowButtons, act Actions, names []string, end bool) myui.Element {
	row := myui.Row(c).
		Height(TitleBarHeight).
		Gap(windowButtonGap).
		AlignItems(myui.Center)
	if end {
		row = row.Padding(0, windowButtonRoom, 0, 0)
	} else {
		row = row.Padding(0, 0, 0, windowButtonRoom)
	}
	return row.Children(func() {
		for _, name := range names {
			windowButton(c, b, act, name)
		}
	})
}

// windowButton is one of the window's buttons. It draws a glyph rather than an
// image, as the strip draws its other buttons, and follows the desktop's
// layout rather than a look of its own.
func windowButton(c *myui.Context, b *WindowButtons, act Actions, name string) {
	t := c.Theme()
	glyph, tip := "", ""
	size := float32(14)
	switch name {
	case "minimize":
		glyph, tip = "\u2013", i18n.T("Minimize")
	case "maximize":
		size = 12
		if b.Maximized {
			glyph, tip = "\u2750", i18n.T("Restore")
		} else {
			glyph, tip = "\u25a1", i18n.T("Maximize")
		}
	case "close":
		glyph, tip = "\u00d7", i18n.T("Close Window")
	default:
		return
	}
	color := t.TextMuted
	el := myui.Box(c).
		Size(windowButtonWidth, 28).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Tooltip(tip)
	if el.Hovered() {
		el = el.Background(t.SurfaceHover)
		color = t.Text
	}
	el.OnClick(func() {
		switch name {
		case "minimize":
			act.MinimizeWindow()
		case "maximize":
			act.ToggleMaximizeWindow()
		case "close":
			act.CloseWindow()
		}
	}).Children(func() {
		myui.Text(c, glyph).FontSize(size).TextColor(color)
	})
}

// stripTabs is the room the tabs take: the buttons that move them, and the
// tabs themselves, which scroll sideways when they outgrow the room.
func stripTabs(c *myui.Context, st State, act Actions) {
	if st.TabsScroll == nil {
		tabs(c, st, act)
		return
	}
	// The buttons take room from the tabs, so they only show once the tabs
	// outgrow it. The viewport is measured every frame whether they show or
	// not, so the frame after a tab no longer fits has them, and a frame
	// they are in keeps them: what they take is what makes the tabs
	// outgrow the strip.
	scrolls := st.TabsScroll.MaxX > 0
	myui.Row(c).Height(TitleBarHeight).Grow(1).Children(func() {
		if scrolls {
			stepButton(c, st, act, -1)
		}
		viewport(c, st, act)
		if scrolls {
			stepButton(c, st, act, 1)
		}
	})
}

// viewport is what the strip shows of the tabs: it scrolls them sideways when
// they outgrow it, which the wheel over it does too.
func viewport(c *myui.Context, st State, act Actions) {
	myui.ScrollHorizontal(c).
		Height(TitleBarHeight).
		Grow(1).
		Scrollbars(false).
		TrackScroll(st.TabsScroll).
		HandleInput(func(ev myui.InputEvent) bool {
			if ev.Kind != myui.InputScroll {
				return false
			}
			by := ev.DX
			if by == 0 {
				by = ev.DY
			}
			if by == 0 {
				return false
			}
			// The strip has no height to scroll in, so the wheel goes
			// sideways: it is the only way to reach the tabs past the
			// edge with a wheel.
			act.ScrollTabs(by)
			return true
		}).
		Children(func() { tabs(c, st, act) })
}

// stepButton moves what the strip shows by a tab's worth, toward the start
// with -1 and toward the end with 1. It fades when there is nothing that way,
// rather than going Disabled, which would take the press with it: the press
// is harmless, the window keeps the offset in the strip.
func stepButton(c *myui.Context, st State, act Actions, dir int) myui.Element {
	t := c.Theme()
	// The heavy angle quotes carry more weight than the single guillemets,
	// which read as specks at this size.
	glyph, tip := "\u276f", i18n.T("Scroll the tabs right")
	if dir < 0 {
		glyph, tip = "\u276e", i18n.T("Scroll the tabs left")
	}
	at := st.TabsScroll
	color := t.Text
	if at == nil || dir < 0 && at.X <= 0 || dir > 0 && at.X >= at.MaxX {
		color = t.TextMuted
		color.A = uint8(float32(color.A) * 0.45)
	}
	return myui.Box(c).
		Size(stepWidth, 28).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Tooltip(tip).
		OnClick(func() { act.ScrollTabs(float32(dir) * tabStep) }).
		Children(func() {
			myui.Text(c, glyph).FontSize(18).TextColor(color)
		})
}

// tabs are the tabs, in order, each a place another tab may be dropped on.
func tabs(c *myui.Context, st State, act Actions) {
	var move *tabMove
	myui.Row(c).Height(TitleBarHeight).Children(func() {
		for i := range st.Tabs {
			el := tab(c, st, act, i)
			if d, ok := myui.Drop[tabDrag](el); ok && move == nil {
				move = &tabMove{from: d.from, to: i}
			}
		}
	})
	if move != nil {
		act.MoveTab(move.from, move.to)
	}
}

// tab is one tab: the dot of the profile it runs, its name, and the button
// that closes it.
func tab(c *myui.Context, st State, act Actions, i int) myui.Element {
	t := c.Theme()
	tb := st.Tabs[i]
	active := i == st.Active
	el := myui.Row(c).
		Height(TitleBarHeight).
		AlignItems(myui.Center).
		Gap(8).
		Padding(0, 8, 0, 12).
		Cursor(myui.CursorPointer).
		OnClick(func() { act.FocusTab(i) }).
		Drag(tabDrag{from: i}).
		ContextMenu(func(m *myui.Menu) { tabMenu(m, st, act, i) })
	name := measuredName(c, tb.Name)
	if st.EqualTabs {
		// Every tab as wide as the others, whatever its name takes.
		el = el.Width(tabWidth)
		name = tabWidth - 60
	}
	// A tab keeps its width when the strip is short of room: it is the strip
	// that scrolls, not the name and the close button that go.
	el = el.Shrink(0)
	if active {
		// The tab in front carries the terminal's background up into the
		// strip, rounded at the top the way Windows Terminal does.
		el = el.Background(t.Background).Radius(8, 8, 0, 0)
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
		if dot := colorOf(tb.Color); dot.A > 0 {
			myui.Box(c).Size(8, 8).Radius(4).Background(dot)
		}
		nameColor := t.TextMuted
		if active {
			nameColor = t.Text
		}
		// The name is a label, not text to select: pressing it starts a drag
		// of the tab, the way pressing the rest of the tab does.
		myui.Text(c, tb.Name).SingleLine().MaxWidth(name).Ellipsis("…").TextColor(nameColor).Unselectable()
		if active || el.Hovered() {
			closeButton(c, act, i, active)
		} else {
			// The button keeps its room, so that a tab does not change width
			// as the pointer comes over it.
			myui.Box(c).Size(16, 16)
		}
	})
	return el
}

// measuredName is how wide a tab draws its name: as wide as the name takes
// in the theme's font, within the room a tab may take.
func measuredName(c *myui.Context, name string) float32 {
	t := c.Theme()
	w, _ := c.MeasureText(0, myui.Span{Text: name, Font: t.Font, Size: t.FontSize})
	return nameWidth(w)
}

// nameWidth keeps a measured name inside the room a tab may take.
func nameWidth(w float32) float32 {
	return min(max(w, minTabName), maxTabName)
}

// tabMenu is the tab's context menu.
func tabMenu(m *myui.Menu, st State, act Actions, i int) {
	if m.Item(i18n.T("Rename")).Chosen() {
		act.BeginRename(i)
	}
	if m.Item(i18n.T("Duplicate Tab")).Chosen() {
		act.DuplicateTab(i)
	}
	m.Separator()
	if m.Item(i18n.T("Close Tab")).Chosen() {
		act.CloseTab(i)
	}
	if m.Item(i18n.T("Close Other Tabs")).Disabled(len(st.Tabs) < 2).Chosen() {
		act.CloseOtherTabs(i)
	}
}

// closeButton closes a tab.
func closeButton(c *myui.Context, act Actions, i int, active bool) myui.Element {
	t := c.Theme()
	color := t.TextMuted
	if active {
		color = t.Text
	}
	return myui.Box(c).
		Size(16, 16).
		Radius(4).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Tooltip(i18n.T("Close Tab")).
		OnClick(func() { act.CloseTab(i) }).
		Children(func() {
			myui.Text(c, "\u00d7").FontSize(13).TextColor(color)
		})
}

// profilesMenu opens a tab that runs one of the profiles.
func profilesMenu(c *myui.Context, st State, act Actions) myui.Element {
	t := c.Theme()
	return myui.Box(c).
		Size(28, 28).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Tooltip(i18n.T("New Tab")).Menu(func(m *myui.Menu) {
		for i, p := range st.Profiles {
			it := m.Item(p.Name)
			if i < 9 {
				shortcut(it, st, action.ProfileAction(i+1))
			}
			if it.Chosen() {
				act.NewTab(p.ID)
			}
		}
		m.Separator()
		it := m.Item(i18n.T("Settings"))
		shortcut(it, st, action.Settings)
		if it.Chosen() {
			act.OpenSettings()
		}
	}).
		Children(func() {
			myui.Text(c, "\u25be").FontSize(12).TextColor(t.TextMuted)
		})
}

// shortcut shows an action's key by a menu item, when the window knows one.
func shortcut(it *myui.MenuItem, st State, action string) {
	if st.Shortcut == nil {
		return
	}
	mods, key, ok := ParseKey(st.Shortcut(action))
	if !ok {
		return
	}
	it.Shortcut(mods, key)
}

// newTabButton opens a tab that runs the default profile.
func newTabButton(c *myui.Context, act Actions) myui.Element {
	t := c.Theme()
	return myui.Box(c).
		Size(28, 28).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Tooltip(i18n.T("New Tab")).
		OnClick(func() { act.NewTab("") }).
		Children(func() {
			myui.Text(c, "+").FontSize(16).TextColor(t.TextMuted)
		})
}
