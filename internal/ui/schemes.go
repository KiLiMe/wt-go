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
		field(c, locales.Scheme(), locales.ColorSchemeNote(), func() {
			myui.Row(c).AlignItems(myui.Center).Gap(10).Children(func() {
				bg, fg := swatchOf(st.Schemes, s.ColorScheme, t)
				schemeSwatch(c, bg, fg, 16)
				myui.Text(c, schemeNamed(*s)).FontSize(13).SingleLine().MaxWidth(220)
				myui.Button(c, locales.SchemeCopy()).OnClick(act.CopyScheme)
			})
		})
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
func swatchOf(schemes []Scheme, name string, t *myui.Theme) (bg, fg myui.Color) {
	for _, sc := range schemes {
		if sc.Name == name {
			return sc.Bg, sc.Fg
		}
	}
	return t.Surface, t.TextMuted
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
		Padding(previewPadY(s), previewPadX(s)).Children(func() {
		// The sample the preview is read off: the prompt with the cursor on it,
		// and the lines under it, which start where the prompt leaves the words.
		// A monospace face puts a letter, a digit and a symbol on the same
		// advance, so the columns of the three lines line up; a face that is not
		// monospace, or a fallback that draws the letters of a script at another
		// width, shows as the columns that do not, which is what the sample is
		// for. The words are the language's own, from its catalog.
		for i, line := range sampleLines() {
			myui.Row(c).AlignItems(myui.Center).Gap(6).Children(func() {
				myui.Text(c, line).TextColor(th.Foreground).Font(textFont(s)).
					FontSize(previewFontSize(s)).LineHeight(previewLineHeight(s))
				if i == 0 {
					cursorBlock(c, s, th)
				}
			})
		}
		// The sixteen colors a program names, in the two rows a terminal keeps
		// them in: the eight it names, and the bright eight below them, so that a
		// color stands over its bright twin. The room between them keeps two
		// neighbors from reading as one band, and the line around each keeps the
		// black of a scheme, which is its background's own color, from vanishing
		// into the page it is drawn on.
		myui.Column(c).Padding(8, 0, 0, 0).Gap(6).Children(func() {
			for _, row := range [][]myui.Color{th.Palette[:8], th.Palette[8:]} {
				myui.Row(c).Gap(6).Children(func() {
					for _, col := range row {
						myui.Box(c).Size(16, 16).Radius(3).Background(col).
							BorderWidth(1).BorderColor(cardLine(t))
					}
				})
			}
		})
	})
}

// sampleLines is what the preview says a terminal says: the prompt with the
// language's first line, and the two lines under it, drawn with the room the
// prompt takes so that the columns of the three line up. That is what the
// sample is for: a face that is not monospace, or a fallback that draws a
// script at another advance, puts them out of line where a person can see it.
func sampleLines() []string {
	return []string{
		"$ " + locales.LineOne(),
		"  " + locales.LineTwo(),
		"  " + locales.LineThree(),
	}
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
