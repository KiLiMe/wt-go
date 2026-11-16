package store

import (
	"strings"
	"testing"
)

// TestValidate covers the settings the window can run with, and the ones it
// cannot: a file that asks for a font no program can draw, or a scrollback no
// machine can hold.
func TestValidate(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatalf("the defaults do not validate: %v", err)
	}
	tests := []struct {
		name string
		edit func(*Settings)
		want string
	}{
		{"font size small", func(s *Settings) { s.Font.Size = 2 }, "font.size"},
		{"font size large", func(s *Settings) { s.Font.Size = 500 }, "font.size"},
		{"weight light", func(s *Settings) { s.Font.Weight = 50 }, "font.weight"},
		{"weight heavy", func(s *Settings) { s.Font.Weight = 1000 }, "font.weight"},
		{"line height", func(s *Settings) { s.Font.LineHeight = 0 }, "font.lineHeight"},
		{"padding", func(s *Settings) { s.Padding.X = -1 }, "padding.x"},
		{"opacity", func(s *Settings) { s.BackgroundOpacity = 1.5 }, "backgroundOpacity"},
		{"opacity too clear", func(s *Settings) { s.BackgroundOpacity = 0.2 }, "backgroundOpacity"},
		{"scrollback", func(s *Settings) { s.ScrollbackMB = 1 << 20 }, "scrollbackMB"},
		{"scroll multiplier", func(s *Settings) { s.ScrollMultiplier = 0 }, "scrollMultiplier"},
		{"contrast", func(s *Settings) { s.MinimumContrast = 21.5 }, "minimumContrast"},
		{"padding color", func(s *Settings) { s.PaddingColor = "extended" }, "paddingColor"},
		{"clipboard read", func(s *Settings) { s.ClipboardRead = "maybe" }, "clipboardRead"},
		{"clipboard write", func(s *Settings) { s.ClipboardWrite = "ask" }, "clipboardWrite"},
		{"panes", func(s *Settings) { s.MaxPanes = 1000 }, "maxPanes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Default()
			tt.edit(&s)
			err := s.Validate()
			if err == nil {
				t.Fatal("validated an out-of-range setting")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want it to name %s", err, tt.want)
			}
		})
	}
}

// TestProfileIDs covers the profiles a file cannot open: none of them refuses
// the file, so the window names them rather than dropping them in silence.
func TestProfileIDs(t *testing.T) {
	got := ProfileIDs([]Profile{{ID: "pwsh"}, {ID: ""}, {ID: "pwsh"}, {ID: "cmd"}})
	if len(got) != 2 {
		t.Fatalf("problems = %q, want the empty id and the repeated one", got)
	}
	if !strings.Contains(got[0], "profiles[1]") || !strings.Contains(got[1], "profiles[2]") {
		t.Errorf("problems = %q, want them to name the two lines", got)
	}
	if p := ProfileIDs(nil); p != nil {
		t.Errorf("problems of no profiles = %q, want none", p)
	}
}

// TestParseRejectsExtremes covers a file that is JSON but not settings: it is
// refused rather than half applied, and answers with no settings rather than
// with defaults a caller could mistake for the file.
func TestParseRejectsExtremes(t *testing.T) {
	s, err := Parse([]byte(`{"maxPanes":1000,"language":"en"}`))
	if err == nil {
		t.Fatal("parsed an out-of-range setting")
	}
	if s.Language != "" || s.MaxPanes != 0 {
		t.Errorf("settings = %+v, want none from a rejected file", s)
	}
}
