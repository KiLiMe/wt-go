package ui

import (
	"fmt"
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
	"wtgo/internal/icons"
	"wtgo/internal/profile"
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
	defaults []string
}

func (a *navActions) AddProfile()                    { a.added++ }
func (a *navActions) HideProfile(id string)          { a.hidden = append(a.hidden, id) }
func (a *navActions) ResetProfile(id string)         { a.reset = append(a.reset, id) }
func (a *navActions) ToggleChangedKeys()             { a.toggled++ }
func (a *navActions) SelectSettingsPage(page string) { a.page = page }
func (a *navActions) SetDefaultProfile(id string)    { a.defaults = append(a.defaults, id) }
func (a *navActions) SetProfileIcon(mark string)     { a.marks = append(a.marks, mark) }

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

// TestProfilesListShowsEveryProfile covers the card the list is drawn in: the
// fields beside it are taller, so the card takes their height, and every row it
// lists is on screen and takes a click. A card with a height of its own, as a
// 100 DIP one, cropped the rows below what fit and left the list of four
// profiles looking empty.
func TestProfilesListShowsEveryProfile(t *testing.T) {
	profiles := []Profile{
		{ID: "pwsh7", Name: "PowerShell 7", Discovered: true},
		{ID: "powershell", Name: "Windows PowerShell", Discovered: true},
		{ID: "bash", Name: "Bash", Discovered: true},
		{ID: "cmd", Name: "Command Prompt", Discovered: true},
	}
	s := store.Default()
	p := store.Profile{ID: "pwsh7", Name: "PowerShell 7", Command: `C:\Program Files\PowerShell\7\pwsh.exe`}
	args := ""
	act := &navActions{}
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: ProfilePage(p.ID), Profile: &p, Args: &args, Profiles: profiles}, act)
	}, 1280, 700)
	tt.Frame()
	for _, profile := range profiles {
		act.page = ""
		if err := tt.Click(profile.Name); err != nil {
			t.Fatalf("%s is not on screen in the list: %v", profile.Name, err)
		}
		if act.page != ProfilePage(profile.ID) {
			t.Fatalf("a click on %s opened %q, want %q", profile.Name, act.page, ProfilePage(profile.ID))
		}
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

// TestProfileListKeepsHiddenProfiles covers the list: it holds every profile
// the file has, the hidden ones too and marked as such, since the page edits
// hiding in the fields beside it and the list is where hiding is undone.
func TestProfileListKeepsHiddenProfiles(t *testing.T) {
	var profiles []Profile
	for i := 0; i < 22; i++ {
		profiles = append(profiles, Profile{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Profile %d", i)})
	}
	profiles = append(profiles, Profile{ID: "gone", Name: "Gone", Hidden: true})
	s := store.Default()
	act := &navActions{}
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageProfiles, Profiles: profiles}, act)
	}, 900, 2000)
	tt.Frame()
	if _, ok := tt.Find("Gone"); !ok {
		t.Fatalf("the list leaves out the hidden profile: %q", tt.Texts())
	}
	marked := 0
	for _, a := range tt.AccessAll() {
		if a.Name == locales.HiddenProfile() {
			marked++
		}
	}
	if marked != 1 {
		t.Fatalf("the list marks %d rows as hidden, want the one it hides", marked)
	}
	// A marked row still takes a click: that is how it comes back.
	if err := tt.Click("Gone"); err != nil {
		t.Fatalf("the hidden profile cannot be opened: %v", err)
	}
	if act.page != ProfilePage("gone") {
		t.Fatalf("page = %q, want %q", act.page, ProfilePage("gone"))
	}
}

// TestProfilePageHidesTheProfileItEdits covers what hides a profile: the
// button under the fields asks for it, and a profile a person wrote keeps the
// delete beside it, which a shell the window finds again cannot take.
func TestProfilePageHidesTheProfileItEdits(t *testing.T) {
	for _, tc := range []struct {
		name       string
		discovered bool
		remove     bool
	}{
		{"a profile the window found", true, false},
		{"a profile a person wrote", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := store.Default()
			p := store.Profile{ID: "p", Name: "P", Command: "pwsh.exe", Discovered: tc.discovered}
			args := ""
			act := &navActions{}
			tt := myui.NewTester(func(c *myui.Context) {
				SettingsTab(c, Settings{Draft: &s, Page: ProfilePage("p"), Profile: &p, Args: &args}, act)
				// Tall enough that the fields and the buttons under them do not
				// share a row: a click on the one in front would land on the
				// page's own footer.
			}, 900, 1200)
			tt.Frame()
			// The fields name this setting too, so the button is the one asked
			// for by its role rather than by a word two controls share.
			hide, ok := tt.AccessByRole(myui.RoleButton, locales.Hidden())
			if !ok {
				t.Fatalf("the page does not offer to hide the profile: %q", tt.Texts())
			}
			tt.ClickAt(hide.Bounds.X+hide.Bounds.W/2, hide.Bounds.Y+hide.Bounds.H/2)
			if len(act.hidden) != 1 || act.hidden[0] != "p" {
				t.Fatalf("hidden = %v, want the profile in front", act.hidden)
			}
			removable := false
			for _, a := range tt.AccessAll() {
				if a.Role == myui.RoleButton && a.Name == locales.Remove() {
					removable = true
				}
			}
			if removable != tc.remove {
				t.Fatalf("the page offers delete = %v, want %v", removable, tc.remove)
			}
		})
	}
}

