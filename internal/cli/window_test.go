package cli

import (
	"bytes"
	"strings"
	"testing"
)

// result runs the command line and answers what it said, and what it
// complained about.
func result(t *testing.T, args ...string) (Result, string) {
	t.Helper()
	var o, e bytes.Buffer
	return Run(args, strings.NewReader(""), &o, &e), e.String()
}

// -d and --dir name the directory a new tab starts in, and leave the window to
// open: nothing is handled.
func TestDirFlagNamesTheStart(t *testing.T) {
	for _, args := range [][]string{{"-d", `C:\src`}, {"--dir", `C:\src`}, {"-d", "/opt/src"}} {
		r, errOut := result(t, args...)
		if r.Handled || r.Code != 0 || r.StartDir != args[1] || errOut != "" {
			t.Fatalf("Run(%q) = %+v, %q, want the window in %q", args, r, errOut, args[1])
		}
	}
	// The last -d wins, as the last flag does everywhere else.
	if r, _ := result(t, "-d", "a", "--dir", "b"); r.StartDir != "b" {
		t.Fatalf("Run(-d a --dir b).StartDir = %q, want b", r.StartDir)
	}
	if r, errOut := result(t); r.Handled || r.StartDir != "" || errOut != "" {
		t.Fatalf("Run() = %+v, %q, want the window where the settings say", r, errOut)
	}
}

// A -d without a directory, or one that would go with a command that answers
// instead of opening a window, is a mistake worth naming.
func TestDirFlagMustStandAlone(t *testing.T) {
	for _, args := range [][]string{{"-d"}, {"--dir"}, {"-d", ""}, {"-d", `C:\src`, "diag"}} {
		r, errOut := result(t, args...)
		if !r.Handled || r.Code != 2 || r.StartDir != "" {
			t.Fatalf("Run(%q) = %+v, want handled with code 2", args, r)
		}
		if !strings.Contains(errOut, "wt-go: ") {
			t.Errorf("Run(%q) said %q, want the reason", args, errOut)
		}
	}
}
