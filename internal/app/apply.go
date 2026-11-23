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

// faceChange is what a terminal has to be told about: the settings that reach
// it, each saying whether the last apply already gave it that one.
type faceChange struct {
	font         bool
	colors       bool
	copyOnSelect bool
	links        bool
	scroll       bool
	contrast     bool
}

// faceChanges is what telling the terminals about now, after was, would say.
// first sends everything, which is what a window's first apply does: the
// terminals of the tabs open then have heard nothing yet.
func faceChanges(now, was store.Settings, first bool) faceChange {
	if first {
		return faceChange{font: true, colors: true, copyOnSelect: true, links: true, scroll: true, contrast: true}
	}
	return faceChange{
		font:         now.Font != was.Font,
		colors:       now.ColorScheme != was.ColorScheme || now.Theme != was.Theme || !store.EqualSchemes(now.Schemes, was.Schemes),
		copyOnSelect: now.CopyOnSelect != was.CopyOnSelect,
		links:        now.Links != was.Links,
		scroll:       now.ScrollMultiplier != was.ScrollMultiplier,
		contrast:     now.MinimumContrast != was.MinimumContrast,
	}
}

// applyWindow gives the window itself what the settings ask for: the keys, the
// profiles and the marks of the tabs that run them, the terminals' face, and
// the frame around them.
func (a *App) applyWindow(s store.Settings) {
	a.keys = bindings(s.Keybindings)
	a.profiles = profile.Resolve(s.Profiles, a.found)
	a.syncTabMarks()
	// A terminal is told what changed and nothing else: the settings page
	// applies the draft as it is edited, so a drag of the opacity, a word typed
	// in a search box that changes a setting, or a language would otherwise hand
	// every terminal a font it has (an atlas to build) and colors it wears (a
	// repaint) on every frame. The font and the colors are worked out only when
	// they are sent, which is also what keeps a scheme of Ghostty's from being
	// read off the disk for nothing.
	change := faceChanges(s, a.applied, !a.faceSent)
	var font terminal.Font
	if change.font {
		font = a.terminalFont()
	}
	var colors *terminal.Theme
	if change.colors {
		colors = theme.Terminal(s)
		a.pane = ui.Face{Background: colors.Background, Foreground: colors.Foreground}
		if a.win != nil {
			if err := a.win.SetBackgroundColor(theme.Hex(colors.Background)); err != nil {
				log.Printf("wt-go: background: %v", err)
			}
		}
	}
	copyOnSelect := terminal.CopyOnSelectNone
	if s.CopyOnSelect {
		copyOnSelect = terminal.CopyOnSelectClipboard
	}
	multiplier := scrollMultiplier(s.ScrollMultiplier)
	for _, tab := range a.panes.Tabs {
		for _, term := range tab.Terminals() {
			if change.font {
				term.SetFont(font)
			}
			if change.colors {
				term.SetTheme(colors, nil)
			}
			if change.copyOnSelect {
				term.SetCopyOnSelect(copyOnSelect)
			}
			if change.links {
				setLinks(term, s.Links.Open, s.Links.Hover)
			}
			if change.scroll {
				term.SetScrollMultiplier(multiplier)
			}
			if change.contrast {
				term.SetMinimumContrast(s.MinimumContrast)
			}
		}
	}
	a.faceSent = true
	if a.win != nil {
		a.win.SetOpacity(opacity(s.BackgroundOpacity))
		a.win.SetAlwaysOnTop(s.AlwaysOnTop)
	}
	a.applied = s.Clone()
	a.redraw()
}
