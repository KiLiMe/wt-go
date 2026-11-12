package ui

import (
	"math"

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
	if contrast(c, ink) > 1.8*contrast(c, paper) {
		return ink
	}
	return paper
}

// contrast is WCAG's contrast ratio of two colors, from 1 to 21.
func contrast(a, b myui.Color) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// luminance is WCAG's relative luminance of c, from 0 to 1.
func luminance(c myui.Color) float64 {
	channel := func(v uint8) float64 {
		f := float64(v) / 255
		if f <= 0.04045 {
			return f / 12.92
		}
		return math.Pow((f+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(c.R) + 0.7152*channel(c.G) + 0.0722*channel(c.B)
}
