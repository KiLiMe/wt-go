package sys

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNativeLibraryEnv covers the path $MYGO_GHOSTTY_VT names: a file is the
// answer; a path with no file behind it is skipped for the copies the plugin
// would otherwise load, since reporting it as found would hide E102.
func TestNativeLibraryEnv(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, nativeName())
	if err := os.WriteFile(file, []byte("not a library"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MYGO_GHOSTTY_VT", file)
	p, err := NativeLibrary()
	if err != nil {
		t.Fatalf("NativeLibrary: %v", err)
	}
	if p != file {
		t.Errorf("NativeLibrary = %q, want %q", p, file)
	}

	// Nothing at the path: the answer is a copy the machine actually has, or
	// an error, never the missing file.
	gone := filepath.Join(dir, "gone", nativeName())
	t.Setenv("MYGO_GHOSTTY_VT", gone)
	p, err = NativeLibrary()
	if err == nil && p == gone {
		t.Errorf("NativeLibrary took %q, which is not a file", gone)
	}
}

// TestTail covers the end of a log the diagnostic reads: a file shorter than
// the limit comes back whole, and a longer one comes back from a line
// boundary, saying it was cut.
func TestTail(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}

	short := write("short.log", "one\ntwo\n")
	text, cut, err := Tail(short, 1<<10)
	if err != nil {
		t.Fatalf("Tail: %v", err)
	}
	if cut {
		t.Error("a file shorter than the limit was cut")
	}
	if text != "one\ntwo\n" {
		t.Errorf("Tail = %q, want the whole file", text)
	}

	// Exactly the limit: nothing was dropped, so nothing is cut.
	exact := write("exact.log", strings.Repeat("a", 64))
	text, cut, err = Tail(exact, 64)
	if err != nil {
		t.Fatalf("Tail: %v", err)
	}
	if cut {
		t.Error("a file exactly the size of the limit was cut")
	}
	if len(text) != 64 {
		t.Errorf("Tail = %d bytes, want the whole file", len(text))
	}

	long := write("long.log", "first line\nsecond line\nlast line\n")
	text, cut, err = Tail(long, 20)
	if err != nil {
		t.Fatalf("Tail: %v", err)
	}
	if !cut {
		t.Error("a file longer than the limit was not cut")
	}
	if !strings.HasSuffix(text, "last line\n") {
		t.Errorf("Tail = %q, want it to end with the file's end", text)
	}
	if strings.HasPrefix(text, "irst") || strings.Contains(text, "econd") {
		t.Errorf("Tail = %q, want it to start at a line boundary", text)
	}

	empty := write("empty.log", "")
	text, cut, err = Tail(empty, 64)
	if err != nil {
		t.Fatalf("Tail: %v", err)
	}
	if text != "" || cut {
		t.Errorf("Tail(empty) = %q, %v, want empty and not cut", text, cut)
	}

	if _, _, err := Tail(filepath.Join(dir, "gone.log"), 64); err == nil {
		t.Error("Tail read a file that is not there")
	}
}
