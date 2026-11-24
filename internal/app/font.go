package app

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/fault"
	"wtgo/internal/store"
	"wtgo/internal/ui"
)

// Boundary: the font a terminal draws with -- the family and size a profile
// names, the names that wait for the picker to open, and the terminal.Font the
// window hands mygo. Measuring which families are monospaced is
// font_monospace.go.

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
		// Keep the rest for glyphs the primary font does not cover, each once: a
		// hand-written list may name the same family twice, and a terminal told
		// of it twice would look for glyphs in it twice.
		if slices.ContainsFunc(candidates, func(c string) bool { return strings.EqualFold(c, name) }) {
			continue
		}
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

// ensureFontFamilies defers measuring the entire picker until its first use:
// what the machine advertises is listed at once, and the census narrows it.
func (a *App) ensureFontFamilies() {
	if a.fonts.asked {
		return
	}
	a.loadFontNames()
	a.fonts.asked = true
	a.fonts.families = advertisedMono(a.fonts.available)
	available := a.fonts.available
	go func() {
		families, mono := census(available)
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
