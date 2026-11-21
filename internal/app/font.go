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
	"wtgo/internal/store"
	"wtgo/internal/ui"
)

// asciiPrintable is what a terminal measures a cell with: the glyphs a shell
// writes most, shaped the way mygo's own terminal shapes them.
const asciiPrintable = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

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
// measure is about: a family the settings leave empty is the default families,
// which is what a terminal draws with then, and a size of zero is 13.
func fontOf(family string, size float32) (string, float32) {
	if strings.TrimSpace(family) == "" {
		family = store.DefaultFontFamily()
	}
	if size <= 0 {
		size = 13
	}
	return family, size
}

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

// census verifies only the fonts advertised as monospaced for the picker.
// It runs off the UI thread and returns measurements for the window's cache.
func census(fs []myui.FontFamily, size float32) (families []string, mono map[monoKey]bool) {
	mono = make(map[monoKey]bool)
	for _, f := range fs {
		if !f.Monospaced {
			continue
		}
		family, at := fontOf(f.Name, size)
		key := monoKey{family: family, size: at}
		ok := monospace(family, at)
		mono[key] = ok
		if ok {
			families = append(families, f.Name)
		}
	}
	slices.Sort(families)
	return families, mono
}

func (a *App) loadFontNames() {
	if a.fonts.installed != nil {
		return
	}
	a.fonts.available = myui.FontFamilies()
	a.fonts.installed = make(map[string]bool, len(a.fonts.available))
	for _, f := range a.fonts.available {
		a.fonts.installed[strings.ToLower(f.Name)] = true
	}
}

// ensureFontFamilies defers measuring the entire picker until its first use.
func (a *App) ensureFontFamilies(size float32) {
	if a.fonts.asked {
		return
	}
	a.loadFontNames()
	a.fonts.asked = true
	available := a.fonts.available
	go func() {
		families, mono := census(available, size)
		a.Do(func() {
			a.fonts.families = families
			if a.fonts.mono == nil {
				a.fonts.mono = make(map[monoKey]bool, len(mono))
			}
			for key, ok := range mono {
				a.fonts.mono[key] = ok
			}
			a.redraw()
		})
	}()
}

// genericFont matches the generic family names accepted by MyGo's text engine.
func genericFont(name string) bool {
	switch name {
	case "monospace", "ui-monospace", "serif", "ui-serif", "sans-serif", "system-ui", "ui-sans-serif", "-apple-system", "blinkmacsystemfont":
		return true
	}
	return false
}

type fontSelection struct {
	family string
	note   ui.FontDiagnostic
}

func (a *App) selectFont(family string, size float32) fontSelection {
	family, size = fontOf(family, size)
	key := monoKey{family: family, size: size}
	if a.fonts.lastKey == key {
		return a.fonts.last
	}
	a.loadFontNames()
	var selected fontSelection
	var candidates []string
	defaults := store.FontNames(store.DefaultFontFamily())
	for _, name := range store.FontNames(family) {
		lower := strings.ToLower(name)
		if !a.fonts.installed[lower] && !genericFont(lower) {
			if !slices.ContainsFunc(defaults, func(guess string) bool { return strings.EqualFold(guess, name) }) {
				selected.note.Missing = append(selected.note.Missing, name)
			}
			continue
		}
		if len(candidates) == 0 && !a.isMonospace(name, size) {
			if selected.note.Proportional == "" {
				selected.note.Proportional = name
			}
			continue
		}
		// Keep the rest for glyphs the primary font does not cover.
		candidates = append(candidates, name)
	}
	if len(candidates) == 0 {
		candidates = []string{store.EmptyFontFamily}
		selected.note.SystemFallback = true
	}
	selected.family = strings.Join(candidates, ", ")
	selected.note.Primary = candidates[0]
	a.fonts.lastKey, a.fonts.last = key, selected
	return selected
}

func (a *App) terminalFont() terminal.Font {
	s := a.settings.Font
	selected := a.selectFont(s.Family, s.Size)
	switch {
	case selected.note.Proportional == "" && !selected.note.SystemFallback:
		a.fontWarned = ""
	case a.fontWarned != s.Family:
		a.fontWarned = s.Family
		reason := fmt.Sprintf("no usable monospace family in %q", s.Family)
		if selected.note.Proportional != "" {
			reason = fmt.Sprintf("skipped non-monospace family %q in %q", selected.note.Proportional, s.Family)
		}
		log.Printf("wt-go: %s %s, using %q as the primary font", fault.BadFont, reason, selected.note.Primary)
	}
	return terminal.Font{Family: selected.family, Size: s.Size, LineHeight: s.LineHeight, Weight: weightOf(s.Weight)}
}
