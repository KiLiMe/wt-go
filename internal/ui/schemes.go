package ui

import (
	"strings"
	"sync"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/icons"
	"wtgo/internal/store"
	"wtgo/locales"
)

// Boundary: the color schemes of the appearance page -- the picker, its rows,
// its search and the cache the search answers from. The terminal that shows the
// draft is schemes_preview.go.

// schemePicker is the color schemes of the appearance page: the scheme in
// force with the button that copies it as one of the settings' own, and the
// search box and the list to pick from or remove. The list is the only place a
// scheme is picked: a control above it that picked too would offer the same
// hundreds of names a second time, without the search that makes them
// findable, and take the room of a row to say what this one says.
func schemePicker(c *myui.Context, s *store.Settings, st Settings, act Actions) {
	// A page drawn without a window to keep its cache for it keeps none: it
	// derives the rows again for each frame, as it did before there was a
	// cache, rather than share one with the frame before it.
	var cache *SchemesCache
	if st.SchemesCache != nil {
		cache = st.SchemesCache
	}
	t := c.Theme()
	card(c, cardFace(t), cardLine(t), func() {
		field(c, "colorScheme", locales.Scheme(), locales.ColorSchemeNote(), func() {
			myui.Row(c).AlignItems(myui.Center).Gap(10).Children(func() {
				bg, fg := swatchOf(cache, st.Schemes, s.ColorScheme, t)
				schemeSwatch(c, bg, fg, 16)
				myui.Text(c, schemeNamed(*s)).FontSize(13).SingleLine().MaxWidth(220)
				myui.Button(c, locales.SchemeCopy()).OnClick(act.CopyScheme)
			})
		}, st.Reveal)
		myui.Column(c).Padding(10, 16).Gap(6).Children(func() {
			names := searchSchemes(cache, st.Schemes, filter(st))
			// A text input edits the string it is handed, so that string has to
			// outlive the frame. Handing it a fresh one every frame would throw
			// away what the user typed on the next frame: with no field to hold
			// the words (nil Filter, e.g. a test that draws the page alone),
			// the search box is left out rather than given nowhere to write.
			if st.Filter != nil {
				box := myui.TextInput(c, st.Filter).Placeholder(locales.SchemeSearch())
				// The keys reach the list through the box: a row is a click, and
				// a person who has typed a name has already found the scheme the
				// search puts first.
				if len(names) > 0 {
					first := names[0].Name
					box.OnSubmit(func() { act.PickScheme(first) })
				}
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
	})
}

// schemeNamed is the scheme in force as the page names it: its own name, or the
// setting's, for a file that names none.
func schemeNamed(s store.Settings) string {
	if s.ColorScheme == "" {
		return locales.Scheme()
	}
	return s.ColorScheme
}

// swatchOf is the colors the page shows the scheme called name by: the ones the
// list carries, and the window's own muted pair for a name the list has none of.
func swatchOf(cache *SchemesCache, schemes []Scheme, name string, t *myui.Theme) (bg, fg myui.Color) {
	sc, ok := findScheme(cache, schemes, name)
	if !ok {
		return t.Surface, t.TextMuted
	}
	return sc.Bg, sc.Fg
}

// findScheme is the scheme the list carries by that name, and whether it carries
// one. The list is hundreds of names long and the page asks for one of them a
// frame: the cache keeps them by name while the list is the one it was made from,
// and a page drawn without a window to keep a cache reads the list.
func findScheme(cache *SchemesCache, schemes []Scheme, name string) (Scheme, bool) {
	if cache == nil {
		for _, sc := range schemes {
			if sc.Name == name {
				return sc, true
			}
		}
		return Scheme{}, false
	}
	cache.Lock()
	defer cache.Unlock()
	// The list is read again only when it is another list: walking it to compare
	// every scheme would cost more than the walk this map saves.
	if cache.names == nil || !sameList(cache.named, schemes) {
		cache.names, cache.named = byName(schemes), schemes
	}
	sc, ok := cache.names[name]
	return sc, ok
}

// sameList reports whether two lists are the one list: the window builds the
// schemes it hands the page once per set of them, and keeps that list, so the
// same slice is the same list.
func sameList(a, b []Scheme) bool {
	if len(a) != len(b) {
		return false
	}
	return len(a) == 0 || &a[0] == &b[0]
}

// byName is the schemes by name: the first of a name that the list carries twice,
// which is the one the list would have shown first.
func byName(all []Scheme) map[string]Scheme {
	by := make(map[string]Scheme, len(all))
	for _, sc := range all {
		if _, ok := by[sc.Name]; !ok {
			by[sc.Name] = sc
		}
	}
	return by
}

// schemeSwatch is a scheme as a chip: its background, and a bar of the color its
// text would be in. The names alone tell a person nothing about the colors they
// pick, and a list of hundreds of them is where that is worth a glance.
func schemeSwatch(c *myui.Context, bg, fg myui.Color, size float32) {
	myui.Box(c).Width(size).Height(size * 0.75).Radius(4).Background(bg).
		BorderWidth(1).BorderColor(fg.Alpha(0.4)).
		AlignItems(myui.Center).Justify(myui.Center).Children(func() {
		myui.Box(c).Width(size * 0.5).Height(2).Radius(1).Background(fg)
	})
}

// filter is what the schemes are searched for.
func filter(st Settings) string {
	if st.Filter == nil {
		return ""
	}
	return *st.Filter
}

// schemeRow is one scheme in the list: its colors and its name, chosen by a
// click, and for a scheme the settings write the button that removes it --
// shown while the row is in front, in room the row keeps either way, so that
// the names do not jump as the pointer moves.
func schemeRow(c *myui.Context, sc Scheme, current string, act Actions) {
	t := c.Theme()
	el := myui.Row(c).Height(28).Padding(0, 4, 0, 8).AlignItems(myui.Center).Gap(8).Radius(6).
		Cursor(myui.CursorPointer).OnClick(func() { act.PickScheme(sc.Name) })
	chosen := current == sc.Name
	if chosen {
		el.Background(t.SurfaceHover)
	}
	shown := sc.Custom && (el.Hovered() || chosen)
	el.Children(func() {
		schemeSwatch(c, sc.Bg, sc.Fg, 14)
		name := myui.Text(c, sc.Name).Grow(1).Shrink(1).SingleLine()
		if chosen {
			// The scheme in force is heavier than the ones around it, so that a
			// list scrolled away and back still says what a save would keep.
			name.FontWeight(600)
		}
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

// SchemesCache is what the rows of the color schemes page were found from: the
// list of schemes, kept by pointer, and the rows last searched out of it. The
// page is drawn again on frames it did not build anew, as the pointer moves over
// it and the field is typed in, and the list is the same until the settings
// change it -- a list of hundreds is copied and searched in full every frame
// otherwise.
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
	// All is the list the rows were found from, nil before they were, and Key
	// the words they were searched for.
	All  []Scheme
	Key  string
	Rows []Scheme
	// names is the list by name, kept while named is the list it was made from.
	names map[string]Scheme
	named []Scheme
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

// reset empties the cache.
func (s *SchemesCache) reset() {
	s.Lock()
	defer s.Unlock()
	s.All, s.Key, s.Rows = nil, "", nil
	s.names, s.named = nil, nil
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
