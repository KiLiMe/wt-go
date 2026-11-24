package app

import (
	"strings"
)

// Boundary: what the strip shows -- which tab is in front, where it stands,
// how far it is scrolled, and the name it wears. The tab itself is tabs.go, its
// drawing is internal/ui/tabs.go.

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
	a.revealTab()
	a.syncTitle()
	a.redraw()
}

// ScrollTabs moves the tabs the strip shows by by DIPs, as the wheel over the
// strip and the buttons that move it ask.
func (a *App) ScrollTabs(by float32) {
	a.tabsScroll.X = max(0, min(a.tabsScroll.X+by, a.tabsScroll.MaxX))
	a.redraw()
}

// revealFrames is how many frames a reveal lasts. The strip shows the buttons
// that scroll it only once it knows the tabs do not fit, and they take their
// room on the frame after that, so the tab in front can be out of view again
// when one frame ends.
const revealFrames = 2

// revealTab asks the strip to bring the tab in front into view, and to keep it
// in view while the strip settles; asking on every frame would keep the user
// from scrolling the strip away from the tab in front.
func (a *App) revealTab() { a.revealTabs = revealFrames }

// BeginRename opens the dialog that renames the tab at i, with the name the
// tab has in its field. Whatever other overlay was open goes: the dialog holds
// the keyboard.
func (a *App) BeginRename(i int) {
	if i < 0 || i >= len(a.panes.Tabs) {
		return
	}
	a.openOverlay(overlayRename)
	a.renaming.Tab = i
	a.renaming.Text = a.panes.Tabs[i].Name
	a.redraw()
}

// RenameTab gives the tab at i a name; "" gives it the name its terminal
// and its profile make.
func (a *App) RenameTab(i int, name string) {
	if i >= 0 && i < len(a.panes.Tabs) {
		a.panes.Tabs[i].Name = strings.TrimSpace(name)
	}
	a.closeOverlay(overlayRename)
	a.syncTitle()
	a.redraw()
}

// CancelRename closes the rename dialog.
func (a *App) CancelRename() {
	a.closeOverlay(overlayRename)
	a.redraw()
}
