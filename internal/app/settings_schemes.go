package app

import (
	myui "github.com/egoist/mygo/ui"
	"strconv"
	"strings"
	"sync"
	"wtgo/internal/store"
	"wtgo/internal/theme"
	"wtgo/internal/ui"
	"wtgo/locales"
)

// schemeList is the color schemes the settings page lists, built once per
// settings the window applied and kept until they change.
func (a *App) schemeList() []ui.Scheme {
	if a.schemes == nil {
		a.schemes = a.uiSchemes(a.settings)
	}
	return a.schemes
}

// uiSchemes are the color schemes the color schemes page lists, each saying
// whether the settings write it -- the page offers to remove those alone -- and
// the colors of a terminal it paints, which the page shows as a chip.
func (a *App) uiSchemes(s store.Settings) []ui.Scheme {
	custom := make(map[string]bool, len(s.Schemes))
	for _, sc := range s.Schemes {
		custom[strings.ToLower(sc.Name)] = true
	}
	names := theme.Schemes(s)
	out := make([]ui.Scheme, 0, len(names))
	for _, name := range names {
		own := custom[strings.ToLower(name)]
		sc := ui.Scheme{Name: name, Custom: own}
		if kept, ok := a.swatches.get(name, own); ok {
			sc = kept
		} else {
			sc.Bg, sc.Fg = schemeColors(s, name)
			a.swatches.put(sc)
		}
		out = append(out, sc)
	}
	return out
}

// schemeColors is the background and the text a terminal paints with the scheme
// called name, and no colors at all for a name that is no scheme.
func schemeColors(s store.Settings, name string) (myui.Color, myui.Color) {
	if th, ok := theme.Scheme(s, name); ok {
		return th.Background, th.Foreground
	}
	return myui.Color{}, myui.Color{}
}

// schemeSwatches are the colors of the schemes the settings do not write, kept
// from one build of the page's list to the next. A scheme this program carries
// and one of Ghostty's are the same while it runs, and looking one up reads the
// user's theme directory and parses its colors: hundreds of them are a good
// part of a frame, and the settings page asks for them again every time the
// settings are applied. A scheme the settings write is looked up every time,
// since its colors are the settings'.
type schemeSwatches struct {
	sync.Mutex
	byName map[string]ui.Scheme
}

// get is the colors kept for name, unless the settings write the scheme: those
// are never kept, being the settings' to change.
func (c *schemeSwatches) get(name string, own bool) (ui.Scheme, bool) {
	if own {
		return ui.Scheme{}, false
	}
	c.Lock()
	defer c.Unlock()
	sc, ok := c.byName[name]
	return sc, ok
}

// put keeps sc, and leaves a scheme the settings write alone.
func (c *schemeSwatches) put(sc ui.Scheme) {
	if sc.Custom {
		return
	}
	c.Lock()
	defer c.Unlock()
	if c.byName == nil {
		c.byName = make(map[string]ui.Scheme)
	}
	c.byName[sc.Name] = sc
}

// PickScheme paints with a color scheme, as a click on the color schemes page
// asks: the choice of the page edits the draft in place, so the window applies
// the scheme as it does any other change.
func (a *App) PickScheme(name string) {
	a.settings.ColorScheme = name
	a.redraw()
}

// CopyScheme makes a color scheme of the settings from the one the terminal
// paints with, and picks it.
func (a *App) CopyScheme() {
	th, ok := theme.Scheme(a.settings, a.settings.ColorScheme)
	if !ok {
		return
	}
	scheme := theme.Stored(th)
	scheme.Name = a.customSchemeName(schemeBaseName(a.settings.ColorScheme))
	a.settings.Schemes = append(a.settings.Schemes, scheme)
	a.settings.ColorScheme = scheme.Name
	a.note = locales.SchemeCopied()
	a.redraw()
}

// customSuffix starts the mark a copied scheme carries on its name.
const customSuffix = " (custom"

// schemeBaseName is the name a copy starts from: the name the copies before it
// added their marks to, without them. A copy copies the scheme in front, which
// may be a copy itself, so the marks nest unless they come off again.
func schemeBaseName(name string) string {
	for {
		i := strings.LastIndex(name, customSuffix)
		if i <= 0 {
			return name
		}
		// What follows the mark is ")" or a number and ")", as
		// customSchemeName writes them.
		n := strings.TrimSpace(strings.TrimSuffix(name[i+len(customSuffix):], ")"))
		if n != "" && !digits(n) {
			return name
		}
		name = name[:i]
	}
}

// digits reports whether s is a run of decimal digits.
func digits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// customSchemeName is a name no scheme holds yet for a copy of the scheme
// named base: its own name with a mark, numbered once that is taken.
func (a *App) customSchemeName(base string) string {
	name := base + customSuffix + ")"

	for i := 2; a.schemeNamed(name); i++ {
		name = base + customSuffix + " " + strconv.Itoa(i) + ")"
	}
	return name
}

// RemoveScheme drops a color scheme the settings write, which is the only kind
// the color schemes page offers to remove: the terminal keeps painting with the
// scheme the one removed was a copy of, while that is still there.
func (a *App) RemoveScheme(name string) {
	out := a.settings.Schemes[:0]
	removed := false
	for _, s := range a.settings.Schemes {
		if strings.EqualFold(s.Name, name) {
			removed = true
			continue
		}
		out = append(out, s)
	}
	if !removed {
		return
	}
	a.settings.Schemes = out
	if strings.EqualFold(a.settings.ColorScheme, name) {
		a.settings.ColorScheme = a.schemeAfterRemoving(name)
	}
	a.note = locales.SchemeRemoved()
	a.redraw()
}

// schemeAfterRemoving is the scheme to paint with once one is gone: the one it
// was a copy of, while that is still there, else the one the defaults name.
func (a *App) schemeAfterRemoving(name string) string {
	for _, candidate := range []string{schemeBaseName(name), store.Default().ColorScheme} {
		if _, ok := theme.Scheme(a.settings, candidate); ok {
			return candidate
		}
	}
	return ""
}

// schemeNamed reports whether the settings hold a color scheme of that name.
func (a *App) schemeNamed(name string) bool {
	for _, s := range a.settings.Schemes {
		if strings.EqualFold(s.Name, name) {
			return true
		}
	}
	return false
}
