package app

import (
	"wtgo/internal/action"
	"wtgo/internal/layout"
	"wtgo/internal/pane"
)

// Boundary: one action id, done -- the switch that turns what a key binding
// names into what the window does. The actions it calls live in their own files.

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
	case action.About:
		a.OpenAbout()
	case action.Palette:
		a.OpenPalette()
	case action.Find:
		a.OpenFind()
	case action.Frames:
		a.showFrames = !a.showFrames
		// A readout that started with the frames of before would say nothing
		// about now: the meter begins again with the frames that follow.
		a.meter = frameMeter{}
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
