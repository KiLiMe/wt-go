package app

import (
	"github.com/egoist/mygo"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
)

// Boundary: a tab's life -- opening it, closing it and the window with the last
// of them. Closing the others is close_others.go, the strip's own state
// tab_strip.go, and an action id tab_actions.go.

// NewTab opens a tab that runs a profile; "" runs the default profile, which
// is what the strip's '+', the keybinding for a new tab and the button of a
// window without tabs open.
func (a *App) NewTab(profileID string) {
	a.openTab(a.tabProfile(profileID), "")
}

// NewSystemTab opens a tab that runs the shell of the machine: what the
// settings call following the device, and what a window opens when the file
// offers no profile at all.
func (a *App) NewSystemTab() {
	a.openTab(a.device, "")
}

// tabProfile is the profile a tab runs: the one it names, and with none the
// default profile. An id the machine does not have, or one since hidden, is the
// shell of the machine rather than another profile of the list.
func (a *App) tabProfile(profileID string) profile.Profile {
	if profileID == "" {
		profileID = a.settings.DefaultProfile
	}
	return profile.Default(a.profiles, profileID, a.device)
}

// CloseTab closes the tab at i, and with the last tab the window, when the
// settings say so. It asks before it drops the changes of the settings tab, and
// before it ends a program that still runs.
func (a *App) CloseTab(i int) {
	if i < 0 || i >= len(a.panes.Tabs) {
		return
	}
	tab := a.panes.Tabs[i]
	if tab == a.settingsTab && a.dirty() {
		a.askCloseSettings(tab)
		return
	}
	list := a.processList()
	if act := a.activity(tab, list); a.asksClose(act) {
		a.askCloseTab(tab, act)
		return
	}
	a.closeTab(tab)
}

// closeTab closes a tab, and with the last tab the window, when the settings
// say so.
func (a *App) closeTab(tab *pane.Tab) {
	i := a.panes.Index(tab)
	if i < 0 {
		return
	}
	if len(a.panes.Tabs) == 1 && a.settings.CloseOnLastTab {
		// The window goes with this tab; whether it asks first is the
		// window's own business (see onClose).
		a.closeWindow()
		return
	}
	a.panes.Remove(i)
	if tab == a.settingsTab {
		a.closeSettingsTab()
	}
	closeTerminals(tab)
	a.forget(tab.Terminals()...)
	a.revealTab()
	a.syncTitle()
	a.redraw()
}

// onClose keeps the window open to ask about it, when the settings say to.
// Every way the window closes arrives here — its own button, Alt+F4, the code
// that asks for it — so this is the one place that asks.
func (a *App) onClose(e *mygo.CloseEvent) {
	if !a.asksCloseWindow() {
		return
	}
	e.PreventDefault()
	a.askCloseWindow()
}

// asksCloseWindow reports whether closing the window asks first: the settings
// say so, and it holds more than one tab — with one, closing the window is
// what closing that tab means, which the close of a tab already asks about.
// A close the window agreed to, as the answer to this question, does not ask.
func (a *App) asksCloseWindow() bool {
	return a.settings.ConfirmCloseWindow && !a.closing && !a.closed && len(a.panes.Tabs) > 1
}

// closeTerminals closes what a tab holds.
func closeTerminals(tab *pane.Tab) {
	for _, term := range tab.Terminals() {
		term.Close()
	}
}

// closeWindow closes the window, which closes its terminals. The close it hears
// asks about it first (see onClose), so this is the way past that question: the
// answer to it, a reset that asked its own, or a close nothing asks about.
func (a *App) closeWindow() {
	a.closing = true
	if a.win == nil || a.closed {
		return
	}
	a.win.Close()
}

// DuplicateTab opens a tab like the one at i: the profile it runs, in the
// directory the pane with the keyboard is in.
func (a *App) DuplicateTab(i int) {
	if i < 0 || i >= len(a.panes.Tabs) {
		return
	}
	tab := a.panes.Tabs[i]
	if tab.Focus == nil || !tab.Focus.Pane() {
		return
	}
	p := profile.Default(a.profiles, tab.ProfileID, a.device)
	a.openTab(p, tab.Focus.Term.Dir())
}
