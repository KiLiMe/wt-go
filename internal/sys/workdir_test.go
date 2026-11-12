package sys

import (
	"os"
	"path/filepath"
	"testing"
)

// WorkDir must report the directory the process runs in, not the home
// directory: that fallback is the whole point.
func TestWorkDirFollowsChdir(t *testing.T) {
	want := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(want); err != nil {
		t.Fatalf("Chdir(%q): %v", want, err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	got := WorkDir()
	if got == "" {
		t.Fatalf("WorkDir is empty after Chdir to %q", want)
	}
	// macOS spells t.TempDir() as /var/... while the kernel reports
	// /private/var/..., so compare where both sides resolve to.
	if resolved(t, got) != resolved(t, want) {
		t.Fatalf("WorkDir = %q, want %q", got, want)
	}
}

// StartDir answers the two words the settings may write and hands every path
// on untouched: a drive letter or a leading ./ must never be read as a word.
func TestStartDir(t *testing.T) {
	want := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(want); err != nil {
		t.Fatalf("Chdir(%q): %v", want, err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	if got := StartDir("current"); resolved(t, got) != resolved(t, want) {
		t.Fatalf("StartDir(%q) = %q, want %q", "current", got, want)
	}

	home, err := os.UserHomeDir()
	if err == nil {
		if got := StartDir("home"); got != home {
			t.Fatalf("StartDir(%q) = %q, want %q", "home", got, home)
		}
	}

	// Nothing written is home on every platform: what a launcher starts a
	// terminal with says nothing about the person (the exe's own folder from
	// Explorer, %SystemRoot%\System32 from the Start menu, a desktop
	// session's cwd), so the cwd is only read when "current" asks for it.
	// No home directory to give is the one case that falls back to the cwd.
	wantEmpty := want
	if home != "" {
		wantEmpty = home
	}
	if got := StartDir(""); resolved(t, got) != resolved(t, wantEmpty) {
		t.Fatalf("StartDir(%q) = %q, want %q", "", got, wantEmpty)
	}

	for _, path := range []string{`C:\code`, "/opt/src", "./current", "sub", `~/x`, "%USERPROFILE%"} {
		if got := StartDir(path); got != path {
			t.Fatalf("StartDir(%q) = %q, want it unchanged", path, got)
		}
	}
}

func resolved(t *testing.T, dir string) string {
	t.Helper()
	out, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", dir, err)
	}
	return out
}
