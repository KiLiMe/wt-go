package ui

import (
	"strings"
	"sync"

	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/icons"
	"wtgo/internal/store"
	"wtgo/internal/theme"
	"wtgo/locales"
)

// schemesPage lists the color schemes, and shows the one in front.
func schemesPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	// A page drawn without a window to keep its cache for it keeps none: it
	// derives the options and the rows again for each frame, as it did before
	// there was a cache, rather than share one with the frame before it.
	var cache *SchemesCache
	if st.SchemesCache != nil {
		cache = st.SchemesCache
	}
	schemesPageFrom(c, s, st, act, cache)
}

// schemesPageFrom lists the schemes a page is handed. The cache is the one the
// window keeps for the page, nil for a page drawn without one.
func schemesPageFrom(c *myui.Context, s *store.Settings, st Settings, act Actions, cache *SchemesCache) {
	t := c.Theme()
	card(c, cardFace(t), cardLine(t), func() {
		field(c, locales.Scheme(), locales.ColorSchemeNote(), func() {
			choice(c, &s.ColorScheme, schemeOptionsFrom(cache, st.Schemes))
		})
		row(c, func() {
			myui.Row(c).Children(func() {
				myui.Button(c, locales.SchemeCopy()).OnClick(act.CopyScheme)
			})
		})
	})
	names := searchSchemes(cache, st.Schemes, filter(st))
	// Not a card: the preview is as wide as its swatches, and a card's padding
	// is what it would have to take that width from.
	myui.Row(c).AlignItems(myui.Start).Gap(24).Children(func() {
		myui.Column(c).Width(320).Gap(6).Children(func() {
			// A text input edits the string it is handed, so that string has to
			// outlive the frame. Handing it a fresh one every frame would throw
			// away what the user typed on the next frame: with no field to hold
			// the words (nil Filter, e.g. a test that draws the page alone),
			// the search box is left out rather than given nowhere to write.
			if st.Filter != nil {
				myui.TextInput(c, st.Filter).Placeholder(locales.SchemeSearch())
			}
			if len(names) == 0 {
				myui.Text(c, locales.SchemeNone()).FontSize(11).TextColor(t.TextMuted)
			}
			myui.Column(c).Gap(2).Children(func() {
				for _, sc := range names {
					schemeRow(c, sc, s.ColorScheme, act)
				}
			})
		})
		if th, ok := theme.Scheme(*s, s.ColorScheme); ok {
			myui.Column(c).Grow(1).Gap(8).Children(func() {
				myui.Text(c, locales.SchemePreview()).FontSize(11).TextColor(t.TextMuted)
				preview(c, th)
			})
		}
	})
}

// filter is what the schemes are searched for.
func filter(st Settings) string {
	if st.Filter == nil {
		return ""
	}
	return *st.Filter
}

// schemeOptions are the schemes a control picks from.
func schemeOptions(schemes []Scheme) []option {
	out := make([]option, 0, len(schemes))
	for _, sc := range schemes {
		out = append(out, option{sc.Name, sc.Name})
	}
	return out
}

// schemeRow is one scheme in the list: its name, chosen by a click, and for a
// scheme the settings write the button that removes it -- shown while the row
// is in front, in room the row keeps either way, so that the names do not jump
// as the pointer moves.
func schemeRow(c *myui.Context, sc Scheme, current string, act Actions) {
	el := myui.Row(c).Height(28).Padding(0, 4, 0, 8).AlignItems(myui.Center).Gap(4).Radius(6).
		Cursor(myui.CursorPointer).OnClick(func() { act.PickScheme(sc.Name) })
	if current == sc.Name {
		el.Background(c.Theme().SurfaceHover)
	}
	shown := sc.Custom && (el.Hovered() || current == sc.Name)
	el.Children(func() {
		myui.Text(c, sc.Name).Grow(1).Shrink(1).SingleLine()
		if !sc.Custom {
			return
		}
		remove := myui.Box(c).Size(20, 20).Radius(4).AlignItems(myui.Center).Justify(myui.Center)
		if !shown {
			// The box keeps its room, so that the names do not move as the
			// pointer goes from row to row, and nothing of the button is there
			// to click: a row that hides its button still picks the scheme.
			remove.Opacity(0).PassThrough()
		} else {
			remove.Cursor(myui.CursorPointer).Label(locales.SchemeRemove()).
				Tooltip(locales.SchemeRemove()).
				OnClick(func() { act.RemoveScheme(sc.Name) })
		}
		remove.Children(func() {
			myui.Icon(c, icons.Close.SVG).Size(12, 12).TextColor(c.Theme().TextMuted)
		})
	})
}

