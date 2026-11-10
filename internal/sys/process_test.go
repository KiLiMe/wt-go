package sys

import "testing"

// tree is the machine a process test walks: 100 started 200 and 400, 200
// started 300.
var tree = []Proc{{Pid: 100, Parent: 1}, {Pid: 200, Parent: 100}, {Pid: 300, Parent: 200}, {Pid: 400, Parent: 100}}

// TestChildren covers the processes one process started.
func TestChildren(t *testing.T) {
	got := Children(tree, 100)
	if len(got) != 2 || got[0] != 200 || got[1] != 400 {
		t.Errorf("Children = %v, want [200 400]", got)
	}
	if got := Children(tree, 300); len(got) != 0 {
		t.Errorf("Children of a leaf = %v, want none", got)
	}
}

// TestDescendants covers the processes under one, which is how the window tells
// a shell at its prompt from one that runs a program.
func TestDescendants(t *testing.T) {
	got := Descendants(tree, 100)
	if len(got) != 3 {
		t.Fatalf("Descendants = %v, want three", got)
	}
	seen := map[int]bool{}
	for _, pid := range got {
		seen[pid] = true
	}
	for _, want := range []int{200, 300, 400} {
		if !seen[want] {
			t.Errorf("Descendants = %v, missing %d", got, want)
		}
	}
	if got := Descendants(tree, 300); len(got) != 0 {
		t.Errorf("Descendants of a leaf = %v, want none", got)
	}
}

// TestDescendantsCycle covers a parent list that loops, which a reused process
// id can make: the walk ends instead of spinning.
func TestDescendantsCycle(t *testing.T) {
	loop := []Proc{{Pid: 2, Parent: 1}, {Pid: 1, Parent: 2}}
	if got := Descendants(loop, 1); len(got) != 1 || got[0] != 2 {
		t.Errorf("Descendants = %v, want [2]", got)
	}
}

// TestSpawned covers the process a terminal started: the one that appeared
// between the two lists as a child of this one.
func TestSpawned(t *testing.T) {
	before := []Proc{{Pid: 100, Parent: 1}, {Pid: 200, Parent: 100}}
	after := []Proc{{Pid: 100, Parent: 1}, {Pid: 200, Parent: 100}, {Pid: 300, Parent: 100}, {Pid: 400, Parent: 1}}
	if pid, ok := Spawned(before, after, 100); !ok || pid != 300 {
		t.Errorf("Spawned = %d, %v, want 300", pid, ok)
	}
	if pid, ok := Spawned(before, after, 400); ok {
		t.Errorf("Spawned under a process with no new child = %d, want none", pid)
	}
	if pid, ok := Spawned(nil, nil, 100); ok {
		t.Errorf("Spawned of nothing = %d, want none", pid)
	}
	// The console the pty runs behind starts beside the program, and is not
	// the program even when it takes the newer id.
	before = []Proc{{Pid: 100, Parent: 1, Name: "wtgo.exe"}}
	after = []Proc{
		{Pid: 100, Parent: 1, Name: "wtgo.exe"},
		{Pid: 300, Parent: 100, Name: "pwsh.exe"},
		{Pid: 500, Parent: 100, Name: "conhost.exe"},
	}
	if pid, ok := Spawned(before, after, 100); !ok || pid != 300 {
		t.Errorf("Spawned past a pty host = %d, %v, want 300", pid, ok)
	}
}
