package sys

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
