package ui

import (
	myui "github.com/egoist/mygo/ui"
	"testing"
	"wtgo/internal/store"
	"wtgo/locales"
)

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
