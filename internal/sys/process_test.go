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

// TestSubtreeOf covers the root and its whole tree, which is what the window
// asks of a machine that lists every process at once.
func TestSubtreeOf(t *testing.T) {
	got := SubtreeOf(tree, 100)
	if len(got) != 4 || got[0].Pid != 100 {
		t.Fatalf("SubtreeOf = %v, want 100 and its three", got)
	}
	seen := map[int]bool{}
	for _, p := range got {
		seen[p.Pid] = true
	}
	for _, want := range []int{100, 200, 300, 400} {
		if !seen[want] {
			t.Errorf("SubtreeOf misses %d", want)
		}
	}
	if got := SubtreeOf(tree, 999); got != nil {
		t.Errorf("SubtreeOf of a process not in the list = %v, want nil", got)
	}
}

// TestSpawned covers the process a terminal started: the one that appeared
// between the two lists as a child of this one.
func TestSpawned(t *testing.T) {
	before := []Proc{{Pid: 100, Parent: 1}, {Pid: 200, Parent: 100}}
	after := []Proc{{Pid: 100, Parent: 1}, {Pid: 200, Parent: 100}, {Pid: 300, Parent: 100}, {Pid: 400, Parent: 1}}
	if p, ok := Spawned(before, after, 100, nil); !ok || p.Pid != 300 {
		t.Errorf("Spawned = %d, %v, want 300", p.Pid, ok)
	}
	if p, ok := Spawned(before, after, 400, nil); ok {
		t.Errorf("Spawned under a process with no new child = %d, want none", p.Pid)
	}
	if p, ok := Spawned(nil, nil, 100, nil); ok {
		t.Errorf("Spawned of nothing = %d, want none", p.Pid)
	}
	// The console the pty runs behind starts beside the program, and is not
	// the program even when it takes the newer id.
	before = []Proc{{Pid: 100, Parent: 1, Name: "wtgo.exe"}}
	after = []Proc{
		{Pid: 100, Parent: 1, Name: "wtgo.exe"},
		{Pid: 300, Parent: 100, Name: "pwsh.exe"},
		{Pid: 500, Parent: 100, Name: "conhost.exe"},
	}
	if p, ok := Spawned(before, after, 100, nil); !ok || p.Pid != 300 || p.Name != "pwsh.exe" {
		t.Errorf("Spawned past a pty host = %d, %v, want 300", p.Pid, ok)
	}
}

// TestSpawnedTogether covers two terminals that started their programs so
// close together that each list holds both: the older program is the one that
// looked first, and the one that looks second passes over what it took.
func TestSpawnedTogether(t *testing.T) {
	before := []Proc{{Pid: 100, Parent: 1}}
	after := []Proc{{Pid: 100, Parent: 1}, {Pid: 300, Parent: 100}, {Pid: 320, Parent: 100}}
	first, ok := Spawned(before, after, 100, nil)
	if !ok || first.Pid != 300 {
		t.Fatalf("the terminal that looked first took %d, want 300", first.Pid)
	}
	second, ok := Spawned(before, after, 100, map[int]bool{first.Pid: true})
	if !ok || second.Pid != 320 {
		t.Errorf("the terminal that looked second took %d, want 320", second.Pid)
	}
}

// TestActivity covers what the processes say a tab does, which is what the
// question before closing it asks about.
func TestActivity(t *testing.T) {
	pwsh := []Proc{{Pid: 100, Parent: 1, Name: "pwsh.exe"}, {Pid: 300, Parent: 100, Name: "pwsh.exe"}}
	idle := ShellProc{Pid: 300, Name: "pwsh.exe"}
	if got := idle.Activity(pwsh); got != Idle {
		t.Errorf("a shell at its prompt = %v, want idle", got)
	}
	busy := append(pwsh, Proc{Pid: 400, Parent: 300, Name: "git.exe"})
	if got := idle.Activity(busy); got != Busy {
		t.Errorf("a shell that runs a program = %v, want busy", got)
	}
	// A program of the background, under a job of the shell, still runs.
	deep := append(busy, Proc{Pid: 500, Parent: 400, Name: "ssh.exe"})
	if got := idle.Activity(deep); got != Busy {
		t.Errorf("a shell with a background job = %v, want busy", got)
	}
	// What ended and waits to be reaped runs nothing, so it does not make the
	// shell busy.
	zombie := append(pwsh, Proc{Pid: 400, Parent: 300, Name: "git.exe", State: 'Z'})
	if got := idle.Activity(zombie); got != Idle {
		t.Errorf("a shell with a zombie under it = %v, want idle", got)
	}
	// The console the pty runs behind is no program of the tab either.
	host := append(pwsh, Proc{Pid: 400, Parent: 300, Name: "OpenConsole.exe"})
	if got := idle.Activity(host); got != Idle {
		t.Errorf("a shell with a pty host under it = %v, want idle", got)
	}
	// Neither is the shell started again under its own name, which is what the
	// shim a package manager puts on PATH leaves under the shell.
	shimmed := append(pwsh, Proc{Pid: 400, Parent: 300, Name: "pwsh.exe"})
	if got := idle.Activity(shimmed); got != Idle {
		t.Errorf("a shell started again under it = %v, want idle", got)
	}
	// What such a shell runs is still a program of the tab.
	under := append(shimmed, Proc{Pid: 500, Parent: 400, Name: "pi.exe"})
	if got := idle.Activity(under); got != Busy {
		t.Errorf("a shell started again that runs a program = %v, want busy", got)
	}
	// A shell that ended, and one whose id went to another program, run
	// nothing.
	gone := ShellProc{Pid: 900, Name: "pwsh.exe"}
	if got := gone.Activity(pwsh); got != Idle {
		t.Errorf("a shell that ended = %v, want idle", got)
	}
	reused := ShellProc{Pid: 300, Name: "bash"}
	if got := reused.Activity(busy); got != Idle {
		t.Errorf("a shell whose id was reused = %v, want idle", got)
	}
	// A shell the window did not learn is one it cannot speak for.
	if got := (ShellProc{}).Activity(pwsh); got != Unknown {
		t.Errorf("a shell with no process = %v, want unknown", got)
	}
}

