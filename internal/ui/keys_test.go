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

// TestParseKeySuperAliases covers the names the Super modifier answers to.
func TestParseKeySuperAliases(t *testing.T) {
	wantMods, wantKey, ok := ParseKey("super+t")
	if !ok || wantMods != myui.Super || wantKey != myui.KeyT {
		t.Fatalf("ParseKey(%q) = %v, %v, %v", "super+t", wantMods, wantKey, ok)
	}
	for _, name := range []string{"super", "win", "meta", "cmd", "command"} {
		mods, key, ok := ParseKey(name + "+t")
		if !ok || mods != wantMods || key != wantKey {
			t.Errorf("ParseKey(%q) = %v, %v, %v, want %v, %v, true", name+"+t", mods, key, ok, wantMods, wantKey)
		}
	}
}

// TestFormatKey covers the names FormatKey writes.
func TestFormatKey(t *testing.T) {
	tests := []struct {
		mods myui.Modifiers
		key  myui.Key
		want string
	}{
		{0, myui.KeyTab, "tab"},
		{myui.Ctrl, myui.KeyT, "ctrl+t"},
		{myui.Ctrl | myui.Super, myui.KeyT, "ctrl+super+t"},
		{myui.Ctrl | myui.Shift | myui.Alt, myui.KeyF4, "ctrl+shift+alt+f4"},
	}
	for _, tt := range tests {
		got, ok := FormatKey(tt.mods, tt.key)
		if !ok || got != tt.want {
			t.Errorf("FormatKey(%v, %v) = %q, %v, want %q, true", tt.mods, tt.key, got, ok, tt.want)
		}
	}
}

// TestFormatKeyRoundTrip covers that a key FormatKey writes parses back to the
// same modifiers and key.
func TestFormatKeyRoundTrip(t *testing.T) {
	tests := []struct {
		mods myui.Modifiers
		key  myui.Key
	}{
		{0, myui.KeyTab},
		{myui.Ctrl, myui.KeyT},
		{myui.Super, myui.KeyEnter},
		{myui.Ctrl | myui.Super, myui.KeyT},
		{myui.Shift | myui.Alt, myui.KeyF12},
		{myui.Ctrl | myui.Shift | myui.Alt | myui.Super, myui.KeyPageDown},
	}
	for _, tt := range tests {
		name, ok := FormatKey(tt.mods, tt.key)
		if !ok {
			t.Fatalf("FormatKey(%v, %v) named no key", tt.mods, tt.key)
		}
		mods, key, ok := ParseKey(name)
		if !ok || mods != tt.mods || key != tt.key {
			t.Errorf("ParseKey(FormatKey(%v, %v)) = ParseKey(%q) = %v, %v, %v", tt.mods, tt.key, name, mods, key, ok)
		}
	}
}

// TestCaptureMods covers the modifiers capture listens for: the eight
// combinations of Ctrl, Shift and Alt, each with and without Super.
func TestCaptureMods(t *testing.T) {
	mods := captureMods()
	if len(mods) != 16 {
		t.Fatalf("len(captureMods()) = %d, want 16", len(mods))
	}
	seen := make(map[myui.Modifiers]bool, len(mods))
	base := myui.Modifiers(0)
	for _, m := range []myui.Modifiers{myui.Ctrl, myui.Shift, myui.Alt} {
		base |= m
	}
	with, without := 0, 0
	for _, m := range mods {
		if seen[m] {
			t.Errorf("captureMods() repeats %v", m)
		}
		seen[m] = true
		if m&myui.Super != 0 {
			with++
		} else {
			without++
		}
		if m&^(base|myui.Super) != 0 {
			t.Errorf("captureMods() has %v outside Ctrl, Shift, Alt and Super", m)
		}
	}
	if with != 8 || without != 8 {
		t.Errorf("captureMods() has %d with Super and %d without, want 8 and 8", with, without)
	}
}
