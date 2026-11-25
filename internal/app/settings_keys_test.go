package app

import (
	"testing"

	"wtgo/internal/action"
	"wtgo/internal/store"
)

// TestSetKeybinding covers recording a key: it replaces the one the action
// had, binds one the defaults do not, and removing the action drops every
// keybinding it has.
func TestSetKeybinding(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	n := len(a.settings.Keybindings)

	a.RecordKeybinding(action.About)
	if a.recording != action.About {
		t.Fatalf("recording = %q, want %q", a.recording, action.About)
	}
	a.RecordKeybinding(action.About)
	if a.recording != "" {
		t.Fatalf("recording = %q, want none", a.recording)
	}

	a.SetKeybinding(action.About, "ctrl+alt+w")
	if got := keyOf(a.settings.Keybindings, action.About); got != "ctrl+alt+w" {
		t.Errorf("key of %q = %q, want %q", action.About, got, "ctrl+alt+w")
	}
	if len(a.settings.Keybindings) != n+1 {
		t.Errorf("the settings hold %d keybindings, want %d", len(a.settings.Keybindings), n+1)
	}
	if a.recording != "" {
		t.Errorf("recording = %q, want none", a.recording)
	}

	a.SetKeybinding(action.About, "ctrl+alt+e")
	if got := keyOf(a.settings.Keybindings, action.About); got != "ctrl+alt+e" {
		t.Errorf("key of %q = %q, want %q", action.About, got, "ctrl+alt+e")
	}
	if len(a.settings.Keybindings) != n+1 {
		t.Errorf("the settings hold %d keybindings, want %d", len(a.settings.Keybindings), n+1)
	}

	a.RemoveKeybinding(action.About)
	if got := keyOf(a.settings.Keybindings, action.About); got != "" {
		t.Errorf("key of %q = %q, want none", action.About, got)
	}
	if len(a.settings.Keybindings) != n {
		t.Errorf("the settings hold %d keybindings, want the %d they started with", len(a.settings.Keybindings), n)
	}

	// An action the file binds twice loses both keys.
	a.settings.Keybindings = append(a.settings.Keybindings, store.Keybinding{Key: "f9", Action: action.TabNew})
	a.RemoveKeybinding(action.TabNew)
	for _, kb := range a.settings.Keybindings {
		if kb.Action == action.TabNew {
			t.Errorf("%q still runs on %q", action.TabNew, kb.Key)
		}
	}
}

// TestResetKeybinding covers going back to the catalog's keys for one action:
// the keys it starts on come back where the file had it, the key the file gave
// it goes, and the actions around it stay where they were.
func TestResetKeybinding(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	a.SetKeybinding(action.TabNew, "f2")
	a.SetKeybinding(action.TabClose, "f3")
	a.ResetKeybinding(action.TabNew)

	var keys []string
	for _, kb := range a.settings.Keybindings {
		if kb.Action == action.TabNew {
			keys = append(keys, kb.Key)
		}
	}
	want := store.DefaultKeybindings()
	var wants []string
	for _, kb := range want {
		if kb.Action == action.TabNew {
			wants = append(wants, kb.Key)
		}
	}
	if len(keys) != len(wants) {
		t.Fatalf("%q runs on %v, want its defaults %v", action.TabNew, keys, wants)
	}
	for i := range wants {
		if keys[i] != wants[i] {
			t.Fatalf("%q runs on %v, want its defaults %v", action.TabNew, keys, wants)
		}
	}
	if got := keyOf(a.settings.Keybindings, action.TabClose); got != "f3" {
		t.Errorf("key of %q = %q, want f3: resetting one action left another alone", action.TabClose, got)
	}
}

// TestSetKeybindingDropsTheOtherKeys covers an action the defaults bind twice:
// recording a key leaves it running on that key alone, so that the key it keep
// invisibly does not still run it.
func TestSetKeybindingDropsTheOtherKeys(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	a.SetKeybinding(action.TabNew, "f2")
	var keys []string
	for _, kb := range a.settings.Keybindings {
		if kb.Action == action.TabNew {
			keys = append(keys, kb.Key)
		}
	}
	if len(keys) != 1 || keys[0] != "f2" {
		t.Fatalf("%q runs on %v, want f2 alone", action.TabNew, keys)
	}
}

// keyOf is the key an action runs on, "" when it runs on none.
func keyOf(all []store.Keybinding, action string) string {
	for _, kb := range all {
		if kb.Action == action {
			return kb.Key
		}
	}
	return ""
}
