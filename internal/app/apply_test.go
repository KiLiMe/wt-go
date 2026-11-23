package app

import (
	"testing"

	"wtgo/internal/store"
)

// TestFaceChanges covers what a terminal is told when the settings change: only
// the settings that reach it, so that the settings page's frame-by-frame apply
// does not hand every terminal a font and colors it already has.
func TestFaceChanges(t *testing.T) {
	base := store.Default()
	none := faceChange{}
	with := func(change func(*store.Settings)) store.Settings {
		s := base
		change(&s)
		return s
	}
	tests := []struct {
		name string
		now  store.Settings
		want faceChange
	}{
		{"nothing", base, none},
		{"the opacity", with(func(s *store.Settings) { s.BackgroundOpacity = 0.4 }), none},
		{"the language", with(func(s *store.Settings) { s.Language = "zh-Hans" }), none},
		{"the font", with(func(s *store.Settings) { s.Font.Size = 17 }), faceChange{font: true}},
		{"the scheme", with(func(s *store.Settings) { s.ColorScheme = "Dracula" }), faceChange{colors: true}},
		{
			"a scheme the file writes",
			with(func(s *store.Settings) {
				s.Schemes = []store.Scheme{{Name: "Mine", Background: "#000000"}}
			}),
			faceChange{colors: true},
		},
		{"the window's theme", with(func(s *store.Settings) { s.Theme = "light" }), faceChange{colors: true}},
		{
			"the links",
			with(func(s *store.Settings) { s.Links.Open = !base.Links.Open }),
			faceChange{links: true},
		},
		{
			"the links' hover hint",
			with(func(s *store.Settings) { s.Links.Hover = !base.Links.Hover }),
			faceChange{links: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := faceChanges(tt.now, base, false); got != tt.want {
				t.Errorf("faceChanges() = %+v, want %+v", got, tt.want)
			}
		})
	}
	// The first apply gives the terminals everything: the tabs open then have
	// heard nothing yet.
	if got := faceChanges(base, store.Settings{}, true); got != (faceChange{true, true, true, true, true, true}) {
		t.Errorf("the first apply sends %+v, want everything", got)
	}
}
