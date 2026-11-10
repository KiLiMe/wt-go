package ui

import (
	"testing"

	myui "github.com/egoist/mygo/ui"
)

// TestParseKey covers the keybindings settings.json writes: the modifiers
// before the key, and the names of the keys.
func TestParseKey(t *testing.T) {
	tests := []struct {
		key  string
		mods myui.Modifiers
		want myui.Key
	}{
		{"ctrl+shift+t", myui.Ctrl | myui.Shift, myui.KeyT},
		{"ctrl+tab", myui.Ctrl, myui.KeyTab},
		{"ctrl+shift+tab", myui.Ctrl | myui.Shift, myui.KeyTab},
		{"ctrl+1", myui.Ctrl, myui.Key1},
		{"ctrl+9", myui.Ctrl, myui.Key9},
		{"alt+f4", myui.Alt, myui.KeyF4},
		{"win+enter", myui.Super, myui.KeyEnter},
		{"ctrl+minus", myui.Ctrl, myui.KeyMinus},
		{" Ctrl+Shift+D ", myui.Ctrl | myui.Shift, myui.KeyD},
	}
	for _, tt := range tests {
		mods, key, ok := ParseKey(tt.key)
		if !ok || mods != tt.mods || key != tt.want {
			t.Errorf("ParseKey(%q) = %v, %v, %v", tt.key, mods, key, ok)
		}
	}
}

// TestParseKeyRejects covers names that bind no key.
func TestParseKeyRejects(t *testing.T) {
	for _, key := range []string{"", "ctrl+", "ctrl+nope", "ctrl+shift+f13", "hyper+t", "ctrl+!"} {
		if _, _, ok := ParseKey(key); ok {
			t.Errorf("ParseKey(%q) bound a key", key)
		}
	}
}
