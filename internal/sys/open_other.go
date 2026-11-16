//go:build !windows

package sys

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

// OpenPath opens path in the program the desktop uses for it: open on macOS,
// xdg-open elsewhere. A missing program is an error the settings page shows.
func OpenPath(path string) error {
	// The path goes in whole: a name that starts with a dash is otherwise read
	// as an option of xdg-open or open, and an option is not the file.
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("sys: open %s: %w", path, err)
	}
	return launch(abs)
}

// OpenURL opens url in the program the desktop associates with its scheme,
// the same way OpenPath opens a file.
func OpenURL(url string) error {
	return launch(url)
}

// launch has the desktop open what: open on macOS, xdg-open elsewhere, and
// nothing waits for the program it starts.
func launch(what string) error {
	command := "xdg-open"
	if runtime.GOOS == "darwin" {
		command = "open"
	}
	program, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("sys: open %s: %w", what, err)
	}
	cmd := exec.Command(program, what)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("sys: open %s: %w", what, err)
	}
	// Nothing waits for the desktop's program: releasing the process keeps the
	// one it leaves behind from becoming a zombie the window holds for the
	// rest of the session.
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("sys: open %s: %w", what, err)
	}
	return nil
}
