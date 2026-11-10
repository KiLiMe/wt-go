package ui

import (
	"strconv"
	"strings"

	myui "github.com/egoist/mygo/ui"
)

// ParseKey is a keybinding as settings.json writes it, "ctrl+shift+t", and
// the key it binds, when it names one: the modifiers, the key, and whether
// the name is one.
func ParseKey(s string) (myui.Modifiers, myui.Key, bool) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(s)), "+")
	key, ok := keyName(parts[len(parts)-1])
	if !ok {
		return 0, 0, false
	}
	var mods myui.Modifiers
	for _, part := range parts[:len(parts)-1] {
		switch strings.TrimSpace(part) {
		case "ctrl", "control":
			mods |= myui.Ctrl
		case "shift":
			mods |= myui.Shift
		case "alt":
			mods |= myui.Alt
		case "win", "super", "meta", "cmd", "command":
			mods |= myui.Super
		default:
			return 0, 0, false
		}
	}
	return mods, key, true
}

// keyName is the key a name writes, "tab" or "t", when it is one.
func keyName(name string) (myui.Key, bool) {
	if key, ok := aliases[name]; ok {
		return key, true
	}
	for _, n := range named {
		if n.Name == name {
			return n.Key, true
		}
	}
	if len(name) == 1 {
		switch c := name[0]; {
		case 'a' <= c && c <= 'z':
			return myui.KeyA + myui.Key(c-'a'), true
		case '0' <= c && c <= '9':
			return myui.Key0 + myui.Key(c-'0'), true
		}
		return 0, false
	}
	if n, err := strconv.Atoi(strings.TrimPrefix(name, "f")); err == nil && n >= 1 && n <= 12 && name[0] == 'f' {
		return myui.KeyF1 + myui.Key(n-1), true
	}
	return 0, false
}

// KeyName is a key, and the name a keybinding writes for it.
type KeyName struct {
	Name string
	Key  myui.Key
}

// named are the keys a keybinding names with a word of its own.
var named = []KeyName{
	{"tab", myui.KeyTab},
	{"enter", myui.KeyEnter},
	{"escape", myui.KeyEscape},
	{"space", myui.KeySpace},
	{"backspace", myui.KeyBackspace},
	{"delete", myui.KeyDelete},
	{"insert", myui.KeyInsert},
	{"home", myui.KeyHome},
	{"end", myui.KeyEnd},
	{"pageup", myui.KeyPageUp},
	{"pagedown", myui.KeyPageDown},
	{"up", myui.KeyUp},
	{"down", myui.KeyDown},
	{"left", myui.KeyLeft},
	{"right", myui.KeyRight},
	{"minus", myui.KeyMinus},
	{"equal", myui.KeyEqual},
	{"comma", myui.KeyComma},
	{"period", myui.KeyPeriod},
	{"slash", myui.KeySlash},
	{"semicolon", myui.KeySemicolon},
	{"quote", myui.KeyQuote},
	{"backslash", myui.KeyBackslash},
}

// aliases are the other names a key answers to.
var aliases = map[string]myui.Key{
	"return": myui.KeyEnter,
	"esc":    myui.KeyEscape,
	"del":    myui.KeyDelete,
	"ins":    myui.KeyInsert,
	"pgup":   myui.KeyPageUp,
	"pgdn":   myui.KeyPageDown,
}

// KeyNames are the keys a keybinding may name, with the name it writes for
// each: the ones the window records.
func KeyNames() []KeyName {
	out := make([]KeyName, 0, len(named)+48)
	out = append(out, named...)
	for c := 'a'; c <= 'z'; c++ {
		out = append(out, KeyName{string(c), myui.KeyA + myui.Key(c-'a')})
	}
	for c := '0'; c <= '9'; c++ {
		out = append(out, KeyName{string(c), myui.Key0 + myui.Key(c-'0')})
	}
	for n := 1; n <= 12; n++ {
		out = append(out, KeyName{"f" + strconv.Itoa(n), myui.KeyF1 + myui.Key(n-1)})
	}
	return out
}

// FormatKey writes a key the way a keybinding does, "ctrl+shift+t", when it
// names one.
func FormatKey(mods myui.Modifiers, key myui.Key) (string, bool) {
	var name string
	for _, n := range KeyNames() {
		if n.Key == key {
			name = n.Name
			break
		}
	}
	if name == "" {
		return "", false
	}
	var parts []string
	for _, m := range []struct {
		mod  myui.Modifiers
		name string
	}{{myui.Ctrl, "ctrl"}, {myui.Shift, "shift"}, {myui.Alt, "alt"}, {myui.Super, "super"}} {
		if mods&m.mod != 0 {
			parts = append(parts, m.name)
		}
	}
	return strings.Join(append(parts, name), "+"), true
}
