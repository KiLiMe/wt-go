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
	page     string
	defaults []string
}

func (a *navActions) AddProfile()                    { a.added++ }
func (a *navActions) ShowHiddenProfiles()            { a.hidden++ }
func (a *navActions) ToggleChangedKeys()             { a.toggled++ }
func (a *navActions) SelectSettingsPage(page string) { a.page = page }
func (a *navActions) SetDefaultProfile(id string)    { a.defaults = append(a.defaults, id) }

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
		// Tall enough for the group the two switches sit in: the sections of
		// the window above it put them past the default height.
	}, 900, 1400)
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
		SettingsTab(c, Settings{Draft: &s, Page: PageAppearance, Validation: reason, Dirty: true}, &navActions{})
		// Tall enough for the first field of the page to be in view under the
		// banner, rather than unfolded out of the scroll.
	}, 900, 1100)
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
	first, ok := tt.Find(locales.Theme())
	if !ok {
		t.Fatal("the page does not show its fields")
	}
	if reasonAt.Y >= first.Y {
		t.Fatalf("the reason at %v is under the field it is about at %v", reasonAt.Y, first.Y)
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

// TestProfilesPageAddsAndOpens covers the page: the profiles are rows of the
// page that lists them beside the one in front, not pages of the column, so
// adding one never grows the column and the fields of one open beside it.
func TestProfilesPageAddsAndOpens(t *testing.T) {
	var profiles []Profile
	for i := 0; i < 22; i++ {
		profiles = append(profiles, Profile{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Profile %d", i)})
	}
	s := store.Default()
	act := &navActions{}
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageProfiles, Profiles: profiles}, act)
	}, 900, 2000)
	tt.Frame()
	add, ok := tt.AccessByRole(myui.RoleButton, locales.AddProfile())
	if !ok {
		t.Fatal("the page has no button that adds a profile")
	}
	if add.Bounds.X < navWidth {
		t.Fatalf("the button at %v is in the column of pages", add.Bounds)
	}
	// The rows scroll under the button, so the first of them stands under it
	// and the last ones are still rows of the list rather than of a page of
	// their own that grew the column.
	first, ok := tt.Find("Profile 0")
	if !ok {
		t.Fatal("the page does not list the first profile")
	}
	if first.X < navWidth {
		t.Fatalf("the first profile at %v is in the column of pages", first)
	}
	if add.Bounds.Y >= first.Y {
		t.Fatalf("the button at %v is not above the list at %v", add.Bounds, first)
	}
	tt.ClickAt(first.X+first.W/2, first.Y+first.H/2)
	if act.page != ProfilePage("p0") {
		t.Fatalf("page = %q, want %q", act.page, ProfilePage("p0"))
	}
}

// settingsSearched draws the settings tab with a box that searches the
// settings, over an empty draft of the profiles page: the column is what the
// search changes, so the page beside it is the one that has nothing to say.
func settingsSearched(t *testing.T, search *string) (*myui.Tester, *navActions) {
	t.Helper()
	s := store.Default()
	act := &navActions{}
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageGeneral, Search: search}, act)
	}, 900, 700)
	tt.Frame()
	return tt, act
}

