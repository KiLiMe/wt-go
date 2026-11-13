package app

import (
	"bytes"
	"testing"

	myui "github.com/egoist/mygo/ui"
)

// TestLegacyBytes checks the control keys written here: the Control letters as
// their control codes, Shift+Tab as its own sequence, and nothing else.
func TestLegacyBytes(t *testing.T) {
	for _, c := range []struct {
		name string
		mods myui.Modifiers
		key  myui.Key
		want []byte
	}{
		{"ctrl+c", myui.Ctrl, myui.KeyC, []byte{0x03}},
		{"ctrl+a", myui.Ctrl, myui.KeyA, []byte{0x01}},
		{"ctrl+z", myui.Ctrl, myui.KeyZ, []byte{0x1a}},
		{"ctrl+shift+a", myui.Ctrl | myui.Shift, myui.KeyA, []byte{0x01}},
		{"shift+tab", myui.Shift, myui.KeyTab, []byte("\x1b[Z")},
		{"letter", 0, myui.KeyC, nil},
		{"tab", 0, myui.KeyTab, nil},
		{"ctrl+alt+c", myui.Ctrl | myui.Alt, myui.KeyC, nil},
		{"ctrl+super+c", myui.Ctrl | myui.Super, myui.KeyC, nil},
		{"ctrl+up", myui.Ctrl, myui.KeyUp, nil},
		{"ctrl+enter", myui.Ctrl, myui.KeyEnter, nil},
	} {
		got, ok := legacyBytes(c.mods, c.key)
		if c.want == nil {
			if ok {
				t.Errorf("%s: wrote %q, want it left to the plugin", c.name, got)
			}
			continue
		}
		if !ok {
			t.Errorf("%s: left to the plugin, want %q", c.name, c.want)
			continue
		}
		if !bytes.Equal(got, c.want) {
			t.Errorf("%s: wrote %q, want %q", c.name, got, c.want)
		}
	}
}
