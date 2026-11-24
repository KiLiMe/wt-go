// Package ui draws the window: the tab strip, which is also the title bar,
// and the panes of the tab in front. It draws what the window hands it and
// asks the window to act through Actions: it reads no settings and starts no
// process.
package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/layout"
)

// Boundary: the frame -- what a window shows, and how the pieces are
// stacked. State and its parts are in state.go, the port to the window in
// actions.go, what the pages and dialogs hold in pages.go and dialogs.go.

// Content is the room the panes take: the window, less the strip it draws
// above them.
func Content(c *myui.Context) layout.Rect {
	w, h := c.Size()
	return layout.Rect{Y: TitleBarHeight, W: w, H: max(h-TitleBarHeight, 0)}
}

// View is the window: the tab strip, which is also the title bar, and the
// panes of the tab in front.
func View(c *myui.Context, st State, act Actions) {
	myui.Column(c).Fill().Children(func() {
		strip(c, st, act)
		content(c, st, act, Content(c))
	})
	rename(c, st, act)
	confirm(c, st, act)
	about(c, st, act)
	findBar(c, st, act)
	palette(c, st, act)
	tabsList(c, st, act)
	frames(c, st)
}
