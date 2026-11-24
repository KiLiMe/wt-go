package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/theme"
)

// colorOf is a color as settings.json writes it, and no color when the string
// is not one: a profile may name any color, and a typo is no color rather
// than a panic.
func colorOf(s string) myui.Color {
	c, _ := theme.Color(s)
	return c
}

var (
	// paper and ink are what a mark is drawn in on a face it did not choose:
	// white, and a near-black that reads as one of the window's own dark
	// themes.
	paper = myui.Hex("#ffffff")
	ink   = myui.Hex("#101010")
)

// inkOn is the color a mark takes on a face of the color c. A shell may wear a
// pale color (#cccccc for cmd, #e5c07b for zsh), where a white mark all but
// disappears; black is taken only when it stands out clearly better, so that
// the blue and orange faces keep the white mark the rest of the window would
// draw on them.
func inkOn(c myui.Color) myui.Color {
	if theme.Contrast(c, ink) > 1.8*theme.Contrast(c, paper) {
		return ink
	}
	return paper
}
