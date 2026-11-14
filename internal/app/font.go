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
	"wtgo/internal/ui"
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

// weightOf is the weight a terminal draws at: mygo reads a zero as the regular
// face and clamps the rest to the weights a family can have, and the settings
// page offers the same nine.
func weightOf(weight int) int {
	if weight <= 0 {
		return 400
	}
	return min(max(weight, 100), 900)
}

// fontOf is the family and size the terminal would draw with, which is what a
// measure is about: mygo draws an empty family with the generic one, and a
// size of zero at 13.
func fontOf(family string, size float32) (string, float32) {
	if strings.TrimSpace(family) == "" {
		family = fallbackFamily
	}
	if size <= 0 {
		size = 13
	}
	return family, size
}

// fontCensus is what the window learned about the machine's fonts: the
// families the font setting offers, and what monospace said of a family at a
// size. Listing the families and shaping the printable ASCII are the two
// slowest measures the window makes, and both answers hold for the session:
// the families a machine has do not change under a running window.
type fontCensus struct {
	// families are the installed monospace families, and asked whether the
	// window has asked for them: the list arrives a frame or two after the page
	// that shows it opens.
	families []string
	// installed are every family the machine has and whether a terminal can draw
	// a grid with it: the font setting says which of the names it holds are
	// missing here, and which of the ones it has are not monospace.
	installed []ui.InstalledFont
	asked     bool
	// mono is what monospace said of a family at a size.
	mono map[monoKey]bool
}

// monoKey is a family measured at a size, which is what monospace answers for.
type monoKey struct {
	family string
	size   float32
}

// monospace reports whether family shapes the printable ASCII at one advance
// at size, which is what a grid of cells needs. The measure is at the regular
// face, since that is where a family is a family: the weights of a monospaced
// one keep the advance of the regular face, and a weight the machine lacks is
// drawn with the nearest one it has. A family that is not installed
// answers for the font the terminal would draw with instead: the system picks a
// proportional one, and every cell then takes the width of its widest glyph,
// which spreads the letters out and makes the columns drift.
func monospace(family string, size float32) bool {
	family, size = fontOf(family, size)
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

// isMonospace is monospace, measured once per family and size: the window asks
// the same question for every tab it opens and every setting it applies, and a
// shape of the ASCII is the longest measure it makes.
func (a *App) isMonospace(family string, size float32) bool {
	family, size = fontOf(family, size)
	key := monoKey{family: family, size: size}
	if ok, known := a.fonts.mono[key]; known {
		return ok
	}
	ok := monospace(family, size)
	if a.fonts.mono == nil {
		a.fonts.mono = make(map[monoKey]bool)
	}
	a.fonts.mono[key] = ok
	return ok
}

// monospaceFamilies are the installed families a terminal can draw with: those
// the text engine calls monospaced and that shape the printable ASCII at one
// advance, sorted. The font setting lists them, so that a family it offers is
// one the terminal takes as it is, without a line about it in the log.
func (a *App) monospaceFamilies(fs []myui.FontFamily, size float32) []string {
	var out []string
	for _, f := range fs {
		if f.Monospaced && a.isMonospace(f.Name, size) {
			out = append(out, f.Name)
		}
	}
	slices.Sort(out)
	return out
}

// installedFonts are every family the machine has, and whether a terminal can
// draw a grid of cells with it: the setting says which of the names in it this
// machine is missing, and a name it has but cannot draw a grid with is one the
// terminal passes over, which is worth a line under the field.
func (a *App) installedFonts(fs []myui.FontFamily, size float32) []ui.InstalledFont {
	out := make([]ui.InstalledFont, 0, len(fs))
	for _, f := range fs {
		out = append(out, ui.InstalledFont{Name: f.Name, Mono: f.Monospaced && a.isMonospace(f.Name, size)})
	}
	slices.SortFunc(out, func(x, y ui.InstalledFont) int { return strings.Compare(x.Name, y.Name) })
	return out
}

// ensureFontFamilies lists the families the font setting offers, which only the
// page that lists them needs: enumerating the machine's fonts and shaping the
// ASCII of each monospaced one costs the first frame 23–26 ms for a menu the
// user may not open. The first ask starts the work, later ones keep it; the
// measures are the window's, so they run on the UI thread once the list is in.
func (a *App) ensureFontFamilies(size float32) {
	if a.fonts.asked {
		return
	}
	a.fonts.asked = true
	go func() {
		fs := myui.FontFamilies()
		a.Do(func() {
			a.fonts.families = a.monospaceFamilies(fs, size)
			a.fonts.installed = a.installedFonts(fs, size)
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
	f := terminal.Font{Family: s.Family, Size: s.Size, LineHeight: s.LineHeight, Weight: weightOf(s.Weight)}
	if a.isMonospace(f.Family, f.Size) {
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
