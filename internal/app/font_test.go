package app

import (
	"slices"
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
)

// TestMonospace covers the families a settings file names: a monospace one is
// drawn as it is, and a proportional one, or a family the machine does not
// have, is not.
func TestMonospace(t *testing.T) {
	cases := []struct {
		name   string
		family string
		want   bool
	}{
		{"the generic family", "monospace", true},
		{"a list that ends in the generic family", "no-such-font-at-all,monospace", true},
		{"no family at all, which mygo draws with the generic one", "", true},
		{"a family no machine has, which resolves to a proportional one", "no-such-font-at-all", false},
		{"a proportional family", "sans-serif", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := monospace(tc.family, 13); got != tc.want {
				t.Errorf("monospace(%q) = %v, want %v", tc.family, got, tc.want)
			}
		})
	}
}

// TestMonospaceFamilies covers the families the font setting lists: the ones
// the text engine calls monospaced that also measure as monospace, sorted.
func TestMonospaceFamilies(t *testing.T) {
	fs := []myui.FontFamily{
		{Name: "monospace", Monospaced: true},
		{Name: "no-such-font-at-all", Monospaced: true},
		{Name: "sans-serif", Monospaced: false},
	}
	a := &App{}
	if got, want := a.monospaceFamilies(fs, 13), []string{"monospace"}; !slices.Equal(got, want) {
		t.Errorf("monospaceFamilies() = %v, want %v", got, want)
	}
}

// TestIsMonospaceCaches covers the measure the window makes per tab and per
// setting: the same family at the same size is measured once, and a family
// written as the empty one is the generic family, which is already measured.
func TestIsMonospaceCaches(t *testing.T) {
	a := &App{}
	if !a.isMonospace("monospace", 13) {
		t.Fatal("monospace(\"monospace\") = false")
	}
	measured := len(a.fonts.mono)
	if !a.isMonospace("", 13) {
		t.Error("an empty family is the generic one, which is monospace")
	}
	if got := len(a.fonts.mono); got != measured {
		t.Errorf("the empty family was measured again: %d keys, want %d", got, measured)
	}
	a.isMonospace("monospace", 20)
	if got := len(a.fonts.mono); got != measured+1 {
		t.Errorf("a size the window has not measured was not measured: %d keys, want %d", got, measured+1)
	}
}

// TestTerminalFont covers the font the terminals draw with: the settings' one
// while it is monospace, and the fallback while it is not.
func TestTerminalFont(t *testing.T) {
	s := store.Default()
	s.Font.Family, s.Font.Size, s.Font.LineHeight = "no-such-font-at-all", 15, 1.4
	a := &App{settings: s}
	f := a.terminalFont()
	if f.Family != fallbackFamily {
		t.Errorf("family = %q, want the fallback %q", f.Family, fallbackFamily)
	}
	if f.Size != 15 || f.LineHeight != 1.4 {
		t.Errorf("font = %+v, want the size and the line height the settings name", f)
	}

	s.Font.Family = "monospace"
	a.settings = s
	if got := a.terminalFont(); got.Family != "monospace" {
		t.Errorf("family = %q, want the one the settings name", got.Family)
	}
}
