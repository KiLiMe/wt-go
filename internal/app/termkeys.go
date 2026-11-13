package app

import (
	"runtime"

	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"
)

// legacyBytes is the bytes a program gets for a control key, in the encoding
// every program reads, and false for the keys that go the plugin's way.
//
// The terminal plugin encodes keys for the Kitty keyboard protocol once a
// program asks for it, and ConPTY carries none of those bytes: a program that
// asks, as pi and other full-screen ones do, hears nothing of Ctrl+C or Ctrl+U
// and its input cannot be edited or cleared. So the keys no mode changes are
// written here instead. See ADR-57.
func legacyBytes(mods myui.Modifiers, key myui.Key) ([]byte, bool) {
	if key == myui.KeyTab && mods == myui.Shift {
		return []byte("\x1b[Z"), true
	}
	// Alt is the text's own prefix, and Ctrl+Alt is AltGr: a character to the
	// plugin, not a control code.
	if mods&myui.Ctrl == 0 || mods&(myui.Alt|myui.Super) != 0 {
		return nil, false
	}
	if key < myui.KeyA || key > myui.KeyZ {
		return nil, false
	}
	// A Control letter is the letter's control code, and it is the same code
	// in every keyboard protocol, so the arrows and the rest are left alone:
	// they follow the program's cursor-key mode, which only the plugin knows.
	return []byte{byte(key-myui.KeyA) + 1}, true
}

// legacyKey writes what legacyBytes gives, and reports whether the key went to
// the program here rather than to the plugin. Only Windows needs it: ConPTY is
// where the Kitty bytes are lost, and elsewhere the plugin's encoding is the
// program's own.
func legacyKey(term *terminal.Terminal, mods myui.Modifiers, key myui.Key) bool {
	if runtime.GOOS != "windows" || term == nil {
		return false
	}
	b, ok := legacyBytes(mods, key)
	if !ok {
		return false
	}
	term.Send(b)
	return true
}
