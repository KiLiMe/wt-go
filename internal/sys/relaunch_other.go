//go:build !windows

package sys

import (
	"fmt"
	"os"
	"os/exec"
)

// Relaunch starts this program again with the arguments it was given, and
// leaves it to outlive this one.
func Relaunch() error {
	path, err := Executable()
	if err != nil {
		return fmt.Errorf("sys: relaunch: %w", err)
	}
	cmd := exec.Command(path, os.Args[1:]...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("sys: relaunch: %w", err)
	}
	return nil
}