// TestReexec covers the launcher a package manager puts on PATH: the program
// the window started runs the program again, and the child is the shell.
func TestReexec(t *testing.T) {
	shim := []Proc{
		{Pid: 100, Parent: 1, Name: "wtgo.exe"},
		{Pid: 300, Parent: 100, Name: "pwsh.exe"},
		{Pid: 320, Parent: 300, Name: "pwsh.exe"},
	}
	if p, ok := Reexec(shim, Proc{Pid: 300, Name: "pwsh.exe"}); !ok || p.Pid != 320 {
		t.Errorf("Reexec = %d, %v, want 320", p.Pid, ok)
	}
	// The shell the launcher started is the program itself.
	if p, ok := Reexec(shim, Proc{Pid: 320, Name: "pwsh.exe"}); ok {
		t.Errorf("Reexec of the program = %d, want none", p.Pid)
	}
	// A child of another name is a program the shell runs, not the shell.
	if p, ok := Reexec(shim, Proc{Pid: 100, Name: "wtgo.exe"}); ok {
		t.Errorf("Reexec of a launcher with another child = %d, want none", p.Pid)
	}
	// A shell that runs two copies of itself may be doing work in them.
	twice := append(shim, Proc{Pid: 340, Parent: 300, Name: "pwsh.exe"})
	if p, ok := Reexec(twice, Proc{Pid: 300, Name: "pwsh.exe"}); ok {
		t.Errorf("Reexec with two children = %d, want none", p.Pid)
	}
}

// TestPrograms covers what a shell says it runs: the programs under it that
// keep the tab open, each once, without the extension Windows adds.
func TestPrograms(t *testing.T) {
	all := []Proc{
		{Pid: 100, Parent: 1, Name: "pwsh.exe"},
		{Pid: 300, Parent: 100, Name: "pwsh.exe"},
		{Pid: 400, Parent: 300, Name: "git.exe"},
		{Pid: 410, Parent: 300, Name: "git.exe"},
		{Pid: 500, Parent: 400, Name: "ssh.exe"},
		{Pid: 600, Parent: 300, Name: "conhost.exe"},
		{Pid: 700, Parent: 300, Name: "sleep.exe", State: 'Z'},
	}
	got := (ShellProc{Pid: 300, Name: "pwsh.exe"}).Programs(all)
	want := []string{"git", "ssh"}
	if len(got) != len(want) {
		t.Fatalf("a shell that runs git and ssh names %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("a shell that runs git and ssh names %v, want %v", got, want)
		}
	}
	if got := (ShellProc{Pid: 300, Name: "pwsh.exe"}).Programs(all[:2]); len(got) != 0 {
		t.Errorf("a shell at its prompt names %v, want none", got)
	}
	// The shell started again under its own name is not named, and what it runs
	// is: a tab whose shell is a shim still says what the program under it does.
	again := append(all[:2], Proc{Pid: 800, Parent: 300, Name: "pwsh.exe"}, Proc{Pid: 900, Parent: 800, Name: "pi.exe"})
	if got := (ShellProc{Pid: 300, Name: "pwsh.exe"}).Programs(again); len(got) != 1 || got[0] != "pi" {
		t.Errorf("a shell started again that runs pi names %v, want [pi]", got)
	}
}
