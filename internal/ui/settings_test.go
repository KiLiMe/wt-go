package ui

import (
	"fmt"
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
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
	// The switch that lists the hidden profiles opens the page, above the list
	// it filters, so a list this long never puts it under the rows it lists.
	sw, ok := tt.AccessByRole(myui.RoleSwitch, locales.ShowHiddenProfiles())
	if !ok {
		t.Fatal("the switch that lists the hidden profiles is missing")
	}
	first := float32(-1)
	for _, a := range tt.AccessAll() {
		if a.Name == "Profile 0" && a.Bounds.X >= navWidth {
			first = a.Bounds.Y
			break
		}
	}
	if first < 0 {
		t.Fatal("the page does not list the first profile")
	}
	if sw.Bounds.Y >= first {
		t.Fatalf("the switch at %v is under the first profile at %v", sw.Bounds, first)
	}
	// The card it opens holds the list, so it is past the window until the
	// page scrolls there, and then a click has to reach the switch itself: its
	// name is on the label at the other end of the row.
	tt.Scroll(600, 400, 0, 600)
	if sw, ok = tt.AccessByRole(myui.RoleSwitch, locales.ShowHiddenProfiles()); !ok {
		t.Fatal("the switch left the page while scrolling")
	}
	tt.ClickAt(sw.Bounds.X+sw.Bounds.W/2, sw.Bounds.Y+sw.Bounds.H/2)
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

// TestActionsPageShowsEveryKeyOfAnAction covers the cell of a row the file
// binds twice, which tab.new is by default: both keys are in it, and the hint
// holds them in full for the cell too narrow to show them.
func TestActionsPageShowsEveryKeyOfAnAction(t *testing.T) {
	s := store.Default()
	s.Keybindings = store.DefaultKeybindings()
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageActions}, &navActions{})
	}, 900, 700)
	tt.Frame()
	if _, ok := tt.Find("ctrl+shift+t · ctrl+t"); !ok {
		t.Fatal("the row of tab.new does not show both of its keys")
	}
}

// TestActionsPageNamesOnlyRealClashes covers the red text of the toolbar: it
// says a key runs more than one action, so it stays away while every key runs
// one, which is what the settings start as.
func TestActionsPageNamesOnlyRealClashes(t *testing.T) {
	s := store.Default()
	s.Keybindings = store.DefaultKeybindings()
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageActions}, &navActions{})
	}, 900, 700)
	tt.Frame()
	if _, ok := tt.Find(locales.KeysConflict()); ok {
		t.Fatal("the page says a key runs two actions while every key runs one")
	}
	s.Keybindings = append(s.Keybindings, store.Keybinding{Key: "ctrl+t", Action: action.TabClose})
	tt.Frame()
	if _, ok := tt.Find(locales.KeysConflict()); !ok {
		t.Fatal("two actions on ctrl+t are not named as a clash")
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

// TestNavLeavesHiddenProfilesOut covers the column: it opens the pages of the
// profiles in use, so a hidden one is listed on the page that brings it back
// alone, whatever the switch above that list says.
func TestNavLeavesHiddenProfilesOut(t *testing.T) {
	tt, _ := settingsOf(t, []Profile{
		{ID: "a", Name: "PowerShell 7"},
		{ID: "b", Name: "Command Prompt", Hidden: true},
	})
	in, ok := tt.Find("PowerShell 7")
	if !ok {
		t.Fatal("the column leaves out a profile in use")
	}
	gone, ok := tt.Find("Command Prompt")
	if !ok {
		// The page lists it, greyed out: the column is what this test reads
		// for it, so the one it found is the page's.
		t.Fatal("the page does not list the hidden profile, so the switch lies")
	}
	if gone.X < navWidth {
		t.Fatalf("the hidden profile at %v is in the column", gone)
	}
	if in.X >= navWidth {
		t.Fatalf("the profile in use at %v is not in the column", in)
	}
}
