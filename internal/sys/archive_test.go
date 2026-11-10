package sys

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestArchiveFile covers keeping a bad copy under a name that says what it is,
// and keeping a second copy when the name is taken.
func TestArchiveFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	now := time.Date(2026, 11, 10, 15, 4, 26, 0, time.Local)

	first, err := archiveFile(path, []byte("{oops"), now)
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	want := filepath.Join(dir, "settings.json.bad-20261110-150426")
	if first != want {
		t.Errorf("path = %q, want %q", first, want)
	}
	data, err := os.ReadFile(first)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "{oops" {
		t.Errorf("archive = %q, want the original", data)
	}

	second, err := archiveFile(path, []byte("{still oops"), now)
	if err != nil {
		t.Fatalf("second archive: %v", err)
	}
	if second == first {
		t.Fatal("the second archive replaced the first")
	}
	if _, err := os.Stat(second); err != nil {
		t.Errorf("second archive: %v", err)
	}
}
