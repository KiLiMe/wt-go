package ui

import (
	"strconv"
	"strings"

	myui "github.com/egoist/mygo/ui"
)

// ParseKey is a keybinding as settings.json writes it, "ctrl+shift+t", and
// the key it binds, when it names one: the modifiers, the key, and whether
// the name is one. A name that ends in a + writes the + key, as "ctrl++"
// does, so that it is not read as an empty key name.
func ParseKey(s string) (myui.Modifiers, myui.Key, bool) {
	name := strings.ToLower(strings.TrimSpace(s))
	var parts []string
	var key myui.Key
	if rest, isPlus := cutPlusKey(name); isPlus {
		key = myui.KeyEqual
		if rest != "" {
			parts = strings.Split(rest, "+")
		}
	} else {
		parts = strings.Split(name, "+")
		var ok bool
		if key, ok = keyName(parts[len(parts)-1]); !ok {
			return 0, 0, false
		}
		parts = parts[:len(parts)-1]
	}
	var mods myui.Modifiers
	for _, part := range parts {
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

// cutPlusKey splits a keybinding that ends in the + key, as "ctrl++",
// "ctrl+shift++" or "+", into the modifiers and the key: the last + of the run
// at the end is the key, and the + before it, when there is one, the separator
// that writes it. It is not enough to end in +: "ctrl+" has no key after that
// separator, and "++" is only the separator and the key with no modifier
// before it, so neither writes the + key.
func cutPlusKey(s string) (string, bool) {
	if s == "+" {
		return "", true
	}
	mods, ok := strings.CutSuffix(s, "++")
	if !ok || mods == "" {
		// "++" is only the separator before the key, so there is no
		// modifier in its place.
		return s, false
	}
	return mods, true
}

// keyName is the key a name writes, "tab" or "t", when it is one.
func keyName(name string) (myui.Key, bool) {
	if key, ok := aliases[name]; ok {
		return key, true
	}
	if key, ok := symbols[name]; ok {
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

// symbols are the keys a keybinding names by the character it types, "=" for
// the equal key, so that a person writing one needs no word for it.
var symbols = map[string]myui.Key{
	"-":  myui.KeyMinus,
	"=":  myui.KeyEqual,
	",":  myui.KeyComma,
	".":  myui.KeyPeriod,
	"/":  myui.KeySlash,
	";":  myui.KeySemicolon,
	"'":  myui.KeyQuote,
	"\\": myui.KeyBackslash,
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
	// "plus" is the key an "equal" is with Shift held; ParseKey keeps it as
	// KeyEqual, so that it is the one spelling of that key, and the one
	// name FormatKey writes for it. It comes first, since KeyNames also
	// answers "equal", which would otherwise be the name written for it.
	{"plus", myui.KeyEqual},
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
