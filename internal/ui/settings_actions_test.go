package ui

import (
	myui "github.com/egoist/mygo/ui"
	"testing"
	"wtgo/internal/action"
	"wtgo/internal/store"
	"wtgo/locales"
)

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
// binds twice: both keys are in it, and the hint holds them in full for the
// cell too narrow to show them.
func TestActionsPageShowsEveryKeyOfAnAction(t *testing.T) {
	s := store.Default()
	s.Keybindings = append(store.DefaultKeybindings(), store.Keybinding{Key: "ctrl+shift+t", Action: action.TabNew})
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageActions}, &navActions{})
	}, 900, 700)
	tt.Frame()
	if _, ok := tt.Find("ctrl+shift+t · ctrl+shift+t"); !ok {
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
	s.Keybindings = append(s.Keybindings,
		store.Keybinding{Key: "ctrl+alt+y", Action: action.TabClose},
		store.Keybinding{Key: "ctrl+alt+y", Action: action.TabNew})
	tt.Frame()
	if _, ok := tt.Find(locales.KeysConflict()); !ok {
		t.Fatal("two actions on one key are not named as a clash")
	}
}
