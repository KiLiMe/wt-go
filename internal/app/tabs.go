package app

import (
	"log"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/action"
	"wtgo/internal/layout"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/ui"
)

// NewTab opens a tab that runs a profile; "" runs the default one.
func (a *App) NewTab(profileID string) {
	p, ok := profile.Default(a.profiles, profileID)
	if !ok {
		return
	}
	a.openTab(p, "")
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
	if a.settings.ConfirmClose && tab.Running() {
		a.askCloseTab(tab)
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
		a.closeWindow()
		closeTerminals(tab)
		return
	}
	a.panes.Remove(i)
	if tab == a.settingsTab {
		a.closeSettingsTab()
	}
	closeTerminals(tab)
	a.syncTitle()
	a.redraw()
}

// closeTerminals closes what a tab holds.
func closeTerminals(tab *pane.Tab) {
	for _, term := range tab.Terminals() {
		term.Close()
	}
}

// closeWindow closes the window, which closes its terminals.
func (a *App) closeWindow() {
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
	p, ok := profile.Default(a.profiles, tab.ProfileID)
	if !ok {
		return
	}
	a.openTab(p, tab.Focus.Term.Dir())
}

// CloseOtherTabs closes every tab but the one at i, asking once when it would
// end programs that still run.
func (a *App) CloseOtherTabs(i int) {
	if i < 0 || i >= len(a.panes.Tabs) {
		return
	}
	tab := a.panes.Tabs[i]
	if a.settings.ConfirmClose && a.othersRunning(tab) {
		a.askCloseOtherTabs(tab)
		return
	}
	a.closeOtherTabs(tab)
}

// othersRunning reports whether a tab but keep still runs its program.
func (a *App) othersRunning(keep *pane.Tab) bool {
	for _, tab := range a.panes.Tabs {
		if tab != keep && tab.Running() {
			return true
		}
	}
	return false
}

// closeOtherTabs closes every tab but keep.
func (a *App) closeOtherTabs(keep *pane.Tab) {
	i := a.panes.Index(keep)
	if i < 0 {
		return
	}
	for _, tab := range a.panes.Keep(i) {
		if tab == a.settingsTab {
			a.closeSettingsTab()
		}
		closeTerminals(tab)
	}
	a.syncTitle()
	a.redraw()
}

// MoveTab moves the tab at from to to.
func (a *App) MoveTab(from, to int) {
	a.panes.Move(from, to)
	a.redraw()
}

// FocusTab brings the tab at i in front.
func (a *App) FocusTab(i int) {
	if i < 0 || i >= len(a.panes.Tabs) {
		return
	}
	a.panes.Active = i
	a.syncTitle()
	a.redraw()
}

// BeginRename opens the dialog that renames the tab at i, with the name the
// tab has in its field.
func (a *App) BeginRename(i int) {
	if i < 0 || i >= len(a.panes.Tabs) {
		return
	}
	a.renaming = &ui.Rename{Open: true, Tab: i, Text: a.panes.Tabs[i].Name}
	a.redraw()
}

// RenameTab gives the tab at i a name; "" gives it the name its terminal
// and its profile make.
func (a *App) RenameTab(i int, name string) {
	if i >= 0 && i < len(a.panes.Tabs) {
		a.panes.Tabs[i].Name = strings.TrimSpace(name)
	}
	a.renaming = nil
	a.syncTitle()
	a.redraw()
}

// CancelRename closes the rename dialog.
func (a *App) CancelRename() {
	a.renaming = nil
	a.redraw()
}

// FocusPane gives the keyboard to a pane of the tab at i, and the tab that
// pane is in the one in front.
func (a *App) FocusPane(i int, n *pane.Node) {
	if i < 0 || i >= len(a.panes.Tabs) || !n.Pane() {
		return
	}
	tab := a.panes.Tabs[i]
	if tab.Focus == n && a.panes.Active == i {
		return
	}
	tab.Focus = n
	a.panes.Active = i
	a.syncTitle()
	a.redraw()
}

// ResizePane moves the line of a split of the tab at i by delta DIPs.
func (a *App) ResizePane(i int, d pane.Divide, delta float32) {
	if i < 0 || i >= len(a.panes.Tabs) {
		return
	}
	a.panes.Tabs[i].Resize(d, delta)
	a.redraw()
}

