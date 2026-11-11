//go:build linux

package app

import "github.com/egoist/mygo"

// titleBarStyle hides the title bar, which makes the tab strip the title bar:
// the strip draws the window's buttons itself (hideWindowButtons), in the
// colors of the terminal rather than the colors of the desktop's theme. GTK
// still decorates the window where the compositor leaves it to its clients,
// which keeps the shadow and the edges that resize the window.
func titleBarStyle() mygo.TitleBarStyle { return mygo.TitleBarHidden }

// hideWindowButtons keeps GTK's title buttons off the window. GTK draws them
// only in a header bar, which the desktop's theme styles: without one there
// is nothing of the title bar that a theme can reach. The buttons the strip
// draws follow the desktop's layout (Theme.WindowButtons), and are none where
// the compositor draws a title bar of its own.
func hideWindowButtons() bool { return true }
