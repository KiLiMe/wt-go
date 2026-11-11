//go:build !windows && !darwin

package sys

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Processes is every process the machine runs.
func Processes() ([]Proc, error) {
	dirs, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	out := make([]Proc, 0, len(dirs))
	for _, d := range dirs {
		pid, err := strconv.Atoi(d.Name())
		if err != nil {
			continue
		}
		p, err := readProc(pid)
		if err != nil {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

// Subtree is root and the processes under it, which the kernel lists itself:
// /proc/<pid>/task/<tid>/children names what each of a process's threads
// started, so a question about one shell costs a handful of small reads where
// a list of the machine costs one read per process it runs.
func Subtree(root int) ([]Proc, error) {
	self, err := readProc(root)
	if err != nil {
		return nil, err
	}
	out := []Proc{self}
	seen := map[int]bool{root: true}
	queue := []int{root}
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]
		for _, child := range childrenOf(pid) {
			if seen[child] {
				continue
			}
			seen[child] = true
			p, err := readProc(child)
			if err != nil {
				// The process ended between the kernel naming it and this read.
				continue
			}
			out = append(out, p)
			queue = append(queue, child)
		}
	}
	return out, nil
}

// childrenOf names the processes pid started, each once: the kernel writes
// them per thread, so every task the process runs is asked.
func childrenOf(pid int) []int {
	dir := filepath.Join("/proc", strconv.Itoa(pid), "task")
	tasks, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []int
	for _, t := range tasks {
		text, err := os.ReadFile(filepath.Join(dir, t.Name(), "children"))
		if err != nil {
			continue
		}
		for _, f := range strings.Fields(string(text)) {
			if n, err := strconv.Atoi(f); err == nil {
				out = append(out, n)
			}
		}
	}
	return out
}

// readProc reads what the kernel says about one process.
func readProc(pid int) (Proc, error) {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return Proc{}, err
	}
	// The command name sits in parentheses and may hold spaces, so the
	// fields after it start past the last ')': the state, then the parent.
	text := string(data)
	i := strings.LastIndexByte(text, ')')
	if i < 0 {
		return Proc{}, os.ErrInvalid
	}
	name := ""
	if open := strings.IndexByte(text, '('); open >= 0 && open < i {
		name = text[open+1 : i]
	}
	fields := strings.Fields(text[i+1:])
	if len(fields) < 2 {
		return Proc{}, os.ErrInvalid
	}
	parent, err := strconv.Atoi(fields[1])
	if err != nil {
		return Proc{}, err
	}
	var state byte
	if fields[0] != "" {
		state = fields[0][0]
	}
	return Proc{Pid: pid, Parent: parent, Name: name, State: state}, nil
}
