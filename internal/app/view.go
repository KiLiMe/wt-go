package app

import (
	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"
	"time"
	"wtgo/internal/pane"
	"wtgo/internal/ui"
)

// Boundary: one frame -- the state the window hands ui, the menu it builds,
// and the keys it binds. It reads the App and never changes it.

// view is the window's frame.
func (a *App) view(c *myui.Context) {
	if a.uiTheme != nil {
		c.SetTheme(a.uiTheme)
	}
	// The popup menus are drawn in the window with the theme's look rather
	// than by the desktop: a menu of the desktop's beside the window's own
	// strip reads as another program's.
	c.SetMenuPresentation(myui.PopoverMenus)
	if a.dismissed() {
		a.dropClosed()
	}
	a.content = ui.Content(c)
	a.apply()
	a.bind(c)
	st := ui.State{
		Active:     a.panes.Active,
		Pane:       a.pane,
		Profiles:   a.menuProfiles(),
		Device:     a.device.Name,
		Rename:     a.renaming,
		EqualTabs:  a.settings.TabWidth == "equal",
		Shortcut:   a.shortcut,
		TabsScroll: &a.tabsScroll,
		RevealTab:  a.revealTabs > 0,
		// The buttons the strip draws for the window, where the window hides
		// the desktop's.
		WindowButtons: a.windowButtons(),
	}
	// The strip brings the tab in front into view for the frames the window
	// asked for, and no more: asking on every frame would keep the user from
	// scrolling the strip away from that tab. Two frames are what it takes,
	// because the buttons that scroll the strip take their room only on the
	// frame after they first appear, and the strip is narrower by then.
	if a.revealTabs > 0 {
		a.revealTabs--
	}
	if a.confirm != nil {
		st.Confirm = &a.confirm.Confirm
	}
	st.About = a.about
	st.Palette = a.paletteState()
	st.TabsList = a.tabsList
	st.Find = a.findState()
	if a.settingsTab != nil {
		a.settingsShown = a.settingsState()
		st.Settings = &a.settingsShown
	}
	for _, tab := range a.panes.Tabs {
		st.Tabs = append(st.Tabs, a.uiTab(tab))
	}
	if a.showFrames {
		a.meter.tick(time.Now())
		st.Frames = a.meter.frames()
	}
	ui.View(c, st, a)
	// The mark over the setting the search found is one frame older than it
	// was drawn: the frame that scrolls to the setting is the first one, and
	// the fade is what the frames after it show.
	a.revealFrame()
	a.syncTitle()
}

// uiTab is a tab as the window draws it: the settings tab draws its pages,
// the others their panes.
func (a *App) uiTab(tab *pane.Tab) ui.Tab {
	out := ui.Tab{Name: a.tabName(tab), Color: tab.Color, Icon: tab.Icon, Panes: tab, Mount: a.mount}
	if tab == a.settingsTab {
		out.Page = a.settingsView
	}
	return out
}

// settingsView is the settings tab's page.
func (a *App) settingsView(c *myui.Context) myui.Element {
	return ui.SettingsTab(c, a.settingsShown, a)
}

// bind gives the window the keys of the tabs, unless the settings wait for a
// key of their own or an overlay has the keyboard. A key the keyboard repeats
// while it stays held runs an action that steps through something, and is left
// to the program for the rest: a held Ctrl+T makes one tab, not fifty.
func (a *App) bind(c *myui.Context) {
	if a.recording != "" || a.top() != overlayNone {
		return
	}
	for _, b := range a.keys {
		b := b
		c.OnShortcut(b.mods, b.key, func() {
			if c.ShortcutRepeated() && !b.repeats {
				return
			}
			a.run(b.action)
		})
	}
}

// mount shows the terminal of a pane.
func (a *App) mount(c *myui.Context, n *pane.Node) myui.Element {
	return terminal.View(c, n.Term)
}

// menuProfiles are the profiles the menu that opens tabs lists: the ones
// the settings show.
func (a *App) menuProfiles() []ui.Profile {
	out := make([]ui.Profile, 0, len(a.profiles))
	for _, p := range a.profiles {
		if p.Hidden {
			continue
		}
		out = append(out, ui.Profile{ID: p.ID, Name: p.Name, Color: p.Color, Icon: p.Icon})
	}
	return out
}
