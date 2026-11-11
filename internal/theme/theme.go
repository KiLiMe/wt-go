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

// base is the colors of the window's theme: dark, unless the settings ask for
// light, or name a theme of their own that is light.
func base(s store.Settings) *terminal.Theme {
	if light(s) {
		return terminal.LightTheme()
	}
	return terminal.DarkTheme()
}

// light reports whether the window is light: the appearance the settings name,
// or the background of the theme of their own. It reads the schemes without
// going through base, which asks it.
func light(s store.Settings) bool {
	switch s.Theme {
	case "light":
		return true
	case "", "system", "dark":
		return false
	}
	if th, known := named(s.Theme); known {
		return luminance(th.Background) >= 0.5
	}
	if sc, own := custom(s, s.Theme); own {
		if bg, ok := Color(sc.Background); ok {
			return luminance(bg) >= 0.5
		}
	}
	return false
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
