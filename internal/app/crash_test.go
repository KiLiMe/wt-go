package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wtgo/internal/sys"
)

// reportCrash must leave a file that says what stopped the program, because the
// window it belonged to is gone.
func TestReportCrash(t *testing.T) {
	settingsDir(t)
	dir, err := sys.Dir()
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}
	reportCrash("wtgo test panic")

	paths, err := filepath.Glob(filepath.Join(dir, "crash-*.log"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no crash log in %s: %v", dir, err)
	}
	newest := paths[0]
	for _, path := range paths[1:] {
		a, errA := os.Stat(path)
		b, errB := os.Stat(newest)
		if errA == nil && errB == nil && a.ModTime().After(b.ModTime()) {
			newest = path
		}
	}
	t.Cleanup(func() { os.Remove(newest) })

	info, err := os.Stat(newest)
	if err != nil {
		t.Fatalf("stat %s: %v", newest, err)
	}
	if time.Since(info.ModTime()) > time.Minute {
		t.Fatalf("%s is old, reportCrash wrote nothing", newest)
	}
	data, err := os.ReadFile(newest)
	if err != nil {
		t.Fatalf("read %s: %v", newest, err)
	}
	if !strings.Contains(string(data), "wtgo test panic") {
		t.Errorf("%s does not hold the reason: %s", newest, data)
	}
}
