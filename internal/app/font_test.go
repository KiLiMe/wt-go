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
	got, mono := census(fs)
	if want := []string{"monospace"}; !slices.Equal(got, want) {
		t.Errorf("census() families = %v, want %v", got, want)
	}
	if len(mono) != 2 {
		t.Errorf("census() measured %d families, want the two advertised as monospaced", len(mono))
	}
}

// TestAdvertisedMono covers the list the picker holds before the census has
// measured anything: the families the text engine calls monospaced, sorted,
// and the others left out.
func TestAdvertisedMono(t *testing.T) {
	fs := []myui.FontFamily{
		{Name: "Grid", Monospaced: true},
		{Name: "Wide", Monospaced: false},
		{Name: "Also", Monospaced: true},
	}
	if got, want := advertisedMono(fs), []string{"Also", "Grid"}; !slices.Equal(got, want) {
		t.Errorf("advertisedMono() = %v, want %v", got, want)
	}
}

// TestIsMonospaceCaches covers the measure the window makes per tab and per
// setting: the same family at the same size is measured once, and a family
// written as the empty one is the default families, which are measured under
// the names they are written as.
func TestIsMonospaceCaches(t *testing.T) {
	a := &App{}
	if !a.isMonospace(store.DefaultFontFamily(), 13) {
		t.Fatalf("monospace(%q) = false", store.DefaultFontFamily())
	}
	measured := len(a.fonts.mono)
	if !a.isMonospace("", 13) {
		t.Error("an empty family is the default families, which are monospace")
	}
	if got := len(a.fonts.mono); got != measured {
		t.Errorf("the empty family was measured again: %d keys, want %d", got, measured)
	}
	a.isMonospace(store.DefaultFontFamily(), 20)
	if got := len(a.fonts.mono); got != measured+1 {
		t.Errorf("a size the window has not measured was not measured: %d keys, want %d", got, measured+1)
	}
}

// TestWeightOf covers what a terminal draws at: a weight of nothing is the
// regular face, and a weight no family has is the nearest one it does.
func TestWeightOf(t *testing.T) {
	tests := []struct{ in, want int }{
		{0, 400}, {-100, 400}, {100, 100}, {350, 350}, {900, 900}, {1000, 900}, {50, 100},
	}
	for _, tt := range tests {
		if got := weightOf(tt.in); got != tt.want {
			t.Errorf("weightOf(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

// TestTerminalFont covers the font the terminals draw with: the settings' one
// while it is monospace, and the fallback while it is not.
func TestTerminalFont(t *testing.T) {
	s := store.Default()
	s.Font.Family, s.Font.Size, s.Font.LineHeight = "no-such-font-at-all", 15, 1.4
	a := &App{settings: s}
	f := a.terminalFont()
	if f.Family != store.EmptyFontFamily {
		t.Errorf("family = %q, want the fallback %q", f.Family, store.EmptyFontFamily)
	}
	if f.Size != 15 || f.LineHeight != 1.4 {
		t.Errorf("font = %+v, want the size and the line height the settings name", f)
	}
	if f.Weight != s.Font.Weight {
		t.Errorf("weight = %d, want the weight the settings name %d", f.Weight, s.Font.Weight)
	}

	s.Font.Family = "monospace"
	a.settings = s
	if got := a.terminalFont(); got.Family != "monospace" {
		t.Errorf("family = %q, want the one the settings name", got.Family)
	}
}

// TestTerminalFontSkipsTheFamilyThatSpreads covers a list whose first family is
// proportional: the terminal passes over it and draws with the next one it can,
// rather than giving up the rest of the list for the generic family.
func TestTerminalFontSkipsTheFamilyThatSpreads(t *testing.T) {
	// A family this machine has that a grid cannot be drawn with, which every
	// machine has one of, and a generic that a grid can.
	wide := ""
	for _, f := range myui.FontFamilies() {
		if !f.Monospaced && !monospace(f.Name, 13) {
			wide = f.Name
			break
		}
	}
	if wide == "" {
		t.Skip("this machine has no proportional family to pass over")
	}
	s := store.Default()
	s.Font.Family = wide + ", monospace"
	a := &App{settings: s}
	if got := a.terminalFont(); got.Family != "monospace" {
		t.Errorf("family = %q, want the family the terminal can draw a grid with", got.Family)
	}
	if a.fontWarned != s.Font.Family {
		t.Errorf("warned about %q, want the list %q said once", a.fontWarned, s.Font.Family)
	}
	// The list itself is drawn with while it shapes as monospace, unread: a
	// list of one family the machine aliases to a monospace font is one.
	s.Font.Family = "monospace"
	a.settings = s
	if got := a.terminalFont(); got.Family != "monospace" || a.fontWarned != "" {
		t.Errorf("family = %q, warned %q; want the list itself, once a list draws", got.Family, a.fontWarned)
	}
}

func TestSelectFontPreservesGlyphFallbacks(t *testing.T) {
	a := &App{fonts: fontCensus{
		installed: map[string]bool{"wide": true, "grid": true, "cjk": true, "emoji": true},
		mono: map[monoKey]bool{
			{"Wide", 13}:    false,
			{"Missing", 13}: true,
			{"GRID", 13}:    true,
		},
	}}
	got := a.selectFont("Wide, Missing, 'GRID', CJK, Absent, Emoji", 13)
	if got.family != "GRID, CJK, Emoji" || got.note.Primary != "GRID" || got.note.Proportional != "Wide" || got.note.SystemFallback {
		t.Errorf("selectFont() = %+v, want GRID with CJK and Emoji as glyph fallbacks", got)
	}
	if want := []string{"Missing", "Absent"}; !slices.Equal(got.note.Missing, want) {
		t.Errorf("missing families = %v, want %v", got.note.Missing, want)
	}
	got = a.selectFont("Wide, Missing", 13)
	if got.family != "monospace" || !got.note.SystemFallback {
		t.Errorf("selectFont() = %+v, want an implicit system fallback", got)
	}
}

func TestSelectFontDropsARepeatedName(t *testing.T) {
	a := &App{fonts: fontCensus{
		installed: map[string]bool{"grid": true, "hack": true},
		mono:      map[monoKey]bool{{"GRID", 13}: true},
	}}
	if got := a.selectFont("GRID, grid, Hack, hack", 13); got.family != "GRID, Hack" {
		t.Errorf("selectFont() = %q, want each family once", got.family)
	}
}

func TestSelectFontReportsOnlyCustomMissingNames(t *testing.T) {
	a := &App{fonts: fontCensus{
		installed: map[string]bool{},
		mono:      map[monoKey]bool{{"monospace", 13}: true},
	}}
	got := a.selectFont("Missing, "+store.DefaultFontFamily(), 13)
	if got.family != "monospace" || got.note.SystemFallback || !slices.Equal(got.note.Missing, []string{"Missing"}) {
		t.Errorf("selectFont() = %+v, want explicit monospace and only the custom missing name", got)
	}
}
