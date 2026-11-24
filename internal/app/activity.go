package app

import (
	"log"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/fault"
	"wtgo/internal/pane"
	"wtgo/internal/sys"
)

// Boundary: what a tab is doing, from one snapshot of the process list -- the
// question the window asks before it closes anything.

// asksClose reports whether the settings ask before a tab that does act closes.
func (a *App) asksClose(act tabActivity) bool {
	return a.settings.ConfirmClose.Ask(act.act != sys.Idle)
}

// tabActivity is what a tab does: whether it runs something, and what that is,
// so that the question the window asks can name it.
type tabActivity struct {
	act      sys.Activity
	programs []string
}

// processList is the machine's process list as one question sees it: the list,
// and whether it could be read at all. A question that covers several tabs reads
// it once, because a scan of the process table costs milliseconds per call.
type processList struct {
	procs []sys.Proc
	known bool
}

// processList reads the machine's processes, and says so in the log when it
// cannot: the answer to a question about a tab is then that the window cannot
// speak for it.
func (a *App) processList() processList {
	procs, err := sys.Processes()
	if err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.NoProcessList, err.Error()))
		return processList{}
	}
	return processList{procs: procs, known: true}
}

// activity is what a tab does, as the processes of the shells it runs say. A
// shell that ended runs nothing, and neither does the console the pty runs
// behind. A shell the window did not learn is one it cannot speak for, and
// saying so beats guessing either way.
func (a *App) activity(tab *pane.Tab, list processList) tabActivity {
	var live []*terminal.Terminal
	for _, term := range tab.Terminals() {
		select {
		case <-term.Done():
		default:
			live = append(live, term)
		}
	}
	if len(live) == 0 {
		return tabActivity{}
	}
	if !list.known {
		return tabActivity{act: sys.Unknown}
	}
	out := tabActivity{}
	for _, term := range live {
		shell, ok := a.shells[term]
		if !ok {
			log.Printf("wt-go: %s", fault.Line(fault.UnknownShell, "a tab has no shell process"))
			out.act = sys.Unknown
			continue
		}
		switch shell.Activity(list.procs) {
		case sys.Busy:
			out.act = sys.Busy
			out.programs = append(out.programs, shell.Programs(list.procs)...)
		case sys.Unknown:
			if out.act == sys.Idle {
				out.act = sys.Unknown
			}
		}
	}
	if out.act != sys.Busy {
		out.programs = nil
	}
	return out
}
