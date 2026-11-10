package sys

import (
	"os"
	"path/filepath"
	"runtime"
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

// TestWriteFileAtomicPerm keeps the permissions of an existing file and uses
// the given ones for a new file. Windows reports synthesised permissions, so
// the check only runs on the other platforms.
func TestWriteFileAtomicPerm(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not keep Unix permission bits")
	}
	dir := t.TempDir()

	existing := filepath.Join(dir, "existing.json")
	if err := os.WriteFile(existing, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(existing, []byte("new"), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}
	assertPerm(t, existing, 0o600)

	fresh := filepath.Join(dir, "fresh.json")
	if err := writeFileAtomic(fresh, []byte("new"), 0o600); err != nil {
		t.Fatalf("write fresh: %v", err)
	}
	assertPerm(t, fresh, 0o600)
}

// TestWriteFileAtomicSymlink writes through a symbolic link and leaves the
// link in place, pointing at the same target.
func TestWriteFileAtomicSymlink(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real.json")
	if err := os.WriteFile(real, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks are not available: %v", err)
	}
	if err := writeFileAtomic(link, []byte("new"), 0o644); err != nil {
		t.Fatalf("write through link: %v", err)
	}

	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("link was replaced by a regular file: mode %v", fi.Mode())
	}
	if got, err := os.Readlink(link); err != nil {
		t.Fatal(err)
	} else if got != real {
		t.Errorf("link points at %q, want %q", got, real)
	}
	data, err := os.ReadFile(real)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Errorf("target = %q, want new", data)
	}
}

// assertPerm checks the permission bits of path.
func assertPerm(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fi.Mode().Perm(); got != want {
		t.Errorf("%s = %v, want %v", filepath.Base(path), got, want)
	}
}
