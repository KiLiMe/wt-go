package app

import (
	"github.com/egoist/mygo/plugins/terminal"
	"log"
	"wtgo/internal/fault"
	"wtgo/internal/layout"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
)

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
		a.forget(term)
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
		a.forget(term)
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
	p := profile.Default(a.profiles, tab.ProfileID, a.device)
	term, err := a.newTerminal(a.termOptions(p, tab.Focus.Term.Dir()))
	if err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.ChildFailed, startFailed(p, err)))
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
