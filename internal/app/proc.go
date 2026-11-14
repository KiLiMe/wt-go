package app

import (
	"log"
	"time"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/fault"
	"wtgo/internal/pane"
	"wtgo/internal/sys"
)

// asksClose reports whether the settings ask before a tab that does act closes.
func (a *App) asksClose(act tabActivity) bool {
	return a.settings.ConfirmClose.Ask(act.act != sys.Idle)
}

// forget drops the processes of the terminals that close, so that the map of
// terminals to processes does not grow with every tab that ever opened.
func (a *App) forget(terms ...*terminal.Terminal) {
	for _, term := range terms {
		delete(a.shells, term)
	}
	gone := make(map[*terminal.Terminal]bool, len(terms))
	for _, term := range terms {
		gone[term] = true
	}
	a.spawnMu.Lock()
	defer a.spawnMu.Unlock()
	for pid, owner := range a.taken {
		if gone[owner] {
			delete(a.taken, pid)
		}
	}
}

// newTerminal starts the program a terminal runs and remembers its process:
// the plugin keeps the process to itself, so the window learns it from the one
// child that starts between the two lists.
func (a *App) newTerminal(opts terminal.Options) (*terminal.Terminal, error) {
	before, berr := sys.Subtree(sys.OwnPID())
	created, err := terminal.New(opts)
	if err != nil {
		return nil, err
	}
	term := created
	if berr == nil {
		// The plugin starts the program in a goroutine of its own, so the
		// child is not there the moment New returns: wait for it off the UI
		// thread, and hand the answer back when it comes.
		go func() {
			if shell, ok := a.spawned(term, before); ok {
				// The first write-back also tells a tab the user closed while the
				// search went on: the second one then finds the terminal gone.
				if !a.recordShell(term, shell) {
					return
				}
				// A launcher on PATH starts the program again, and the program
				// is the shell: the launcher under it would keep the tab busy
				// for as long as it is open, so the window moves on to the
				// child as soon as it shows.
				if real, ok := a.reexec(shell, term); ok {
					a.moveShell(term, shell.Pid, real)
				}
			}
		}()
	}
	return term, nil
}

// recordShell remembers the process a terminal started, and reports whether it
// was there to remember: a terminal that ended or left the strip while the
// search went on is given back, so that neither map keeps a terminal no tab
// shows. The write-back reaches the UI thread a moment later, and the tab may
// have closed in between: that is why it asks again there rather than trusting
// the look it took before.
func (a *App) recordShell(term *terminal.Terminal, shell sys.ShellProc) bool {
	if a.stale(term) {
		a.forget(term)
		return false
	}
	recorded := false
	a.Do(func() {
		if a.stale(term) {
			a.forget(term)
			return
		}
		a.shells[term] = shell
		recorded = true
	})
	return recorded
}

// moveShell moves a terminal on to the program its launcher started, and only
// from the launcher it knew: a terminal that closed or that another write-back
// already moved is left as it is.
func (a *App) moveShell(term *terminal.Terminal, from int, to sys.ShellProc) {
	if a.stale(term) {
		a.forget(term)
		return
	}
	a.Do(func() {
		if a.stale(term) {
			a.forget(term)
			return
		}
		if known, ok := a.shells[term]; ok && known.Pid == from {
			a.shells[term] = to
		}
	})
}

// stale reports whether a terminal is one the window no longer keeps: its
// program ended, or no pane of the strip holds it.
func (a *App) stale(term *terminal.Terminal) bool {
	select {
	case <-term.Done():
		return true
	default:
	}
	return !a.hasTerminal(term)
}

// hasTerminal reports whether a tab of the strip holds the terminal.
func (a *App) hasTerminal(term *terminal.Terminal) bool {
	for _, tab := range a.panes.Tabs {
		for _, t := range tab.Terminals() {
			if t == term {
				return true
			}
		}
	}
	return false
}

