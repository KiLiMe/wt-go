// Package theme is the colors the window and the terminal draw with: the
// scheme the settings name, and the colors of a scheme as settings.json
// writes them. It is pure: no file is read and no process is started.
package theme

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
)

// Terminal is the colors the terminal draws with: the scheme the settings
// name, else the one of the window's theme.
func Terminal(s store.Settings) *terminal.Theme {
	if t, ok := Scheme(s, s.ColorScheme); ok {
		return t
	}
	return base(s)
}

// Color is a color as settings.json writes it, "#rgb", "#rgba", "#rrggbb" or
// "#rrggbbaa", and no color when the string is not one. A scheme may name any
// color, and a typo is no color rather than a panic.
func Color(s string) (myui.Color, bool) {
	s = strings.TrimSpace(s)
	if len(s) < 4 || len(s) > 9 || s[0] != '#' {
		return myui.Color{}, false
	}
	switch len(s) {
	case 4, 5, 7, 9:
	default:
		return myui.Color{}, false
	}
	for i := 1; i < len(s); i++ {
		if !isHexDigit(s[i]) {
			return myui.Color{}, false
		}
	}
	return myui.Hex(s), true
}

// Hex is c as mygo writes a window's background, "#rrggbb".
func Hex(c myui.Color) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// base is the colors the terminal falls back on: dark, unless the settings ask
// for light, or name a theme of their own that is light. A window that follows
// the desktop falls back on dark, which is what the terminal did before the
// window and the terminal were told apart.
func base(s store.Settings) *terminal.Theme {
	if dark, known := Dark(s); known && !dark {
		return terminal.LightTheme()
	}
	return terminal.DarkTheme()
}

// Dark reports whether the window is dark, and whether the settings say so at
// all: "light" and "dark" say outright, and a theme of their own says through
// the background it paints. "system" is none of them — the desktop answers
// that one — and neither is a name no scheme has.
func Dark(s store.Settings) (dark, known bool) {
	switch s.Theme {
	case "light":
		return false, true
	case "dark":
		return true, true
	case "", "system":
		return false, false
	}
	// A scheme the file writes wins over the one of that name it is written
	// over, the way Scheme resolves them.
	if sc, own := custom(s, s.Theme); own {
		if bg, ok := Color(sc.Background); ok {
			return luminance(bg) < 0.5, true
		}
	}
	if th, known := named(s.Theme); known {
		return luminance(th.Background) < 0.5, true
	}
	return false, false
}

// override sets *c to the color s writes, and keeps it when s writes none.
func override(c *myui.Color, s string) {
	if v, ok := Color(s); ok {
		*c = v
	}
}

// isHexDigit reports whether c is a hexadecimal digit.
func isHexDigit(c byte) bool {
	switch {
	case '0' <= c && c <= '9', 'a' <= c && c <= 'f', 'A' <= c && c <= 'F':
		return true
	}
	return false
}
