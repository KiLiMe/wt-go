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

// schemePicker is the color schemes of the appearance page: the scheme in
// force, the button that copies it as one of the settings' own, the search box,
// and the list to pick from or remove. It is the page the settings had of its
// own, drawn as a card of the page that previews what the schemes look like.
func schemePicker(c *myui.Context, s *store.Settings, st Settings, act Actions) {
	// A page drawn without a window to keep its cache for it keeps none: it
	// derives the options and the rows again for each frame, as it did before
	// there was a cache, rather than share one with the frame before it.
	var cache *SchemesCache
	if st.SchemesCache != nil {
		cache = st.SchemesCache
	}
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
		myui.Column(c).Padding(10, 16).Gap(6).Children(func() {
			// A text input edits the string it is handed, so that string has to
			// outlive the frame. Handing it a fresh one every frame would throw
			// away what the user typed on the next frame: with no field to hold
			// the words (nil Filter, e.g. a test that draws the page alone),
			// the search box is left out rather than given nowhere to write.
			if st.Filter != nil {
				myui.TextInput(c, st.Filter).Placeholder(locales.SchemeSearch())
			}
			names := searchSchemes(cache, st.Schemes, filter(st))
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

// preview is the preview at the head of the appearance page: a terminal drawn
// as the draft would draw it, so that a change to the face, the typesetting,
// the cursor or the scheme shows as it is made rather than after a save.
func preview(c *myui.Context, s *store.Settings) {
	th := theme.Terminal(*s)
	t := c.Theme()
	// The window is translucent over what is behind it, which the preview has
	// no window behind: the scheme's own background at the same share of the
	// face the page paints on is what a person sees of it.
	bg := th.Background.MixOklab(t.Background, 1-float32(clampOpacity(s.BackgroundOpacity)))
	myui.Box(c).BorderWidth(1).BorderColor(cardLine(t)).Radius(8).Background(bg).
		Padding(previewPadX(s), previewPadY(s)).Children(func() {
		myui.Row(c).AlignItems(myui.Center).Gap(6).Children(func() {
			myui.Text(c, locales.LineOne()).TextColor(th.Foreground).Font(textFont(s)).FontSize(previewFontSize(s))
			cursorBlock(c, s, th)
		})
		myui.Text(c, locales.LineTwo()).TextColor(th.Foreground).Font(textFont(s)).FontSize(previewFontSize(s)).
			LineHeight(previewLineHeight(s))
		myui.Row(c).Gap(4).Wrap().Children(func() {
			for _, col := range th.Palette {
				myui.Box(c).Size(16, 16).Radius(3).Background(col)
			}
		})
	})
}

// clampOpacity holds a share to what the window can show, since the draft is
// drawn while it is typed and may name one the file would refuse.
func clampOpacity(v float64) float64 {
	return min(max(v, 0), 1)
}

// previewFontSize is the size the preview draws the terminal's lines with: the
// draft's own, held to a range that leaves the preview its room, so that a size
// typed into the file does not make the preview a card of lines alone.
func previewFontSize(s *store.Settings) float32 {
	return min(max(s.Font.Size, previewFontMin), previewFontMax)
}

// previewFontMin and previewFontMax are how small and how large the preview
// draws a line.
const (
	previewFontMin float32 = 9
	previewFontMax float32 = 20
)

// previewLineHeight is the draft's line height, held to what the control that
// sets it offers.
func previewLineHeight(s *store.Settings) float32 {
	return min(max(s.Font.LineHeight, 0.8), 2.4)
}

// textFont is the family the preview draws with: the draft's, unless it names
// a family this machine has not been asked for yet, which is none.
func textFont(s *store.Settings) string {
	if s.Font.Family == "" {
		return "monospace"
	}
	return s.Font.Family
}

// previewPadX and previewPadY are the room the preview keeps around its lines,
// held to what the preview has to give: a terminal's padding is measured for a
// window, and the preview is a card of one.
func previewPadX(s *store.Settings) float32 { return min(max(s.Padding.X, 0), 48) }
func previewPadY(s *store.Settings) float32 { return min(max(s.Padding.Y, 0), 24) }

// cursorBlock is the cursor of the draft, in its shape and the scheme's cursor
// color, beside the text it would sit after.
func cursorBlock(c *myui.Context, s *store.Settings, th *terminal.Theme) {
	// The line is as tall as the letters of the preview, so the cursor is too.
	h := previewFontSize(s) + 4
	shape := myui.Box(c).Height(h).Background(th.Cursor)
	switch s.CursorShape {
	case "block":
		shape.Width(previewFontSize(s))
	case "underline":
		shape.Width(previewFontSize(s)).Height(2)
	default:
		shape.Width(2)
	}
	// A blink is a moment, and the preview is a frame: the cursor is drawn
	// either way, so the switch says what is drawn and only the shape shows.
}
