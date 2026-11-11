// Package icons is the window's icons: SVGs from Remix Icon
// (https://remixicon.com, github.com/Remix-Design/RemixIcon, v4.8.0,
// Apache License 2.0; see LICENSE and NOTICE in this directory), embedded in
// the binary and drawn by mygo's ui.Icon, in the color of the text around
// them.
//
// A glyph will not do: "+" and "\u00d7" are missing from some fonts, or wide in the fonts
// that draw them, which leaves a button blank or its mark off its box. To add
// an icon, copy icons/<Category>/<name>.svg from the Remix Icon v4.8.0 tag
// into svg/ unchanged, and add a variable below with its file name.
package icons

import (
	"embed"

	"github.com/egoist/mygo/ui"
)

//go:embed svg/*.svg
var files embed.FS

// Icon is one parsed icon: the name of the file it came from, and the SVG
// that ui.Icon draws.
type Icon struct {
	Name string
	SVG  *ui.SVG
}

// load parses the icon name, whose file is in svg/. A missing file is a
// mistake in this package, which fails at start rather than drawing nothing.
func load(name string) *Icon {
	data, err := files.ReadFile("svg/" + name + ".svg")
	if err != nil {
		panic("icons: no svg/" + name + ".svg")
	}
	return &Icon{Name: name, SVG: ui.MustParseSVG(data)}
}

// The icons the window draws, by what they mean.
var (
	Add     = load("add-line")          // the tab strip's new tab
	Chevron = load("arrow-down-s-line") // the menu of the profiles, the match after
	Up      = load("arrow-up-s-line")   // the find bar's match before
	Left    = load("arrow-left-s-line") // the tabs the strip scrolled past
	Right   = load("arrow-right-s-line")
	Close   = load("close-line") // a close button
)
