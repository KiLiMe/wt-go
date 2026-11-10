package ui

import (
	"reflect"
	"testing"

	"wtgo/internal/action"
	"wtgo/internal/store"
)

// TestGroups covers the Actions page's grouping: the actions the window knows
// come first, in the order the catalog lists them, and a keybinding of an
// action the window does not know comes last, under Other.
func TestGroups(t *testing.T) {
	all := []store.Keybinding{
		{Key: "alt+shift+w", Action: action.PaneCloseOthers},
		{Key: "ctrl+nope", Action: "no.such.action"},
		{Key: "ctrl+comma", Action: action.Settings},
		{Key: "ctrl+t", Action: action.TabNew},
		{Key: "ctrl+shift+f", Action: action.Find},
	}
	got := groups(all)
	want := []actionGroup{
		{title: action.GroupTab, rows: []int{3}},
		{title: action.GroupPane, rows: []int{0}},
		{title: action.GroupFind, rows: []int{4}},
		{title: action.GroupWindow, rows: []int{2}},
		{title: "Other", rows: []int{1}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("groups() = %+v, want %+v", got, want)
	}
	if len(groups(nil)) != 0 {
		t.Errorf("groups(nil) = %+v, want none", groups(nil))
	}
}
