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
