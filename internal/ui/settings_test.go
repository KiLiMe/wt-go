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
	added    int
	hidden   int
	toggled  int
	defaults []string
}

func (a *navActions) AddProfile()                 { a.added++ }
func (a *navActions) ShowHiddenProfiles()         { a.hidden++ }
func (a *navActions) ToggleChangedKeys()          { a.toggled++ }
func (a *navActions) SetDefaultProfile(id string) { a.defaults = append(a.defaults, id) }

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
	if len(act.defaults) != 0 {
		t.Fatalf("drawing the page asked for %v, want nothing", act.defaults)
	}
	// A choice asks the window for it rather than writing the draft, which
	// the window is the one to do. The select opens on a press of its
	// trigger, which shows the option in force, the device, and a press on
	// another option closes the popup and names it. The press goes on the
	// option itself: a press on the row already in force changes nothing, so
	// the select does not report it.
	trigger, ok := tt.Find(locales.SystemProfile())
	if !ok {
		t.Fatal("the select for the default profile is missing")
	}
	x, y := trigger.X+trigger.W/2, trigger.Y+trigger.H/2
	tt.ClickAt(x, y)
	if n := countTexts(tt.Texts(), locales.SystemProfile()); n != 2 {
		t.Fatalf("the select opened %d choices for the device, want 2", n)
	}
	option, ok := tt.Find(known[0].Name)
	if !ok {
		t.Fatalf("the popup does not offer %q", known[0].Name)
	}
	// The window is told the profile's id, not the name the select shows.
	tt.ClickAt(option.X+option.W/2, option.Y+option.H/2)
	if len(act.defaults) != 1 || act.defaults[0] != known[0].ID {
		t.Fatalf("asked for %v, want %q", act.defaults, known[0].ID)
	}
	if s.DefaultProfile != "gone" {
		t.Errorf("the page wrote the draft itself: DefaultProfile = %q", s.DefaultProfile)
	}
}

// TestStartupPageAsksBeforeTheWindowCloses covers the switch of
// confirmCloseWindow: it sits under the switch that closes the window, since
// with that one off there is no window close to ask about, and a click edits
// the draft.
func TestStartupPageAsksBeforeTheWindowCloses(t *testing.T) {
	s := store.Default()
	s.ConfirmCloseWindow = false
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageStartup}, &navActions{})
	}, 900, 700)
	tt.Frame()
	sw := switchOf(t, tt, locales.ConfirmCloseWindow())
	last := switchOf(t, tt, locales.CloseOnLastTab())
	if sw.Bounds.Y <= last.Bounds.Y {
		t.Fatalf("the question at %v is not under the switch it goes with at %v", sw.Bounds, last.Bounds)
	}
	tt.ClickAt(sw.Bounds.X+sw.Bounds.W/2, sw.Bounds.Y+sw.Bounds.H/2)
	if !s.ConfirmCloseWindow {
		t.Fatal("a click on the switch left the draft as it was")
	}
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

// TestSettingsBannerShowsTheRefusal covers what the page says when a save
// refused the draft: the reason is above the fields, in the danger color, so it
// is not cut off by the foot of the page or hidden by whatever the last action
// said, and the foot says what to do about it.
func TestSettingsBannerShowsTheRefusal(t *testing.T) {
	s := store.Default()
	const reason = "settings: font.size must be between 4 and 200"
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageProfiles, Validation: reason, Dirty: true}, &navActions{})
	}, 900, 700)
	tt.Frame()
	reasonAt, ok := tt.Find(reason)
	if !ok {
		t.Fatalf("the page does not say why the save was refused: %q", tt.Texts())
	}
	if !tt.HasText(locales.SaveRefusedNote()) {
		t.Fatalf("the foot does not say what to do about the refusal: %q", tt.Texts())
	}
	if tt.HasText(locales.UnsavedChanges()) {
		t.Error("the foot says the draft is unsaved rather than what is wrong with it")
	}
	// Above the fields: past the column of pages, and above the first field
	// of the page the save was refused on.
	if reasonAt.X < navWidth {
		t.Fatalf("the reason at %v is in the column of pages", reasonAt)
	}
	font, ok := tt.Find(locales.Font())
	if !ok {
		t.Fatal("the page does not show its fields")
	}
	if reasonAt.Y >= font.Y {
		t.Fatalf("the reason at %v is under the field it is about at %v", reasonAt.Y, font.Y)
	}
}

// TestSettingsWithoutARefusalShowsNone covers the page of a draft nothing is
// wrong with: no banner, and the foot says the draft is unsaved.
func TestSettingsWithoutARefusalShowsNone(t *testing.T) {
	s := store.Default()
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageProfiles, Dirty: true}, &navActions{})
	}, 900, 700)
	tt.Frame()
	if tt.HasText(locales.SaveRefusedNote()) {
		t.Fatal("the page says a save was refused while none was")
	}
	if !tt.HasText(locales.UnsavedChanges()) {
		t.Fatalf("the foot does not say the draft is unsaved: %q", tt.Texts())
	}
}

// TestProfilePageHintsAnEmptyCommand covers a profile that runs the shell of
// the machine: the command field says what an empty one does, and the hop a
// program of a person's own starts from.
func TestProfilePageHintsAnEmptyCommand(t *testing.T) {
	s := store.Default()
	p := store.Profile{ID: "p1", Name: "Mine"}
	args := ""
	hint := locales.CommandEmptyHint(store.DefaultShellHint())
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: ProfilePage(p.ID), Profile: &p, Args: &args}, &navActions{})
	}, 900, 700)
	tt.Frame()
	if !tt.HasText(hint) {
		t.Fatalf("an empty command says nothing about what it runs: %q", tt.Texts())
	}
	if tt.HasText(locales.DiscoveredProfileNote()) {
		t.Fatal("a profile of a person's own is said to be the window's")
	}

	p.Command = "pwsh.exe"
	tt.Frame()
	if tt.HasText(hint) {
		t.Fatal("a command the profile names still reads as empty")
	}
}

// TestProfilePageLeavesTheCommandOfADiscoveredProfile covers the profile the
// window found: its command is the path the window knows, so the page does not
// hint at emptying it and does not offer to change it.
func TestProfilePageLeavesTheCommandOfADiscoveredProfile(t *testing.T) {
	s := store.Default()
	p := store.Profile{ID: "cmd", Name: "Command Prompt", Command: `C:\Windows\System32\cmd.exe`, Discovered: true}
	args := ""
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: ProfilePage(p.ID), Profile: &p, Args: &args}, &navActions{})
	}, 900, 700)
	tt.Frame()
	if !tt.HasText(locales.DiscoveredProfileNote()) {
		t.Fatalf("the page does not say the command is the window's: %q", tt.Texts())
	}
	if tt.HasText(locales.CommandEmptyHint(store.DefaultShellHint())) {
		t.Fatal("a discovered profile reads as one that runs the system shell")
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
