//go:build linux

package sys

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestSubtreeSeesAChild covers the child list the kernel keeps: the window
// reads the subtree of its own process to find the shell a terminal started,
// so a child it started itself has to be in there.
func TestSubtreeSeesAChild(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Skipf("cannot start a child: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()
	time.Sleep(20 * time.Millisecond)

	procs, err := Subtree(os.Getpid())
	if err != nil {
		t.Fatalf("Subtree: %v", err)
	}
	if len(procs) == 0 || procs[0].Pid != os.Getpid() {
		t.Fatalf("Subtree does not start with this process: %v", procs)
	}
	found := false
	for _, p := range procs {
		if p.Pid == cmd.Process.Pid {
			found = true
			if p.Parent != os.Getpid() {
				t.Errorf("child's parent = %d, want %d", p.Parent, os.Getpid())
			}
		}
	}
	if !found {
		t.Fatalf("Subtree does not hold the child %d: %v", cmd.Process.Pid, procs)
	}

	if _, err := Subtree(1 << 30); err == nil {
		t.Error("Subtree of a process that is not there reported no error")
	}
}
