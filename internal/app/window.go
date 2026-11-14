package app

import (
	"github.com/egoist/mygo"
	myui "github.com/egoist/mygo/ui"
	"wtgo/internal/theme"
	"wtgo/internal/ui"
)

// opacity is the opacity of the window the settings ask for, between the 0.3
// that keeps the text readable and the 1 that hides nothing.
func opacity(v float64) float64 {
	switch {
	case v <= 0, v > 1:
		return 1
	case v < 0.3:
		return 0.3
	}
	return v
}

// background is the color the window fills what the page has not painted yet:
// the terminal's own, which the tab in front carries up into the strip.
func (a *App) background() myui.Color {
	return theme.Terminal(a.settings).Background
}

// appearance gives the window the appearance the settings ask for: dark and
// light are the settings', system is the machine's, and a theme of its own is
// the one that theme is. The window draws in that theme, or in mygo's own for
// the appearance; the terminal's colors are the terminal's business, so that a
// light window can hold a dark terminal without a seam between them.
func (a *App) appearance() {
	// A theme of the settings' own says both what the window draws with and
	// which way the desktop's own menus and dialogs go.
	if t, ok := theme.UITheme(a.settings, a.settings.Theme); ok {
		a.uiTheme = t
		darkMode(t.Dark)
		return
	}
	dark, known := theme.Dark(a.settings)
	if known {
		darkMode(dark)
	} else {
		// Following the desktop: ask it, then draw in the colors its own
		// windows, menus and dialogs are drawn in.
		mygo.Theme.SetSource(mygo.ThemeSystem)
		dark = mygo.Theme.IsDark()
	}
	a.uiTheme = theme.Native(dark)
}

// darkMode asks the desktop's own windows, menus and dialogs to be dark or
// light, which is the appearance the settings name.
func darkMode(dark bool) {
	if dark {
		mygo.Theme.SetSource(mygo.ThemeDark)
		return
	}
	mygo.Theme.SetSource(mygo.ThemeLight)
}

// windowButtons are the buttons the strip draws where the window hides the
// desktop's: the ones the desktop's layout names, at the sides it puts them,
// and whether the window is maximized. It is nil where mygo draws them, or
// where the compositor draws a title bar of its own.
func (a *App) windowButtons() *ui.WindowButtons {
	if !hideWindowButtons() {
		return nil
	}
	// A strip that draws none is the same as a strip that hides them, and nil
	// says so without building one every frame.
	if len(a.buttons.Start) == 0 && len(a.buttons.End) == 0 {
		return nil
	}
	return &ui.WindowButtons{Start: a.buttons.Start, End: a.buttons.End, Maximized: a.maximized}
}

// watchWindowButtons keeps the strip's copy of the desktop's buttons in step:
// the layout changes when the user picks another one, and the window maximizes
// and restores as the user asks it to.
func (a *App) watchWindowButtons() {
	a.buttons = mygo.Theme.WindowButtons()
	a.maximized = a.win.IsMaximized()
	mygo.Theme.OnUpdated(func() {
		a.Do(func() {
			a.buttons = mygo.Theme.WindowButtons()
			a.redraw()
		})
	})
	a.win.OnMaximize(func() {
		a.Do(func() {
			a.maximized = true
			a.redraw()
		})
	})
	a.win.OnUnmaximize(func() {
		a.Do(func() {
			a.maximized = false
			a.redraw()
		})
	})
}

// MinimizeWindow minimizes the window, as its own button would.
func (a *App) MinimizeWindow() {
	a.Do(func() { a.win.Minimize() })
}

// ToggleMaximizeWindow maximizes the window, or restores it, as its own button
// would.
func (a *App) ToggleMaximizeWindow() {
	a.Do(func() { a.win.ToggleMaximize() })
}

// CloseWindow closes the window, as its own close button would, so that the
// shells the window runs still close with it.
func (a *App) CloseWindow() {
	a.Do(func() { a.win.Close() })
}
