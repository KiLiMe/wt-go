package app

import (
	"testing"

	"wtgo/internal/store"
	"wtgo/internal/theme"
)

// TestBackgroundIsTheTerminals covers the color the window fills what nothing
// has painted: the terminal's own, whatever theme the window draws with, so
// that the surface shows through nowhere but behind a pane.
func TestBackgroundIsTheTerminals(t *testing.T) {
	s := store.Default()
	s.Theme = "light"
	s.ColorScheme = "Campbell"
	a := &App{settings: s}
	th, ok := theme.Scheme(s, s.ColorScheme)
	if !ok {
		t.Fatalf("%q is no scheme", s.ColorScheme)
	}
	if got := a.background(); got != th.Background {
		t.Errorf("the window fills %v, want the terminal's %v", got, th.Background)
	}
	// A theme of the window's own is the window's, not the fill behind it.
	a.uiTheme = theme.Native(true)
	if got := a.background(); got != th.Background {
		t.Errorf("a themed window fills %v, want the terminal's %v", got, th.Background)
	}
}
