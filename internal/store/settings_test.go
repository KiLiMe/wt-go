package store

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
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
	if !s.CloseOnLastTab || !s.ConfirmCloseWindow || s.MaxPanes != 4 || s.ScrollbackMB != 10 {
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
	if s.ScrollMultiplier != 1 || s.PaddingColor != "background" ||
		s.ClipboardRead != "deny" || s.ClipboardWrite != "allow" {
		t.Errorf("the settings added later lost their defaults: %+v", s)
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

// TestKnownKeysAreTheFields covers the two lists of keys: one missing from
// knownKeys is kept in Extra and written after its field, which would give a
// file a key twice.
func TestKnownKeysAreTheFields(t *testing.T) {
	typ := reflect.TypeOf(Settings{})
	var want []string
	for i := range typ.NumField() {
		f := typ.Field(i)
		// Extra is the keys no field names, which is not one of them.
		if f.Name == "Extra" {
			continue
		}
		want = append(want, strings.Split(f.Tag.Get("json"), ",")[0])
	}
	slices.Sort(want)
	got := make([]string, 0, len(knownKeys))
	for k := range knownKeys {
		got = append(got, k)
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("knownKeys = %v, want the fields %v", got, want)
	}
}

// TestParseFallsBackUnknownNames covers a name no setting takes: it is read as
// the default, as docs/配置参考.md promises, rather than refused the way a
// range is. The file is not rejected, so the settings around a typo still apply.
func TestParseFallsBackUnknownNames(t *testing.T) {
	tests := []struct {
		name string
		file string
		get  func(Settings) string
		want string
	}{
		{"theme", `{"theme":"darkk"}`, func(s Settings) string { return s.Theme }, Default().Theme},
		{"cursor shape", `{"cursorShape":"filledBox"}`, func(s Settings) string { return s.CursorShape }, Default().CursorShape},
		{"padding color", `{"paddingColor":"window"}`, func(s Settings) string { return s.PaddingColor }, Default().PaddingColor},
		{"right click", `{"rightClick":"ignore"}`, func(s Settings) string { return s.RightClick }, Default().RightClick},
		{"clipboard read", `{"clipboardRead":"maybe"}`, func(s Settings) string { return s.ClipboardRead }, Default().ClipboardRead},
		{"clipboard write", `{"clipboardWrite":"ask"}`, func(s Settings) string { return s.ClipboardWrite }, Default().ClipboardWrite},
		{"new tab position", `{"newTabPosition":"middle"}`, func(s Settings) string { return s.NewTabPosition }, Default().NewTabPosition},
		{"tab width", `{"tabWidth":"titleLength"}`, func(s Settings) string { return s.TabWidth }, Default().TabWidth},
		{"start on", `{"startOn":"restore"}`, func(s Settings) string { return s.StartOn }, Default().StartOn},
		{"confirm close", `{"confirmClose":"maybe"}`, func(s Settings) string { return string(s.ConfirmClose) }, string(Default().ConfirmClose)},
		{"language", `{"language":"not a tag"}`, func(s Settings) string { return s.Language }, ""},
		{"empty theme", `{"theme":""}`, func(s Settings) string { return s.Theme }, Default().Theme},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, err := Parse([]byte(tc.file))
			if err != nil {
				t.Fatalf("refused a name to fall back on: %v", err)
			}
			if got := tc.get(s); got != tc.want {
				t.Errorf("%s = %q, want the default %q", tc.name, got, tc.want)
			}
			// The fall back leaves a set of settings that validates and reads the
			// same again, so that writing them and starting from them is a loop.
			data, err := Marshal(s)
			if err != nil {
				t.Fatal(err)
			}
			again, err := Parse(data)
			if err != nil {
				t.Fatalf("re-reading the settings it wrote: %v", err)
			}
			if !Equal(s, again) {
				t.Errorf("re-reading changed them: %+v, %+v", s, again)
			}
		})
	}
}

