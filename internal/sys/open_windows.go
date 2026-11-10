//go:build windows

package sys

import "os/exec"

// OpenPath opens path in the program the desktop uses for it, which Explorer
// is on Windows.
func OpenPath(path string) error {
	return exec.Command("explorer.exe", path).Start()
}
