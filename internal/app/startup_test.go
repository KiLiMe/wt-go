package app

import (
	"strings"
	"testing"

	"wtgo/internal/profile"
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

// TestNewAppAlwaysKnowsTheDevice covers a file that is already there: the
// shells come from it and are not probed again, but the device is a fact about
// the machine, so the settings page still names the shell it runs.
func TestNewAppAlwaysKnowsTheDevice(t *testing.T) {
	settingsDir(t)
	if first := newApp(""); len(first.found) == 0 {
		t.Skip("no shell on this machine")
	}
	a := newApp("")
	if len(a.found) != 0 {
		t.Errorf("probed %d shells with a file in place", len(a.found))
	}
	if a.device.ID == "" || a.device.Command == "" {
		t.Fatalf("device = %+v, want the shell of this machine", a.device)
	}
	if want := profile.Name(a.device.ID, a.device.Command); a.device.Name != want {
		t.Errorf("device name = %q, want %q", a.device.Name, want)
	}
	if got := a.settingsState().Device; got != a.device.Name {
		t.Errorf("settings page names the device %q, want %q", got, a.device.Name)
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
