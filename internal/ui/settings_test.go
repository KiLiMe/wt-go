package ui

import (
	"fmt"
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/i18n"
	"wtgo/internal/store"
)

// navActions is the window as the settings column sees it: it counts the
// profiles it asks to add.
type navActions struct {
	Actions
	added  int
	hidden int
}

func (a *navActions) AddProfile()         { a.added++ }
func (a *navActions) ShowHiddenProfiles() { a.hidden++ }

// settingsOf draws the settings tab over a list of profiles.
func settingsOf(t *testing.T, profiles []Profile) (*myui.Tester, *navActions) {
	t.Helper()
	s := store.Default()
	act := &navActions{}
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageProfiles, Profiles: profiles}, act)
	}, 900, 700)
	tt.Frame()
	return tt, act
}

// TestNavAddsProfiles covers the row that adds a profile: it holds at the
// bottom of the column, so a list long enough to scroll never puts it out of
// reach, and a click on it asks for a new profile.
func TestNavAddsProfiles(t *testing.T) {
	var profiles []Profile
	for i := 0; i < 18; i++ {
		profiles = append(profiles, Profile{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Profile %d", i)})
	}
	tt, act := settingsOf(t, profiles)
	if err := tt.Click(i18n.T("Add profile")); err != nil {
		t.Fatalf("the add row scrolled away: %v", err)
	}
	if act.added != 1 {
		t.Fatalf("added = %d, want 1", act.added)
	}
	if err := tt.Click("Profile 17"); err == nil {
		t.Fatal("the last profile is in view, so the column did not scroll")
	}
	if err := tt.Click(i18n.T("Show hidden profiles")); err != nil {
		t.Fatalf("the checkbox scrolled away: %v", err)
	}
	if act.hidden != 1 {
		t.Fatalf("hidden = %d, want 1", act.hidden)
	}
}

// TestNavHoldsUnderTheList covers the short list: the row that adds a profile
// sits at the bottom of the column, under the profiles rather than beside them.
func TestNavHoldsUnderTheList(t *testing.T) {
	tt, _ := settingsOf(t, []Profile{{ID: "a", Name: "PowerShell 7"}})
	last, ok := tt.Find("PowerShell 7")
	if !ok {
		t.Fatal("the profile is missing from the column")
	}
	add, ok := tt.Find(i18n.T("Add profile"))
	if !ok {
		t.Fatal("the add row is missing from the column")
	}
	if add.Y <= last.Y+last.H {
		t.Fatalf("add row at %v is not under the profile at %v", add.Y, last.Y)
	}
}
