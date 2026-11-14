package ui

import (
	mygoi18n "github.com/egoist/mygo/i18n"
	myui "github.com/egoist/mygo/ui"
	"slices"
	"testing"
	"wtgo/internal/action"
	"wtgo/locales"
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

// TestStripProfiles covers what the strip's profile menu lists: the profiles of
// the window, and with none the one that follows the device, so that a window
// without a profile still opens a tab.
func TestStripProfiles(t *testing.T) {
	got := stripProfiles(State{})
	if len(got) != 1 || got[0].ID != "" || got[0].Name != locales.SystemProfile() {
		t.Fatalf("stripProfiles with no profile = %+v, want the one that follows the device", got)
	}
	// The entry says which shell it opens, the way the settings page names it.
	got = stripProfiles(State{Device: "PowerShell 7"})
	if len(got) != 1 || got[0].Name != locales.SystemProfileOf("PowerShell 7") {
		t.Fatalf("stripProfiles = %+v, want the device named", got)
	}
	ps := []Profile{{ID: "pwsh7", Name: "PowerShell 7"}}
	if got := stripProfiles(State{Profiles: ps}); len(got) != 1 || got[0].ID != "pwsh7" {
		t.Fatalf("stripProfiles = %+v, want the profiles of the window", got)
	}
}

// TestTitleBarMenu covers the menu the strip opens on a right press away from
// the tabs: the window's own menu, which the desktop shows on a right press of
// the title bar it draws, and Windows Terminal offers on its strip.
func TestTitleBarMenu(t *testing.T) {
	act := &windowActions{}
	st := &State{Tabs: []Tab{{Name: "PowerShell 7"}}}
	tt := myui.NewTester(func(c *myui.Context) { strip(c, *st, act) }, 900, TitleBarHeight*2)
	tt.Frame()
	// The room that moves the window, past the buttons that follow the tabs.
	tt.RightClickAt(700, TitleBarHeight/2)
	menu := tt.Menu()
	for _, label := range []string{locales.Restore(), locales.Minimize(), locales.Maximize(),
		locales.CloseWindow(), locales.Settings()} {
		if !slices.Contains(menu, label) {
			t.Errorf("a right press on the empty strip opens %q, want %q in it", menu, label)
		}
	}
	if err := tt.ChooseMenuItem(locales.CloseWindow()); err != nil {
		t.Fatal(err)
	}
	if want := []string{"close"}; !slices.Equal(act.asked, want) {
		t.Errorf("the menu asked for %q, want %q", act.asked, want)
	}
	// A maximized window offers the item that restores it, not the one that
	// maximizes a window already maximized.
	st.WindowButtons = &WindowButtons{Maximized: true}
	tt.Frame()
	tt.RightClickAt(700, TitleBarHeight/2)
	if err := tt.ChooseMenuItem(locales.Restore()); err != nil {
		t.Fatal(err)
	}
	if want := []string{"close", "maximize"}; !slices.Equal(act.asked, want) {
		t.Errorf("the menu asked for %q, want %q", act.asked, want)
	}
}

// TestStripSpeaksEachLanguage covers the words the strip's menus say in every
// language the catalogs ship, the source language and the one translated from
// it: they are the catalogs' own words, so a window in English says them in
// English on a machine that speaks Chinese, and one in Chinese says them in
// Chinese. A language that has no word for one says the source language's,
// which is why the source language is the one every key is filed under.
func TestStripSpeaksEachLanguage(t *testing.T) {
	for _, tc := range []struct {
		locale string
		window []string
		close  string
		tab    []string
	}{
		{
			locale: "en",
			window: []string{"Restore", "Minimize", "Maximize", "-", "Close Window", "Settings"},
			close:  "Close Window",
			tab:    []string{"Rename", "Duplicate Tab", "Close Tab"},
		},
		{
			locale: "zh-Hans",
			window: []string{"还原", "最小化", "最大化", "-", "关闭窗口", "设置"},
			close:  "关闭窗口",
			tab:    []string{"重命名", "复制标签", "关闭标签"},
		},
	} {
		t.Run(tc.locale, func(t *testing.T) {
			speakIn(t, tc.locale)
			act := &windowActions{}
			st := &State{Tabs: []Tab{{Name: "PowerShell 7"}}}
			tt := myui.NewTester(func(c *myui.Context) { strip(c, *st, act) }, 900, TitleBarHeight*2)
			tt.Frame()
			tt.RightClickAt(700, TitleBarHeight/2)
			if menu := tt.Menu(); !slices.Equal(menu, tc.window) {
				t.Errorf("a right press on the empty strip opens %q, want %q", menu, tc.window)
			}
			// The menu is worked through the words the language says, not
			// through the keys behind them.
			if err := tt.ChooseMenuItem(tc.close); err != nil {
				t.Fatal(err)
			}
			if want := []string{"close"}; !slices.Equal(act.asked, want) {
				t.Errorf("the menu asked for %q, want %q", act.asked, want)
			}

			tt.Frame()
			name, ok := tt.Find("PowerShell 7")
			if !ok {
				t.Fatalf("the strip does not draw the tab: %q", tt.Texts())
			}
			tt.RightClickAt(name.X+name.W/2, name.Y+name.H/2)
			menu := tt.Menu()
			tt.CloseMenu()
			for _, label := range tc.tab {
				if !slices.Contains(menu, label) {
					t.Errorf("a right press on the tab opens %q, want %q in it", menu, label)
				}
			}
		})
	}
}

// speakIn draws the rest of the test in the language of tag and puts the
// language back afterwards: the message functions say the language the app
// sets, and no test in this package sets one, so they say the source language
// unless a test says otherwise.
func speakIn(t *testing.T, tag string) {
	t.Helper()
	mygoi18n.SetLocaleFunc(func() string { return tag })
	t.Cleanup(func() { mygoi18n.SetLocaleFunc(nil) })
}

// TestTabMenu covers the tab's menu: a right click anywhere on the tab opens
// it, over the name too, which is most of the tab.
func TestTabMenu(t *testing.T) {
	tt, _, _ := stripOf(t, 1, 0, 900)
	name, ok := tt.Find("PowerShell 7")
	if !ok {
		t.Fatal("the tab does not draw its name")
	}
	at := []struct {
		what string
		x, y float32
	}{
		{"the name", name.X + name.W/2, name.Y + name.H/2},
		{"the tab's left end", stripGutter + 4, TitleBarHeight / 2},
	}
	for _, p := range at {
		tt.RightClickAt(p.x, p.y)
		menu := tt.Menu()
		tt.CloseMenu()
		if !slices.Contains(menu, locales.Rename()) {
			t.Errorf("a right click on %s opens %q, want the tab's menu", p.what, menu)
		}
	}
}

// TestTabMenuOfSettings covers the menu of the tab that is not a terminal: it
// offers neither renaming nor a copy, which only a tab that runs a shell has.
func TestTabMenuOfSettings(t *testing.T) {
	at := &myui.ScrollState{}
	settings := Tab{Name: locales.Settings(), Page: func(c *myui.Context) myui.Element { return myui.Box(c) }}
	st := &State{Tabs: []Tab{{Name: "PowerShell 7"}, settings}, TabsScroll: at}
	tt := myui.NewTester(func(c *myui.Context) { strip(c, *st, &scrollActions{at: at}) }, 900, TitleBarHeight*2)
	tt.Frame()
	for _, tc := range []struct {
		name string
		own  bool
	}{{"PowerShell 7", true}, {settings.Name, false}} {
		box, ok := tt.Find(tc.name)
		if !ok {
			t.Fatalf("the strip does not draw %q: %q", tc.name, tt.Texts())
		}
		tt.RightClickAt(box.X+box.W/2, box.Y+box.H/2)
		menu := tt.Menu()
		tt.CloseMenu()
		for _, label := range []string{locales.Rename(), locales.DuplicateTab()} {
			if got := slices.Contains(menu, label); got != tc.own {
				t.Errorf("a right click on %q opens %q, want %q in it: %v", tc.name, menu, label, tc.own)
			}
		}
		if !slices.Contains(menu, locales.CloseTab()) {
			t.Errorf("a right click on %q opens %q, want the tab's menu", tc.name, menu)
		}
	}
}
