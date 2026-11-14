//go:build windows

package sys

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// createBreakawayFromJob lets the process this one starts leave the job this
// one holds, so that closing this window, which ends the job, leaves the new
// window running.
const createBreakawayFromJob = 0x01000000

// Relaunch starts this program again with the arguments it was given, and
// leaves it to outlive this one: the job this process holds kills what stays
// in it, so the new process breaks away from it.
func Relaunch() error {
	path, err := Executable()
	if err != nil {
		return fmt.Errorf("sys: relaunch: %w", err)
	}
	cmd := exec.Command(path, os.Args[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createBreakawayFromJob}
	err = cmd.Start()
	if err == nil {
		return nil
	}
	// Windows refuses CREATE_BREAKAWAY_FROM_JOB when the job this process holds
	// does not allow it (a nested job, as under a taskbar or a terminal that
	// already runs in one). A window that does not start at all is worse than
	// one that the job takes with it, so the retry drops the flag; the window
	// comes back, and the next launch may get the breakaway it asked for.
	if !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return fmt.Errorf("sys: relaunch: %w", err)
	}
	cmd = exec.Command(path, os.Args[1:]...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("sys: relaunch: %w", err)
	}
	return nil
}
