package ui

import (
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
)

// TestShortcut covers the menu's shortcut hint: an action the window binds a
// key to gets it, and an action it does not, or a key that is not one, gets
// nothing and does not panic.
func TestShortcut(t *testing.T) {
	st := State{Shortcut: func(id string) string {
		switch id {
		case action.ProfileAction(2):
			return "ctrl+shift+2"
		case action.Settings:
			return "ctrl+comma"
		case "broken":
			return "ctrl+nope"
		}
		return ""
	}}
	it := &myui.MenuItem{}
	shortcut(it, st, action.ProfileAction(2))
	shortcut(it, st, action.Settings)
	shortcut(it, st, action.ProfileAction(9))
	shortcut(it, st, "broken")

	shortcut(it, State{}, action.Settings)
}
