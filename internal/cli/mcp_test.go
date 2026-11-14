package cli

import (
	"runtime"
	"strings"
	"testing"

	"wtgo/internal/mcp"
	"wtgo/internal/sys"
)

// logTool is the wtgo_log tool of the server wt-go offers.
func logTool(t *testing.T) mcp.Tool {
	t.Helper()
	for _, tool := range Server().Tools {
		if tool.Name == "wtgo_log" {
			return tool
		}
	}
	t.Fatal("Server() has no wtgo_log")
	return mcp.Tool{}
}

// sepDir is a directory of this test's own, for a path the tool has to refuse.
// It is built by joining, not by path/filepath: a package's code may not reach
// the files the app loads, and its tests hold to that too, so that the test of
// the rule is not what makes an exception to it.
func sepDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// join is path/filepath.Join for the paths this test builds: the same count of
// separators, on either platform the test runs on.
func join(elem ...string) string {
	var sep string
	if runtime.GOOS == "windows" {
		sep = `\`
	} else {
		sep = "/"
	}
	return strings.Join(elem, sep)
}

// A path the client sends is read only when it names a log wt-go wrote: the
// tool answers with the path itself, so a refusal is visible as an error.
func TestLogToolRefusesOutsidePaths(t *testing.T) {
	tool := logTool(t)
	paths := []string{
		join(sepDir(t), "wt-go.log"),
		join(sepDir(t), "settings.json"),
	}
	if runtime.GOOS == "windows" {
		paths = append(paths, `C:\Windows\win.ini`)
	} else {
		paths = append(paths, "/etc/passwd")
	}
	if dir, err := sys.Dir(); err == nil {
		paths = append(paths, join(dir, "..", "wt-go.log"))
	}
	for _, p := range paths {
		if sys.IsLogPath(p) {
			t.Fatalf("IsLogPath(%q) = true, the test picked a log", p)
		}
		text, err := tool.Run(map[string]any{"path": p})
		if err == nil {
			t.Errorf("Run(%q) = %q, want a refusal", p, text)
			continue
		}
		if !strings.Contains(err.Error(), "not a wt-go log") {
			t.Errorf("Run(%q) = %v, want the refusal to say so", p, err)
		}
	}
}

// A path inside the wt-go directory is read: a log that is there and not
// there answers with the OS error, which is what the caller has to see.
func TestLogToolReadsLogPath(t *testing.T) {
	dir, err := sys.Dir()
	if err != nil {
		t.Skipf("Dir: %v", err)
	}
	p := join(dir, "crash-19700101-000000.log")
	if !sys.IsLogPath(p) {
		t.Fatalf("IsLogPath(%q) = false", p)
	}
	text, err := logTool(t).Run(map[string]any{"path": p, "lines": 1.0})
	if err != nil {
		// The file almost certainly does not exist; what matters is that the
		// refusal was the read failing, not the path check.
		if strings.Contains(err.Error(), "not a wt-go log") {
			t.Errorf("Run(%q) = %v, want it to be read", p, err)
		}
		return
	}
	if !strings.HasPrefix(text, p+"\n\n") {
		t.Errorf("Run(%q) = %q, want the path first", p, text)
	}
}
