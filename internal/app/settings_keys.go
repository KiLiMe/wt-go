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
		out := make([]store.Keybinding, 0, len(a.settings.Keybindings)+1)
		found := false
		for _, kb := range a.settings.Keybindings {
			if kb.Action != action {
				out = append(out, kb)
				continue
			}
			if !found {
				// The first keybinding of the action keeps its place in the
				// file, so that the list does not reorder as keys change.
				out = append(out, store.Keybinding{Key: key, Action: action})
				found = true
			}
		}
		if !found {
			out = append(out, store.Keybinding{Key: key, Action: action})
		}
		a.settings.Keybindings = out
	}
	a.recording = ""
	a.redraw()
}

// RemoveKeybinding drops the keybindings of an action, so that it runs on no
// key at all.
func (a *App) RemoveKeybinding(action string) {
	out := a.settings.Keybindings[:0]
	for _, kb := range a.settings.Keybindings {
		if kb.Action != action {
			out = append(out, kb)
		}
	}
	a.settings.Keybindings = out
	a.recording = ""
	a.redraw()
}

// ResetKeybinding gives an action the keys the catalog starts it on, in the
// place of the first key it ran on: one row at a time, so that the rest of the
// file, and where the action sat in it, stay as they are.
func (a *App) ResetKeybinding(action string) {
	rest := make([]store.Keybinding, 0, len(a.settings.Keybindings))
	at := -1
	for _, kb := range a.settings.Keybindings {
		if kb.Action == action {
			if at < 0 {
				at = len(rest)
			}
			continue
		}
		rest = append(rest, kb)
	}
	if at < 0 {
		at = len(rest)
	}
	out := make([]store.Keybinding, 0, len(rest)+1)
	out = append(out, rest[:at]...)
	for _, kb := range store.DefaultKeybindings() {
		if kb.Action == action {
			out = append(out, kb)
		}
	}
	out = append(out, rest[at:]...)
	a.settings.Keybindings = out
	a.recording = ""
	a.redraw()
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
