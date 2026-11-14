package ui

import (
	myui "github.com/egoist/mygo/ui"
	"wtgo/internal/action"
	"wtgo/internal/icons"
	"wtgo/locales"
)

// titleBarMenu is the menu the strip opens on a right press: the window's own
// menu, which the desktop shows on a right press of the title bar it draws, and
// Windows Terminal offers on its strip as well. The two items that ask the
// desktop to move or to size the window are left out: mygo offers no move the
// pointer is not already dragging, and an item that did nothing would be a
// worse lie than its absence.
func titleBarMenu(m *myui.Menu, st State, act Actions) {
	maximized := st.WindowButtons != nil && st.WindowButtons.Maximized
	if it := m.Item(locales.Restore()).Disabled(!maximized); it.Chosen() {
		act.ToggleMaximizeWindow()
	}
	if m.Item(locales.Minimize()).Chosen() {
		act.MinimizeWindow()
	}
	if it := m.Item(locales.Maximize()).Disabled(maximized); it.Chosen() {
		act.ToggleMaximizeWindow()
	}
	m.Separator()
	closer := m.Item(locales.CloseWindow())
	if mods, key, ok := ParseKey("alt+f4"); ok {
		// The desktop takes this key before the window ever sees it, and it is
		// what a person would press anyway: the menu names it rather than
		// binding it.
		closer.Shortcut(mods, key)
	}
	if closer.Chosen() {
		act.CloseWindow()
	}
	settings := m.Item(locales.Settings())
	shortcut(settings, st, action.Settings)
	if settings.Chosen() {
		act.OpenSettings()
	}
}

// tabMenu is the tab's context menu. The tab that is not a terminal, the
// settings, has no name of its own to keep and nothing to open a second time:
// its menu offers only closing it, so it opens without the line that would
// separate what it does not offer.
func tabMenu(m *myui.Menu, st State, act Actions, i int) {
	if i >= 0 && i < len(st.Tabs) && st.Tabs[i].Page == nil {
		if m.Item(locales.Rename()).Chosen() {
			act.BeginRename(i)
		}
		if m.Item(locales.DuplicateTab()).Chosen() {
			act.DuplicateTab(i)
		}
		m.Separator()
	}
	if m.Item(locales.CloseTab()).Chosen() {
		act.CloseTab(i)
	}
	if m.Item(locales.CloseOtherTabs()).Disabled(len(st.Tabs) < 2).Chosen() {
		act.CloseOtherTabs(i)
	}
}

// stripProfiles are what the menu of the strip lists. With every profile
// deleted or hidden it lists one entry, following the device, which is the
// shell the window runs then anyway: a menu with nothing to open a tab on
// reads as a window that cannot open one. That entry names the device the way
// the settings page does, so the menu says which shell it opens.
func stripProfiles(st State) []Profile {
	if len(st.Profiles) > 0 {
		return st.Profiles
	}
	return []Profile{{Name: deviceLabel(st.Device)}}
}

// profilesMenu opens a tab that runs one of the profiles.
func profilesMenu(c *myui.Context, st State, act Actions) myui.Element {
	base := myui.Box(c).
		Size(stripButtonSize, stripButtonSize).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Label(locales.Profiles()).
		Tooltip(locales.Profiles()).Menu(func(m *myui.Menu) {
		for i, p := range stripProfiles(st) {
			it := m.Item(p.Name)
			if p.ID == "" {
				// Following the device is no numbered profile of the file.
				if it.Chosen() {
					act.NewSystemTab()
				}
				continue
			}
			if i < 9 {
				shortcut(it, st, action.ProfileAction(i+1))
			}
			if it.Chosen() {
				act.NewTab(p.ID)
			}
		}
		m.Separator()
		it := m.Item(locales.Settings())
		shortcut(it, st, action.Settings)
		if it.Chosen() {
			act.OpenSettings()
		}
		if about := m.Item(locales.About()); about.Chosen() {
			act.OpenAbout()
		}
	})
	return stripButton(c, base).Children(func() {
		myui.Icon(c, icons.Chevron.SVG).Size(14, 14).TextColor(c.Theme().Text)
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
