package sys

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteCrashLog(t *testing.T) {
	path, err := WriteCrashLog("boom", []byte("goroutine 1 [running]:"))
	if err != nil {
		t.Fatalf("WriteCrashLog: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })

	if !strings.HasPrefix(filepath.Base(path), "crash-") || !strings.HasSuffix(path, ".log") {
		t.Errorf("path = %q, want crash-*.log", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	for _, want := range []string{"boom", "goroutine 1 [running]:"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("%s does not hold %q: %s", path, want, data)
		}
	}
}

// TestWriteCrashLogKeepsBoth covers two crashes in the same second: the second
// takes a numbered name rather than overwriting the first, and both still name
// a log IsLogPath reads.
func TestWriteCrashLogKeepsBoth(t *testing.T) {
	first, err := WriteCrashLog("first", []byte("one"))
	if err != nil {
		t.Fatalf("WriteCrashLog: %v", err)
	}
	t.Cleanup(func() { os.Remove(first) })
	// The same second, without waiting for the clock: the name the first took
	// is taken, so the next write has to move on from it.
	taken, err := WriteCrashLog("second", []byte("two"))
	if err != nil {
		t.Fatalf("WriteCrashLog: %v", err)
	}
	t.Cleanup(func() { os.Remove(taken) })
	if taken == first {
		t.Fatal("the second crash overwrote the first")
	}
	if !IsLogPath(taken) {
		t.Errorf("IsLogPath(%q) = false, want the numbered name to still be a log", taken)
	}
	for path, want := range map[string]string{first: "first", taken: "second"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(data), want) {
			t.Errorf("%s does not hold %q: %s", path, want, data)
		}
	}
}

// TestWriteCrashLogGivesUp covers a second that is full of crashes: the write
// says it could not, rather than overwriting one of the eight it kept or
// naming a log IsLogPath would not read.
func TestWriteCrashLogGivesUp(t *testing.T) {
	dir, err := Dir()
	if err != nil {
		t.Skipf("Dir: %v", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// The limit is a second of the clock's, so the test asks for the same
	// second's names until the clock has not moved: a crash that lands after
	// midnight of the name would be a different second, and proves nothing.
	// Names the crash before took are left alone, so this test may not run
	// beside another crash log writer.
	for {
		name := "crash-" + time.Now().Format("20060102-150405")
		for i := 0; i < crashTries; i++ {
			path := filepath.Join(dir, name+crashSuffix(i)+".log")
			if err := os.WriteFile(path, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { os.Remove(path) })
		}
		if time.Now().Format("20060102-150405") != strings.TrimPrefix(name, "crash-") {
			// The second turned while the names were written: they belong to a
			// second a crash no longer asks for, and are cleaned up anyway.
			continue
		}
		p, err := WriteCrashLog("overflow", nil)
		if err == nil {
			t.Cleanup(func() { os.Remove(p) })
			t.Fatalf("WriteCrashLog took %q with every name of the second taken", p)
		}
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error = %v, want it to name the log it could not write", err)
		}
		return
	}
}
