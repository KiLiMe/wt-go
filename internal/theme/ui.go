package theme

import (
	"math"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
)

// uiThemes are the themes the window draws with, in the order the settings
// page offers them: schemes this program carries or Ghostty ships, whose
// colors the window takes as its own. Any scheme a setting names is a theme of
// the window; these are the ones the page lists.
var uiThemes = []string{
	"Nord",
	"Catppuccin Mocha",
	"Catppuccin Latte",
	"Gruvbox Dark",
	"Gruvbox Light",
	"Rose Pine",
	"Rose Pine Dawn",
	"Kanagawa Wave",
	"Dracula",
	"Solarized Dark",
	"Solarized Light",
}

// UIThemes are the names of the themes the settings page offers.
func UIThemes() []string { return uiThemes }

// UITheme is the theme the window draws with when the settings name one: the
// colors of the scheme called name, with the surfaces, borders and muted text
// a window needs derived from them. The appearance the settings name ("system",
// "light", "dark") is none, and so is a name no scheme has: those windows
// draw with Native.
func UITheme(s store.Settings, name string) (*myui.Theme, bool) {
	switch strings.ToLower(name) {
	case "", "system", "light", "dark":
		return nil, false
	}
	th, ok := Scheme(s, name)
	if !ok {
		return nil, false
	}
	return window(th), true
}

// Native is the theme of a window that names no theme of its own: the one mygo
// draws for an appearance, light or dark. Its colors are the desktop's own, so
// that the window reads as one of the system's windows next to the system's
// menus and dialogs, rather than as a shade of the terminal it holds.
func Native(dark bool) *myui.Theme {
	if dark {
		return myui.DarkTheme()
	}
	return myui.LightTheme()
}

// window is the theme of a window drawn in a scheme's colors: the scheme's
// background and text, its cursor for the accent, and the shades between the
// two that a window needs for its surfaces, borders and secondary text. They
// are derived rather than written per theme, so that a scheme this program has
// never seen makes a window too.
func window(th *terminal.Theme) *myui.Theme {
	bg, fg := th.Background, th.Foreground
	accent := accentOf(th)
	dark := luminance(bg) < 0.5
	t := *myui.LightTheme()
	if dark {
		t = *myui.DarkTheme()
	}
	t.Dark = dark
	t.Background = bg
	// The surfaces are far enough from the background to read as surfaces:
	// the strip has to stand out behind a tab that carries the background
	// itself, which a 7% shade of a #0c0c0c background does not do.
	t.Surface = bg.MixOklab(fg, 0.16)
	t.SurfaceHover = bg.MixOklab(fg, 0.22)
	t.SurfacePressed = bg.MixOklab(fg, 0.28)
	t.Border = bg.MixOklab(fg, 0.32)
	t.Text = fg
	t.TextMuted = fg.MixOklab(bg, 0.38)
	t.Accent = accent
	t.AccentHover = accent.MixOklab(fg, 0.2)
	t.AccentPressed = accent.MixOklab(fg, 0.32)
	t.AccentText = ink(accent, bg, fg)
	t.Selection = accent.Alpha(0.3)
	t.Focus = accent.Alpha(0.55)
	t.Scrollbar = fg.Alpha(0.35)
	return &t
}

// accentOf is the color a window of the scheme accents with: the blue of its
// palette, which every scheme writes, else the cursor where that is a color of
// its own, else its text. Cursors are mostly shades of gray, which would make
// a window with no accent at all.
func accentOf(th *terminal.Theme) myui.Color {
	if colored(th.Palette[4]) {
		return th.Palette[4]
	}
	if colored(th.Cursor) {
		return th.Cursor
	}
	return th.Foreground
}

// colored reports whether c is more than a shade of gray.
func colored(c myui.Color) bool {
	hi, lo := c.R, c.R
	for _, v := range [2]uint8{c.G, c.B} {
		hi, lo = max(hi, v), min(lo, v)
	}
	return hi-lo > 24
}

// ink is the color of text on top of c: the window's background or its text,
// whichever reads better there.
func ink(c, bg, fg myui.Color) myui.Color {
	if contrast(c, bg) >= contrast(c, fg) {
		return bg
	}
	return fg
}

// contrast is how far apart two colors are, as WCAG writes it: 1 for the same
// color, 21 for black and white.
func contrast(a, b myui.Color) float32 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// luminance is how bright c is, from 0 for black to 1 for white.
func luminance(c myui.Color) float32 {
	return 0.2126*linear(c.R) + 0.7152*linear(c.G) + 0.0722*linear(c.B)
}

// linear is one channel of a color as light adds up, rather than as it is
// stored.
func linear(v uint8) float32 {
	f := float32(v) / 255
	if f <= 0.04045 {
		return f / 12.92
	}
	return float32(math.Pow(float64((f+0.055)/1.055), 2.4))
}
