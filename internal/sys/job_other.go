//go:build !windows

package sys

// HoldChildren does nothing on a POSIX system. Windows keeps a job object and
// kills it when this process ends, however it ends, so a pane's shell goes
// with it. POSIX has no such object: a shell ends when its terminal hangs up,
// and mygo ends it with the process group when it closes the terminal, but
// nothing guarantees that a killed process takes its children along. A shell
// that ignores SIGHUP, or one that escaped with setsid, can outlive this
// process; only an extension point for SysProcAttr (Pdeathsig) could close
// that gap.
func HoldChildren() error { return nil }
