package ui

import (
	"fmt"
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
	"wtgo/locales"
)

// navActions is the window as the settings column sees it: it counts the
// profiles it asks to add, and the keybinding changes it is asked for.
type navActions struct {
	Actions
	added   int
	hidden  int
	toggled int
}

func (a *navActions) AddProfile()         { a.added++ }
func (a *navActions) ShowHiddenProfiles() { a.hidden++ }
func (a *navActions) ToggleChangedKeys()  { a.toggled++ }

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
	if err := tt.Click(locales.AddProfile()); err != nil {
		t.Fatalf("the add row scrolled away: %v", err)
	}
	if act.added != 1 {
		t.Fatalf("added = %d, want 1", act.added)
	}
	if err := tt.Click("Profile 17"); err == nil {
		t.Fatal("the last profile is in view, so the column did not scroll")
	}
	// The switch that lists the hidden profiles is on the page, under the list
	// it filters, so a list this long puts it past the window until it scrolls.
	if err := tt.Click(locales.ShowHiddenProfiles()); err == nil {
		t.Fatal("the switch is in view, want it under the list on the page")
	}
	tt.Scroll(600, 400, 0, 4000)
	if err := tt.Click(locales.ShowHiddenProfiles()); err != nil {
		t.Fatalf("the switch did not come into view: %v", err)
	}
	if act.hidden != 1 {
		t.Fatalf("hidden = %d, want 1", act.hidden)
	}
}

// TestStartupPageShowsTheSystemForAProfileThatIsGone covers the select of the
// default profile: a file that names one the page does not offer runs the
// system's shell, so the select says so rather than naming a profile that is
// not there, and drawing the page leaves the draft as the file wrote it.
func TestStartupPageShowsTheSystemForAProfileThatIsGone(t *testing.T) {
	s := store.Default()
	s.DefaultProfile = "gone"
	act := &navActions{}
	known := []Profile{{ID: "pwsh7", Name: "PowerShell 7"}}
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageStartup, Known: known}, act)
	}, 900, 700)
	tt.Frame()
	if tt.HasText("gone") {
		t.Fatal("the select names a profile the page does not offer")
	}
	if s.DefaultProfile != "gone" {
		t.Fatalf("drawing the page changed the draft: DefaultProfile = %q", s.DefaultProfile)
	}
}

// TestActionsPageToolbar covers what the toolbar of the actions page offers: a
// click on the switch asks for the changed actions alone, the groups that fold
// numbered actions away say how many they hold back, and an action the file
// binds as the catalog does is not marked changed.
func TestActionsPageToolbar(t *testing.T) {
	s := store.Default()
	act := &navActions{}
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageActions}, act)
	}, 900, 700)
	tt.Frame()
	if err := tt.Click(locales.OnlyChanged(0)); err != nil {
		t.Fatalf("the switch for the changed actions is missing: %v", err)
	}
	if act.toggled != 1 {
		t.Fatalf("toggled = %d, want 1", act.toggled)
	}
	// The settings bind every action as the catalog does, so the only rows
	// folded away are the numbered ones past the third: the groups of the tabs
	// and of the profiles hold back six each.
	if _, ok := tt.Find(locales.ShowNumbered(6)); !ok {
		t.Fatal("the bar that unfolds the numbered actions is missing")
	}
	if _, ok := tt.Find(locales.Changed()); ok {
		t.Fatal("an action the file binds as the catalog does is marked changed")
	}
}

// TestNavPutsDefaultsAboveTheProfiles covers the order of the column: the page
// that holds the defaults and the profiles comes before the heading that names
// the list of profiles.
func TestNavPutsDefaultsAboveTheProfiles(t *testing.T) {
	tt, _ := settingsOf(t, []Profile{{ID: "a", Name: "PowerShell 7"}})
	def, ok := tt.Find(locales.Defaults())
	if !ok {
		t.Fatal("the defaults page is missing from the column")
	}
	head, ok := tt.Find(locales.Profiles())
	if !ok {
		t.Fatal("the profiles heading is missing from the column")
	}
	// Both are in the column, not the heading of the page the settings draw.
	if def.X >= navWidth || head.X >= navWidth {
		t.Fatalf("defaults at %v or heading at %v is not in the column", def.X, head.X)
	}
	if def.Y >= head.Y {
		t.Fatalf("the defaults page at %v is not above the heading at %v", def.Y, head.Y)
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
	add, ok := tt.Find(locales.AddProfile())
	if !ok {
		t.Fatal("the add row is missing from the column")
	}
	if add.Y <= last.Y+last.H {
		t.Fatalf("add row at %v is not under the profile at %v", add.Y, last.Y)
	}
}
