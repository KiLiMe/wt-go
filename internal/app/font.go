// The font a terminal draws with: the settings' one, and what the window does
// with a family that is not a monospace font.
package app

import (
	"fmt"
	"log"
	"math"
	"slices"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/fault"
)

// asciiPrintable is what a terminal measures a cell with: the glyphs a shell
// writes most, shaped the way mygo's own terminal shapes them.
const asciiPrintable = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

// fallbackFamily is what a terminal draws with when the settings name a family
// that is not monospace: the generic one every system resolves to a monospace
// font.
const fallbackFamily = "monospace"

// advanceSpread is how far the advances of a family may differ, in DIPs, and
// still be one cell: a monospace font shapes them all the same, and hinting
// moves them by a fraction of a pixel at most.
const advanceSpread = 0.5

// monospace reports whether family shapes the printable ASCII at one advance
// at size, which is what a grid of cells needs. A family that is not installed
// answers for the font the terminal would draw with instead: the system picks a
// proportional one, and every cell then takes the width of its widest glyph,
// which spreads the letters out and makes the columns drift.
func monospace(family string, size float32) bool {
	if strings.TrimSpace(family) == "" {
		// mygo draws an empty family with the generic one.
		family = fallbackFamily
	}
	if size <= 0 {
		// And a size of zero at 13.
		size = 13
	}
	glyphs := myui.Shape(asciiPrintable, myui.Font{Family: family, Size: size})
	if len(glyphs) == 0 {
		return false
	}
	lo, hi := float32(math.MaxFloat32), float32(0)
	for _, g := range glyphs {
		// A ligature is one glyph holding several runes, of as many cells: the
		// width of a cell is what it advances divided by what it holds.
		cell := g.Advance / float32(max(g.Runes, 1))
		lo, hi = min(lo, cell), max(hi, cell)
	}
	return hi-lo < advanceSpread
}

// monospaceFamilies are the installed families a terminal can draw with: those
// the text engine calls monospaced and that shape the printable ASCII at one
// advance, sorted. The font setting lists them, so that a family it offers is
// one the terminal takes as it is, without a line about it in the log.
func monospaceFamilies(fs []myui.FontFamily, size float32) []string {
	var out []string
	for _, f := range fs {
		if f.Monospaced && monospace(f.Name, size) {
			out = append(out, f.Name)
		}
	}
	slices.Sort(out)
	return out
}

// loadFontFamilies lists the families the font setting offers, off the UI
// thread: the text engine takes long enough to list them for a frame to show it.
// The measures are the window's, though, so they wait for it.
func (a *App) loadFontFamilies(size float32) {
	go func() {
		fs := myui.FontFamilies()
		a.Do(func() {
			a.families = monospaceFamilies(fs, size)
			a.redraw()
		})
	}()
}

// terminalFont is the font the terminals draw with: the one the settings name,
// except that a family which is not monospace gives way to the fallback, which
// the log says once per family so that a settings file naming a font the
// machine does not have leaves a line behind.
func (a *App) terminalFont() terminal.Font {
	s := a.settings.Font
	f := terminal.Font{Family: s.Family, Size: s.Size, LineHeight: s.LineHeight}
	if monospace(f.Family, f.Size) {
		a.fontWarned = ""
		return f
	}
	if a.fontWarned != s.Family {
		a.fontWarned = s.Family
		log.Printf("wt-go: %s", fault.Line(fault.BadFont, fmt.Sprintf("%q is not monospace, drawing with %q", s.Family, fallbackFamily)))
	}
	f.Family = fallbackFamily
	return f
}
