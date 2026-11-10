package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/i18n"
	"wtgo/internal/store"
)

// TitleBarHeight is the height of the tab strip, which is the title bar: the
// window keeps it clear of its window controls, and the strip runs under
// them.
const TitleBarHeight = 36

// tabWidth is how wide a tab is while the settings make them all alike.
const tabWidth float32 = 176

// tabDrag is a tab on its way to another place in the strip.
type tabDrag struct {
	from int
}

// tabMove is a tab dropped on another place: the tab at from belongs at to.
type tabMove struct {
	from, to int
}

// strip is the title bar: the tabs, the menu that opens tabs, the button that
// opens one, and the room that moves the window.
func strip(c *myui.Context, st State, act Actions) {
	t := c.Theme()
	bar := c.TitleBar()
	myui.Row(c).
		Height(TitleBarHeight).
		Background(t.Surface).
		Padding(0, bar.Right, 0, bar.Left).
		DragWindow().
		Children(func() {
			tabs(c, st, act)
			profilesMenu(c, st, act)
			newTabButton(c, act)
			myui.Spacer(c)
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
	name := float32(220)
	if st.EqualTabs {
		// Every tab as wide as the others: the strip cannot measure the
		// names, so they share the width the settings give them.
		el = el.Width(tabWidth)
		name = tabWidth - 60
	}
	if active {
		// The tab in front carries the terminal's background up into the
		// strip, rounded at the top the way Windows Terminal does.
		el = el.Background(t.Background).Radius(8, 8, 0, 0)
	} else if el.Hovered() {
		el = el.Background(t.SurfaceHover)
	}
	el.Children(func() {
		if dot := colorOf(tb.Color); dot.A > 0 {
			myui.Box(c).Size(8, 8).Radius(4).Background(dot)
		}
		nameColor := t.TextMuted
		if active {
			nameColor = t.Text
		}
		myui.Text(c, tb.Name).SingleLine().MaxWidth(name).TextColor(nameColor)
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
				shortcut(it, st, store.ProfileAction(i+1))
			}
			if it.Chosen() {
				act.NewTab(p.ID)
			}
		}
		m.Separator()
		it := m.Item(i18n.T("Settings"))
		shortcut(it, st, store.ActionSettings)
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
