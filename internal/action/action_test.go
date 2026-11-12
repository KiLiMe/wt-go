package action_test

import (
	"strings"
	"testing"

	"wtgo/internal/action"
)

// TestCatalog covers the actions the window has: unique ids, a title and a
// group each, and the numbered ones spelled one through nine.
func TestCatalog(t *testing.T) {
	seen := map[string]bool{}
	for _, a := range action.All() {
		if seen[a.ID] {
			t.Errorf("%q is listed twice", a.ID)
		}
		seen[a.ID] = true
		if a.TitleKey == "" || a.GroupKey == "" {
			t.Errorf("%q has no title or group: %+v", a.ID, a)
		}
	}
	for n := 1; n <= action.MaxNumber; n++ {
		for _, id := range []string{action.GotoAction(n), action.ProfileAction(n)} {
			if !seen[id] {
				t.Errorf("%q is not listed", id)
			}
		}
	}
	for _, id := range []string{action.Find, action.Palette, action.Settings} {
		if !seen[id] {
			t.Errorf("%q is not listed", id)
		}
	}
}

// TestTitle covers the titles: the number an action names, the id of an action
// the window does not know, and the source language the window is in until the
// settings name another: what a language makes of them is locales' to check.
func TestTitle(t *testing.T) {
	if got := action.Title(action.GotoAction(3)); got != "Switch to Tab 3" {
		t.Errorf("Title(tab.goto.3) = %q", got)
	}
	if got := action.Title(action.ProfileAction(9)); got != "Open Profile 9" {
		t.Errorf("Title(profile.9) = %q", got)
	}
	if got := action.Title(action.Find); got != "Find" {
		t.Errorf("Title(find) = %q", got)
	}
	if got := action.Title("no.such.action"); got != "no.such.action" {
		t.Errorf("Title of an action the window does not know = %q", got)
	}
	if got := action.Group("no.such.action"); got != "" {
		t.Errorf("Group of an action the window does not know = %q", got)
	}
	if got := action.Title(action.GotoAction(2)); !strings.Contains(got, "2") {
		t.Errorf("Title(tab.goto.2) kept no number: %q", got)
	}
}

// TestGotoAction covers the tab.goto.n actions.
func TestGotoAction(t *testing.T) {
	for n := 1; n <= action.MaxNumber; n++ {
		id := action.GotoAction(n)
		i, ok := action.GotoIndex(id)
		if !ok || i != n-1 {
			t.Errorf("GotoIndex(%q) = %d, %v", id, i, ok)
		}
	}
	for _, id := range []string{"tab.new", "tab.goto.", "tab.goto.x", "tab.goto.0"} {
		if _, ok := action.GotoIndex(id); ok {
			t.Errorf("GotoIndex(%q) found a tab", id)
		}
	}
	if _, ok := action.ProfileIndex("profile.3"); !ok {
		t.Error("ProfileIndex(profile.3) found no profile")
	}
	if _, ok := action.ProfileIndex("profile.x"); ok {
		t.Error("ProfileIndex(profile.x) found a profile")
	}
}
