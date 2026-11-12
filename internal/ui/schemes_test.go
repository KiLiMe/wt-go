package ui

import (
	"slices"
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
	"wtgo/locales"
)

// schemeActions is the window as the color schemes page sees it: it records the
// schemes the page asks to remove.
type schemeActions struct {
	Actions
	removed []string
}

func (a *schemeActions) RemoveScheme(name string) { a.removed = append(a.removed, name) }

// schemesOf draws the page over the settings and schemes it gives.
func schemesOf(t *testing.T, s *store.Settings, schemes []Scheme) (*myui.Tester, *schemeActions) {
	t.Helper()
	act := &schemeActions{}
	tt := myui.NewTester(func(c *myui.Context) {
		schemesPage(c, Settings{Draft: s, Schemes: schemes}, act)
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
	if s.ColorScheme != "Paper" {
		t.Errorf("scheme = %q, want the clicked %q", s.ColorScheme, "Paper")
	}
	if len(act.removed) != 0 {
		t.Errorf("removed %v, want none", act.removed)
	}
}
