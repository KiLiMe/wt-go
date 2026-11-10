package sys

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteFileAtomic covers a write that replaces the file and leaves no
// temporary file behind.
func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "settings.json")
	if err := writeFileAtomic(path, []byte("one"), 0o644); err != nil {
		t.Fatalf("write one: %v", err)
	}
	if err := writeFileAtomic(path, []byte("two"), 0o644); err != nil {
		t.Fatalf("write two: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "two" {
		t.Errorf("file = %q, want two", data)
	}
	tmp, err := filepath.Glob(filepath.Join(dir, "nested", ".settings.json.tmp-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tmp) != 0 {
		t.Errorf("temporary files left: %v", tmp)
	}
}
