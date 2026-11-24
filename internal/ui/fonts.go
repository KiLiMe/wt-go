// The font setting's list of families: what the names in it are worth on this
// machine, and the one line under the field that says what the terminal will
// draw with.
package ui

import (
	"strings"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
	"wtgo/locales"
)

// LineHeightMin and LineHeightMax are the line heights the field offers and the
// preview draws with: a line of a terminal is drawn at a multiple of the font
// size rather than at a size of its own, so its bounds are its own.
const (
	LineHeightMin float32 = 0.8
	LineHeightMax float32 = 2.4
)

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

// fontNote presents the terminal's resolved font for the appearance draft.
func fontNote(list string, font FontDiagnostic) (text string, danger bool) {
	if strings.TrimSpace(list) == "" || font.Primary == "" {
		return "", false
	}
	switch {
	case font.Proportional != "":
		return locales.FontWideFirst(font.Proportional, font.Primary), true
	case font.SystemFallback:
		return locales.FontNoneUsable(), true
	case len(font.Missing) > 0:
		return locales.FontMissing(strings.Join(font.Missing, ", "), font.Primary), false
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
