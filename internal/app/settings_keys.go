package app

import (
	"wtgo/internal/store"
)

// RecordKeybinding waits for a key of an action; asking again for the one it
// waits for stops it.
func (a *App) RecordKeybinding(action string) {
	if a.recording == action {
		a.recording = ""
	} else {
		a.recording = action
	}
	a.redraw()
}

// SetKeybinding gives an action a key, which ends the wait for one: it
// replaces the keys the action had with that one, or adds the keybinding it did
// not have. An action runs on one key, so the keys it had go: leaving one
// behind would keep running it, invisibly, on a key the page shows as free.
func (a *App) SetKeybinding(action, key string) {
	if key != "" {
		a.withKeys(action, []store.Keybinding{{Key: key, Action: action}})
	}
	a.recording = ""
	a.redraw()
}

// RemoveKeybinding drops the keybindings of an action, so that it runs on no
// key at all.
func (a *App) RemoveKeybinding(action string) {
	a.withKeys(action, nil)
	a.recording = ""
	a.redraw()
}

// ResetKeybinding gives an action the keys the catalog starts it on, in the
// place of the first key it ran on: one row at a time, so that the rest of the
// file, and where the action sat in it, stay as they are.
func (a *App) ResetKeybinding(action string) {
	var keys []store.Keybinding
	for _, kb := range store.DefaultKeybindings() {
		if kb.Action == action {
			keys = append(keys, kb)
		}
	}
	a.withKeys(action, keys)
	a.recording = ""
	a.redraw()
}

// withKeys rebuilds the list with keys in place of the ones action had, at the
// place of the first: the list does not reorder as an action's keys change, so
// the file a person reads stays where it was. An action bound nowhere takes its
// keys at the end, being new to the list.
func (a *App) withKeys(action string, keys []store.Keybinding) {
	out := make([]store.Keybinding, 0, len(a.settings.Keybindings)+len(keys))
	placed := false
	for _, kb := range a.settings.Keybindings {
		if kb.Action != action {
			out = append(out, kb)
			continue
		}
		if !placed {
			out = append(out, keys...)
			placed = true
		}
	}
	if !placed {
		out = append(out, keys...)
	}
	a.settings.Keybindings = out
}

// ToggleChangedKeys narrows the actions page to the actions the file binds on
// keys of its own, or lists them all again.
func (a *App) ToggleChangedKeys() {
	a.onlyChangedKeys = !a.onlyChangedKeys
	a.redraw()
}

// ToggleNumberedKeys unfolds the numbered actions the actions page folds away,
// or folds them again.
func (a *App) ToggleNumberedKeys() {
	a.numberedKeys = !a.numberedKeys
	a.redraw()
}

// ResetKeybindings goes back to the keybindings the window starts with.
func (a *App) ResetKeybindings() {
	a.settings.Keybindings = store.DefaultKeybindings()
	a.recording = ""
	a.redraw()
}
