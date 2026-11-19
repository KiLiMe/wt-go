package ui

import (
	"fmt"
	myui "github.com/egoist/mygo/ui"
	"testing"
	"wtgo/internal/store"
	"wtgo/locales"
)

// navActions is the window as the settings column sees it: it counts the
// profiles it asks to add, and the keybinding changes it is asked for.
type navActions struct {
	Actions
	added    int
	hidden   []string
	reset    []string
	marks    []string
	toggled  int
	page     string
	key      string
	defaults []string
}

func (a *navActions) AddProfile() { a.added++ }

func (a *navActions) HideProfile(id string) { a.hidden = append(a.hidden, id) }

func (a *navActions) ResetProfile(id string) { a.reset = append(a.reset, id) }

func (a *navActions) ToggleChangedKeys() { a.toggled++ }

func (a *navActions) SelectSettingsPage(page string) { a.page = page }

// RevealSetting is what a hit of the search over the settings asks for: the
// page that holds the setting, and the setting itself, which the page marks.
func (a *navActions) RevealSetting(page, key string) {
	a.page = page
	a.key = key
}

func (a *navActions) SetDefaultProfile(id string) { a.defaults = append(a.defaults, id) }

func (a *navActions) SetProfileIcon(mark string) { a.marks = append(a.marks, mark) }

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
	// The head of the card holds what adds a profile above the list it opens,
	// so a list this long never puts it under the rows it lists.
	head, ok := tt.Find(locales.AddProfile())
	if !ok {
		t.Fatal("the page has no button that adds a profile")
	}
	tt.Scroll(600, 400, 0, 600)
	after, ok := tt.Find(locales.AddProfile())
	if !ok {
		t.Fatal("the button that adds a profile left the page while scrolling")
	}
	if first, ok := tt.Find("Profile 0"); ok && after.Y >= first.Y {
		t.Fatalf("the button at %v is under the row at %v", after, first)
	}
	if err := tt.Click(locales.AddProfile()); err != nil {
		t.Fatalf("the button that adds a profile cannot be clicked: %v", err)
	}
	if act.added != 2 {
		t.Fatalf("added = %d, want 2", act.added)
	}
	_ = head
}

// switchOf is the switch of the field a title labels: the field is one group
// to a reader, so the switch inside its box is the control it names.
func switchOf(t *testing.T, tt *myui.Tester, title string) myui.AccessInfo {
	t.Helper()
	g, ok := tt.Access(title)
	if !ok {
		t.Fatalf("the field %q is missing", title)
	}
	for _, a := range tt.AccessAll() {
		if a.Role != myui.RoleSwitch {
			continue
		}
		inside := a.Bounds.X >= g.Bounds.X && a.Bounds.Y >= g.Bounds.Y &&
			a.Bounds.X+a.Bounds.W <= g.Bounds.X+g.Bounds.W &&
			a.Bounds.Y+a.Bounds.H <= g.Bounds.Y+g.Bounds.H
		if inside {
			return a
		}
	}
	t.Fatalf("the field %q has no switch", title)
	return myui.AccessInfo{}
}

// TestNavListsTheSixPages covers the column: it holds the pages and nothing
// else, so that the profiles, which are one page with the one in front beside
// the list, never grow it.
func TestNavListsTheSixPages(t *testing.T) {
	tt, act := settingsOf(t, []Profile{
		{ID: "a", Name: "PowerShell 7", Discovered: true},
		{ID: "b", Name: "Command Prompt"},
	})
	want := []string{
		locales.General(), locales.Interaction(), locales.Appearance(),
		locales.Profiles(), locales.Actions(), locales.Advanced(),
	}
	var last float32
	for _, w := range want {
		at, ok := tt.Find(w)
		if !ok {
			t.Fatalf("the page %q is missing from the column", w)
		}
		if at.X >= navWidth {
			t.Fatalf("the page %q at %v is not in the column", w, at)
		}
		if at.Y <= last {
			t.Fatalf("the page %q at %v is not under the one before it", w, at.Y)
		}
		last = at.Y
	}
	// Nothing else stands in the column: the list of profiles and the button
	// that adds one belong to the page beside it.
	for _, info := range tt.AccessAll() {
		if info.Bounds.X+info.Bounds.W <= 0 || info.Bounds.X >= navWidth {
			continue
		}
		if contains([]string{locales.AddProfile(), locales.Defaults(), "PowerShell 7", "Command Prompt"}, info.Name) {
			t.Fatalf("the column holds %q, which is not a page", info.Name)
		}
	}
	if err := tt.Click(locales.Profiles()); err != nil {
		t.Fatal(err)
	}
	if act.page != PageProfiles {
		t.Fatalf("page = %q, want %q", act.page, PageProfiles)
	}
}

// contains reports whether a list of names holds one.
func contains(all []string, name string) bool {
	for _, a := range all {
		if a == name {
			return true
		}
	}
	return false
}
