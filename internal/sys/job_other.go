//go:build !windows

package sys

// HoldChildren does nothing on a POSIX system, and returns no error on
// purpose: the window calls it once at startup, and a platform where nothing
// can be held is where nothing can fail.
//
// Windows keeps a job object and kills it when this process ends, however it
// ends, so a pane's shell goes with it. POSIX has no such object, and no call
// to make here would stand in for one: a shell ends when its terminal hangs
// up. mygo runs each pty as its own process group and mygo ends that group
// when it closes the terminal, which covers the window that closes a tab, the
// window that exits, and the shell that ends on its own. What it does not
// cover is this process dying without running any of it — a crash, or a signal
// that is not handled — and then a shell that ignores SIGHUP, or one that
// escaped its group with setsid, can outlive it. Closing that gap is not this
// function's to close: it would take a SysProcAttr per child (Pdeathsig on
// Linux, which the process that starts the pty, not this one, would have to
// set) and nothing of the sort on macOS, which is why the gap is written down
// here rather than papered over.
func HoldChildren() error { return nil }
