package app

import (
	"testing"

	"wtgo/internal/store"
	"wtgo/internal/theme"
	"wtgo/internal/ui"
)

// TestSchemeBaseName covers the name a copy starts from: the marks earlier
// copies left on it come off, and a name that only looks like one keeps its
// own.
func TestSchemeBaseName(t *testing.T) {
	tests := []struct{ name, want string }{
		{"Campbell", "Campbell"},
		{"Campbell (custom)", "Campbell"},
		{"Campbell (custom 2)", "Campbell"},
		{"Campbell (custom 12)", "Campbell"},
		{"Campbell (custom) (custom)", "Campbell"},
		{"Campbell (custom) (custom 3)", "Campbell"},
		{"Campbell (custom) (custom) (custom)", "Campbell"},
		{"(custom)", "(custom)"},
		{"Paper (custom look)", "Paper (custom look)"},
		{"Paper (custom x)", "Paper (custom x)"},
	}
	for _, tt := range tests {
		if got := schemeBaseName(tt.name); got != tt.want {
			t.Errorf("schemeBaseName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestCopySchemeNamesCopies covers a copy: it is named after the scheme behind
// the copy, so copying copies does not grow a chain of marks, and a name the
// settings already write is numbered.
func TestCopySchemeNamesCopies(t *testing.T) {
	settingsDir(t)
	s := store.Default()
	s.Schemes = []store.Scheme{own("Campbell (custom)")}
	a, _ := settingsApp(t, s)
	a.settings.ColorScheme = "Campbell"

	a.CopyScheme()
	want := "Campbell (custom 2)"
	if a.settings.ColorScheme != want {
		t.Fatalf("copy = %q, want %q", a.settings.ColorScheme, want)
	}
	if !a.schemeNamed(want) {
		t.Errorf("the settings do not write %q", want)
	}

	a.CopyScheme()
	want = "Campbell (custom 3)"
	if a.settings.ColorScheme != want {
		t.Fatalf("copy of a copy = %q, want %q", a.settings.ColorScheme, want)
	}
	if len(a.settings.Schemes) != 3 {
		t.Errorf("the settings write %d schemes, want three", len(a.settings.Schemes))
	}
}

// TestRemoveScheme covers removing a scheme the settings write: it goes, the
// terminal takes the scheme it was a copy of, the defaults' when that is gone
// too, and a name the settings do not write asks for nothing.
func TestRemoveScheme(t *testing.T) {
	settingsDir(t)
	s := store.Default()
	s.Schemes = []store.Scheme{own("Paper"), own("Paper (custom)")}
	s.ColorScheme = "Paper (custom)"
	a, _ := settingsApp(t, s)

	a.RemoveScheme("Dracula")
	if a.note != "" {
		t.Errorf("note = %q, want none for a scheme the settings do not write", a.note)
	}
	if !a.schemeNamed("Paper") || !a.schemeNamed("Paper (custom)") {
		t.Fatal("a scheme the settings do not write took one of theirs")
	}

	a.RemoveScheme("Paper (custom)")
	if a.schemeNamed("Paper (custom)") {
		t.Error("the scheme is still there")
	}
	if a.settings.ColorScheme != "Paper" {
		t.Errorf("scheme = %q, want the %q it copied", a.settings.ColorScheme, "Paper")
	}
	if a.note == "" {
		t.Error("no note about the removal")
	}

	a.RemoveScheme("Paper")
	if want := store.Default().ColorScheme; a.settings.ColorScheme != want {
		t.Errorf("scheme with every copy gone = %q, want %q", a.settings.ColorScheme, want)
	}
}

// own is a scheme the settings write, with the few colors the theme needs.
func own(name string) store.Scheme {
	return store.Scheme{
		Name:       name,
		Background: "#101010",
		Foreground: "#eeeeee",
		Palette:    []string{"#000000"},
	}
}

// The colors a scheme is shown by are the scheme's own, and a scheme the
// settings write is looked up again every time: the page keeps the colors of
// the hundreds that do not change, and must not keep the ones that do.
func TestSchemeSwatches(t *testing.T) {
	s := store.Default()
	s.Schemes = []store.Scheme{{Name: "Paper", Background: "#123456", Foreground: "#abcdef"}}
	a := &App{settings: s}
	colors := func(list []ui.Scheme, name string) (string, string) {
		for _, sc := range list {
			if sc.Name == name {
				return theme.Hex(sc.Bg), theme.Hex(sc.Fg)
			}
		}
		t.Fatalf("the list holds no scheme called %q", name)
		return "", ""
	}

	list := a.uiSchemes(s)
	if bg, fg := colors(list, "Paper"); bg != "#123456" || fg != "#abcdef" {
		t.Errorf("Paper = %s, %s, want the colors the settings write", bg, fg)
	}
	campbell, _ := colors(list, "Campbell")

	s.Schemes[0].Background = "#654321"
	list = a.uiSchemes(s)
	if bg, _ := colors(list, "Paper"); bg != "#654321" {
		t.Errorf("Paper = %s, want the color the settings now write: the page kept a scheme it should not", bg)
	}
	if bg, _ := colors(list, "Campbell"); bg != campbell {
		t.Errorf("Campbell = %s, want the color it had %s: a scheme nothing writes does not change", bg, campbell)
	}
	if bg, _ := colors(list, "Paper"); bg == "#123456" {
		t.Error("Paper kept its old color")
	}
}

// The page asks for the schemes once per settings revision: building them
// sorts 617 names, which a frame should not do twice.
func TestSchemeListIsBuiltOnce(t *testing.T) {
	a := &App{settings: store.Default()}
	first := a.schemeList()
	if len(first) == 0 {
		t.Fatal("no schemes")
	}
	if second := a.schemeList(); &second[0] != &first[0] {
		t.Fatal("the list was rebuilt for the same settings")
	}
}

// A setting that does not reach the list of color schemes does not build it
// again: the settings page applies the draft on every frame of a drag, and a
// drag of the opacity has nothing to do with a list of names.
func TestSchemeListIgnoresOtherSettings(t *testing.T) {
	a := &App{settings: store.Default()}
	first := a.schemeList()
	if len(first) == 0 {
		t.Fatal("no schemes")
	}
	a.settings.BackgroundOpacity = 0.5
	a.settings.Font.Size = 17
	if second := a.schemeList(); &second[0] != &first[0] {
		t.Fatal("the list was rebuilt for a setting that does not reach it")
	}
	a.settings.Schemes = append(a.settings.Schemes, store.Scheme{Name: "Mine", Background: "#101010"})
	if third := a.schemeList(); &third[0] == &first[0] {
		t.Fatal("the list was kept while the file wrote a new scheme")
	}
}

// The schemes the settings page offers come from the theme package, which
// sorts the 617 names it is built with and reads a theme for each of their
// colors; once per frame is a cost the page should not pay, so this pins what
// one call buys with the swatches the process already holds.
func BenchmarkUISchemes(b *testing.B) {
	s := store.Default()
	a := &App{settings: s}
	a.uiSchemes(s) // the colors of the schemes that do not change are kept
	b.ReportAllocs()
	for b.Loop() {
		a.uiSchemes(s)
	}
}
