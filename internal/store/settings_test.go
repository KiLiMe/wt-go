package store

import "testing"

// TestDefaultFontFamily covers the font a terminal draws with before the
// settings name one, which the settings page also shows as its placeholder.
func TestDefaultFontFamily(t *testing.T) {
	if got := Default(); got.Font.Family != DefaultFontFamily() {
		t.Errorf("default font family = %q, want %q", got.Font.Family, DefaultFontFamily())
	}
}

// TestDefault covers the settings a window starts with: the values
// docs/settings.md promises, and the keys the tab actions need.
func TestDefault(t *testing.T) {
	s := Default()
	if s.Language != "zh-CN" || s.Theme != "dark" || s.ColorScheme != "Campbell" {
		t.Errorf("language, theme, scheme = %q, %q, %q", s.Language, s.Theme, s.ColorScheme)
	}
	if s.Font.Family != DefaultFontFamily() || s.Font.Size != 13 || s.Font.LineHeight != 1.2 {
		t.Errorf("font = %+v", s.Font)
	}
	if s.Padding.X != 12 || s.Padding.Y != 8 {
		t.Errorf("padding = %+v", s.Padding)
	}
	if !s.CloseOnLastTab || s.MaxPanes != 4 || s.ScrollbackMB != 10 {
		t.Errorf("behavior = %+v", s)
	}
	if len(s.Keybindings) == 0 {
		t.Fatal("no default keybindings")
	}
	if s.Profiles != nil || s.Schemes != nil {
		t.Errorf("defaults name profiles or schemes: %v, %v", s.Profiles, s.Schemes)
	}
}

// TestParseKeepsDefaults covers a file that names some settings: the ones it
// leaves out keep the default rather than turning into zero values.
func TestParseKeepsDefaults(t *testing.T) {
	s, err := Parse([]byte(`{"language":"en","font":{"size":14},"closeOnLastTab":false}`))
	if err != nil {
		t.Fatal(err)
	}
	if s.Language != "en" || s.Font.Size != 14 {
		t.Errorf("named settings = %q, %v", s.Language, s.Font.Size)
	}
	if s.Font.Family != DefaultFontFamily() {
		t.Errorf("font family = %q, want the default", s.Font.Family)
	}
	if s.CloseOnLastTab {
		t.Error("closeOnLastTab stayed true")
	}
	if s.Theme != "dark" || len(s.Keybindings) == 0 {
		t.Errorf("unnamed settings lost their defaults: %+v", s)
	}
}

// TestParseRejects covers a file that is not JSON.
func TestParseRejects(t *testing.T) {
	if _, err := Parse([]byte("{oops")); err == nil {
		t.Fatal("parsed a broken file")
	}
}

// TestMarshalRoundTrip covers the file the app writes on first run.
func TestMarshalRoundTrip(t *testing.T) {
	b, err := Marshal(Default())
	if err != nil {
		t.Fatal(err)
	}
	if b[len(b)-1] != '\n' {
		t.Error("no trailing newline")
	}
	s, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	if s.Font != Default().Font || len(s.Keybindings) != len(Default().Keybindings) {
		t.Errorf("round trip changed the settings: %+v", s)
	}
}

// TestGotoAction covers the tab.goto.n actions.
func TestGotoAction(t *testing.T) {
	for n := 1; n <= 9; n++ {
		action := GotoAction(n)
		i, ok := GotoIndex(action)
		if !ok || i != n-1 {
			t.Errorf("GotoIndex(%q) = %d, %v", action, i, ok)
		}
	}
	for _, action := range []string{"tab.new", "tab.goto.", "tab.goto.x", "tab.goto.0"} {
		if _, ok := GotoIndex(action); ok {
			t.Errorf("GotoIndex(%q) found a tab", action)
		}
	}
}
