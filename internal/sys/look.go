package sys

import (
	"os"
	"os/exec"
	"strings"
)

// LookCommand is where the machine keeps a command, and whether it has it. A
// profile that names a shell this machine does not have is a check of its own.
func LookCommand(name string) (string, bool) {
	if name == "" {
		return "", false
	}
	// A command with a path in it is looked up as it is, like exec.Command
	// would run it.
	if strings.ContainsAny(name, `\\/:`) {
		if st, err := os.Stat(name); err == nil && !st.IsDir() {
			return name, true
		}
		return "", false
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return "", false
	}
	return path, true
}
