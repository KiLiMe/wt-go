package sys

import "strings"

// Proc is a process the machine runs: its id, the id of the process that
// started it, its name, and what it does.
type Proc struct {
	Pid    int
	Parent int
	Name   string
	// State is what the system says the process does: 'R' running, 'S'
	// sleeping, 'Z' a zombie that has ended and waits to be reaped. It is 0
	// where the system does not say, as on Windows, which lists no zombie.
	State byte
}

// ShellProc is the program a terminal runs, as the window learned it: the id
// the process list gave it, and the name it had, so that a number the system
// hands out again is not taken for it.
type ShellProc struct {
	Pid  int
	Name string
}

// Activity is what the processes under a shell say about the tab that runs it.
type Activity int

const (
	// Idle is a shell at its prompt: it runs, and nothing runs under it.
	Idle Activity = iota
	// Busy is a shell that runs something else: a command in front, a job in
	// the background, a helper a startup file left behind.
	Busy
	// Unknown is a shell the window cannot read the processes of. Asking is
	// the only safe answer, and the question does not claim to know.
	Unknown
)

// String makes an activity print as itself, for the tests and the log.
func (a Activity) String() string {
	switch a {
	case Idle:
		return "idle"
	case Busy:
		return "busy"
	}
	return "unknown"
}

// Activity reports what the processes say a shell does: Busy while something
// other than the shell runs under it, Idle while only the shell does, and
// Unknown when the window does not know which process the shell is.
//
// A shell that ended, and one whose id the system handed to another program,
// have nothing under them, so they are Idle. Neither is a process that ended
// and waits to be reaped, or the console the pty runs behind: they run nothing,
// so they leave a shell at its prompt Idle.
func (s ShellProc) Activity(all []Proc) Activity {
	if s.Pid <= 0 {
		return Unknown
	}
	self, ok := findProc(all, s.Pid)
	if !ok {
		return Idle
	}
	if s.Name != "" && !sameProgram(self.Name, s.Name) {
		return Idle
	}
	if len(s.Programs(all)) > 0 {
		return Busy
	}
	return Idle
}

// Programs names what a shell runs besides itself: the programs under it that
// keep a tab from closing quietly, each once, in the order the tree lists them.
// It is what the window tells the user when it asks, so that the question says
// which program it means.
func (s ShellProc) Programs(all []Proc) []string {
	if s.Pid <= 0 {
		return nil
	}
	var out []string
	seen := make(map[string]bool)
	for _, pid := range Descendants(all, s.Pid) {
		p, ok := findProc(all, pid)
		if !ok {
			continue
		}
		if p.State == 'Z' || ptyHost(p.Name) {
			continue
		}
		name := programName(p.Name)
		if seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

// findProc is the process with an id, when the list holds it.
func findProc(all []Proc, pid int) (Proc, bool) {
	for _, p := range all {
		if p.Pid == pid {
			return p, true
		}
	}
	return Proc{}, false
}

// sameProgram reports whether two names name the same program, which the
// systems write with and without ".exe", in either case.
func sameProgram(a, b string) bool {
	return strings.EqualFold(programName(a), programName(b))
}

// Children are the processes parent started.
func Children(all []Proc, parent int) []int {
	var out []int
	for _, p := range all {
		if p.Parent == parent {
			out = append(out, p.Pid)
		}
	}
	return out
}

// Descendants are the processes under root: its children, their children, and
// so on down. It is how the window tells a shell at its prompt, which has none,
// from a shell that runs a program.
func Descendants(all []Proc, root int) []int {
	kids := make(map[int][]int, len(all))
	for _, p := range all {
		kids[p.Parent] = append(kids[p.Parent], p.Pid)
	}
	seen := map[int]bool{root: true}
	queue := append([]int(nil), kids[root]...)
	var out []int
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]
		if seen[pid] {
			continue
		}
		seen[pid] = true
		out = append(out, pid)
		queue = append(queue, kids[pid]...)
	}
	return out
}

// Spawned is the program a terminal begins with: a child of parent that
// started between the two lists, and that taken does not name. The pty the
// terminal reads was made by a host process of its own, which starts beside the
// program and is not it, so the host is passed over.
//
// Two terminals may start their programs so close together that each list holds
// both: the older one is then the one that looked first, and the one that looks
// second is told which programs the window already gave away. It reports false
// when no program started.
func Spawned(before, after []Proc, parent int, taken map[int]bool) (Proc, bool) {
	old := make(map[int]bool, len(before))
	for _, p := range before {
		old[p.Pid] = true
	}
	var oldest Proc
	for _, p := range after {
		if old[p.Pid] || p.Parent != parent || ptyHost(p.Name) || taken[p.Pid] {
			continue
		}
		if oldest.Pid == 0 || p.Pid < oldest.Pid {
			oldest = p
		}
	}
	return oldest, oldest.Pid != 0
}

// Reexec is the process p runs when p is a launcher that starts the program
// again: the shims a package manager puts on PATH do that, and the program the
// user sees is then the child, not the shim. It is the one child of p that has
// the name of p, and false when p has none or several, so that a shell that
// runs a second copy of itself for a command is not taken for a launcher.
func Reexec(all []Proc, p Proc) (Proc, bool) {
	var only Proc
	found := 0
	for _, c := range all {
		if c.Parent != p.Pid || !sameProgram(c.Name, p.Name) {
			continue
		}
		only = c
		found++
	}
	return only, found == 1
}

// ptyHost reports whether a process is the console the pty runs behind rather
// than the program the terminal shows.
func ptyHost(name string) bool {
	switch strings.ToLower(programName(name)) {
	case "conhost", "openconsole":
		return true
	}
	return false
}

// programName is a process name without the extension Windows gives programs.
// The systems do not agree on the case of it: a name may come as PING.EXE.
func programName(name string) string {
	if n := len(name); n > 4 && strings.EqualFold(name[n-4:], ".exe") {
		return name[:n-4]
	}
	return name
}
