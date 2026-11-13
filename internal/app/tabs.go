package app

import (
	"log"
	"strings"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
	"wtgo/internal/fault"
	"wtgo/internal/layout"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/sys"
	"wtgo/internal/ui"
)

// NewTab opens a tab that runs a profile; "" runs the default profile, which
// is what the strip's '+', the keybinding for a new tab and the button of a
// window without tabs open.
func (a *App) NewTab(profileID string) {
	a.openTab(a.tabProfile(profileID), "")
}

// NewSystemTab opens a tab that runs the shell of the machine: what the
// settings call following the system, and what every tab runs when the file
// offers no profile at all.
func (a *App) NewSystemTab() {
	a.openTab(profile.Profile{}, "")
}

// tabProfile is the profile a tab runs: the one it names, and with none the
// default profile. An id the machine does not have, or one since hidden, is the
// system's own shell rather than another profile of the list.
func (a *App) tabProfile(profileID string) profile.Profile {
	if profileID == "" {
		profileID = a.settings.DefaultProfile
	}
	return profile.Default(a.profiles, profileID)
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
		a.closeWindow()
		closeTerminals(tab)
		a.forget(tab.Terminals()...)
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

// closeTerminals closes what a tab holds.
func closeTerminals(tab *pane.Tab) {
	for _, term := range tab.Terminals() {
		term.Close()
	}
}

// newTerminal starts the program a terminal runs and remembers its process:
// the plugin keeps the process to itself, so the window learns it from the one
// child that starts between the two lists.
func (a *App) newTerminal(opts terminal.Options) (*terminal.Terminal, error) {
	before, berr := sys.Subtree(sys.OwnPID())
	// The keys a program that asks for the Kitty keyboard protocol never gets
	// are written here instead: see legacyBytes.
	var term *terminal.Terminal
	opts.OnKey = func(mods myui.Modifiers, key myui.Key) bool { return legacyKey(term, mods, key) }
	created, err := terminal.New(opts)
	if err != nil {
		return nil, err
	}
	term = created
	if berr == nil {
		// The plugin starts the program in a goroutine of its own, so the
		// child is not there the moment New returns: wait for it off the UI
		// thread, and hand the answer back when it comes.
		go func() {
			if shell, ok := a.spawned(term, before); ok {
				a.Do(func() { a.shells[term] = shell })
				// A launcher on PATH starts the program again, and the program
				// is the shell: the launcher under it would keep the tab busy
				// for as long as it is open, so the window moves on to the
				// child as soon as it shows.
				if real, ok := a.reexec(shell, term); ok {
					a.Do(func() {
						if known, ok := a.shells[term]; ok && known.Pid == shell.Pid {
							a.shells[term] = real
						}
					})
				}
			}
		}()
	}
	return term, nil
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
	p := profile.Default(a.profiles, tab.ProfileID)
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
	p := profile.Default(a.profiles, tab.ProfileID)
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
