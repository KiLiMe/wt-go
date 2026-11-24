package app

import (
	"wtgo/internal/pane"
	"wtgo/internal/sys"
)

// Boundary: closing every tab but one -- the single question the window asks
// about them, and what it does with the answer.

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
