package sys

import "strings"

// Proc is a process the machine runs: its id, the id of the process that
// started it, and its name.
type Proc struct {
	Pid    int
	Parent int
	Name   string
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
// started between the two lists. The pty the terminal reads was made by a host
// process of its own, which starts beside the program and is not it, so the
// host is passed over. It reports false when no program started, and takes the
// newest when the lists name more than one, which is rare enough that a wrong
// answer only makes the window ask before it closes.
func Spawned(before, after []Proc, parent int) (int, bool) {
	old := make(map[int]bool, len(before))
	for _, p := range before {
		old[p.Pid] = true
	}
	newest := 0
	for _, p := range after {
		if old[p.Pid] || p.Parent != parent || ptyHost(p.Name) {
			continue
		}
		if p.Pid > newest {
			newest = p.Pid
		}
	}
	return newest, newest != 0
}

// ptyHost reports whether a process is the console the pty runs behind rather
// than the program the terminal shows.
func ptyHost(name string) bool {
	switch strings.ToLower(strings.TrimSuffix(name, ".exe")) {
	case "conhost", "openconsole":
		return true
	}
	return false
}
