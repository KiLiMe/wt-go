package app

import (
	"testing"

	"wtgo/internal/action"
	"wtgo/internal/store"
)

// TestHeldKeys covers which keys keep firing while they stay held: the ones
// that step through the tabs and the panes, and no others. A held Ctrl+T
// making a tab per repeat is what the terminal filled up with.
func TestHeldKeys(t *testing.T) {
	repeats := map[string]bool{}
	for _, b := range bindings(store.DefaultKeybindings()) {
		repeats[b.action] = b.repeats
	}
	steps := map[string]bool{
		action.TabNext:        true,
		action.TabPrev:        true,
		action.PaneFocusLeft:  true,
		action.PaneFocusRight: true,
		action.PaneFocusUp:    true,
		action.PaneFocusDown:  true,
		action.TabNew:         false,
		action.TabClose:       false,
		action.TabDuplicate:   false,
		action.PaneSplitAuto:  false,
		action.PaneClose:      false,
		action.Find:           false,
		action.Palette:        false,
		action.Settings:       false,
	}
	for id, want := range steps {
		got, bound := repeats[id]
		if !bound {
			continue // no default key binds it, so no key can repeat it either
		}
		if got != want {
			t.Errorf("%q repeats = %v, want %v", id, got, want)
		}
	}
}
