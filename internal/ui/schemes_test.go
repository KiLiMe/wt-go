package ui

import (
	"slices"
	"strings"
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
	"wtgo/locales"
)

// schemeActions is the window as the color schemes page sees it: it records the
// schemes the page asks to remove, and the ones it asks to pick.
type schemeActions struct {
	Actions
	removed []string
	picked  []string
}

func (a *schemeActions) RemoveScheme(name string) { a.removed = append(a.removed, name) }
func (a *schemeActions) PickScheme(name string)   { a.picked = append(a.picked, name) }

// schemesOf draws the page over the settings and schemes it gives.
func schemesOf(t *testing.T, s *store.Settings, schemes []Scheme) (*myui.Tester, *schemeActions) {
	t.Helper()
	return schemesOfFilter(t, s, schemes, nil)
}

// schemesOfFilter draws the page over the settings and schemes it gives, with
// filter as the words the search box edits and writes back into.
func schemesOfFilter(t *testing.T, s *store.Settings, schemes []Scheme, filter *string) (*myui.Tester, *schemeActions) {
	t.Helper()
	act := &schemeActions{}
	tt := myui.NewTester(func(c *myui.Context) {
		schemesPage(c, Settings{Draft: s, Schemes: schemes, Filter: filter}, act)
	}, 900, 700)
	tt.Frame()
	return tt, act
}

// TestSchemeRowRemoves covers the button a scheme the settings write carries:
// it shows while the pointer is over its row, and a click on it asks to remove
// that scheme alone, leaving the pick where it was.
func TestSchemeRowRemoves(t *testing.T) {
	s := store.Default()
	s.ColorScheme = "Campbell"
	s.Schemes = []store.Scheme{{
		Name:       "Paper",
		Background: "#fafafa",
		Foreground: "#111111",
		Palette:    []string{"#000000"},
	}}
	schemes := []Scheme{
		{Name: "Paper", Custom: true},
		{Name: "Campbell"},
	}
	tt, act := schemesOf(t, &s, schemes)

	label := locales.SchemeRemove()
	if _, ok := tt.Find(label); ok {
		t.Fatalf("the button shows while the pointer is elsewhere: %q", tt.Texts())
	}

	name, ok := tt.Find("Paper")
	if !ok {
		t.Fatalf("the list does not show the scheme: %q", tt.Texts())
	}
	tt.Move(name.X+name.W/2, name.Y+name.H/2)
	tt.Frame()
	if err := tt.Click(label); err != nil {
		t.Fatalf("the row under the pointer offers no way to remove it: %v", err)
	}
	if want := []string{"Paper"}; !slices.Equal(act.removed, want) {
		t.Errorf("removed %v, want %v", act.removed, want)
	}
	if s.ColorScheme != "Campbell" {
		t.Errorf("scheme = %q, want the pick left at %q", s.ColorScheme, "Campbell")
	}
	if len(act.picked) != 0 {
		t.Errorf("picked %v, want none: the button removes, it does not pick", act.picked)
	}
}

