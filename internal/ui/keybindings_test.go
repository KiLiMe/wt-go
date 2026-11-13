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
	if last := got[len(got)-1]; last.title != action.GroupOther || len(last.rows) != 1 || last.rows[0].action != "no.such.action" {
		t.Errorf("the last group = %+v, want the unknown action under Other", last)
	}
	if n := rowCount(groups(nil)); n != len(action.All()) {
		t.Errorf("groups(nil) lists %d rows, want the catalog's %d", n, len(action.All()))
	}
}

// TestChangedActions covers what the page marks as a person's own: the
// actions the file binds on keys other than the catalog's, and the actions it
// binds that the window has none of.
func TestChangedActions(t *testing.T) {
	all := []store.Keybinding{{Key: "ctrl+comma", Action: action.Settings}}
	for _, kb := range store.DefaultKeybindings() {
		// One action on a key of its own alone, and the rest as they start.
		if kb.Action == action.TabNew {
			continue
		}
		if kb.Action != action.Settings {
			all = append(all, kb)
		}
	}
	all = append(all,
		store.Keybinding{Key: "f12", Action: action.TabNew},
		store.Keybinding{Key: "ctrl+nope", Action: "no.such.action"},
	)
	changed := changedActions(all)
	if changed[action.Settings] {
		t.Errorf("%q is marked changed, though the file binds its default key", action.Settings)
	}
	if !changed[action.TabNew] {
		t.Errorf("%q is not marked changed, though the file binds f12 alone", action.TabNew)
	}
	if !changed["no.such.action"] {
		t.Errorf("an action the window does not know is not marked changed")
	}
	if len(changed) != 2 {
		t.Errorf("changed = %v, want the two actions of the file", changed)
	}
}

// TestClashingKeys covers what the page says in red: a key the file binds to
// more than one action, and a key it binds once.
func TestClashingKeys(t *testing.T) {
	clashing := clashingKeys([]store.Keybinding{
		{Key: "ctrl+t", Action: action.TabNew},
		{Key: "ctrl+t", Action: action.TabClose},
		{Key: "ctrl+w", Action: action.TabClose},
	})
	if got := clashing["ctrl+t"]; got != 2 {
		t.Errorf("ctrl+t runs %d actions, want 2", got)
	}
	if got := clashing["ctrl+w"]; got != 1 {
		t.Errorf("ctrl+w runs %d actions, want 1", got)
	}
}

// TestListedRows covers what the page draws of a group: the numbered actions
// past the third folded away while the page is short, and all of them back once
// the page is asked for the changed ones alone or searches.
func TestListedRows(t *testing.T) {
	rows := []entry{
		{action: action.TabGoto + "1", row: -1},
		{action: action.TabGoto + "2", row: -1},
		{action: action.TabGoto + "3", row: -1},
		{action: action.TabGoto + "4", row: -1},
		{action: action.TabGoto + "5", row: -1},
		{action: action.Settings, row: -1},
	}
	changed := map[string]bool{action.TabGoto + "5": true}

	out, folded, more := listedRows(rows, nil, changed, Settings{}, "")
	if len(out) != 4 || folded != 2 || more != 2 {
		t.Errorf("folded = %d rows, %d folded, %d more, want 4, 2, 2", len(out), folded, more)
	}

	out, folded, more = listedRows(rows, nil, changed, Settings{KeysExpanded: true}, "")
	if len(out) != 6 || folded != 0 || more != 2 {
		t.Errorf("unfolded = %d rows, %d folded, %d more, want 6, 0, 2", len(out), folded, more)
	}

	out, folded, _ = listedRows(rows, nil, changed, Settings{OnlyChanged: true}, "")
	if len(out) != 1 || out[0].action != action.TabGoto+"5" || folded != 0 {
		t.Errorf("changed only = %v, want the one changed action", out)
	}

	out, _, _ = listedRows(rows, nil, changed, Settings{}, "tab.goto.2")
	if len(out) != 1 || out[0].action != action.TabGoto+"2" {
		t.Errorf("search = %v, want the one action its id names", out)
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
		if !seen[a.GroupKey] {
			seen[a.GroupKey] = true
			want = append(want, a.GroupKey)
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
