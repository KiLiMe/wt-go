//go:build !windows

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
		data, err := os.ReadFile(filepath.Join("/proc", d.Name(), "stat"))
		if err != nil {
			continue
		}
		// The command name sits in parentheses and may hold spaces, so the
		// fields after it start past the last ')': the state, then the parent.
		text := string(data)
		i := strings.LastIndexByte(text, ')')
		if i < 0 {
			continue
		}
		name := ""
		if open := strings.IndexByte(text, '('); open >= 0 && open < i {
			name = text[open+1 : i]
		}
		fields := strings.Fields(text[i+1:])
		if len(fields) < 2 {
			continue
		}
		parent, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		out = append(out, Proc{Pid: pid, Parent: parent, Name: name})
	}
	return out, nil
}
