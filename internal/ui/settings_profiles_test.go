package ui

import (
	"fmt"
	myui "github.com/egoist/mygo/ui"
	"testing"
	"wtgo/internal/icons"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/locales"
)

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
// TestProfilesPageFitsANarrowWindow covers the two halves of the profiles
// page: the list of profiles and the fields of the one in front sit beside each
// other while the fields have room, and the fields take a line of their own
// under the list when they do not, rather than be squeezed narrower than the
// controls in them, which draws them past the card.
func TestProfilesPageFitsANarrowWindow(t *testing.T) {
	for _, width := range []int{900, 620, 480} {
		s := store.Default()
		p := store.Profile{ID: "p", Name: "A name long enough to be cut", Command: "pwsh.exe"}
		args := ""
		tt := myui.NewTester(func(c *myui.Context) {
			SettingsTab(c, Settings{Draft: &s, Page: ProfilePage(p.ID), Profile: &p, Args: &args}, &navActions{})
		}, width, 900)
		tt.Frame()
		for _, a := range tt.AccessAll() {
			if a.Role != myui.RoleTextField {
				continue
			}
			if a.Bounds.X+a.Bounds.W > float32(width) {
				t.Fatalf("at %v the field %q is drawn to %v, past the window",
					width, a.Name, a.Bounds.X+a.Bounds.W)
			}
		}
	}
}

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
