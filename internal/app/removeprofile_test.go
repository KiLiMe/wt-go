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

	a.ShowProfile("cmd")
	if a.settings.Profiles[0].Hidden {
		t.Errorf("profiles = %+v, want it shown again", a.settings.Profiles)
	}

	if !a.showHidden {
		t.Error("showHidden = false, want the page to list hidden profiles by default")
	}
	a.ShowHiddenProfiles()
	if a.showHidden {
		t.Error("showHidden = true, want the switch to hide them again")
	}
}