// TestAppearancePreviewKeepsTheHead covers the preview the appearance page
// draws: the setting that changes it is often just under it, so the preview
// keeps the head of the page while the fields scroll past, rather than leaving
// a person to scroll up after every change.
func TestAppearancePreviewKeepsTheHead(t *testing.T) {
	s := store.Default()
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageAppearance}, &navActions{})
	}, 900, 700)
	tt.Frame()
	line, ok := tt.Find(locales.LineTwo())
	if !ok {
		t.Fatalf("the appearance page draws no preview: %q", tt.Texts())
	}
	size, ok := tt.Find(locales.FontSize())
	if !ok {
		t.Fatalf("the appearance page lists no font size: %q", tt.Texts())
	}
	// The wheel turns over the fields, which is where a person's pointer is.
	tt.Scroll(600, 400, 0, 400)
	tt.Frame()
	held, ok := tt.Find(locales.LineTwo())
	if !ok {
		t.Fatalf("the preview is gone once the page scrolls: %q", tt.Texts())
	}
	if held.Y < 0 {
		t.Fatalf("the preview line at %v is above the page it should hold", held.Y)
	}
	if held.Y >= line.Y {
		t.Fatalf("the preview line went from %v to %v: it scrolled with the fields", line.Y, held.Y)
	}
	// The fields under it did scroll, so what held the preview is not the page
	// having nothing to scroll.
	if moved, ok := tt.Find(locales.FontSize()); ok && moved.Y >= size.Y {
		t.Fatalf("the font size went from %v to %v: the page did not scroll", size.Y, moved.Y)
	}
	// It paints above the fields that roll under it, and the pointer finds it
	// there rather than the field behind it.
	if err := tt.Click(locales.LineTwo()); err != nil {
		t.Fatalf("the preview cannot be clicked where it holds the head: %v", err)
	}
}

// TestProfileReset covers the button a built-in profile gets: a click asks the
// window to reset that profile, and a profile a person wrote has no such
// button, since there is nothing on this machine to build it from.
func TestProfileReset(t *testing.T) {
	draw := func(p store.Profile) (*myui.Tester, *navActions) {
		s := store.Default()
		args := ""
		act := &navActions{}
		listed := []Profile{{ID: p.ID, Name: p.Name, Discovered: p.Discovered}}
		tt := myui.NewTester(func(c *myui.Context) {
			SettingsTab(c, Settings{
				Draft: &s, Page: ProfilePage(p.ID), Profile: &p, Args: &args, Profiles: listed,
			}, act)
		}, 1100, 1700)
		tt.Frame()
		return tt, act
	}

	found := store.Profile{ID: "pwsh7", Name: "PowerShell 7", Command: `C:\pwsh.exe`, Discovered: true}
	tt, act := draw(found)
	if err := tt.Click(locales.ResetProfile()); err != nil {
		t.Fatalf("clicking reset: %v", err)
	}
	if len(act.reset) != 1 || act.reset[0] != "pwsh7" {
		t.Errorf("reset = %v, want the profile in front", act.reset)
	}

	mine := store.Profile{ID: "mine", Name: "我的", Command: "cmd.exe"}
	tt, _ = draw(mine)
	if _, ok := tt.Find(locales.ResetProfile()); ok {
		t.Error("a profile the person wrote has a reset button")
	}
}

// TestProfileIcon covers the field that picks the mark of a profile: a square
// is one of the built-in shapes, and the first square is the mark the command
// asks for, which is the file's "".
func TestProfileIcon(t *testing.T) {
	s := store.Default()
	p := store.Profile{ID: "p", Name: "P", Command: "pwsh.exe", Discovered: true}
	args, glyph := "", ""
	act := &navActions{}
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: ProfilePage("p"), Profile: &p, Args: &args, Glyph: &glyph}, act)
	}, 900, 1200)
	tt.Frame()

	click := func(name string) {
		t.Helper()
		a, ok := tt.Access(name)
		if !ok {
			t.Fatalf("the page does not offer %q: %q", name, tt.Texts())
		}
		tt.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	}
	click("Bash · arch")
	click(locales.IconAuto())
	want := []string{"bash:arch", ""}
	if len(act.marks) != len(want) {
		t.Fatalf("marks = %v, want %v", act.marks, want)
	}
	for i := range want {
		if act.marks[i] != want[i] {
			t.Fatalf("marks = %v, want %v", act.marks, want)
		}
	}
	// Every shape the profile page offers is one the tab strip can draw.
	for _, pr := range profile.Presets() {
		if icons.Preset(pr.Key) == nil {
			t.Errorf("no icon is drawn for %q", pr.Key)
		}
	}
}
