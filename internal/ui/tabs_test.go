package ui

import (
	"slices"
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
	"wtgo/internal/i18n"
)

// TestShortcut covers the menu's shortcut hint: an action the window binds a
// key to gets it, and an action it does not, or a key that is not one, gets
// nothing and does not panic.
func TestShortcut(t *testing.T) {
	st := State{Shortcut: func(id string) string {
		switch id {
		case action.ProfileAction(2):
			return "ctrl+shift+2"
		case action.Settings:
			return "ctrl+comma"
		case "broken":
			return "ctrl+nope"
		}
		return ""
	}}
	it := &myui.MenuItem{}
	shortcut(it, st, action.ProfileAction(2))
	shortcut(it, st, action.Settings)
	shortcut(it, st, action.ProfileAction(9))
	shortcut(it, st, "broken")

	shortcut(it, State{}, action.Settings)
}

// TestNameWidth covers the room a tab's name takes: a short name keeps a
// tab clickable, a long one stops growing.
func TestNameWidth(t *testing.T) {
	tests := []struct{ measured, want float32 }{
		{0, minTabName},
		{12, minTabName},
		{minTabName, minTabName},
		{120, 120},
		{maxTabName, maxTabName},
		{900, maxTabName},
	}
	for _, tt := range tests {
		if got := nameWidth(tt.measured); got != tt.want {
			t.Errorf("nameWidth(%v) = %v, want %v", tt.measured, got, tt.want)
		}
	}
}

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
	closeAt, ok := tt.Find(i18n.T("Close Window"))
	if !ok {
		t.Fatalf("the strip draws no close button: %q", tt.Texts())
	}
	minAt, ok := tt.Find(i18n.T("Minimize"))
	if !ok {
		t.Fatalf("the strip draws no minimize button: %q", tt.Texts())
	}
	maxAt, ok := tt.Find(i18n.T("Maximize"))
	if !ok {
		t.Fatalf("the strip draws no maximize button: %q", tt.Texts())
	}
	tabs, _ := tt.Find("PowerShell 7")
	if !(closeAt.X < tabs.X && tabs.X < minAt.X && minAt.X < maxAt.X) {
		t.Errorf("the buttons sit at %v, %v and %v around the tabs at %v", closeAt.X, minAt.X, maxAt.X, tabs.X)
	}
	for _, label := range []string{i18n.T("Close Window"), i18n.T("Minimize"), i18n.T("Maximize")} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	if want := []string{"close", "minimize", "maximize"}; !slices.Equal(act.asked, want) {
		t.Errorf("the buttons asked for %q, want %q", act.asked, want)
	}

	// A maximized window draws the button that restores it.
	tt, _ = draw(&WindowButtons{End: []string{"maximize"}, Maximized: true})
	if !tt.HasText(i18n.T("Restore")) || tt.HasText(i18n.T("Maximize")) {
		t.Errorf("a maximized window draws %q, want the button that restores it", tt.Texts())
	}

	// A window that keeps the desktop's buttons draws none of its own.
	tt, _ = draw(nil)
	for _, label := range []string{i18n.T("Close Window"), i18n.T("Minimize"), i18n.T("Maximize")} {
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

func (a *windowActions) MinimizeWindow()       { a.asked = append(a.asked, "minimize") }
func (a *windowActions) ToggleMaximizeWindow() { a.asked = append(a.asked, "maximize") }
func (a *windowActions) CloseWindow()          { a.asked = append(a.asked, "close") }

// scrollActions is the window as a strip sees it: it keeps where the strip
// shows the tabs the way the window does, and nothing else answers.
type scrollActions struct {
	Actions
	at    *myui.ScrollState
	moved []float32
}

func (a *scrollActions) ScrollTabs(by float32) {
	a.moved = append(a.moved, by)
	a.at.X = max(0, min(a.at.X+by, a.at.MaxX))
}

// stripOf draws the title bar of n tabs named like a shell, in a window of the
// size it gives.
func stripOf(t *testing.T, n int, active int, width int) (*myui.Tester, *scrollActions, *myui.ScrollState) {
	t.Helper()
	at := &myui.ScrollState{}
	act := &scrollActions{at: at}
	st := &State{Tabs: make([]Tab, n), Active: active, TabsScroll: at}
	for i := range st.Tabs {
		st.Tabs[i].Name = "PowerShell 7"
	}
	tt := myui.NewTester(func(c *myui.Context) { strip(c, *st, act) }, width, TitleBarHeight*2)
	tt.Frame()
	return tt, act, at
}

// TestStripFits covers a strip the tabs do not outgrow: it does not scroll,
// it shows no buttons to move the tabs, and every tab is as wide as it would
// be alone.
func TestStripFits(t *testing.T) {
	tt, _, at := stripOf(t, 2, 0, 900)
	if at.MaxX != 0 {
		t.Fatalf("two tabs in 900 DIPs scroll: MaxX = %v", at.MaxX)
	}
	if tt.HasText("\u276e") || tt.HasText("\u276f") {
		t.Errorf("the strip shows buttons to move tabs that all fit: %q", tt.Texts())
	}
	one, _, _ := stripOf(t, 1, 0, 900)
	name, ok := one.Find("PowerShell 7")
	if !ok {
		t.Fatal("the tab does not draw its name")
	}
	pair, ok := tt.Find("PowerShell 7")
	if !ok {
		t.Fatal("the tabs do not draw their names")
	}
	if pair.W != name.W {
		t.Errorf("a tab of two draws its name %v wide, a tab alone %v", pair.W, name.W)
	}
}

// TestStripScrolls covers a strip the tabs outgrow: the tab in front comes
// into view, the wheel over the strip moves the tabs sideways, and the button
// at the start brings them back.
func TestStripScrolls(t *testing.T) {
	at := &myui.ScrollState{}
	act := &scrollActions{at: at}
	st := &State{Tabs: make([]Tab, 20), Active: 19, TabsScroll: at, RevealTab: true}
	for i := range st.Tabs {
		st.Tabs[i].Name = "PowerShell 7"
	}
	tt := myui.NewTester(func(c *myui.Context) { strip(c, *st, act) }, 900, TitleBarHeight*2)
	tt.Frame()
	if at.MaxX <= 0 {
		t.Fatalf("twenty tabs in 900 DIPs do not outgrow the strip")
	}
	if at.X <= 0 {
		t.Fatalf("the tab in front did not come into view: X = %v", at.X)
	}
	// The window asks once, as it changes the tab in front.
	st.RevealTab = false

	// The buttons take room from the tabs, so they only show once the tabs
	// outgrow the strip: a frame later, as the strip is measured first.
	tt.Frame()
	if !tt.HasText("\u276e") || !tt.HasText("\u276f") {
		t.Errorf("the strip does not show buttons to move the tabs: %q", tt.Texts())
	}

	// The wheel has no height to scroll in, so it moves the tabs sideways.
	before := at.X
	tt.Scroll(400, TitleBarHeight/2, 0, -99)
	tt.Frame()
	if at.X >= before {
		t.Fatalf("the wheel did not move the tabs: X = %v, was %v", at.X, before)
	}
	if len(act.moved) == 0 || act.moved[0] != -99 {
		t.Errorf("the wheel asked for %v, want -99", act.moved)
	}

	// The button at the start shows the tabs before them.
	before = at.X
	tt.ClickAt(stepWidth/2, TitleBarHeight/2)
	tt.Frame()
	if at.X >= before {
		t.Errorf("the button at the start did not move the tabs: X = %v, was %v", at.X, before)
	}
}
