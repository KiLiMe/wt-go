package ui

import (
	"testing"

	"wtgo/internal/action"
	"wtgo/internal/store"
)

// TestGroups covers the Actions page's listing: every action of the catalog
// has a row, in the order the catalog lists it, with the row of its keybinding
// when the file binds one and none when it does not; a keybinding of an action
// the window does not know comes last, under Other.
func TestGroups(t *testing.T) {
	all := []store.Keybinding{
		{Key: "alt+shift+w", Action: action.PaneCloseOthers},
		{Key: "ctrl+nope", Action: "no.such.action"},
		{Key: "ctrl+comma", Action: action.Settings},
		{Key: "ctrl+t", Action: action.TabNew},
		{Key: "ctrl+shift+f", Action: action.Find},
	}
	got := groups(all)
	bound := map[string]int{
		action.PaneCloseOthers: 0,
		"no.such.action":       1,
		action.Settings:        2,
		action.TabNew:          3,
		action.Find:            4,
	}
	seen := map[string]bool{}
	rows := 0
	for _, g := range got {
		for _, e := range g.rows {
			rows++
			if seen[e.action] {
				t.Errorf("%q is listed twice", e.action)
			}
			seen[e.action] = true
			want, ok := bound[e.action]
			if !ok {
				want = -1
			}
			if e.row != want {
				t.Errorf("row of %q = %d, want %d", e.action, e.row, want)
			}
		}
	}
	if want := len(action.All()) + 1; rows != want {
		t.Errorf("the page lists %d rows, want the catalog's %d and the unknown one", rows, want)
	}
	if last := got[len(got)-1]; last.title != "Other" || len(last.rows) != 1 || last.rows[0].action != "no.such.action" {
		t.Errorf("the last group = %+v, want the unknown action under Other", last)
	}
	if n := rowCount(groups(nil)); n != len(action.All()) {
		t.Errorf("groups(nil) lists %d rows, want the catalog's %d", n, len(action.All()))
	}
}

// rowCount is how many rows the groups hold.
func rowCount(groups []actionGroup) int {
	n := 0
	for _, g := range groups {
		n += len(g.rows)
	}
	return n
}

// TestGroupsOrder covers the order the groups come in: the catalog's, not the
// file's, so that the page does not move rows around as the file changes.
func TestGroupsOrder(t *testing.T) {
	got := groups([]store.Keybinding{
		{Key: "ctrl+comma", Action: action.Settings},
		{Key: "ctrl+t", Action: action.TabNew},
	})
	var titles []string
	for _, g := range got {
		titles = append(titles, g.title)
	}
	var want []string
	seen := map[string]bool{}
	for _, a := range action.All() {
		if !seen[a.Group] {
			seen[a.Group] = true
			want = append(want, a.Group)
		}
	}
	if len(titles) != len(want) {
		t.Fatalf("groups = %v, want %v", titles, want)
	}
	for i := range want {
		if titles[i] != want[i] {
			t.Fatalf("groups = %v, want %v", titles, want)
		}
	}
}
