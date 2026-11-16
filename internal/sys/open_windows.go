//go:build windows

package sys

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

// OpenPath opens path in the program the desktop uses for it, which Explorer
// is on Windows.
func OpenPath(path string) error {
	return exec.Command("explorer.exe", path).Start()
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
