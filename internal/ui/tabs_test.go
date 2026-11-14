package ui

import (
	"image/color"
	"slices"
	"strings"
	"testing"

	mygoi18n "github.com/egoist/mygo/i18n"
	myui "github.com/egoist/mygo/ui"

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

// TestStripButtons covers the buttons after the last tab: they follow it
// rather than the far end of the strip, and the one that opens a tab comes
// before the menu that picks the profile it runs.
func TestStripButtons(t *testing.T) {
	tt, _, _ := stripOf(t, 1, 0, 900)
	name, ok := tt.Find("PowerShell 7")
	if !ok {
		t.Fatal("the tab does not draw its name")
	}
	open, ok := tt.Find(locales.NewTab())
	if !ok {
		t.Fatalf("the strip draws no button to open a tab: %q", tt.Texts())
	}
	menu, ok := tt.Find(locales.Profiles())
	if !ok {
		t.Fatalf("the strip draws no menu of profiles: %q", tt.Texts())
	}
	if !(name.X < open.X && open.X < menu.X) {
		t.Errorf("the buttons sit at %v and %v after a name at %v", open.X, menu.X, name.X)
	}
	// They belong beside the tab they open next to, not at the far end of a
	// wide window, where the buttons that move the window are.
	if menu.X > 300 {
		t.Errorf("the buttons sit at %v and %v in a 900 DIP strip, want them beside the tab at %v",
			open.X, menu.X, name.X)
	}
}

// TestTabFillsStrip covers the tab in the strip: it is the strip's whole
// height, as Windows Terminal's is, so that the surface above it does not
// push the pane down, and it starts in from the window's edge.
func TestTabFillsStrip(t *testing.T) {
	tt, _, _ := stripOf(t, 1, 0, 900)
	img := tt.Image()
	px := func(x, y int) color.RGBA { return img.RGBAAt(x, y) }
	// The far end of the strip is the surface, with nothing drawn over it.
	surface := px(890, 2)
	if surface.A != 0xff {
		t.Fatalf("the strip draws %v at its far end", surface)
	}
	if got := px(int(stripGutter)/2, TitleBarHeight-1); got != surface {
		t.Errorf("the strip draws %v left of the tab, want the surface %v", got, surface)
	}
	// The tab reaches both edges of the strip: no surface above or below it.
	for _, y := range []int{0, TitleBarHeight - 1} {
		if got := px(int(stripGutter)+40, y); got == surface {
			t.Errorf("the strip draws the surface %v at height %d, over the tab", got, y)
		}
	}
}

// TestTabTakesThePanesFace covers the tab in front: it takes the color the
// window says the pane below it paints, not a shade of the window's theme, so
// that a light window holding a dark terminal joins the two rather than seaming
// between them. The name on it reads in the pane's text color.
func TestTabTakesThePanesFace(t *testing.T) {
	pane := Face{Background: myui.RGB(0x0c, 0x0c, 0x0c), Foreground: myui.RGB(0xcc, 0xcc, 0xcc)}
	st := State{Tabs: []Tab{{Name: "PowerShell 7"}, {Name: "cmd"}}, Active: 1, Pane: pane}
	at := &myui.ScrollState{}
	tt := myui.NewTester(func(c *myui.Context) { strip(c, st, &scrollActions{at: at}) }, 900, TitleBarHeight*2)
	img := tt.Image()
	rgba := func(c myui.Color) color.RGBA { return color.RGBA{R: c.R, G: c.G, B: c.B, A: 0xff} }
	px := func(x, y int) color.RGBA { return img.RGBAAt(x, y) }
	surface := px(int(stripGutter)-1, 2)
	if surface == rgba(pane.Background) {
		t.Fatalf("the strip draws the pane's background at its far end: the test proves nothing")
	}
	// The top of a tab is above the text in it, and the tab in front is the
	// one whose name the second tab carries.
	name, ok := tt.Find("cmd")
	if !ok {
		t.Fatalf("the strip draws no name for the tab in front: %q", tt.Texts())
	}
	if got := px(int(name.X)+2, 2); got != rgba(pane.Background) {
		t.Errorf("the tab in front draws %v, want the pane's %v", got, pane.Background)
	}
	first, ok := tt.Find("PowerShell 7")
	if !ok {
		t.Fatalf("the strip draws no name for the tab behind: %q", tt.Texts())
	}
	if got := px(int(first.X)+2, 2); got != surface {
		t.Errorf("the tab behind draws %v, want the surface %v", got, surface)
	}
}

// TestTabOfAPageTakesThemesOwn covers the tab in front when it draws a page of
// its own, the settings: it has no pane of a terminal's below it, so it takes
// the theme's own background rather than the pane's.
func TestTabOfAPageTakesThemesOwn(t *testing.T) {
	pane := Face{Background: myui.RGB(0x0c, 0x0c, 0x0c), Foreground: myui.RGB(0xcc, 0xcc, 0xcc)}
	page := func(c *myui.Context) myui.Element { return myui.Box(c) }
	st := State{Tabs: []Tab{{Name: "Settings", Page: page}}, Active: 0, Pane: pane}
	at := &myui.ScrollState{}
	tt := myui.NewTester(func(c *myui.Context) { strip(c, st, &scrollActions{at: at}) }, 900, TitleBarHeight*2)
	name, ok := tt.Find("Settings")
	if !ok {
		t.Fatalf("the strip draws no name for the tab in front: %q", tt.Texts())
	}
	if got := tt.Image().RGBAAt(int(name.X)+2, 2); got == (color.RGBA{R: 0x0c, G: 0x0c, B: 0x0c, A: 0xff}) {
		t.Errorf("the tab of a page draws the pane's background %v, want the theme's own", got)
	}
}

// TestTabWidths covers the room a tab takes: as wide as its name asks for
// while the strip has room, narrower as more tabs come, and never narrower than
// the parts of a tab need -- past that the strip scrolls instead of squeezing a
// name and a close button away.
func TestTabWidths(t *testing.T) {
	alone := tabRoomIn(t, 1, 2400)
	if alone > maxTabWidth {
		t.Errorf("one tab alone takes %v DIPs, want no more than the widest tab, %v", alone, maxTabWidth)
	}
	if alone < minTabWidth {
		t.Errorf("one tab alone takes %v DIPs, want no less than its parts, %v", alone, minTabWidth)
	}
	crowded := tabRoomIn(t, 6, 900)
	if !(crowded < alone) {
		t.Errorf("six tabs take %v DIPs each, one tab alone %v: the tabs do not narrow as more come",
			crowded, alone)
	}
	// Forty tabs cannot fit however narrow they get: they stop at the room
	// their parts need, and the strip scrolls the rest.
	tt, _, at := stripOf(t, 40, 0, 900)
	if at.MaxX <= 0 {
		t.Fatalf("forty tabs in 900 DIPs fit: MaxX = %v", at.MaxX)
	}
	if got := tabRoomIn(t, 40, 900); got != minTabWidth {
		t.Errorf("forty tabs take %v DIPs each, want the room their parts need, %v", got, minTabWidth)
	}
	// A maximized window has room to spare even for long names: those tabs
	// each take the widest a tab gets, and stop there.
	long := "Windows PowerShell (Administrator)"
	if got := tabRoomOfNamed(t, 3, 1936, long); got != maxTabWidth {
		t.Errorf("three long names in a 1936 DIP strip take %v DIPs each, want %v", got, maxTabWidth)
	}
	if tt.HasText("\u2026") {
		t.Errorf("the name of a tab is ellipsized before the strip scrolls")
	}
}

// tabRoomIn draws n tabs in a strip of the width it gives, and says how wide
// one of them is.
func tabRoomIn(t *testing.T, n, width int) float32 {
	t.Helper()
	tt, _, _ := stripOf(t, n, 0, width)
	name, ok := tt.Find("PowerShell 7")
	if !ok {
		t.Fatalf("%d tabs draw no name: %q", n, tt.Texts())
	}
	return name.W + tabChrome
}

// tabRoomOfNamed is tabRoomIn for tabs that carry the name it gives.
func tabRoomOfNamed(t *testing.T, n, width int, name string) float32 {
	t.Helper()
	tt, _, _ := stripOfNamed(t, n, 0, width, name)
	found, ok := tt.Find(name)
	if !ok {
		t.Fatalf("%d tabs name %q draw no name: %q", n, name, tt.Texts())
	}
	return found.W + tabChrome
}

// TestTabsShare covers the room a tab is given out of the strip's: the strip
// split among the tabs, less the lines between them, and the lines beside the
// tab in front left out, that tab being the dark one they would cut.
func TestTabsShare(t *testing.T) {
	for _, tc := range []struct {
		tabs, active int
		room, want   float32
	}{
		// Four tabs and the three places a line may take: the tab in front
		// takes no line beside it, and that line keeps its room anyway, so the
		// share is the same whichever tab is in front.
		{4, 1, 301, 74.5},
		{4, 0, 301, 74.5},
		{4, 3, 301, 74.5},
	} {
		st := State{Tabs: make([]Tab, tc.tabs), Active: tc.active}
		if got := tabsShare(st, tc.room); got != tc.want {
			t.Errorf("with tab %d of %d in front, %v DIPs split to %v a tab, want %v",
				tc.active, tc.tabs, tc.room, got, tc.want)
		}
	}
	if got := tabsShare(State{}, 900); got != 0 {
		t.Errorf("a strip with no tabs splits %v DIPs, want none", got)
	}
}

// TestStripDividers covers the lines the strip draws: one DIP wide, sixteen
// high and centered in the strip, between two tabs that are both behind and on
// both sides of the button that opens a tab — but none beside the tab in front,
// which is the dark rounded one the line would cut, and that counts for the
// line on the left of the button too when that button sits against it.
func TestStripDividers(t *testing.T) {
	for _, tc := range []struct {
		tabs, active, want int
	}{
		// One tab, and it is in front: only the line past the button.
		{1, 0, 1},
		// Two tabs: the one behind keeps its lines, the one in front keeps none.
		{2, 0, 2},
		{2, 1, 1},
		// Three tabs: the line between two of them behind, plus the button's.
		{3, 0, 3},
		{3, 1, 2},
		{3, 2, 2},
	} {
		tt, _, _ := stripOf(t, tc.tabs, tc.active, 900)
		open, ok := tt.Find(locales.NewTab())
		if !ok {
			t.Fatal("the strip draws no button to open a tab")
		}
		img := tt.Image()
		px := func(x, y int) color.RGBA { return img.RGBAAt(x, y) }
		mid := int(TitleBarHeight) / 2
		surface := px(890, 2)
		var found []int
		// The window's own buttons are past the one that picks a profile, so
		// this stops before them.
		for x := 1; x < int(open.X)+60; x++ {
			if px(x, mid) == surface || px(x-1, mid) != surface || px(x+1, mid) != surface {
				continue
			}
			// A mark of the strip is as high as a tab's, so the rows above and
			// below the line are the surface again.
			if px(x, 6) != surface || px(x, 23) != surface {
				continue
			}
			if px(x, 7) == surface || px(x, 22) == surface {
				continue
			}
			found = append(found, x)
		}
		if len(found) != tc.want {
			t.Errorf("with tab %d of %d in front the strip draws %d lines, want %d: %v",
				tc.active, tc.tabs, len(found), tc.want, found)
		}
	}
}

// TestStripButtonsStayPut covers where the buttons after the tabs stand: the
// line on the left of the button that opens a tab is not drawn beside the tab
// in front, and keeps its room anyway, so that the buttons do not move as
// another tab comes to the front. Windows Terminal's strip holds them still the
// same way: with either of two tabs in front its buttons stand in the same
// place, and only that line comes and goes.
//
// The tabs are of different lengths, and one set of them is short enough that
// the strip gives each tab the width its name asks for rather than an equal
// share: the tab in front names itself in bold, which is a few DIPs wider, and
// a strip whose buttons followed the tabs' own width would move them by that
// much every time the front tab changed.
func TestStripButtonsStayPut(t *testing.T) {
	sets := [][]string{
		{"PowerShell 7", "PowerShell 7", "PowerShell 7"},
		{"PowerShell 7"},
		{"PowerShell 7", "π - wt-go", "长一点的标签名字在这里", "pwsh"},
	}
	for _, names := range sets {
		for _, width := range []int{520, 900, 1200} {
			for _, buttons := range []bool{false, true} {
				var want [2]float32
				for active := range names {
					tt := stripOfEach(t, names, active, width, buttons)
					open, ok := tt.Find(locales.NewTab())
					if !ok {
						t.Fatal("the strip draws no button to open a tab")
					}
					menu, ok := tt.Find(locales.Profiles())
					if !ok {
						t.Fatal("the strip draws no menu of profiles")
					}
					got := [2]float32{open.X, menu.X}
					if active == 0 {
						want = got
						continue
					}
					if got != want {
						t.Errorf("at %d DIPs with tab %d of %d in front, the buttons stand at %v, want %v",
							width, active, len(names), got, want)
					}
				}
			}
		}
	}
}

// stripOfEach is a strip of tabs carrying the names given, with the window's
// own buttons where buttons is set.
func stripOfEach(t *testing.T, names []string, active int, width int, buttons bool) *myui.Tester {
	t.Helper()
	at := &myui.ScrollState{}
	act := &scrollActions{at: at}
	st := &State{Tabs: make([]Tab, len(names)), Active: active, TabsScroll: at}
	if buttons {
		st.WindowButtons = &WindowButtons{End: []string{"minimize", "maximize", "close"}}
	}
	for i := range st.Tabs {
		st.Tabs[i].Name = names[i]
	}
	tt := myui.NewTester(func(c *myui.Context) { strip(c, *st, act) }, width, TitleBarHeight*2)
	tt.Frame()
	return tt
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
func (a *windowActions) OpenSettings()         { a.asked = append(a.asked, "settings") }

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
	return stripOfNamed(t, n, active, width, "PowerShell 7")
}

// stripOfNamed is stripOf for tabs that carry the name it gives.
func stripOfNamed(t *testing.T, n int, active int, width int, name string) (*myui.Tester, *scrollActions, *myui.ScrollState) {
	t.Helper()
	at := &myui.ScrollState{}
	act := &scrollActions{at: at}
	st := &State{Tabs: make([]Tab, n), Active: active, TabsScroll: at}
	for i := range st.Tabs {
		st.Tabs[i].Name = name
	}
	tt := myui.NewTester(func(c *myui.Context) { strip(c, *st, act) }, width, TitleBarHeight*2)
	tt.Frame()
	return tt, act, at
}

// TestTabNameFillsItsRoom covers the room a tab gives its name: the whole of
// it, so that the button that closes the tab sits on the tab's edge rather
// than a gap short of it, where a short name leaves nothing drawn.
func TestTabNameFillsItsRoom(t *testing.T) {
	tt, _, _ := stripOfNamed(t, 1, 0, 900, "pwsh")
	name, ok := tt.Find("pwsh")
	if !ok {
		t.Fatalf("the tab draws no name: %q", tt.Texts())
	}
	if name.W != minTabName {
		t.Errorf("the name takes %v DIPs of the %v a tab gives it: the close button stands %v short of the edge",
			name.W, minTabName, minTabName-name.W)
	}
}

// TestMinWindowWidth covers the floor the window is held to: at it the strip
// still holds a tab, and the buttons that move the tabs and the desktop's own
// buttons are drawn beside it rather than over it.
func TestMinWindowWidth(t *testing.T) {
	width := int(MinWindowWidth)
	tt, _, at := stripOf(t, 20, 19, width)
	if at.MaxX <= 0 {
		t.Fatalf("twenty tabs of %v DIPs each fit a window of %d: MaxX = %v", minTabWidth, width, at.MaxX)
	}
	if !tt.HasText(locales.ScrollTabsLeft()) || !tt.HasText(locales.ScrollTabsRight()) {
		t.Errorf("the strip hides the buttons that move the tabs: %q", tt.Texts())
	}
	if !tt.HasText("PowerShell 7") {
		t.Errorf("the strip draws no tab of the twenty: %q", tt.Texts())
	}
}

// TestMinWindowHeight covers the floor of the height: the title bar, then a
// pane that still holds a line of output below its padding, rather than a
// sliver of a window.
func TestMinWindowHeight(t *testing.T) {
	if MinWindowHeight != TitleBarHeight+3*16+2*8 {
		t.Errorf("the shortest window is %v DIPs tall, want the title bar %v and three rows with the padding",
			MinWindowHeight, TitleBarHeight)
	}
	if MinWindowHeight <= TitleBarHeight+2*8+16 {
		t.Errorf("the pane of the shortest window is %v DIPs tall: less than a line of output with its padding",
			MinWindowHeight-TitleBarHeight)
	}
}

// TestStripFits covers a strip the tabs do not outgrow: it does not scroll,
// it shows no buttons to move the tabs, and every tab is as wide as it would
// be alone.
func TestStripFits(t *testing.T) {
	tt, _, at := stripOf(t, 2, 0, 900)
	if at.MaxX != 0 {
		t.Fatalf("two tabs in 900 DIPs scroll: MaxX = %v", at.MaxX)
	}
	if tt.HasText(locales.ScrollTabsLeft()) || tt.HasText(locales.ScrollTabsRight()) {
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
	if !tt.HasText(locales.ScrollTabsLeft()) || !tt.HasText(locales.ScrollTabsRight()) {
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

// TestMarkIcon covers the three kinds of mark a tab wears: a shape drawn as it
// is, a character in the square of the profile's color, and the plain terminal
// mark, with and without a color, in front and behind.
func TestMarkIcon(t *testing.T) {
	tt := myui.NewTester(func(c *myui.Context) {
		myui.Row(c).Children(func() {
			markIcon(c, "bash:arch", "#4eaa25", 16, true)
			markIcon(c, "\U0001F92C", "#3b78ff", 16, true)
			markIcon(c, "\U0001F92C", "", 16, false)
			markIcon(c, "", "#cccccc", 16, true)
			markIcon(c, "", "", 16, false)
		})
	}, 200, 24)
	tt.Frame()
	// The character is written rather than drawn, so the frame says it twice:
	// once on the profile's color and once on the quiet square.
	if got := strings.Count(strings.Join(tt.Texts(), ""), "\U0001F92C"); got != 2 {
		t.Fatalf("the character is drawn %d times, want 2: %q", got, tt.Texts())
	}
}

// TestMarkIconPreset paints a preset on its own: it is a picture, so the frame
// holds the tones it was drawn with, not the one flat color a glyph would be.
func TestMarkIconPreset(t *testing.T) {
	tt := myui.NewTester(func(c *myui.Context) {
		markIcon(c, "bash:ubuntu", "", 32, true)
	}, 32, 32)
	tt.Frame()
	tones := map[color.RGBA]struct{}{}
	img := tt.Image()
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			tones[color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}] = struct{}{}
		}
	}
	if len(tones) < 4 {
		t.Fatalf("the preset drew %d colors, want its own", len(tones))
	}
}
