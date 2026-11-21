// The font setting's list of families: what the names in it are worth on this
// machine, and the one line under the field that says what the terminal will
// draw with.
package ui

import (
	"strings"
	"sync"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
	"wtgo/locales"
)

// genericFamilies are the names a text engine resolves to a font of its own,
// which the setting may hold in place of a family this machine has: a list
// naming one is not a list naming a font that is missing.
var genericFamilies = []string{"monospace", "serif", "sans-serif", "system-ui"}

// fontKind is what this machine makes of one name of the list.
type fontKind int

const (
	// fontMissing is a family this machine does not have, which the terminal
	// skips.
	fontMissing fontKind = iota
	// fontWide is a family the machine has, drawn with letters of their own
	// width: a terminal needs one cell per character.
	fontWide
	// fontMono is a family the terminal draws a grid of cells with.
	fontMono
	// fontGeneric is one of the generic names, which every system resolves.
	fontGeneric
)

// fontEntry is one name of the list and what this machine makes of it.
type fontEntry struct {
	name string
	kind fontKind
}

// fontEntries are the names of a family list and what this machine makes of
// each.
func fontEntries(list string, fonts []InstalledFont) []fontEntry {
	known, mono := fontMaps(fonts)
	names := store.FontNames(list)
	out := make([]fontEntry, 0, len(names))
	for _, name := range names {
		switch {
		case mono[name]:
			out = append(out, fontEntry{name, fontMono})
		case known[name]:
			out = append(out, fontEntry{name, fontWide})
		case slicesContain(genericFamilies, name):
			out = append(out, fontEntry{name, fontGeneric})
		default:
			out = append(out, fontEntry{name, fontMissing})
		}
	}
	return out
}

// fontMaps are the machine's families by name: the ones it has, and the ones it
// draws a grid with. The appearance page notes a family on every frame it
// draws, and the machine's fonts do not change while the window runs, so the
// two maps are made once for the list the window holds.
func fontMaps(fonts []InstalledFont) (known, mono map[string]bool) {
	fontKnown.Lock()
	defer fontKnown.Unlock()
	if fontKnown.known != nil && sameFonts(fontKnown.from, fonts) {
		return fontKnown.known, fontKnown.mono
	}
	known = make(map[string]bool, len(fonts))
	mono = make(map[string]bool, len(fonts))
	for _, f := range fonts {
		known[f.Name] = true
		if f.Mono {
			mono[f.Name] = true
		}
	}
	fontKnown.from, fontKnown.known, fontKnown.mono = fonts, known, mono
	return known, mono
}

// fontKnown is what the maps of the last list were made from, for the frames
// that ask with the same one.
var fontKnown struct {
	sync.Mutex
	from  []InstalledFont
	known map[string]bool
	mono  map[string]bool
}

// sameFonts reports whether two lists are the machine's same families: the
// window builds the list once and hands the page that one, so the slices are
// the same, and a page given another list is answered anew.
func sameFonts(a, b []InstalledFont) bool {
	if len(a) != len(b) {
		return false
	}
	return len(a) == 0 || &a[0] == &b[0]
}

// slicesContain reports whether one of the names is the one, compared without
// case: a family name is a name, and "monospace" is the same word in any case.
func slicesContain(names []string, name string) bool {
	for _, n := range names {
		if strings.EqualFold(n, name) {
			return true
		}
	}
	return false
}

// fontNote is what a list of families will draw with, said under the field. The
// terminal draws with the first name the machine has, so a first name that is
// not monospace is the one case worth a warning: the columns drift then, and
// the window draws with the generic family instead. Names the machine does not
// have are skipped without a word from the terminal, and said quietly here,
// since a list written on another machine is bound to hold a few -- except the
// ones the default list names too, which are guesses rather than choices.
func fontNote(list string, fonts []InstalledFont) (text string, danger bool) {
	entries := fontEntries(list, fonts)
	if len(entries) == 0 {
		// The field says what an empty list is worth in its placeholder, which
		// the setting fills in as soon as it is submitted.
		return "", false
	}
	guessed := store.FontNames(store.DefaultFontFamily())
	var missing []string
	var first *fontEntry
	for i := range entries {
		if entries[i].kind == fontMissing {
			// A name the default list makes too is a guess about what a machine
			// will have, and a machine that has none of the guesses is not
			// missing anything a person asked for: the line is for the names a
			// list of one's own holds.
			if !slicesContain(guessed, entries[i].name) {
				missing = append(missing, entries[i].name)
			}
			continue
		}
		if first == nil {
			first = &entries[i]
		}
	}
	switch {
	case first == nil:
		return locales.FontNoneUsable(), true
	case first.kind == fontWide:
		return locales.FontWideFirst(first.name), true
	case len(missing) > 0:
		return locales.FontMissing(strings.Join(missing, ", ")), false
	}
	return "", false
}

// fontChoices are what the list's menu offers: the families this machine has a
// terminal can draw with, then the generic name, which is not a family of this
// machine but a word every system resolves.
func fontChoices(families []string) []option {
	out := make([]option, 0, len(families)+1)
	for _, f := range families {
		out = append(out, option{f, f})
	}
	return append(out, option{store.EmptyFontFamily, locales.FontGeneric()})
}

// pickFamily is the menu's choice: a family the list already holds moves to the
// front of it, since the family a terminal draws with is the first one the
// machine has, and one the list does not hold is added there too.
func pickFamily(list *string, name string) {
	rest := store.FontNames(*list)
	kept := make([]string, 0, len(rest)+1)
	kept = append(kept, name)
	for _, f := range rest {
		if !strings.EqualFold(f, name) {
			kept = append(kept, f)
		}
	}
	*list = strings.Join(kept, ", ")
}

// fillFamily names the default families where a submitted field holds none,
// which is what the terminal draws with anyway: the field says it, so that the
// file does not have to be read by a person to know.
func fillFamily(list *string) {
	if strings.TrimSpace(*list) == "" {
		*list = store.DefaultFontFamily()
	}
}

// fontPicker is the list's field and the menu beside it: the field takes any
// names at all, since the file takes them, and the menu is the shortcut for the
// ones this machine has.
func fontPicker(c *myui.Context, list *string, families []string) {
	myui.Row(c).Gap(8).AlignItems(myui.Center).Children(func() {
		myui.TextInput(c, list).
			Width(controlWidth - 40).
			Placeholder(store.DefaultFontFamily()).
			OnSubmit(func() { fillFamily(list) })
		myui.MenuButton(c, "\uff0b", func(m *myui.Menu) {
			for _, o := range fontChoices(families) {
				it := m.Item(o.Label)
				if o.Value != store.EmptyFontFamily {
					it.Checked(slicesContain(store.FontNames(*list), o.Value))
				}
				if it.Chosen() {
					pickFamily(list, o.Value)
				}
			}
		}).Tooltip(locales.FontAdd())
	})
}
