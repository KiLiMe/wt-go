package app

import (
	"slices"
	"strings"
	"testing"

	"wtgo/internal/profile"
	"wtgo/internal/store"
)

// TestResetProfile covers the button that puts a built-in profile back: the
// program and the arguments the probe has now, the name and the dot the table
// gives that shell, and a hidden profile on the list again. The start
// directory stays, because that is the person's choice.
func TestResetProfile(t *testing.T) {
	settingsDir(t)
	a := newApp("")
	a.settings.Profiles = []store.Profile{{
		ID:         "pwsh7",
		Name:       "我的 pwsh",
		Command:    `C:\gone\pwsh.exe`,
		Args:       []string{"-NoLogo"},
		Dir:        `D:\work`,
		Color:      "#123456",
		Hidden:     true,
		Discovered: true,
	}}
	a.profiles = profile.Resolve(a.settings.Profiles, a.found)
	a.editProfile("pwsh7")

	a.resetProfile(a.settings.Profiles[0], profile.Found{
		ID:      "pwsh7",
		Command: `C:\Program Files\PowerShell\7\pwsh.exe`,
		Args:    []string{"-NoLogo", "-NoProfile"},
	}, true)

	p := a.settings.Profiles[0]
	if want := profile.Name("pwsh7", p.Command); p.Name != want {
		t.Errorf("name = %q, want the table's %q", p.Name, want)
	}
	if want := profile.Color("pwsh7"); p.Color != want {
		t.Errorf("color = %q, want the table's %q", p.Color, want)
	}
	if p.Command != `C:\Program Files\PowerShell\7\pwsh.exe` {
		t.Errorf("command = %q, want the one the probe found", p.Command)
	}
	if want := []string{"-NoLogo", "-NoProfile"}; !slices.Equal(p.Args, want) {
		t.Errorf("args = %q, want %q", p.Args, want)
	}
	if p.Hidden {
		t.Error("a profile the machine has is hidden, so the menu would not offer it")
	}
	if p.Dir != `D:\work` {
		t.Errorf("dir = %q, want the directory the person chose", p.Dir)
	}
	if a.editing == nil || a.editing.Name != p.Name || a.editing.Hidden {
		t.Errorf("the page's copy is %+v, want the reset profile", a.editing)
	}
	if want := strings.Join(p.Args, "\n"); a.editingArgs != want {
		t.Errorf("arguments on the page = %q, want %q", a.editingArgs, want)
	}
	if !a.dirty() {
		t.Error("the draft is not dirty, so the reset could not be saved")
	}
}

// TestResetProfileGone covers a reset whose shell this machine no longer has:
// the profile is hidden, since a row that runs nothing is what the menu would
// otherwise offer, and the page says so rather than going quiet.
func TestResetProfileGone(t *testing.T) {
	settingsDir(t)
	a := newApp("")
	a.settings.Profiles = []store.Profile{{
		ID:         "nu",
		Name:       "Nushell",
		Command:    `C:\gone\nu.exe`,
		Color:      "#7c6cf5",
		Discovered: true,
	}}
	a.profiles = profile.Resolve(a.settings.Profiles, a.found)
	a.editProfile("nu")

	a.resetProfile(a.settings.Profiles[0], profile.Found{}, false)

	if !a.settings.Profiles[0].Hidden {
		t.Errorf("profiles = %+v, want the one with no program hidden", a.settings.Profiles)
	}
	if !a.editing.Hidden {
		t.Error("the page's copy is not hidden, so its switch would read off after hiding")
	}
	if a.note == "" {
		t.Error("the page says nothing about why the profile is hidden now")
	}
}
