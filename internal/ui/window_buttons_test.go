package ui

import (
	myui "github.com/egoist/mygo/ui"
	"slices"
	"testing"
	"wtgo/locales"
)

// TestWindowButtons covers the buttons the strip draws for a window that hides
// the desktop's: the layout's buttons at the sides it names, the button that
// restores a maximized window, and a click that asks the window for each.
func TestWindowButtons(t *testing.T) {
	draw := func(b *WindowButtons) (*myui.Tester, *windowActions) {
		act := &windowActions{}
		st := &State{Tabs: []Tab{{Name: "PowerShell 7"}}, WindowButtons: b}
		tt := myui.NewTester(func(c *myui.Context) { strip(c, *st, act) }, 900, TitleBarHeight*2)
		tt.Frame()
		return tt, act
	}
	// Close at the start, as a desktop that puts its buttons there names it,
	// and the other two at the end.
	tt, act := draw(&WindowButtons{Start: []string{"close"}, End: []string{"minimize", "maximize"}})
	// The buttons are named rather than written, so that a test finds them
	// where a screen reader would.
	closeAt, ok := tt.Find(locales.CloseWindow())
	if !ok {
		t.Fatalf("the strip draws no close button: %q", tt.Texts())
	}
	minAt, ok := tt.Find(locales.Minimize())
	if !ok {
		t.Fatalf("the strip draws no minimize button: %q", tt.Texts())
	}
	maxAt, ok := tt.Find(locales.Maximize())
	if !ok {
		t.Fatalf("the strip draws no maximize button: %q", tt.Texts())
	}
	tabs, _ := tt.Find("PowerShell 7")
	if !(closeAt.X < tabs.X && tabs.X < minAt.X && minAt.X < maxAt.X) {
		t.Errorf("the buttons sit at %v, %v and %v around the tabs at %v", closeAt.X, minAt.X, maxAt.X, tabs.X)
	}
	for _, label := range []string{locales.CloseWindow(), locales.Minimize(), locales.Maximize()} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	if want := []string{"close", "minimize", "maximize"}; !slices.Equal(act.asked, want) {
		t.Errorf("the buttons asked for %q, want %q", act.asked, want)
	}

	// A maximized window draws the button that restores it.
	tt, _ = draw(&WindowButtons{End: []string{"maximize"}, Maximized: true})
	if !tt.HasText(locales.Restore()) || tt.HasText(locales.Maximize()) {
		t.Errorf("a maximized window draws %q, want the button that restores it", tt.Texts())
	}

	// A window that keeps the desktop's buttons draws none of its own.
	tt, _ = draw(nil)
	for _, label := range []string{locales.CloseWindow(), locales.Minimize(), locales.Maximize()} {
		if tt.HasText(label) {
			t.Errorf("the strip draws %q for a window that keeps the desktop's buttons", label)
		}
	}
}

// windowActions is the window as the strip's buttons see it: it records what
// each button asked for.
type windowActions struct {
	Actions
	asked []string
}

func (a *windowActions) MinimizeWindow() { a.asked = append(a.asked, "minimize") }

func (a *windowActions) ToggleMaximizeWindow() { a.asked = append(a.asked, "maximize") }

func (a *windowActions) CloseWindow() { a.asked = append(a.asked, "close") }

func (a *windowActions) OpenSettings() { a.asked = append(a.asked, "settings") }