// closePane closes the pane with the keyboard of the tab at i, and the tab
// with its last pane. Closing the last pane is closing the tab, so it goes
// through the questions a tab close asks.
func (a *App) closePane(i int) {
	if i < 0 || i >= len(a.panes.Tabs) {
		return
	}
	tab := a.panes.Tabs[i]
	if tab.Root == nil || tab.Count() == 1 {
		a.CloseTab(i)
		return
	}
	if term, _ := tab.Close(); term != nil {
		term.Close()
	}
	a.syncTitle()
	a.redraw()
}

// closeOtherPanes closes every pane of the tab at i but the one with the
// keyboard.
func (a *App) closeOtherPanes(i int) {
	if i < 0 || i >= len(a.panes.Tabs) {
		return
	}
	tab := a.panes.Tabs[i]
	if tab.Focus == nil {
		return
	}
	for _, term := range closeOthers(tab) {
		term.Close()
	}
	a.syncTitle()
	a.redraw()
}

// closeOthers drops every pane of tab but the one with the keyboard, and
// hands back the terminals it dropped for the caller to close.
func closeOthers(tab *pane.Tab) []*terminal.Terminal {
	keep := tab.Focus
	var dropped []*terminal.Terminal
	for _, n := range tab.Panes() {
		if n == keep {
			continue
		}
		if term, _ := tab.CloseAt(n); term != nil {
			dropped = append(dropped, term)
		}
	}
	return dropped
}

// splitPane divides the pane with the keyboard in two, the new pane running
// the profile of the tab, in the directory that pane is in.
func (a *App) splitPane(dir layout.Dir) {
	tab, ok := a.panes.Front()
	if !ok || tab.Focus == nil || !tab.Focus.Pane() {
		return
	}
	if a.settings.MaxPanes > 0 && tab.Count() >= a.settings.MaxPanes {
		return
	}
	p, ok := profile.Default(a.profiles, tab.ProfileID)
	if !ok {
		return
	}
	term, err := terminal.New(a.termOptions(p, tab.Focus.Term.Dir()))
	if err != nil {
		log.Printf("wt-go: %s: %v", p.Name, err)
		return
	}
	tab.Split(dir, term)
	a.syncTitle()
	a.redraw()
}

// moveFocus gives the keyboard to the pane next to it in dir.
func (a *App) moveFocus(dir pane.Direction) {
	tab, ok := a.panes.Front()
	if !ok || tab.Focus == nil || !tab.Move(dir, a.content) {
		return
	}
	a.syncTitle()
	a.redraw()
}

// act does the action a keybinding names.
func (a *App) act(id string) {
	switch id {
	case action.TabNew:
		a.NewTab("")
	case action.TabClose:
		if _, ok := a.panes.Front(); ok {
			a.CloseTab(a.panes.Active)
		}
	case action.TabNext:
		a.stepTab(1)
	case action.TabPrev:
		a.stepTab(-1)
	case action.TabDuplicate:
		a.DuplicateTab(a.panes.Active)
	case action.PaneSplitAuto:
		if tab, ok := a.panes.Front(); ok {
			a.splitPane(tab.Auto(a.content))
		}
	case action.PaneSplitRight:
		a.splitPane(layout.SideBySide)
	case action.PaneSplitDown:
		a.splitPane(layout.Stacked)
	case action.PaneFocusLeft:
		a.moveFocus(pane.Left)
	case action.PaneFocusRight:
		a.moveFocus(pane.Right)
	case action.PaneFocusUp:
		a.moveFocus(pane.Up)
	case action.PaneFocusDown:
		a.moveFocus(pane.Down)
	case action.PaneClose:
		a.closePane(a.panes.Active)
	case action.PaneCloseOthers:
		a.closeOtherPanes(a.panes.Active)
	case action.Settings:
		a.OpenSettings()
	case action.Palette:
		a.OpenPalette()
	case action.Find:
		a.OpenFind()
	default:
		if i, ok := action.GotoIndex(id); ok {
			a.FocusTab(i)
			return
		}
		if n, ok := action.ProfileIndex(id); ok {
			a.newProfileTab(n)
		}
	}
}

// newProfileTab opens a tab that runs the profile at index i of the ones the
// menus list, as Windows Terminal's Control+Shift+number does.
func (a *App) newProfileTab(i int) {
	shown := a.menuProfiles()
	if i < 0 || i >= len(shown) {
		return
	}
	a.NewTab(shown[i].ID)
}

// stepTab brings the tab delta places along the strip in front, wrapping
// around at either end.
func (a *App) stepTab(delta int) {
	n := len(a.panes.Tabs)
	if n == 0 {
		return
	}
	a.FocusTab((a.panes.Active + delta + n) % n)
}
