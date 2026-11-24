package app

import (
	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/sys"
)

// Boundary: the terminals the window keeps -- making one, remembering the
// process it runs, and dropping both when it closes. Following a launcher on to
// the shell it started is adopt.go; asking what a tab is doing is activity.go.

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
	// A printed path opens from where the shell says it is now, so the
	// directory is asked for at the click, when term is no longer nil.
	var term *terminal.Terminal
	opts.OpenLink = func(link string) bool { return a.openLink(link, linkDir(term, opts.Dir)) }
	before, berr := sys.Subtree(sys.OwnPID())
	term, err := terminal.New(opts)
	if err != nil {
		return nil, err
	}
	// A tab opened while the setting says no links starts with them off, so
	// that its first frame underlines nothing.
	setLinks(term, a.settings.Links)
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
