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
	cmd := exec.Command(program, path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("sys: open %s: %w", path, err)
	}
	// Nothing waits for the desktop's program: releasing the process keeps the
	// one it leaves behind from becoming a zombie the window holds for the
	// rest of the session.
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("sys: open %s: %w", path, err)
	}
	return nil
}
