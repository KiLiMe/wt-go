//go:build windows

package sys

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

// OpenPath opens path in the program the desktop uses for it, which Explorer
// is on Windows. The path goes in whole: explorer.exe reads a name that starts
// with a dash or a slash as an option, and an option is not the file.
func OpenPath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("sys: open %s: %w", path, err)
	}
	// The one in %SystemRoot% is the one this Windows ships; a bare name is
	// whatever the PATH decides, which is not something to trust with a click.
	explorer := "explorer.exe"
	if root := os.Getenv("SystemRoot"); root != "" {
		explorer = filepath.Join(root, "explorer.exe")
	}
	cmd := exec.Command(explorer, abs)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("sys: open %s: %w", path, err)
	}
	// Nothing waits for Explorer: releasing the process keeps the window from
	// holding it for the rest of the session.
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("sys: open %s: %w", path, err)
	}
	return nil
}

// OpenURL opens url in the program the desktop associates with its scheme.
// Explorer.exe would start one too, but it only reports that it started: a
// scheme with no program behind it, the usual reason a click in a terminal
// does nothing, would leave no trace anywhere. ShellExecuteW answers with a
// code that is 32 or less when it found no program.
func OpenURL(url string) error {
	verb, err := syscall.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	target, err := syscall.UTF16PtrFromString(url)
	if err != nil {
		return err
	}
	const swShownormal = 1
	r, _, _ := shellExecuteW.Call(0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(target)),
		0, 0, swShownormal)
	if r <= 32 {
		return fmt.Errorf("sys: open %s: error %d", url, r)
	}
	return nil
}

var shellExecuteW = syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteW")