// spawned waits for the process a terminal started: the child of this process
// that was not there before, and that no other terminal took. A shell that
// reads its startup files takes a while to appear, so the window keeps looking
// for a few seconds, less and less often.
func (a *App) spawned(term *terminal.Terminal, before []sys.Proc) (sys.ShellProc, bool) {
	start := time.Now()
	deadline := start.Add(3 * time.Second)
	wait := 10 * time.Millisecond
	for time.Now().Before(deadline) {
		// A terminal that is already gone has no child coming, and the tab it
		// was in may have closed: there is nothing left to wait for.
		select {
		case <-term.Done():
			return sys.ShellProc{}, false
		default:
		}
		// The children of this process are the ones the kernel names for it, so
		// looking again costs a few small reads instead of a list of every
		// process the machine runs, which this would throw away each time.
		if after, err := sys.Subtree(sys.OwnPID()); err == nil {
			if p, ok := sys.Spawned(before, after, sys.OwnPID(), a.claimed()); ok && a.claim(p.Pid, term) {
				return sys.ShellProc{Pid: p.Pid, Name: p.Name}, true
			}
		}
		select {
		case <-term.Done():
			// The program ended, or could not start: no process is coming, so
			// the wait is over before its three seconds are.
			log.Printf("wt-go: %s", fault.Line(fault.UnknownShell, "the terminal ended before a child process appeared"))
			return sys.ShellProc{}, false
		case <-time.After(wait):
		}
		if wait < 100*time.Millisecond {
			wait *= 2
		}
	}
	log.Printf("wt-go: %s", fault.Line(fault.UnknownShell, "no child process appeared within "+time.Since(start).Round(time.Millisecond).String()))
	return sys.ShellProc{}, false
}

// shellSettle is how long the window watches a program it started for one that
// runs the program again. A launcher starts its program at once, so this is a
// moment rather than a wait, and a shell that is the program itself costs one
// look at the process list.
const shellSettle = 250 * time.Millisecond

// reexec follows a shell through the launcher that started it, and reports the
// program the tab really runs. It reports false when the program the window
// started is the shell itself.
func (a *App) reexec(shell sys.ShellProc, term *terminal.Terminal) (sys.ShellProc, bool) {
	deadline := time.Now().Add(shellSettle)
	wait := 10 * time.Millisecond
	cur := sys.Proc{Pid: shell.Pid, Name: shell.Name}
	moved := false
	for time.Now().Before(deadline) {
		// The terminal ended, or its tab closed: the window is not looking
		// for a program that no tab shows.
		if a.stale(term) {
			return shell, false
		}
		if all, err := sys.Subtree(cur.Pid); err == nil {
			if next, ok := sys.Reexec(all, cur); ok {
				if a.adopt(cur.Pid, next.Pid, term) {
					cur, moved = next, true
				}
			} else if moved {
				break
			}
		}
		select {
		case <-term.Done():
			return shell, false
		case <-time.After(wait):
		}
		if wait < 100*time.Millisecond {
			wait *= 2
		}
	}
	if !moved {
		return shell, false
	}
	return sys.ShellProc{Pid: cur.Pid, Name: cur.Name}, true
}

// claimed is the processes the window already gave to terminals.
func (a *App) claimed() map[int]bool {
	a.spawnMu.Lock()
	defer a.spawnMu.Unlock()
	out := make(map[int]bool, len(a.taken))
	for pid := range a.taken {
		out[pid] = true
	}
	return out
}

// claim takes a process for a terminal, and reports whether no other terminal
// had it.
func (a *App) claim(pid int, term *terminal.Terminal) bool {
	a.spawnMu.Lock()
	defer a.spawnMu.Unlock()
	if a.taken == nil {
		a.taken = make(map[int]*terminal.Terminal)
	}
	if _, ok := a.taken[pid]; ok {
		return false
	}
	a.taken[pid] = term
	return true
}

// adopt adds the process a terminal's shell started to what that terminal
// took, and reports false when the terminal no longer holds the one it came
// from, as when the tab closed while the window was looking. The launcher
// stays taken too: it belongs to the same tab, and leaving it free would let
// the search of a tab that starts at the same time take it for its own shell.
func (a *App) adopt(from, to int, term *terminal.Terminal) bool {
	a.spawnMu.Lock()
	defer a.spawnMu.Unlock()
	if a.taken[from] != term {
		return false
	}
	a.taken[to] = term
	return true
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
