package app

import (
	"fmt"
	"log"

	"wtgo/internal/action"
	"wtgo/internal/fault"
	"wtgo/internal/store"
	"wtgo/internal/ui"
)

// bindings are the keybindings the settings name, as the keys they bind.
func bindings(keybindings []store.Keybinding) []binding {
	out := make([]binding, 0, len(keybindings))
	for _, kb := range keybindings {
		mods, key, ok := ui.ParseKey(kb.Key)
		if !ok {
			log.Printf("wt-go: %s", fault.Line(fault.BadKeybinding, fmt.Sprintf("keybinding %q", kb.Key)))
			continue
		}
		out = append(out, binding{mods: mods, key: key, action: kb.Action, repeats: action.Repeats(kb.Action)})
	}
	return out
}

// shortcut is the key the settings bind to an action, as a keybinding
// writes it, and "" when none binds it.
func (a *App) shortcut(action string) string {
	for _, b := range a.keys {
		if b.action != action {
			continue
		}
		key, ok := ui.FormatKey(b.mods, b.key)
		if !ok {
			return ""
		}
		return key
	}
	return ""
}