// search is the schemes that hold what the search asks for, at most the rows a
// page shows at once: the schemes number in the hundreds.
func search(schemes []Scheme, words string) []Scheme {
	words = strings.ToLower(strings.TrimSpace(words))
	out := make([]Scheme, 0, min(len(schemes), schemeRows))
	for _, sc := range schemes {
		if words != "" && !strings.Contains(strings.ToLower(sc.Name), words) {
			continue
		}
		out = append(out, sc)
		if len(out) == schemeRows {
			break
		}
	}
	return out
}

// schemeRows is how many rows the color schemes page shows at once.
const schemeRows = 60

// SchemesCache is what the options and the rows of the color schemes page were
// found from: the list of schemes, kept by pointer, and the list the options
// were never built from. The page is drawn again on frames it did not build
// anew, as the pointer moves over it and the field is typed in, and the list is
// the same until the settings change it -- a list of hundreds is copied and
// searched in full every frame otherwise.
//
// The list is the one the window builds, and the window builds a new one as the
// settings change, so the pointer is what tells the list apart. The page keeps
// its cache from frame to frame, and it is drawn on the goroutine that draws
// its window: the lock is for the two windows a process may draw.
type SchemesCache struct {
	sync.Mutex
	// _ keeps `==` from comparing two caches equal, which would say the
	// settings page holds the one and not the other.
	_ [0]func()
	// All is the list the options and the rows were found from, nil before
	// either was, and Options the options it was found as.
	All     []Scheme
	Key     string
	Rows    []Scheme
	Options []option
}

// search is search, with the answer of the last call kept while the list of
// schemes and the words searched for are the same. A page drawn without a
// window to keep a cache keeps the answer for the frame alone.
func searchSchemes(cache *SchemesCache, all []Scheme, words string) []Scheme {
	if cache == nil {
		return search(all, words)
	}
	cache.Lock()
	defer cache.Unlock()
	if cache.Rows != nil && cache.Key == words && sameSchemes(cache.All, all) {
		return cache.Rows
	}
	// search finds at least an empty list, and a nil one would stand for a
	// search never made.
	cache.Rows, cache.Key = search(all, words), words
	cache.All = all
	return cache.Rows
}

// schemeOptionsFrom are the schemes a control picks from, built once for the
// list, since the page draws the control on every frame it is in front.
func schemeOptionsFrom(schemes *SchemesCache, all []Scheme) []option {
	if schemes == nil {
		return schemeOptions(all)
	}
	schemes.Lock()
	defer schemes.Unlock()
	if schemes.Options != nil && sameSchemes(schemes.All, all) {
		return schemes.Options
	}
	schemes.Options, schemes.All = schemeOptions(all), all
	return schemes.Options
}

// reset empties the cache.
func (s *SchemesCache) reset() {
	s.Lock()
	defer s.Unlock()
	s.All, s.Key, s.Rows, s.Options = nil, "", nil, nil
}

// sameSchemes reports whether two lists name the same schemes, in the same
// order, each written or found as the other says. The page is handed the list
// the window built once for the settings, so its length and its names are what
// tell a rebuilt list apart from the one the cache kept.
func sameSchemes(a, b []Scheme) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// preview shows the colors of a scheme: its background, foreground, cursor and
// selection, and the sixteen colors programs ask for.
func preview(c *myui.Context, th *terminal.Theme) {
	t := c.Theme()
	myui.Column(c).Gap(8).Children(func() {
		myui.Box(c).Padding(8).Radius(6).Background(th.Background).Children(func() {
			myui.Text(c, "The quick brown fox").TextColor(th.Foreground).SingleLine()
			myui.Text(c, "jumps over the lazy dog").TextColor(th.Foreground).SingleLine()
		})
		myui.Row(c).Gap(6).Wrap().Children(func() {
			for _, col := range []myui.Color{th.Background, th.Foreground, th.Cursor, th.Selection} {
				myui.Box(c).Size(26, 26).Radius(4).Background(col).BorderWidth(1).BorderColor(t.Border)
			}
		})
		myui.Row(c).Gap(4).Wrap().Children(func() {
			for _, col := range th.Palette {
				myui.Box(c).Size(22, 22).Radius(3).Background(col)
			}
		})
	})
}
