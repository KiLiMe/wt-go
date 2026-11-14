package app

import (
	"testing"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/ui"
)

// TestKnownProfilesSkipsHidden covers the list the settings offer: a hidden
// profile runs the system's shell, so the page does not offer it.
func TestKnownProfilesSkipsHidden(t *testing.T) {
	ps := []profile.Profile{{ID: "a", Name: "A"}, {ID: "b", Name: "B", Hidden: true}}
	got := knownProfiles(ps)
	if len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("knownProfiles = %+v, want A alone", got)
	}
}

// TestAddProfileDropsTheArgumentsOfTheLastOne covers a new profile opening on
// the page the profile before it had: the lines it held are not this profile's
// arguments, so the page starts empty and nothing is unsaved.
func TestAddProfileDropsTheArgumentsOfTheLastOne(t *testing.T) {
	settingsDir(t)
	s := store.Default()
	s.Profiles = []store.Profile{{ID: "p1", Name: "One", Command: "pwsh", Args: []string{"-NoLogo", "-NoProfile"}}, {ID: "p2", Name: "Two"}}
	a, _ := settingsApp(t, s)
	a.settingsTab = &pane.Tab{Name: "Settings"}

	a.SelectSettingsPage(ui.ProfilePage("p1"))
	if a.editingArgs != "-NoLogo\n-NoProfile" {
		t.Fatalf("editingArgs = %q, want the profile's lines", a.editingArgs)
	}
	a.editingArgs = "-NoLogo\n-NoProfile\n-Draft"
	if !a.dirty() {
		t.Fatal("a draft over the file is not dirty")
	}

	a.AddProfile()
	if a.editingArgs != "" {
		t.Errorf("editingArgs = %q, want none: the last profile's lines are not this one's", a.editingArgs)
	}
	// The profile itself is new, which is unsaved; what the page holds is not.
	a.settings = a.saved.Clone()
	if a.dirty() {
		t.Error("the page holds a line the new profile does not")
	}
}

// TestDuplicateProfileShowsTheCopiedArguments covers a copy opening on the
// page: it shows the arguments it copied, not the ones the page held before.
func TestDuplicateProfileShowsTheCopiedArguments(t *testing.T) {
	settingsDir(t)
	s := store.Default()
	s.Profiles = []store.Profile{
		{ID: "p1", Name: "One", Command: "pwsh", Args: []string{"-NoLogo"}},
		{ID: "p2", Name: "Two", Command: "bash", Args: []string{"-l", "-i"}},
		{ID: "p3", Name: "Three"},
	}
	a, _ := settingsApp(t, s)

	a.SelectSettingsPage(ui.ProfilePage("p1"))

	a.DuplicateProfile("p2")
	if want := "-l\n-i"; a.editingArgs != want {
		t.Errorf("editingArgs = %q, want %q", a.editingArgs, want)
	}
	// The page opened is the copy's, not the one it was on.
	if a.settingsPage != ui.ProfilePage(a.editing.ID) {
		t.Errorf("page = %q, want the copy's %q", a.settingsPage, ui.ProfilePage(a.editing.ID))
	}
	if a.editing.Name == "Two" {
		t.Error("the page edits the profile itself, not a copy")
	}
	// The copy itself is new, which is unsaved; what the page holds is not.
	a.settings = a.saved.Clone()
	if a.dirty() {
		t.Error("the page holds a line the copy does not")
	}
}

// TestDiscardSettingsLeavesAnUnsavedProfile covers discarding a draft with a
// profile the file does not hold: the profile goes, and the page it was on
// goes with it, rather than editing nothing.
func TestDiscardSettingsLeavesAnUnsavedProfile(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	before := len(a.settings.Profiles)

	a.AddProfile()
	if a.editing == nil {
		t.Fatal("the new profile is not open on the page")
	}

	a.DiscardSettings()
	if a.editing != nil {
		t.Error("a profile the file does not hold is still open")
	}
	if a.settingsPage != ui.PageProfiles {
		t.Errorf("page = %q, want %q", a.settingsPage, ui.PageProfiles)
	}
	if len(a.settings.Profiles) != before {
		t.Errorf("the settings hold %d profiles, want the file's %d", len(a.settings.Profiles), before)
	}
	if a.dirty() {
		t.Error("discarding left a draft")
	}
}

// TestProfileIcon covers the mark the profile page picks: a shape is written
// to the file as it is, a character typed in the box is written as the one
// character it is, and an emptied box is a profile that wears no character of
// its own. The page edits a copy of the profile, so the file changes on save.
func TestProfileIcon(t *testing.T) {
	settingsDir(t)
	a := newApp("")
	a.settings.Profiles = []store.Profile{{
		ID:         "pwsh7",
		Name:       "PowerShell 7",
		Command:    `C:\Windows\pwsh.exe`,
		Discovered: true,
	}}
	a.profiles = profile.Resolve(a.settings.Profiles, a.found)
	a.editProfile("pwsh7")

	a.SetProfileIcon("bash:arch")
	if got := a.editing.Icon; got != "bash:arch" {
		t.Errorf("icon = %q, want the shape the page picked", got)
	}
	if a.editingGlyph != "" {
		t.Errorf("the character box holds %q, want it cleared by a shape", a.editingGlyph)
	}
	if !a.dirty() {
		t.Error("the page does not know it changed the mark")
	}
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "bash:arch" {
		t.Errorf("saved icon = %q, want the shape the page picked", got)
	}

	a.editProfile("pwsh7")
	if a.editingGlyph != "" {
		t.Error("the character box opens on text, want it empty for a shape")
	}
	a.editingGlyph = "\U0001F92C"
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "\U0001F92C" {
		t.Errorf("saved icon = %q, want the character the page typed", got)
	}

	a.editProfile("pwsh7")
	if a.editingGlyph != "\U0001F92C" {
		t.Errorf("the box opens on %q, want the character the file holds", a.editingGlyph)
	}
	a.editingGlyph = "\U0001F92C\U0001F92C"
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "\U0001F92C" {
		t.Errorf("saved icon = %q, want the one character of the two typed", got)
	}

	// An emptied box is a profile that wears no character: it keeps a shape it
	// holds, and goes back to its command's mark otherwise.
	a.SetProfileIcon("cmd:matrix")
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "cmd:matrix" {
		t.Errorf("saved icon = %q, want the shape the page picked", got)
	}
	a.editProfile("pwsh7")
	a.editingGlyph = ""
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "cmd:matrix" {
		t.Errorf("saved icon = %q, want the shape an empty box keeps", got)
	}
	a.SetProfileIcon("")
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "" {
		t.Errorf("saved icon = %q, want the command's mark", got)
	}
}
