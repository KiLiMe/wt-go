package app

import (
	"testing"

	"wtgo/internal/profile"
	"wtgo/internal/store"
)

// TestRemoveProfile covers what removing a profile does: a profile a person
// wrote goes for good, and one the window found is hidden, since the next
// start would find it again, and can be shown again from the page.
func TestRemoveProfile(t *testing.T) {
	settingsDir(t)
	a := newApp("")
	a.settings.Profiles = []store.Profile{
		{ID: "cmd", Name: "Command Prompt", Command: `C:\Windows\system32\cmd.exe`, Discovered: true},
		{ID: "mine", Name: "我的", Command: "cmd.exe"},
	}
	a.profiles = profile.Resolve(a.settings.Profiles, a.found)

	a.RemoveProfile("cmd")
	if len(a.settings.Profiles) != 2 || !a.settings.Profiles[0].Hidden {
		t.Fatalf("found profile = %+v, want it hidden", a.settings.Profiles)
	}

	a.RemoveProfile("mine")
	if len(a.settings.Profiles) != 1 || a.settings.Profiles[0].ID != "cmd" {
		t.Fatalf("profiles = %+v, want the written one gone", a.settings.Profiles)
	}

	a.HideProfile("cmd")
	if !a.settings.Profiles[0].Hidden {
		t.Errorf("profiles = %+v, want it hidden", a.settings.Profiles)
	}
	if len(a.settings.Profiles) != 1 {
		t.Errorf("profiles = %+v, want the hidden profile kept in the list", a.settings.Profiles)
	}
}

// TestHideProfile covers the button that hides the profile in front: it writes
// the draft the page edits, so the list marks the profile and the file takes
// the hiding on save, and the profile stays in front rather than the page
// moving to another one.
func TestHideProfile(t *testing.T) {
	settingsDir(t)
	a := newApp("")
	a.settings.Profiles = []store.Profile{
		{ID: "mine", Name: "我的", Command: "cmd.exe"},
		{ID: "other", Name: "别的", Command: "cmd.exe"},
	}
	a.profiles = profile.Resolve(a.settings.Profiles, a.found)
	a.editProfile("mine")

	a.HideProfile("mine")
	if !a.settings.Profiles[0].Hidden {
		t.Errorf("profiles = %+v, want the one in front hidden", a.settings.Profiles)
	}
	if !a.editing.Hidden {
		t.Error("the page's copy is not hidden, so its switch would read on after hiding")
	}
	if !a.dirty() {
		t.Error("the draft is not dirty, so the hiding could not be saved")
	}
	if a.editing == nil || a.editing.ID != "mine" {
		t.Errorf("the page left the profile it hid: %+v", a.editing)
	}
}
