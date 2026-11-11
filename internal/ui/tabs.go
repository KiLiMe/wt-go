package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
	"wtgo/internal/i18n"
	"wtgo/internal/icons"
)

// TitleBarHeight is the height of the tab strip, which is the title bar: the
// window keeps it clear of its window controls, and the strip runs under
// them. It is the height of Windows Terminal's, so that a window beside one
// keeps the two bars level.
const TitleBarHeight = 40

// tabHeight is the height of a tab, less than the strip's: the tab in front
// floats clear of the window's top edge, so that its rounded top corners are
// drawn rather than cut off, the way Windows Terminal's are.
const tabHeight float32 = 32

// tabLine is how thick the line along the top of the tab in front is, and
// tabLineInset how far it stops short of the tab's ends: the tab's top corners
// are rounded, and a line run to them would show outside the round.
const (
	tabLine      float32 = 2
	tabLineInset float32 = 4
)

// stripGutter is the room the strip keeps between the window's left edge and
// the first tab, for the same reason.
const stripGutter float32 = 8

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

// windowIcon is how wide the icon of a window button is: a square that fills
// the icon area of the desktop's own buttons.
const windowIcon float32 = 10

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
// the tabs with the buttons that open one, and the room that moves the window.
func strip(c *myui.Context, st State, act Actions) {
	t := c.Theme()
	bar := c.TitleBar()
	myui.Row(c).
		Height(TitleBarHeight).
		Background(t.Surface).
		Padding(0, bar.Right, 0, bar.Left+stripGutter).
		DragWindow().
		Children(func() {
			if b := st.WindowButtons; b != nil && len(b.Start) > 0 {
				windowButtonGroup(c, b, act, b.Start, false)
			}
			stripTabs(c, st, act)
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

// windowButton is one of the window's buttons. It draws each mark itself, in
// 1px lines at the size the title bar asks: the glyphs for these are missing
// from some fonts, or wide in the fonts that draw them, and the strokes of an
// icon set come out bold at this size.
func windowButton(c *myui.Context, b *WindowButtons, act Actions, name string) {
	t := c.Theme()
	label := ""
	var icon func(color myui.Color, bg myui.Color)
	switch name {
	case "minimize":
		label = i18n.T("Minimize")
		icon = func(color myui.Color, bg myui.Color) {
			myui.Box(c).Width(windowIcon).Height(1).Background(color)
		}
	case "maximize":
		label = i18n.T("Maximize")
		if b.Maximized {
			label = i18n.T("Restore")
			icon = func(color myui.Color, bg myui.Color) {
				myui.Box(c).Size(windowIcon+1, windowIcon+1).Children(func() {
					myui.Box(c).Absolute().Left(4).Top(0).Size(windowIcon-3, windowIcon-3).Radius(1).Border(1, color)
					myui.Box(c).Absolute().Left(0).Top(4).Size(windowIcon-3, windowIcon-3).Radius(1).Border(1, color).Background(bg)
				})
			}
		} else {
			icon = func(color myui.Color, bg myui.Color) {
				myui.Box(c).Size(windowIcon-1, windowIcon-1).Radius(1).Border(1, color)
			}
		}
	case "close":
		label = i18n.T("Close Window")
		icon = func(color myui.Color, bg myui.Color) {
			// A cross of 1px lines, as Windows draws it: the icon set's close
			// is a 2-unit stroke on a 24 grid, which comes out bold beside a
			// 1px minus and box.
			myui.Box(c).Size(windowIcon, windowIcon).Draw(func(p *myui.Painter, r myui.Rect) {
				const in = 0.5
				p.Line(r.X+in, r.Y+in, r.X+r.W-in, r.Y+r.H-in, 1, color)
				p.Line(r.X+r.W-in, r.Y+in, r.X+in, r.Y+r.H-in, 1, color)
			})
		}
	default:
		return
	}
	bg := t.Surface
	color := t.TextMuted
	el := myui.Box(c).
		Size(windowButtonWidth, 28).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Label(label).
		Tooltip(label)
	if el.Hovered() {
		bg = t.SurfaceHover
		color = t.Text
		if name == "close" {
			// The button that takes the window away turns the red Windows
			// gives it: the one button of the three that does something the
			// others do not is the one worth a color of its own.
			bg = t.Danger
		}
		el = el.Background(bg)
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
		icon(color, bg)
	})
}

// stripTabs is the room the tabs take: the buttons that move them, and the
// tabs themselves, which scroll sideways when they outgrow the room.
func stripTabs(c *myui.Context, st State, act Actions) {
	if st.TabsScroll == nil {
		myui.Row(c).Height(TitleBarHeight).Grow(1).AlignItems(myui.Center).Children(func() {
			tabs(c, st, act)
			stripButtons(c, st, act)
		})
		return
	}
	// The buttons take room from the tabs, so they only show once the tabs
	// outgrow it. The viewport is measured every frame whether they show or
	// not, so the frame after a tab no longer fits has them, and a frame
	// they are in keeps them: what they take is what makes the tabs
	// outgrow the strip.
	scrolls := st.TabsScroll.MaxX > 0
	myui.Row(c).Height(TitleBarHeight).Grow(1).AlignItems(myui.Center).Children(func() {
		if scrolls {
			stepButton(c, st, act, -1)
		}
		viewport(c, st, act)
		if scrolls {
			stepButton(c, st, act, 1)
		}
		// The buttons that open a tab sit past the tabs rather than in them:
		// however many tabs there are, a tab is a press away rather than a
		// scroll to the end away.
		stripButtons(c, st, act)
	})
}

// viewport is what the strip shows of the tabs: it scrolls them sideways when
// they outgrow it, which the wheel over it does too.
func viewport(c *myui.Context, st State, act Actions) {
	myui.ScrollHorizontal(c).
		Height(TitleBarHeight).
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
	// The arrows come from the icon set: the angle quotes that would stand
	// for them read as specks at this size, and some fonts have no such
	// glyph at all.
	ic, tip := icons.Right, i18n.T("Scroll the tabs right")
	if dir < 0 {
		ic, tip = icons.Left, i18n.T("Scroll the tabs left")
	}
	at := st.TabsScroll
	color := t.Text
	if at == nil || dir < 0 && at.X <= 0 || dir > 0 && at.X >= at.MaxX {
		color = t.TextMuted
		color.A = uint8(float32(color.A) * 0.45)
	}
	el := myui.Box(c).
		Size(stepWidth, 28).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Label(tip).
		Tooltip(tip).
		OnClick(func() { act.ScrollTabs(float32(dir) * tabStep) })
	if el.Hovered() {
		el = el.Background(t.SurfaceHover)
	}
	return el.Children(func() {
		myui.Icon(c, ic.SVG).Size(16, 16).TextColor(color)
	})
}

// tabs are the tabs, in order, each a place another tab may be dropped on,
// and the buttons that open a tab after the last of them.
func tabs(c *myui.Context, st State, act Actions) {
	var move *tabMove
	myui.Row(c).Height(TitleBarHeight).AlignItems(myui.End).Gap(4).Children(func() {
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

// stripButtons are the buttons after the last tab: the one that opens a tab
// and the menu that picks the profile it runs, with the line Windows Terminal
// draws between them. They scroll with the tabs, so that they stay beside the
// tab they open next to rather than drift to the far end of a wide window.
func stripButtons(c *myui.Context, st State, act Actions) myui.Element {
	t := c.Theme()
	return myui.Row(c).Height(TitleBarHeight).AlignItems(myui.Center).Gap(2).Children(func() {
		newTabButton(c, act)
		myui.Box(c).Width(1).Height(16).Margin(0, 4).Background(t.Border)
		profilesMenu(c, st, act)
	})
}

// tab is one tab: the dot of the profile it runs, its name, and the button
// that closes it.
func tab(c *myui.Context, st State, act Actions, i int) myui.Element {
	t := c.Theme()
	tb := st.Tabs[i]
	active := i == st.Active
	menu := func(m *myui.Menu) { tabMenu(m, st, act, i) }
	el := myui.Row(c).
		Height(tabHeight).
		AlignItems(myui.Center).
		Gap(8).
		Padding(0, 8, 0, 12).
		Cursor(myui.CursorPointer).
		OnClick(func() { act.FocusTab(i) }).
		Drag(tabDrag{from: i}).
		ContextMenu(menu)
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
		// strip, rounded at the top the way Windows Terminal does, and draws
		// the accent along its top edge: the tab joins the pane below it, so
		// the line is what tells the two apart at a glance.
		el = el.Background(t.Background).Radius(8, 8, 0, 0).
			Draw(func(p *myui.Painter, r myui.Rect) {
				p.Fill(myui.Rect{X: r.X + tabLineInset, Y: r.Y, W: r.W - 2*tabLineInset, H: tabLine},
					tabAccent(t), 1)
			})
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
		// of the tab, the way pressing the rest of the tab does. It carries
		// the tab's menu itself: an unselectable element ends the search for
		// one above it, so a menu only on the row would not open over the
		// name, the widest part of the tab.
		myui.Text(c, tb.Name).
			SingleLine().MaxWidth(name).Ellipsis("…").TextColor(nameColor).
			Unselectable().ContextMenu(menu)
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

// tabAccent is the color of the line along the top of the tab in front: the
// accent of the theme, a shade toward the text, which a line two DIPs high
// needs -- the accent itself reads as a filled button's face, too near the
// background to stand on its own at that size.
func tabAccent(t *myui.Theme) myui.Color {
	return t.Accent.MixOklab(t.Text, 0.2)
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

// tabMenu is the tab's context menu. The tab that is not a terminal, the
// settings, has no name of its own to keep and nothing to open a second time:
// its menu offers only closing it, so it opens without the line that would
// separate what it does not offer.
func tabMenu(m *myui.Menu, st State, act Actions, i int) {
	if i >= 0 && i < len(st.Tabs) && st.Tabs[i].Page == nil {
		if m.Item(i18n.T("Rename")).Chosen() {
			act.BeginRename(i)
		}
		if m.Item(i18n.T("Duplicate Tab")).Chosen() {
			act.DuplicateTab(i)
		}
		m.Separator()
	}
	if m.Item(i18n.T("Close Tab")).Chosen() {
		act.CloseTab(i)
	}
	if m.Item(i18n.T("Close Other Tabs")).Disabled(len(st.Tabs) < 2).Chosen() {
		act.CloseOtherTabs(i)
	}
}

// closeButton closes a tab. It takes the surface under the pointer, so that
// it reads as the button it is rather than as a mark on the tab.
func closeButton(c *myui.Context, act Actions, i int, active bool) myui.Element {
	t := c.Theme()
	color := t.TextMuted
	if active {
		color = t.Text
	}
	el := myui.Box(c).
		Size(16, 16).
		Radius(4).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Tooltip(i18n.T("Close Tab")).
		OnClick(func() { act.CloseTab(i) })
	if el.Hovered() {
		el = el.Background(t.SurfaceHover)
		color = t.Text
	}
	return el.Children(func() {
		myui.Icon(c, icons.Close.SVG).Size(14, 14).TextColor(color)
	})
}

// profilesMenu opens a tab that runs one of the profiles.
func profilesMenu(c *myui.Context, st State, act Actions) myui.Element {
	base := myui.Box(c).
		Size(28, 28).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Label(i18n.T("Profiles")).
		Tooltip(i18n.T("Profiles")).Menu(func(m *myui.Menu) {
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
	})
	return stripButton(c, base).Children(func() {
		myui.Icon(c, icons.Chevron.SVG).Size(14, 14).TextColor(stripIcon(c, base))
	})
}

// stripButton is one of the buttons beside the tabs: a square in the room of a
// glyph and a half, which takes the surface under the pointer, as the window's
// own buttons do, so that a mark in the strip reads as a button.
func stripButton(c *myui.Context, el myui.Element) myui.Element {
	if el.Hovered() {
		return el.Background(c.Theme().SurfaceHover)
	}
	return el
}

// stripIcon is the color of a strip button's mark: the text color while the
// pointer is over the button, the muted one otherwise.
func stripIcon(c *myui.Context, el myui.Element) myui.Color {
	if el.Hovered() {
		return c.Theme().Text
	}
	return c.Theme().TextMuted
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
	base := myui.Box(c).
		Size(28, 28).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Label(i18n.T("New Tab")).
		Tooltip(i18n.T("New Tab")).
		OnClick(func() { act.NewTab("") })
	return stripButton(c, base).Children(func() {
		myui.Icon(c, icons.Add.SVG).Size(14, 14).TextColor(stripIcon(c, base))
	})
}
