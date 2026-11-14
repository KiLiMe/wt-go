package app

import (
	"errors"
	"testing"

	"wtgo/internal/action"
	"wtgo/internal/cli"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/ui"
)

// TestOneOverlayAtATime covers the overlays of the window: opening one closes
// the one before it, so that the window never shows two and the keys cannot
// reach both.
func TestOneOverlayAtATime(t *testing.T) {
	a := &App{settings: store.Default(), saved: store.Default()}
	asked := 0
	a.ask(ui.Confirm{Title: "close?"}, func() { asked++ }, nil)
	if a.top() != overlayConfirm {
		t.Fatalf("top = %v, want the question", a.top())
	}

	a.OpenPalette()
	if a.confirm != nil {
		t.Error("the question stayed open under the palette")
	}
	if a.top() != overlayPalette {
		t.Fatalf("top = %v, want the palette", a.top())
	}
	// The answer that the question would have acted on is gone, so a click
	// that arrived late does not act.
	a.ConfirmYes()
	if asked != 0 {
		t.Errorf("the question asked while closed acted %d times", asked)
	}
	if a.palette == nil {
		t.Error("the question's answer closed the palette")
	}

	a.OpenAbout()
	if a.palette != nil {
		t.Error("the palette stayed open under the dialog")
	}
	if a.top() != overlayAbout {
		t.Fatalf("top = %v, want the about dialog", a.top())
	}
	a.CloseAbout()
	if a.top() != overlayNone {
		t.Fatalf("top = %v, want none", a.top())
	}
}

// TestAboutSaysWhichBuild covers the dialog's lines: the build that answered,
// what it was built on, and the two files a bug report asks to look at. A path
// the machine cannot name is left out rather than shown half.
func TestAboutSaysWhichBuild(t *testing.T) {
	self := cli.Self{Version: "v0.4.0", Revision: "a8e3625", Go: "go1.25.0", Platform: "windows/amd64"}
	settings := func() (string, error) { return `C:\Users\me\AppData\Local\wt-go\settings.json`, nil }
	logPath := func() (string, error) { return `C:\Users\me\AppData\Local\wt-go\wt-go.log`, nil }
	a := aboutOf(self, "0.9.0", settings, logPath)
	if a.Version != "v0.4.0" {
		t.Errorf("version = %q", a.Version)
	}
	if a.Revision != "a8e3625" {
		t.Errorf("revision = %q", a.Revision)
	}
	if a.Mygo != "0.9.0" {
		t.Errorf("mygo = %q", a.Mygo)
	}
	if a.Go != "go1.25.0" || a.Platform != "windows/amd64" {
		t.Errorf("runtime = %q %q", a.Go, a.Platform)
	}
	if a.Settings != `C:\Users\me\AppData\Local\wt-go\settings.json` {
		t.Errorf("settings = %q", a.Settings)
	}
	if a.Log != `C:\Users\me\AppData\Local\wt-go\wt-go.log` {
		t.Errorf("log = %q", a.Log)
	}

	// A folder that cannot be named is E103: the dialog says less, not a
	// path of its own invention.
	broken := func() (string, error) { return "", errors.New("no LOCALAPPDATA") }
	a = aboutOf(self, "", broken, broken)
	if a.Settings != "" || a.Log != "" {
		t.Errorf("settings, log = %q, %q, want both empty", a.Settings, a.Log)
	}
}

// TestStartOnNewOpensTheMachineShell covers startOn: a window told to start on
// a new tab runs the shell of this machine even where defaultProfile names
// another profile.
func TestStartOnNewOpensTheMachineShell(t *testing.T) {
	profiles := []profile.Profile{
		{ID: "pwsh7", Name: "PowerShell 7"},
		{ID: "zsh", Name: "zsh"},
	}
	found := []profile.Found{{ID: "pwsh7", Command: `C:\Program Files\PowerShell\7\pwsh.exe`}}
	for _, tc := range []struct {
		start, want string
	}{
		{store.StartDefault, "zsh"},
		{store.StartNew, "pwsh7"},
	} {
		a := &App{profiles: profiles, device: profile.Device(found)}
		a.settings.DefaultProfile = "zsh"
		a.settings.StartOn = tc.start
		if got := a.firstTab().ID; got != tc.want {
			t.Errorf("startOn %q: firstTab = %q, want %q", tc.start, got, tc.want)
		}
	}
}

// TestKeysGoToTheOverlayOnTop covers the keys: an overlay that holds the
// keyboard keeps the window's actions from them until it closes.
func TestKeysGoToTheOverlayOnTop(t *testing.T) {
	a := &App{settings: store.Default(), saved: store.Default()}
	a.ask(ui.Confirm{Title: "close?"}, nil, nil)
	a.run(action.Palette)
	if a.palette != nil {
		t.Error("a key opened the palette over the question")
	}
	if a.top() != overlayConfirm {
		t.Errorf("top = %v, want the question", a.top())
	}
}

// TestFindAndPalette covers the two overlays that do not hold the keyboard the
// way a dialog does: one opened over the other closes it, and the window's keys
// still reach the window over the bar that asks the terminal.
func TestFindAndPalette(t *testing.T) {
	a := &App{settings: store.Default(), saved: store.Default()}
	a.openOverlay(overlayFind)
	a.openOverlay(overlayPalette)
	if a.find != nil {
		t.Error("the bar stayed open under the palette")
	}
	if a.top() != overlayPalette {
		t.Fatalf("top = %v, want the palette", a.top())
	}

	a.openOverlay(overlayFind)
	if a.palette != nil {
		t.Error("the palette stayed open under the bar")
	}
	a.run(action.Find)
	if a.palette != nil {
		t.Error("a key opened the palette over the bar")
	}
	if a.top() != overlayFind {
		t.Errorf("top = %v, want the bar", a.top())
	}
}

// TestGoingToTheSettingsClosesTheOverlays covers leaving the terminal view: the
// settings tab is not a terminal, so nothing that asks a terminal stays open
// over it.
func TestGoingToTheSettingsClosesTheOverlays(t *testing.T) {
	a := &App{settings: store.Default(), saved: store.Default()}
	a.openOverlay(overlayRename)
	a.OpenSettings()
	if a.top() != overlayNone {
		t.Fatalf("top = %v, want none", a.top())
	}
	if a.settingsTab == nil {
		t.Error("the settings did not open")
	}
}