// TestParseKeepsKnownNames covers the names a file may write, which the fall
// back leaves alone: a theme named by a scheme, a language with no catalog yet,
// and a directory, which is a path rather than a word of a list.
func TestParseKeepsKnownNames(t *testing.T) {
	s, err := Parse([]byte(`{"theme":"light","language":"fr-FR","startDirectory":"D:/src"}`))
	if err != nil {
		t.Fatal(err)
	}
	if s.Theme != "light" || s.Language != "fr-FR" || s.StartDirectory != "D:/src" {
		t.Errorf("theme, language, startDirectory = %q, %q, %q", s.Theme, s.Language, s.StartDirectory)
	}
	// A scheme of the file's own is a theme a person may name, so a file that
	// names one is kept as it is.
	s, err = Parse([]byte(`{"theme":"Paper","schemes":[{"name":"Paper","background":"#fafafa"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if s.Theme != "Paper" {
		t.Errorf("theme = %q, want the scheme the file holds", s.Theme)
	}
	// A theme left out of the file, like one no scheme has any more, is the
	// default rather than "" — which internal/theme would read as "nobody
	// says," and the settings page would show as nothing.
	s, err = Parse([]byte(`{"font":{"size":14}}`))
	if err != nil {
		t.Fatal(err)
	}
	if s.Theme != Default().Theme {
		t.Errorf("an unnamed theme = %q, want the default %q", s.Theme, Default().Theme)
	}
}

// TestParseKeepsUnknownKeys covers a file a person or a later version extended:
// the keys no field names are kept through a change and a save, where a window
// that read them as nothing would drop them the first time someone turned a
// switch in the settings page.
func TestParseKeepsUnknownKeys(t *testing.T) {
	file := `{
  "theme": "light",
  "unknownField": 7,
  "font": {"size": 14, "hinting": "full"},
  "profiles": [{"id": "x", "name": "X", "notes": [1, 2]}]
}`
	s, err := Parse([]byte(file))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Extra) != 1 {
		t.Fatalf("Extra = %v, want the one top-level key no field names", s.Extra)
	}

	s.Font.Size = 16
	b, err := Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("%v\n%s", err, b)
	}
	if got := string(out["unknownField"]); got != "7" {
		t.Errorf("unknownField = %s, want it kept as 7", got)
	}
	if got := string(out["theme"]); got != `"light"` {
		t.Errorf("theme = %s, want the light the file said", got)
	}
	if !strings.Contains(string(b), `"size": 16`) {
		t.Errorf("font.size was not written:\n%s", b)
	}

	// The keys of the objects a field does name are that field's business, so a
	// key inside font is lost with the rest of the font the window rewrote. The
	// file is still settings after a save, and the key it kept is read again.
	again, err := Parse(b)
	if err != nil {
		t.Fatalf("re-reading the file it wrote: %v\n%s", err, b)
	}
	if again.Font.Size != 16 || string(again.Extra["unknownField"]) != "7" {
		t.Errorf("re-reading = %+v", again)
	}
}

// TestMarshalKnownKeysWin covers a setting the window changed: the key the file
// wrote second is written as the field holds it now, not as it was read.
func TestMarshalKnownKeysWin(t *testing.T) {
	s := Default()
	s.Extra = map[string]json.RawMessage{"theme": json.RawMessage(`"light"`)}
	b, err := Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if got := string(out["theme"]); got != `"dark"` {
		t.Errorf("theme = %s, want the field's dark", got)
	}
}

// TestEqualIgnoresUnknownKeys covers what the window compares to know whether it
// has unsaved work: a file whose only difference from the window's copy is a key
// neither of them understands holds nothing to save.
func TestEqualIgnoresUnknownKeys(t *testing.T) {
	a := Default()
	b := a.Clone()
	b.Extra = map[string]json.RawMessage{"unknownField": json.RawMessage(`1`)}
	if !Equal(a, b) {
		t.Error("a key no field names makes the settings different")
	}
	b.Theme = "light"
	if Equal(a, b) {
		t.Error("a setting the window shows makes no difference")
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

// TestFontWeight covers the weight a terminal draws at: the regular face before
// the file names one, and the regular face for a file that leaves the key out,
// which is what a zero says and what mygo reads a zero as.
func TestFontWeight(t *testing.T) {
	if got := Default().Font.Weight; got != 400 {
		t.Errorf("default weight = %d, want 400", got)
	}
	read, err := Parse([]byte(`{"font":{"size":14}}`))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if read.Font.Weight != 400 {
		t.Errorf("a file without a weight reads as %d, want 400", read.Font.Weight)
	}
	read, err = Parse([]byte(`{"font":{"weight":300}}`))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if read.Font.Weight != 300 {
		t.Errorf("weight = %d, want the one the file names", read.Font.Weight)
	}
}

// TestFillEmptyFont covers a family the settings leave empty: an empty family
// is not a font, and the file says what the window draws with instead of
// leaving it to be read as the generic one. The window fills it in on save, and
// a file read from disk that leaves it out is filled the same way.
func TestFillEmptyFont(t *testing.T) {
	s := Default()
	s.Font.Family = "  "
	s.FillEmptyFont()
	if s.Font.Family != EmptyFontFamily {
		t.Errorf("family = %q, want %q", s.Font.Family, EmptyFontFamily)
	}
	// A family the file names stays as it is, with the spaces a person wrote.
	s.Font.Family = " Consolas , monospace "
	s.FillEmptyFont()
	if s.Font.Family != " Consolas , monospace " {
		t.Errorf("a family the file names was rewritten: %q", s.Font.Family)
	}
	read, err := Parse([]byte(`{"font":{"family":""}}`))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if read.Font.Family != EmptyFontFamily {
		t.Errorf("a file that leaves the family empty reads as %q, want %q", read.Font.Family, EmptyFontFamily)
	}
}
