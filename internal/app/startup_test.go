package app

import (
	"strings"
	"testing"

	"wtgo/internal/sys"
)

// TestNewAppWritesTheFoundShells covers the first start: the shells the
// machine has end up in settings.json, once, as the window writes them.
func TestNewAppWritesTheFoundShells(t *testing.T) {
	settingsDir(t)
	a := newApp()
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
	b := newApp()
	if string(before) != string(b.savedData) {
		t.Errorf("the second start rewrote settings.json")
	}
}
