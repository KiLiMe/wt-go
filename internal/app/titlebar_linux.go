//go:build linux

package app

import "github.com/egoist/mygo"

// titleBarStyle leaves the title bar to the platform. mygo's hidden style
// undecorates the window, which on Wayland lets the compositor draw a title
// bar but on X11 leaves neither a border nor a button: there the window
// manager draws it only when the window asks for decorations.
func titleBarStyle() mygo.TitleBarStyle { return mygo.TitleBarDefault }
