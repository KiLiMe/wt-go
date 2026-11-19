package ui

import (
	"sort"
	"strconv"
	"strings"

	myui "github.com/egoist/mygo/ui"

	"wtgo/locales"
)

// tabsListWidth is how wide the list of tabs is at most: as wide as the palette,
// since it is the same kind of list over the same window.
const tabsListWidth float32 = 600

// TabsList is every tab, listed. The strip shows the tabs it has room for, and
// scrolling it is one way to the rest; the list is the other, where a tab is
// found by the name it was given rather than by how far it has been scrolled
// out of sight.
type TabsList struct {
	// Open is whether the list shows; the list closes it.
	Open bool
	// Filter is what the field holds.
	Filter string
	// Selected is the row the keys pick, 0 the first.
	Selected int
	// scroll keeps the list where the keys put it, and reveal says the row the
	// keys picked is to show.
	scroll myui.ScrollState
	reveal bool
	// matched is what filtering the tabs by Filter last found, kept with the two
	// things it was found from, so that a frame that changes neither of them does
	// not score every tab again.
	matched []int
	filter  string
	names   []string
	known   bool
}

// matchedTabs is the tabs that answer the filter, by their place in the strip.
func (l *TabsList) matchedTabs(st State) []int {
	names := make([]string, len(st.Tabs))
	for i, tb := range st.Tabs {
		names[i] = tb.Name
	}
	if l.known && l.filter == l.Filter && sameNames(l.names, names) {
		return l.matched
	}
	items := MatchTabs(l.Filter, names)
	if items == nil {
		// nil stands for an answer never found, and the next frame still finds
		// it here.
		items = []int{}
	}
	l.matched = items
	l.filter = l.Filter
	l.names = append(l.names[:0], names...)
	l.known = true
	return items
}

// sameNames reports whether two lists name the same tabs, in the same order.
func sameNames(a, b []string) bool {
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

// MatchTabs is the tabs that answer a filter, by their place in the strip: the
// ones whose name holds what is typed to its letters in order, in order, and
// every tab for an empty filter. A tab found by a letter of its name is found
// no matter how long the name was or where in the strip the tab sits.
func MatchTabs(filter string, names []string) []int {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		out := make([]int, len(names))
		for i := range names {
			out[i] = i
		}
		return out
	}
	type scored struct {
		at    int
		score int
	}
	hits := make([]scored, 0, len(names))
	for i, name := range names {
		if score := fuzzy(filter, name); score >= 0 {
			hits = append(hits, scored{i, score})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	out := make([]int, len(hits))
	for i, h := range hits {
		out[i] = h.at
	}
	return out
}

// tabsList draws the list of every tab, and brings the one picked in front.
func tabsList(c *myui.Context, st State, act Actions) {
	l := st.TabsList
	if l == nil || !l.Open {
		return
	}
	t := c.Theme()
	room := Content(c)
	tabs := l.matchedTabs(st)
	l.Selected = min(max(l.Selected, 0), max(len(tabs)-1, 0))
	// The row the keys picked, and the one a click picks: both are a place in the
	// strip by the time the frame is through.
	selected := -1
	if len(tabs) > 0 {
		selected = tabs[l.Selected]
	}
	reveal, chosen := l.reveal, -1
	l.reveal = false
	myui.Modal(c, &l.Open, func() {
		myui.Column(c).Width(min(tabsListWidth, max(room.W-32, 0))).Children(func() {
			myui.Row(c).Height(48).Padding(0, 16).AlignItems(myui.Center).Children(func() {
				in := myui.TextInputBase(c.Key("tabFilter"), &l.Filter).
					Grow(1).
					Placeholder(locales.TabListPlaceholder()).
					AutoFocus()
				if in.Changed() {
					l.Selected, l.reveal = 0, true
				}
				if c.Shortcut(0, myui.KeyDown) || c.Shortcut(myui.Ctrl, myui.KeyN) {
					l.Selected, l.reveal = min(l.Selected+1, len(tabs)-1), true
				}
				if c.Shortcut(0, myui.KeyUp) || c.Shortcut(myui.Ctrl, myui.KeyP) {
					l.Selected, l.reveal = max(l.Selected-1, 0), true
				}
				if (in.Submitted() || c.Shortcut(0, myui.KeyEnter)) && selected >= 0 {
					chosen = selected
				}
			})
			myui.Divider(c)
			myui.Scroll(c.Key("tabList")).
				Height(PaletteHeight).
				Padding(palettePadding).
				TrackScroll(&l.scroll).
				Children(func() {
					if len(tabs) == 0 {
						myui.Text(c, locales.TabListNoMatch()).
							FontSize(12).
							TextColor(t.TextMuted).
							Padding(10, 10)
					}
					for _, i := range tabs {
						row := tabsListRow(c, st, i, i == selected)
						if i == selected && reveal {
							row.ScrollIntoView()
						}
						if row.Hovered() && row.Clicked() {
							chosen = i
						}
					}
				})
		})
	})
	if chosen >= 0 {
		l.Open = false
		act.FocusTab(chosen)
	}
}

// tabsListRow is one tab in the list: the number the strip answers to, the mark
// of the profile it runs, its whole name, and the mark of the tab in front. The
// name wraps rather than ends in an ellipsis: a list that cuts the names short
// is the strip again, and the strip is what it was opened to see past.
func tabsListRow(c *myui.Context, st State, i int, selected bool) myui.Element {
	t := c.Theme()
	tb := st.Tabs[i]
	active := i == st.Active
	el := myui.Row(c).
		MinHeight(paletteRowHeight).
		Padding(6, 10).
		Gap(10).
		AlignItems(myui.Center).
		Radius(6).
		Cursor(myui.CursorPointer).
		Children(func() {
			myui.Text(c, strconv.Itoa(i+1)).FontSize(11).TextColor(t.TextMuted).Width(18)
			tabIcon(c, tb, active)
			name := myui.Text(c, tb.Name).FontSize(13).TextColor(t.Text).Wrap().Grow(1)
			if active {
				name.FontWeight(600)
			}
		})
	if active {
		// The tab in front is the one the list is opened from, so it is the one
		// the list marks: the strip's own seam is not there to say it.
		el.Children(func() {
			myui.Box(c).Size(6, 6).Radius(3).Background(t.Accent)
		})
	}
	switch {
	case selected:
		el = el.Background(t.SurfaceHover)
	case el.Hovered():
		el = el.Background(t.Surface)
	}
	return el
}
