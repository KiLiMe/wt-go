package store

import (
	"encoding/json"
	"testing"
)

// TestDefaultFontFamily covers the font a terminal draws with before the
// settings name one, which the settings page also shows as its placeholder.
func TestDefaultFontFamily(t *testing.T) {
	if got := Default(); got.Font.Family != DefaultFontFamily() {
		t.Errorf("default font family = %q, want %q", got.Font.Family, DefaultFontFamily())
	}
}

// TestDefault covers the settings a window starts with: the values
// docs/使用说明.md promises, and the keys the tab actions need.
func TestDefault(t *testing.T) {
	s := Default()
	if s.Language != "" || s.Theme != "dark" || s.ColorScheme != "Campbell" {
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

// TestParseLegacyLanguage covers the language files of earlier versions named by
// region: they keep it, under the tag the catalogs are filed under, instead of
// following the system's.
func TestParseLegacyLanguage(t *testing.T) {
	for _, c := range []struct{ file, want string }{
		{`{"language":"zh-CN"}`, "zh-Hans"},
		{`{"language":"en-US"}`, "en"},
		{`{"language":"zh-Hans"}`, "zh-Hans"},
		{`{"language":""}`, ""},
		{`{}`, ""},
	} {
		s, err := Parse([]byte(c.file))
		if err != nil {
			t.Fatalf("%s: %v", c.file, err)
		}
		if s.Language != c.want {
			t.Errorf("%s: language = %q, want %q", c.file, s.Language, c.want)
		}
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

// TestConfirmAsk covers the three answers: only "running" looks at what the
// tab holds, "always" asks about a tab that lives, "never" about none.
func TestConfirmAsk(t *testing.T) {
	tests := []struct {
		mode    Confirm
		running bool
		want    bool
	}{
		{ConfirmRunning, true, true},
		{ConfirmRunning, false, false},
		{ConfirmAlways, true, true},
		{ConfirmAlways, false, true},
		{ConfirmNever, true, false},
		{ConfirmNever, false, false},
	}
	for _, tc := range tests {
		if got := tc.mode.Ask(tc.running); got != tc.want {
			t.Errorf("%q.Ask(%v) = %v, want %v", tc.mode, tc.running, got, tc.want)
		}
	}
}

// TestConfirmJSON covers the words a file writes, and the true and false the
// setting held before it had three answers.
func TestConfirmJSON(t *testing.T) {
	tests := []struct {
		data string
		want Confirm
	}{
		{`"running"`, ConfirmRunning},
		{`"always"`, ConfirmAlways},
		{`"never"`, ConfirmNever},
		{`true`, ConfirmRunning},
		{`false`, ConfirmNever},
	}
	for _, tc := range tests {
		var got Confirm
		if err := json.Unmarshal([]byte(tc.data), &got); err != nil {
			t.Errorf("reading %s: %v", tc.data, err)
			continue
		}
		if got != tc.want {
			t.Errorf("reading %s = %q, want %q", tc.data, got, tc.want)
		}
		data, err := json.Marshal(got)
		if err != nil || string(data) != `"`+string(tc.want)+`"` {
			t.Errorf("writing %q = %s, %v, want %q", got, data, err, tc.want)
		}
	}
	var bad Confirm
	if err := json.Unmarshal([]byte(`3`), &bad); err == nil {
		t.Errorf("reading 3 = %q, want an error", bad)
	}
}
