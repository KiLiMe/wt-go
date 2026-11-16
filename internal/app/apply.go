package app

import (
	"github.com/egoist/mygo/plugins/terminal"
	"log"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/theme"
	"wtgo/internal/ui"
)

// syncTabMarks gives the tabs already open the marks their profiles now name.
// A tab copies its profile's name, color and mark when it opens, and the
// settings page edits the profiles under the tabs that run them: without this a
// person saves a new icon and every tab but the one in front keeps the old one.
// A profile the settings no longer offer -- removed, or hidden, which is not one
// of the window's -- leaves its tab with the mark it opened with, so that a
// window running a shell that is gone does not turn into a plain one.
func (a *App) syncTabMarks() {
	for _, tab := range a.panes.Tabs {
		if tab.ProfileID == "" {
			continue
		}
		for _, p := range a.profiles {
			if p.ID == tab.ProfileID {
				tab.ProfileName, tab.Color, tab.Icon = p.Name, p.Color, p.Icon
				break
			}
		}
	}
}

// apply gives the window what the settings ask for, and does nothing while
// they are what it gave it last time: the settings tab edits the settings the
// window shows, so that a change shows as it is made.
func (a *App) apply() {
	if store.Equal(a.settings, a.applied) {
		return
	}
	s := a.settings
	// The words the pages speak in, and the desktop's own menus, dialogs and
	// title bar, are the main loop's to set, and each one a test has no loop
	// for: what the window itself takes is applyWindow.
	locale(s.Language)
	a.appearance()
	a.applyWindow(s)
}

// applyWindow gives the window itself what the settings ask for: the keys, the
// profiles and the marks of the tabs that run them, the terminals' face, and
// the frame around them.
func (a *App) applyWindow(s store.Settings) {
	a.keys = bindings(s.Keybindings)
	a.profiles = profile.Resolve(s.Profiles, a.found)
	a.syncTabMarks()
	font := a.terminalFont()
	colors := theme.Terminal(s)
	a.pane = ui.Face{Background: colors.Background, Foreground: colors.Foreground}
	copyOnSelect := terminal.CopyOnSelectNone
	if s.CopyOnSelect {
		copyOnSelect = terminal.CopyOnSelectClipboard
	}
	multiplier := scrollMultiplier(s.ScrollMultiplier)
	for _, tab := range a.panes.Tabs {
		for _, term := range tab.Terminals() {
			term.SetFont(font)
			term.SetTheme(colors, nil)
			term.SetCopyOnSelect(copyOnSelect)
			term.SetLinks(s.Links.Open)
			term.SetLinkHoverHint(s.Links.Underline)
			term.SetScrollMultiplier(multiplier)
			term.SetMinimumContrast(s.MinimumContrast)
		}
	}
	if a.win != nil {
		if err := a.win.SetBackgroundColor(theme.Hex(colors.Background)); err != nil {
			log.Printf("wt-go: background: %v", err)
		}
		a.win.SetOpacity(opacity(s.BackgroundOpacity))
		a.win.SetAlwaysOnTop(s.AlwaysOnTop)
	}
	a.applied = s.Clone()
	// What the settings page lists is built from the settings, so a new set is
	// a new list: it is built again when the page asks.
	a.schemes = nil
	a.redraw()
}