// TestSettingsSearchFindsTheFieldOfAPage covers the box over the column of
// pages: a word finds the setting it names, in either language it is written
// in and by the word a person may know it by, and a click on what it found
// opens the page that holds the field.
func TestSettingsSearchFindsTheFieldOfAPage(t *testing.T) {
	for _, tc := range []struct{ words, want string }{
		{"光标", PageAppearance},
		{"cursor", PageAppearance},
		{"opacity", PageAppearance},
		{"scrollback", PageInteraction},
		{"paste", PageInteraction},
		{"剪贴板", PageInteraction},
		{"tab", PageGeneral},
		{"theme", PageAppearance},
		{"scheme", PageAppearance},
		{"keybinding", PageActions},
		{"reset", PageAdvanced},
	} {
		t.Run(tc.words, func(t *testing.T) {
			search := tc.words
			tt, act := settingsSearched(t, &search)
			// The column lists what it found instead of the six pages, and a
			// word every page shares, as "tab" does, finds more than the pages
			// a person sees: the first hit is where the search looks first.
			if tt.HasText(locales.NoMatchingSettings()) {
				t.Fatalf("%q found nothing: %q", tc.words, tt.Texts())
			}
			hits := searchSettingsIndex(tc.words)
			if len(hits) == 0 {
				t.Fatalf("%q found nothing", tc.words)
			}
			found := false
			for _, h := range hits {
				if h.page == tc.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("%q does not find a setting of %q: %v", tc.words, tc.want, hits)
			}
			// A click on the first hit asks for the page it sits on.
			at, ok := tt.Find(hits[0].name)
			if !ok {
				t.Fatalf("the column does not list %q, which it found: %q", hits[0].name, tt.Texts())
			}
			if at.X >= navWidth {
				t.Fatalf("%q at %v is not in the column", hits[0].name, at)
			}
			tt.ClickAt(at.X+at.W/2, at.Y+at.H/2)
			if act.page != hits[0].page {
				t.Fatalf("page = %q, want %q", act.page, hits[0].page)
			}
		})
	}
}

// TestSettingsSearchOpensTheAppearancePage covers the word a person is most
// likely to type: the cursor is a setting of the appearance page, and a click
// on it opens that page.
func TestSettingsSearchOpensTheAppearancePage(t *testing.T) {
	search := "cursor"
	tt, act := settingsSearched(t, &search)
	// The hit is not the only row there: every other field holding the word
	// lists under it, so what is checked here is the one asked for.
	at, ok := tt.Find(locales.Cursor())
	if !ok {
		t.Fatalf("the search does not find the cursor: %q", tt.Texts())
	}
	if !tt.HasText(locales.Appearance()) {
		t.Fatalf("the row does not say which page it is on: %q", tt.Texts())
	}
	tt.ClickAt(at.X+at.W/2, at.Y+at.H/2)
	if act.page != PageAppearance {
		t.Fatalf("page = %q, want %q", act.page, PageAppearance)
	}
}

// TestSettingsSearchSaysWhenNothingMatches covers a word no setting holds: the
// column says so rather than go empty.
func TestSettingsSearchSaysWhenNothingMatches(t *testing.T) {
	search := "zzzz"
	tt, _ := settingsSearched(t, &search)
	if !tt.HasText(locales.NoMatchingSettings()) {
		t.Fatalf("the column says nothing of a search that found nothing: %q", tt.Texts())
	}
	// A hit is not the only row there: every row under the column's clip is
	// a text of the frame, so a page name is only a row when it has no box.
	if at, ok := tt.Find(locales.General()); ok && at.X < navWidth {
		t.Fatal("the column lists the pages under a search that found nothing")
	}
}

