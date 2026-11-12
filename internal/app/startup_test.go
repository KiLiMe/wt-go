package app

import (
	"strings"
	"testing"

	"wtgo/internal/sys"
)

// TestStartDirectoryPrefersTheFlag: -d is for this run only, and beats the
// setting without changing it.
func TestStartDirectoryPrefersTheFlag(t *testing.T) {
	settingsDir(t)
	a := newApp(`C:\src`)
	if got := a.startDirectory(); got != `C:\src` {
		t.Fatalf("startDirectory() = %q, want the flag", got)
	}
	a.settings.StartDirectory = "home"
	if got := a.startDirectory(); got != `C:\src` {
		t.Errorf("startDirectory() = %q, want the flag over the setting", got)
	}
	a.startDir = ""
	if got := a.startDirectory(); got != "home" {
		t.Errorf("startDirectory() = %q, want the setting once the flag is gone", got)
	}
}

// TestNewAppWritesTheFoundShells covers the first start: the shells the
// machine has end up in settings.json, once, as the window writes them.
func TestNewAppWritesTheFoundShells(t *testing.T) {
	settingsDir(t)
	a := newApp("")
	if len(a.found) == 0 {
		t.Skip("no shell on this machine")
	}
	data, ok, err := sys.ReadSettings()
	if err != nil || !ok {
		t.Fatalf("read: %v, %v", err, ok)
	}
	first := string(data)
	for _, f := range a.found {
		if !strings.Contains(first, `"id": "`+f.ID+`"`) {
			t.Errorf("settings.json has no %q: %s", f.ID, first)
		}
	}

	// A second start finds the file already says what the machine has, and
	// leaves it alone.
	before := a.savedData
	b := newApp("")
	if string(before) != string(b.savedData) {
		t.Errorf("the second start rewrote settings.json")
	}
}
