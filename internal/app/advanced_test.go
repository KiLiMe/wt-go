package app

import (
	"bytes"
	"testing"

	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
)

// TestNewAppOnlyProbesWithoutAFile covers a file that is already there: the
// shells come from it, and the machine is not probed again, so a profile a
// person edited keeps the command until they ask for a probe.
func TestNewAppOnlyProbesWithoutAFile(t *testing.T) {
	settingsDir(t)
	s := store.Default()
	s.Profiles = []store.Profile{{
		ID: "pwsh7", Name: "Mine", Command: `C:\nowhere\pwsh.exe`, Discovered: true,
	}}
	data, err := store.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if err := sys.WriteSettings(data); err != nil {
		t.Fatal(err)
	}
	a := newApp("")
	if len(a.found) != 0 {
		t.Errorf("probed %d shells with a file in place", len(a.found))
	}
	if len(a.profiles) != 1 || a.profiles[0].Command != `C:\nowhere\pwsh.exe` {
		t.Fatalf("profiles = %+v, want the one the file names", a.profiles)
	}
	after, _, err := sys.ReadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, data) {
		t.Errorf("the start rewrote the file:\n%s", after)
	}
}

// TestAdoptAgainTakesTheProbedCommand covers what a probe does to a profile
// the file already has: the command becomes the one that is installed now,
// while the name, color and directory a person chose stay.
func TestAdoptAgainTakesTheProbedCommand(t *testing.T) {
	s := store.Settings{Profiles: []store.Profile{{
		ID: "pwsh7", Name: "Mine", Command: `C:\old\pwsh.exe`, Dir: `D:\code`, Color: "#ff0000", Discovered: true,
	}}}
	found := []profile.Found{{ID: "pwsh7", Command: `C:\new\pwsh.exe`, Args: []string{"-NoLogo"}}}
	if !adoptAgain(&s, found) {
		t.Fatal("a changed command is a change")
	}
	p := s.Profiles[0]
	if p.Command != `C:\new\pwsh.exe` || len(p.Args) != 1 || p.Args[0] != "-NoLogo" {
		t.Errorf("command = %q %q, want the one probed", p.Command, p.Args)
	}
	if p.Name != "Mine" || p.Dir != `D:\code` || p.Color != "#ff0000" {
		t.Errorf("the probe took what a person chose: %+v", p)
	}
	if adoptAgain(&s, found) {
		t.Error("probing twice with the same machine changed something")
	}
}

// TestAdoptAgainHidesAndShows covers the shells that come and go: one whose
// program is gone is hidden rather than dropped, and one that is back is shown
// again.
func TestAdoptAgainHidesAndShows(t *testing.T) {
	s := store.Settings{Profiles: []store.Profile{
		{ID: "pwsh7", Command: `C:\a\pwsh.exe`, Discovered: true},
		{ID: "cmd", Command: `C:\b\cmd.exe`, Discovered: true},
	}}
	if !adoptAgain(&s, []profile.Found{{ID: "cmd", Command: `C:\b\cmd.exe`}}) {
		t.Fatal("a shell that is gone is a change")
	}
	if !s.Profiles[0].Hidden {
		t.Error("the shell that is gone is shown")
	}
	if s.Profiles[1].Hidden {
		t.Error("the shell that is there is hidden")
	}
	if !adoptAgain(&s, []profile.Found{
		{ID: "cmd", Command: `C:\b\cmd.exe`},
		{ID: "pwsh7", Command: `C:\a\pwsh.exe`},
	}) {
		t.Fatal("a shell that is back is a change")
	}
	if s.Profiles[0].Hidden {
		t.Error("the shell that is back is still hidden")
	}
}

// TestOpenSettingsFileWritesTheFileFirst covers the foot's button: a window
// that never wrote settings.json writes the settings it holds before opening
// it, so that the file the desktop opens is there to edit.
func TestOpenSettingsFileWritesTheFileFirst(t *testing.T) {
	settingsDir(t)
	s := store.Default()
	s.Theme = "light"
	a := &App{settings: s.Clone(), saved: s.Clone()}
	if err := a.ensureSettingsFile(); err != nil {
		t.Fatal(err)
	}
	data, ok, err := sys.ReadSettings()
	if err != nil || !ok {
		t.Fatalf("the file is not there: %v, %v", err, ok)
	}
	got, err := store.Parse(data)
	if err != nil {
		t.Fatalf("the file it wrote is not settings: %v", err)
	}
	if got.Theme != "light" {
		t.Fatalf("the file holds theme %q, want the draft's", got.Theme)
	}
	// What it wrote is what the window has: a file it wrote itself is not a
	// draft to save again.
	if !bytes.Equal(a.savedData, data) {
		t.Fatal("the window did not take the file it wrote as saved")
	}
	// A file that is there is left alone.
	other := []byte(`{"theme":"dark"}`)
	if err := sys.WriteSettings(other); err != nil {
		t.Fatal(err)
	}
	if err := a.ensureSettingsFile(); err != nil {
		t.Fatal(err)
	}
	if data, _, err := sys.ReadSettings(); err != nil || !bytes.Equal(data, other) {
		t.Fatalf("the file the person edited was overwritten: %q, %v", data, err)
	}
}

// TestResetDeletesTheFileAndRestarts covers the reset: the file goes, the
// program starts again, and the window closes so that there are not two.
func TestResetDeletesTheFileAndRestarts(t *testing.T) {
	settingsDir(t)
	a := newApp("")
	if _, ok, err := sys.ReadSettings(); err != nil || !ok {
		t.Fatalf("the first start wrote no file: %v, %v", err, ok)
	}
	restarted := false
	a.relaunch = func() error {
		restarted = true
		return nil
	}
	a.reset()
	if _, ok, err := sys.ReadSettings(); err != nil || ok {
		t.Fatalf("settings.json is still there: %v, %v", err, ok)
	}
	if !restarted {
		t.Error("the window did not start again")
	}
}