// TestSettingsSearchKeepsWhatIsTyped covers the box itself: what is typed goes
// into the string the page is handed, the frames after it still read it there,
// and emptying it lists the six pages again.
func TestSettingsSearchKeepsWhatIsTyped(t *testing.T) {
	search := ""
	tt, _ := settingsSearched(t, &search)
	// A text input shows nothing to find by while it is empty, so it is
	// reached where its field is drawn.
	box, ok := tt.AccessByRole(myui.RoleTextField, "")
	if !ok {
		t.Fatalf("the column has no search box: %q", tt.Texts())
	}
	if box.Bounds.X >= navWidth {
		t.Fatalf("the search box at %v is not over the column", box.Bounds)
	}
	x, y := box.Bounds.X+box.Bounds.W/2, box.Bounds.Y+box.Bounds.H/2
	tt.Press(x, y)
	tt.Release(x, y)
	tt.Frame()
	if box, ok = tt.AccessByRole(myui.RoleTextField, ""); !ok || !box.Focused {
		t.Fatalf("a click at (%v, %v) leaves the search box without the keys", x, y)
	}
	tt.Type("cursor")
	if search != "cursor" {
		t.Fatalf("search = %q, want what was typed %q", search, "cursor")
	}
	tt.Frame()
	tt.Frame()
	if search != "cursor" {
		t.Errorf("search = %q, want the next frames kept what was typed", search)
	}
	for _, w := range []string{"cursor", "CURSOR"} {
		if got := searchSettingsIndex(w); len(got) == 0 {
			t.Errorf("%q found nothing", w)
		}
	}

	// Emptying the box lists the pages again: the word comes out a letter at a
	// time, which is what the field draws as it is taken out.
	tt.Press(x, y)
	tt.Release(x, y)
	for range "cursor" {
		tt.Key(0, myui.KeyBackspace)
	}
	tt.Frame()
	if search != "" {
		t.Fatalf("search = %q, the box kept what was taken out of it", search)
	}
	for _, w := range []string{
		locales.General(), locales.Interaction(), locales.Appearance(),
		locales.Profiles(), locales.Actions(), locales.Advanced(),
	} {
		if _, ok := tt.Find(w); !ok {
			t.Fatalf("the page %q is missing from the column once the box is empty: %q", w, tt.Texts())
		}
	}
	if tt.HasText(locales.NoMatchingSettings()) {
		t.Fatal("the column says nothing matches while it lists the pages")
	}
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

// TestProfileListLeavesHiddenProfilesOut covers the list: it holds the
// profiles the settings offer and leaves out the ones they hide, until the
// switch above it lists them too, greyed out, so that one comes back from here.
func TestProfileListLeavesHiddenProfilesOut(t *testing.T) {
	// A list long enough to scroll, so that the switch stands where it holds:
	// above the rows it filters rather than scrolled away under them.
	var profiles []Profile
	for i := 0; i < 22; i++ {
		profiles = append(profiles, Profile{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Profile %d", i)})
	}
	listed := profiles
	s := store.Default()
	// The window keeps the switch and hands the page the profiles it lists: the
	// test stands in for it, adding the hidden one once the switch is on.
	act := &navActions{}
	redraw := func() *myui.Tester {
		t.Helper()
		tt := myui.NewTester(func(c *myui.Context) {
			SettingsTab(c, Settings{Draft: &s, Page: PageProfiles, Profiles: listed}, act)
		}, 900, 2000)
		tt.Frame()
		return tt
	}
	tt := redraw()
	if _, ok := tt.Find("gone"); ok {
		t.Fatalf("the list shows a hidden profile while the switch says it hides them: %q", tt.Texts())
	}
	sw, ok := tt.AccessByRole(myui.RoleSwitch, locales.ShowHiddenProfiles())
	if !ok {
		t.Fatal("the list has no switch to show the hidden profiles")
	}
	if sw.Bounds.X < navWidth {
		t.Fatalf("the switch at %v is in the column of pages", sw.Bounds)
	}
	if first, ok := tt.Find("Profile 0"); !ok {
		t.Fatal("the list does not show its first row")
	} else if first.Y <= sw.Bounds.Y {
		t.Fatalf("the switch at %v is under the row at %v", sw.Bounds, first)
	}
	tt.ClickAt(sw.Bounds.X+sw.Bounds.W/2, sw.Bounds.Y+sw.Bounds.H/2)
	if act.hidden != 1 {
		t.Fatalf("hidden = %d, want 1", act.hidden)
	}
	listed = append(append([]Profile(nil), profiles...), Profile{ID: "gone", Name: "gone", Hidden: true})
	tt = redraw()
	if _, ok := tt.Find("gone"); !ok {
		t.Fatalf("the list does not show a hidden profile the switch asked for: %q", tt.Texts())
	}
}
