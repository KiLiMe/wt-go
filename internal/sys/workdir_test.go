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

func resolved(t *testing.T, dir string) string {
	t.Helper()
	out, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", dir, err)
	}
	return out
}