// TestSchemeSearchKeepsWhatIsTyped covers the search box: what is typed into
// it goes into the string the page is handed, and the frame drawn after that
// still finds it there, both in the box and in the list it narrows.
func TestSchemeSearchKeepsWhatIsTyped(t *testing.T) {
	s := store.Default()
	s.ColorScheme = "Campbell"
	filter := ""
	tt, _ := schemesOfFilter(t, &s, []Scheme{{Name: "Paper"}, {Name: "Campbell"}}, &filter)
	// The pick above the list shows the chosen scheme too, so counting the
	// name is what tells the list apart: it shows twice before the search.
	count := func() int { return countTexts(tt.Texts(), "Campbell") }
	if got := count(); got != 2 {
		t.Fatalf("%q shows %d times, want the pick and the list: %q", "Campbell", got, tt.Texts())
	}

	// A text input shows nothing to find by while it is empty, so it is
	// reached where its field is drawn.
	box, ok := tt.AccessByRole(myui.RoleTextField, "")
	if !ok {
		t.Fatalf("the page shows no search box: %q", tt.Texts())
	}
	x, y := box.Bounds.X+box.Bounds.W/2, box.Bounds.Y+box.Bounds.H/2
	tt.Press(x, y)
	tt.Release(x, y)
	tt.Frame()
	if box, ok = tt.AccessByRole(myui.RoleTextField, ""); !ok || !box.Focused {
		t.Fatalf("a click at (%v, %v) leaves the search box without the keys", x, y)
	}
	tt.Type("Pap")
	if filter != "Pap" {
		t.Fatalf("filter = %q, want what was typed %q", filter, "Pap")
	}

	// A page that gave the box a string of its own every frame would show
	// the words for the frame they were typed in and lose them after that,
	// leaving the scheme the search does not ask for in the list.
	tt.Frame()
	tt.Frame()
	if filter != "Pap" {
		t.Errorf("filter = %q, want the next frames kept what was typed %q", filter, "Pap")
	}
	if got := count(); got != 1 {
		t.Errorf("%q shows %d times, want only the pick: the list was not narrowed: %q", "Campbell", got, tt.Texts())
	}
	// The value of a text field is what it holds, not its name: a field's
	// name is its label, and this one is labeled by nothing on screen.
	box, ok = tt.AccessByRole(myui.RoleTextField, "")
	if !ok || !strings.Contains(box.Value, "Pap") {
		t.Errorf("the search box reads %q, want the typed %q", box.Value, "Pap")
	}
}

// countTexts is how many of the texts are s.
func countTexts(texts []string, s string) int {
	n := 0
	for _, text := range texts {
		if text == s {
			n++
		}
	}
	return n
}

// TestSchemesWithoutFilterDraws covers a page drawn with no string to search
// with: the box would have nowhere to write what is typed, so it is left out,
// and the page still draws and still picks schemes.
func TestSchemesWithoutFilterDraws(t *testing.T) {
	s := store.Default()
	s.ColorScheme = "Campbell"
	tt, act := schemesOf(t, &s, []Scheme{{Name: "Paper"}, {Name: "Campbell"}})

	// The search box can only be there if it has a string to edit: one made
	// up for the frame would take what is typed down with it.
	if _, ok := tt.AccessByRole(myui.RoleTextField, ""); ok {
		t.Error("the page offers a search box with no string to search with")
	}
	name, ok := tt.Find("Paper")
	if !ok {
		t.Fatalf("the list does not show the scheme: %q", tt.Texts())
	}
	tt.ClickAt(name.X+name.W/2, name.Y+name.H/2)
	if want := []string{"Paper"}; !slices.Equal(act.picked, want) {
		t.Errorf("picked %v, want %v", act.picked, want)
	}
	if s.ColorScheme != "Campbell" {
		t.Errorf("scheme = %q, want the draft left as it was: %q", s.ColorScheme, "Campbell")
	}
	if len(act.removed) != 0 {
		t.Errorf("removed %v, want none", act.removed)
	}
}

// TestSchemeRowPicks covers the rest of a row: a click on the name picks the
// scheme, and a scheme the settings do not write has no button to remove it.
func TestSchemeRowPicks(t *testing.T) {
	s := store.Default()
	s.ColorScheme = "Campbell"
	tt, act := schemesOf(t, &s, []Scheme{{Name: "Paper"}, {Name: "Campbell"}})

	name, ok := tt.Find("Paper")
	if !ok {
		t.Fatalf("the list does not show the scheme: %q", tt.Texts())
	}
	tt.Move(name.X+name.W/2, name.Y+name.H/2)
	tt.Frame()
	if _, ok := tt.Find(locales.SchemeRemove()); ok {
		t.Error("a scheme the settings do not write offers to be removed")
	}
	if err := tt.Click("Paper"); err != nil {
		t.Fatal(err)
	}
	if want := []string{"Paper"}; !slices.Equal(act.picked, want) {
		t.Errorf("picked %v, want %v", act.picked, want)
	}
	if s.ColorScheme != "Campbell" {
		t.Errorf("scheme = %q, want the draft left as it was: %q", s.ColorScheme, "Campbell")
	}
	if len(act.removed) != 0 {
		t.Errorf("removed %v, want none", act.removed)
	}
}
