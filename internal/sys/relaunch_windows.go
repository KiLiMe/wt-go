//go:build windows

package sys

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
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
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("sys: relaunch: %w", err)
	}
	return nil
}
