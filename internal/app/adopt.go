package app

import (
	"log"
	"time"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/fault"
	"wtgo/internal/sys"
)

// Boundary: a terminal's shell as the machine started it -- the child to wait
// for, the re-exec that puts a launcher behind it, and the process each terminal
// claims so that two terminals never claim one.

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
