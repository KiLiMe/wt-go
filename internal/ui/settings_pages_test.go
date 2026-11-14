package ui

import (
	myui "github.com/egoist/mygo/ui"
	"testing"
	"wtgo/internal/store"
	"wtgo/locales"
)

// TestStartupPageShowsTheSystemForAProfileThatIsGone covers the select of the
// default profile: a file that names one the page does not offer runs the
// system's shell, so the select says so rather than naming a profile that is
// not there, and drawing the page leaves the draft as the file wrote it.
func TestStartupPageShowsTheSystemForAProfileThatIsGone(t *testing.T) {
	s := store.Default()
	s.DefaultProfile = "gone"
	act := &navActions{}
	known := []Profile{{ID: "pwsh7", Name: "PowerShell 7"}}
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageStartup, Known: known}, act)
	}, 900, 700)
	tt.Frame()
	if tt.HasText("gone") {
		t.Fatal("the select names a profile the page does not offer")
	}
	if s.DefaultProfile != "gone" {
		t.Fatalf("drawing the page changed the draft: DefaultProfile = %q", s.DefaultProfile)
	}
	if len(act.defaults) != 0 {
		t.Fatalf("drawing the page asked for %v, want nothing", act.defaults)
	}
	// A choice asks the window for it rather than writing the draft, which
	// the window is the one to do. The select opens on a press of its
	// trigger, which shows the option in force, the device, and a press on
	// another option closes the popup and names it. The press goes on the
	// option itself: a press on the row already in force changes nothing, so
	// the select does not report it.
	trigger, ok := tt.Find(locales.SystemProfile())
	if !ok {
		t.Fatal("the select for the default profile is missing")
	}
	x, y := trigger.X+trigger.W/2, trigger.Y+trigger.H/2
	tt.ClickAt(x, y)
	if n := countTexts(tt.Texts(), locales.SystemProfile()); n != 2 {
		t.Fatalf("the select opened %d choices for the device, want 2", n)
	}
	option, ok := tt.Find(known[0].Name)
	if !ok {
		t.Fatalf("the popup does not offer %q", known[0].Name)
	}
	// The window is told the profile's id, not the name the select shows.
	tt.ClickAt(option.X+option.W/2, option.Y+option.H/2)
	if len(act.defaults) != 1 || act.defaults[0] != known[0].ID {
		t.Fatalf("asked for %v, want %q", act.defaults, known[0].ID)
	}
	if s.DefaultProfile != "gone" {
		t.Errorf("the page wrote the draft itself: DefaultProfile = %q", s.DefaultProfile)
	}
}

// TestStartupPageAsksBeforeTheWindowCloses covers the switch of
// confirmCloseWindow: it sits under the switch that closes the window, since
// with that one off there is no window close to ask about, and a click edits
// the draft.
func TestStartupPageAsksBeforeTheWindowCloses(t *testing.T) {
	s := store.Default()
	s.ConfirmCloseWindow = false
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageStartup}, &navActions{})
		// Tall enough for the group the two switches sit in: the sections of
		// the window above it put them past the default height.
	}, 900, 1400)
	tt.Frame()
	sw := switchOf(t, tt, locales.ConfirmCloseWindow())
	last := switchOf(t, tt, locales.CloseOnLastTab())
	if sw.Bounds.Y <= last.Bounds.Y {
		t.Fatalf("the question at %v is not under the switch it goes with at %v", sw.Bounds, last.Bounds)
	}
	tt.ClickAt(sw.Bounds.X+sw.Bounds.W/2, sw.Bounds.Y+sw.Bounds.H/2)
	if !s.ConfirmCloseWindow {
		t.Fatal("a click on the switch left the draft as it was")
	}
}

// TestSettingsBannerShowsTheRefusal covers what the page says when a save
// refused the draft: the reason is above the fields, in the danger color, so it
// is not cut off by the foot of the page or hidden by whatever the last action
// said, and the foot says what to do about it.
func TestSettingsBannerShowsTheRefusal(t *testing.T) {
	s := store.Default()
	const reason = "settings: font.size must be between 4 and 200"
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageAppearance, Validation: reason, Dirty: true}, &navActions{})
		// Tall enough for the first field of the page to be in view under the
		// banner, rather than unfolded out of the scroll.
	}, 900, 1100)
	tt.Frame()
	reasonAt, ok := tt.Find(reason)
	if !ok {
		t.Fatalf("the page does not say why the save was refused: %q", tt.Texts())
	}
	if !tt.HasText(locales.SaveRefusedNote()) {
		t.Fatalf("the foot does not say what to do about the refusal: %q", tt.Texts())
	}
	if tt.HasText(locales.UnsavedChanges()) {
		t.Error("the foot says the draft is unsaved rather than what is wrong with it")
	}
	// Above the fields: past the column of pages, and above the first field
	// of the page the save was refused on.
	if reasonAt.X < navWidth {
		t.Fatalf("the reason at %v is in the column of pages", reasonAt)
	}
	first, ok := tt.Find(locales.Theme())
	if !ok {
		t.Fatal("the page does not show its fields")
	}
	if reasonAt.Y >= first.Y {
		t.Fatalf("the reason at %v is under the field it is about at %v", reasonAt.Y, first.Y)
	}
}

// TestSettingsWithoutARefusalShowsNone covers the page of a draft nothing is
// wrong with: no banner, and the foot says the draft is unsaved.
func TestSettingsWithoutARefusalShowsNone(t *testing.T) {
	s := store.Default()
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageProfiles, Dirty: true}, &navActions{})
	}, 900, 700)
	tt.Frame()
	if tt.HasText(locales.SaveRefusedNote()) {
		t.Fatal("the page says a save was refused while none was")
	}
	if !tt.HasText(locales.UnsavedChanges()) {
		t.Fatalf("the foot does not say the draft is unsaved: %q", tt.Texts())
	}
}
