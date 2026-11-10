package app

import (
	"testing"

	"wtgo/internal/action"
	"wtgo/internal/store"
	"wtgo/internal/ui"
)

// TestPaletteCommands covers what the palette offers: the actions the window
// has, each with a title, and the numbered ones listed once, by what they
// open.
func TestPaletteCommands(t *testing.T) {
	a := &App{settings: store.Default(), saved: store.Default()}
	seen := map[string]bool{}
	for _, cmd := range a.paletteCommands() {
		if cmd.Title == "" {
			t.Errorf("%q is offered with no title", cmd.Action)
		}
		if seen[cmd.Action] {
			t.Errorf("%q is offered twice", cmd.Action)
		}
		seen[cmd.Action] = true
	}
	for _, id := range []string{action.TabNew, action.PaneClose, action.Settings, action.Palette, action.Find} {
		if !seen[id] {
			t.Errorf("%q is not offered", id)
		}
	}
	// The numbered actions open a tab, and the window starts with none.
	if seen[action.GotoAction(1)] || seen[action.ProfileAction(1)] {
		t.Errorf("the palette lists a tab or a profile the window does not have")
	}
}

// TestRun covers the palette's run: the palette closes, and the action the
// palette picked is what the window does.
func TestRun(t *testing.T) {
	a := &App{settings: store.Default(), saved: store.Default()}
	a.palette = &ui.Palette{Open: true}
	a.Run(action.Settings)
	if a.palette != nil {
		t.Error("the palette stayed open")
	}
	if a.settingsTab == nil {
		t.Error("Run(settings.open) opened no settings tab")
	}
}
