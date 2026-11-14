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
		{"ctrl+0", myui.Ctrl, myui.Key0},
		{"ctrl+-", myui.Ctrl, myui.KeyMinus},
		{"minus", 0, myui.KeyMinus},
		{"=", 0, myui.KeyEqual},
		{"equal", 0, myui.KeyEqual},
		{"ctrl+=", myui.Ctrl, myui.KeyEqual},
		{"ctrl+equal", myui.Ctrl, myui.KeyEqual},
		{" Ctrl+Shift+D ", myui.Ctrl | myui.Shift, myui.KeyD},
		// The + key ends the name, so it is written as the separator and
		// then the key: "ctrl++" is Ctrl and +, not Ctrl and an empty name.
		{"+", 0, myui.KeyEqual},
		{"ctrl++", myui.Ctrl, myui.KeyEqual},
		{"ctrl+shift++", myui.Ctrl | myui.Shift, myui.KeyEqual},
		{"ctrl+plus", myui.Ctrl, myui.KeyEqual},
		{" Ctrl++ ", myui.Ctrl, myui.KeyEqual},
	}
	for _, tt := range tests {
		mods, key, ok := ParseKey(tt.key)
		if !ok || mods != tt.mods || key != tt.want {
			t.Errorf("ParseKey(%q) = %v, %v, %v", tt.key, mods, key, ok)
		}
	}
}

// TestParseKeyRejects covers names that bind no key: a name that ends in a +
// with no key after it, as "ctrl+", names no key, and neither does "++",
// which is only the separator and the key with no modifier before it.
func TestParseKeyRejects(t *testing.T) {
	for _, key := range []string{"", "ctrl+", "ctrl+shift+", "++", "+a", "ctrl+nope", "ctrl+shift+f13", "hyper+t", "ctrl+!"} {
		if mods, k, ok := ParseKey(key); ok {
			t.Errorf("ParseKey(%q) bound %v, %v", key, mods, k)
		}
	}
}

// TestFormatKeyRoundTrip checks that the one name FormatKey writes for a key
// parses back to that key, so that the settings page and the default
// keybindings agree on how the + and - keys are spelled.
func TestFormatKeyRoundTrip(t *testing.T) {
	tests := []struct {
		mods myui.Modifiers
		key  myui.Key
		want string
	}{
		{myui.Ctrl, myui.KeyEqual, "ctrl+plus"},
		{myui.Ctrl | myui.Shift, myui.KeyEqual, "ctrl+shift+plus"},
		{0, myui.KeyEqual, "plus"},
		{myui.Ctrl, myui.KeyMinus, "ctrl+minus"},
	}
	for _, tt := range tests {
		got, ok := FormatKey(tt.mods, tt.key)
		if !ok || got != tt.want {
			t.Errorf("FormatKey(%v, %v) = %q, %v", tt.mods, tt.key, got, ok)
			continue
		}
		mods, key, parsed := ParseKey(got)
		if !parsed || mods != tt.mods || key != tt.key {
			t.Errorf("ParseKey(%q) = %v, %v, %v", got, mods, key, parsed)
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

// TestFormatKeyAllRoundTrip covers that every key FormatKey writes parses back
// to the same modifiers and key.
func TestFormatKeyAllRoundTrip(t *testing.T) {
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
		{myui.Ctrl, myui.KeyEqual},
		{0, myui.KeyMinus},
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

// TestParseKeySymbols covers the keys named by the character they type.
func TestParseKeySymbols(t *testing.T) {
	tests := []struct {
		key  string
		want myui.Key
	}{
		{"-", myui.KeyMinus},
		{"=", myui.KeyEqual},
		{",", myui.KeyComma},
		{".", myui.KeyPeriod},
		{"/", myui.KeySlash},
		{";", myui.KeySemicolon},
		{"'", myui.KeyQuote},
		{"\\", myui.KeyBackslash},
	}
	for _, tt := range tests {
		if _, key, ok := ParseKey(tt.key); !ok || key != tt.want {
			t.Errorf("ParseKey(%q) = %v, %v, want %v", tt.key, key, ok, tt.want)
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
