package app

import (
	"log"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/action"
	"wtgo/internal/fault"
	"wtgo/internal/layout"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/sys"
)

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

// asksClose reports whether the settings ask before a tab that does act closes.

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

// CloseOtherTabs closes every tab but the one at i, asking once when it would
// end programs that still run.
func (a *App) CloseOtherTabs(i int) {
	if i < 0 || i >= len(a.panes.Tabs) {
		return
	}
	tab := a.panes.Tabs[i]
	if act := a.othersActivity(tab); a.asksClose(act) {
		a.askCloseOtherTabs(tab, act)
		return
	}
	a.closeOtherTabs(tab)
}

// othersActivity is what the tabs but keep do, one answer for the question
// that closes them all: a program that still runs outranks a tab the window
// cannot speak for.
func (a *App) othersActivity(keep *pane.Tab) tabActivity {
	out := tabActivity{}
	seen := make(map[string]bool)
	// One list for every tab the question covers: a scan per tab would read the
	// whole process table once per tab, for one answer.
	list := a.processList()
	for _, tab := range a.panes.Tabs {
		if tab == keep {
			continue
		}
		act := a.activity(tab, list)
		if act.act == sys.Busy {
			out.act = sys.Busy
			for _, name := range act.programs {
				if !seen[name] {
					seen[name] = true
					out.programs = append(out.programs, name)
				}
			}
			continue
		}
		if act.act == sys.Unknown && out.act == sys.Idle {
			out.act = sys.Unknown
		}
	}
	if out.act != sys.Busy {
		out.programs = nil
	}
	return out
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
		a.forget(tab.Terminals()...)
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
