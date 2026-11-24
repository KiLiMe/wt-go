package app

import (
	"math"
	"slices"

	myui "github.com/egoist/mygo/ui"
)

// Boundary: measuring a family, once per size, to say whether it shapes the
// printable ASCII on a grid -- the census the font picker draws from. Nothing
// here picks the font: font.go does.

// asciiPrintable is what a terminal measures a cell with: the glyphs a shell
// writes most, shaped the way mygo's own terminal shapes them.
const asciiPrintable = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

// advanceSpread is how far the advances of a family may differ, in DIPs, and
// still be one cell: a monospace font shapes them all the same, and hinting
// moves them by a fraction of a pixel at most.
const advanceSpread = 0.5

// fontCensus keeps the installed names and the measurements shared by the
// terminal and the appearance page.
type fontCensus struct {
	installed map[string]bool
	available []myui.FontFamily
	families  []string
	asked     bool
	mono      map[monoKey]bool
	lastKey   monoKey
	last      fontSelection
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
// drawn with the nearest one it has. What a name the machine does not have
// answers is what fontconfig aliases it to, which may be a monospace font or a
// proportional one, so a missing family is not a family that measures false:
// the census is what says which names a machine has, and this says of a name it
// draws with whether the cells of a grid hold.
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

// advertisedMono is what the text engine calls monospaced, which is the list
// the picker can show before anything is measured: listing a machine's families
// is one call, and measuring each of them is a shape of the ASCII apiece, so a
// menu waiting for the census would hold the word "monospace" alone until it
// finished. The census is what narrows this list to the families that measure
// as what they advertise.
func advertisedMono(fs []myui.FontFamily) []string {
	families := make([]string, 0, len(fs))
	for _, f := range fs {
		if f.Monospaced {
			families = append(families, f.Name)
		}
	}
	slices.Sort(families)
	return families
}

// census verifies only the fonts advertised as monospaced for the picker.
// It runs off the UI thread and returns measurements for the window's cache.
//
// What it answers is a property of the family and not of a size: the advance of
// a monospace face is one cell at every size a terminal draws, and the sizes a
// terminal does draw at are measured where they are asked for. A census that
// took the settings' size would re-measure the whole machine each time a person
// tried another one, and would answer under a key nothing else asks for.
func census(fs []myui.FontFamily) (families []string, mono map[monoKey]bool) {
	names := advertisedMono(fs)
	mono = make(map[monoKey]bool, len(names))
	for _, name := range names {
		family, size := fontOf(name, 0)
		key := monoKey{family: family, size: size}
		ok := monospace(family, size)
		mono[key] = ok
		if ok {
			families = append(families, name)
		}
	}
	slices.Sort(families)
	return families, mono
}
