//go:build !windows

package sys

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenPath opens path in the program the desktop uses for it: open on macOS,
// xdg-open elsewhere. A missing program is an error the settings page shows.
func OpenPath(path string) error {
	command := "xdg-open"
	if runtime.GOOS == "darwin" {
		command = "open"
	}
	program, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("sys: open %s: %w", path, err)
	}
	return exec.Command(program, path).Start()
}
