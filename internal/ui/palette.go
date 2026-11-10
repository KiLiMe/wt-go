package ui

import (
	"sort"
	"strings"
	"unicode"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/i18n"
)

// The palette's size: eight rows of the list, and the padding above and
// below it.
const (
	paletteRowHeight float32 = 30
	palettePadding   float32 = 6

	// PaletteHeight is how tall the list is before it scrolls.
	PaletteHeight float32 = 8*paletteRowHeight + 2*palettePadding
)

// Palette is the command palette: open while the window lists what it does,
// and what the user typed and picked.
type Palette struct {
	// Open is whether the palette shows; the palette closes it.
	Open bool
	// Filter is what the field holds.
	Filter string
	// Selected is the row the keys pick, 0 the first.
	Selected int
	// Commands are what the palette offers, which the window builds for the
	// frame: the window knows the tabs and the keys, the palette only lists.
	Commands []Command
	// scroll keeps the list where the keys put it, and reveal says the row
	// the keys picked is to show.
	scroll myui.ScrollState
	reveal bool
}

// Command is one thing the palette offers.
type Command struct {
	// Action is what running it does.
	Action string
	// Title is what it is called, in the window's language.
	Title string
	// Group is what it belongs to: the word the filter finds it by, as
	// "Pane", and the heading it lists under.
	Group string
	// Key is the key that runs it, as a keybinding writes it, "" for none.
	Key string
}

// palette lists what the window does, and runs what is picked.
func palette(c *myui.Context, st State, act Actions) {
	p := st.Palette
	if p == nil || !p.Open {
		return
	}
	t := c.Theme()
	room := Content(c)
	items := Match(p.Filter, p.Commands)
	p.Selected = min(max(p.Selected, 0), max(len(items)-1, 0))
	reveal, chosen := p.reveal, ""
	p.reveal = false
	myui.Modal(c, &p.Open, func() {
		myui.Column(c).
			Width(min(600, max(room.W-32, 0))).
			Children(func() {
				myui.Row(c).
					Height(48).
					Padding(0, 16).
					AlignItems(myui.Center).
					Children(func() {
						in := myui.TextInputBase(c.Key("filter"), &p.Filter).
							Grow(1).
							Placeholder(i18n.T("Type the name of a command")).
							AutoFocus()
						if in.Changed() {
							p.Selected, p.reveal = 0, true
						}
						if c.Shortcut(0, myui.KeyDown) || c.Shortcut(myui.Ctrl, myui.KeyN) {
							p.Selected, p.reveal = min(p.Selected+1, len(items)-1), true
						}
						if c.Shortcut(0, myui.KeyUp) || c.Shortcut(myui.Ctrl, myui.KeyP) {
							p.Selected, p.reveal = max(p.Selected-1, 0), true
						}
						if (in.Submitted() || c.Shortcut(0, myui.KeyEnter)) && len(items) > 0 {
							chosen = items[p.Selected].Action
						}
					})
				myui.Divider(c)
				myui.Scroll(c.Key("commands")).
					Height(PaletteHeight).
					Padding(palettePadding).
					TrackScroll(&p.scroll).
					Children(func() {
						if len(items) == 0 {
							myui.Text(c, i18n.T("No command matches.")).
								FontSize(12).
								TextColor(t.TextMuted).
								Padding(10, 10)
						}
						for i, item := range items {
							row := paletteRow(c, item, i == p.Selected)
							if i == p.Selected && reveal {
								row.ScrollIntoView()
							}
							if row.Hovered() && row.Clicked() {
								chosen = item.Action
							}
						}
					})
			})
	})
	if chosen != "" {
		p.Open = false
		act.Run(chosen)
	}
}

// paletteRow is one command: its title, the group it belongs to, and the key
// that runs it.
func paletteRow(c *myui.Context, item Command, selected bool) myui.Element {
	t := c.Theme()
	el := myui.Row(c).
		Height(paletteRowHeight).
		Padding(0, 10).
		Gap(10).
		AlignItems(myui.Center).
		Radius(6).
		Cursor(myui.CursorPointer).
		Children(func() {
			myui.Text(c, item.Title).FontSize(13).TextColor(t.Text).SingleLine().Grow(1)
			myui.Text(c, item.Group).FontSize(11).TextColor(t.TextMuted).SingleLine()
			if item.Key != "" {
				myui.Text(c, item.Key).FontSize(11).TextColor(t.TextMuted).SingleLine()
			}
		})
	if selected {
		el = el.Background(t.SurfaceHover)
	}
	return el
}

// Match is the commands that answer a filter: the ones whose title holds what
// is typed, in order, before the ones a word of the group or the action names,
// and every command for an empty filter.
func Match(filter string, cmds []Command) []Command {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return append([]Command{}, cmds...)
	}
	type scored struct {
		cmd   Command
		score int
	}
	hits := make([]scored, 0, len(cmds))
	for _, cmd := range cmds {
		score := match(filter, cmd)
		if score < 0 {
			continue
		}
		hits = append(hits, scored{cmd, score})
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	out := make([]Command, len(hits))
	for i, h := range hits {
		out[i] = h.cmd
	}
	return out
}

// match scores how well a filter finds a command, -1 for one it does not: the
// title as letters in order, which counts for more, or the group and the words
// of the action, each the start of a word.
func match(filter string, cmd Command) int {
	score := fuzzy(filter, cmd.Title)
	if score >= 0 {
		// A title that holds what is typed to its letters in order is the
		// command the reader most likely means.
		score += 50
	}
	return max(score, words(filter, cmd.Group+" "+strings.ReplaceAll(cmd.Action, ".", " ")))
}

// fuzzy scores how well query matches s, as its letters in order; start of
// words and runs of letters score more, and -1 is no match.
func fuzzy(query, s string) int {
	if query == "" {
		return 0
	}
	q := []rune(strings.ToLower(query))
	r := []rune(strings.ToLower(s))
	score, qi, run := 0, 0, 0
	for i := 0; i < len(r) && qi < len(q); i++ {
		if q[qi] == ' ' {
			qi++
			continue
		}
		if r[i] != q[qi] {
			run = 0
			continue
		}
		bonus := 1
		if i == 0 || !unicode.IsLetter(r[i-1]) && !unicode.IsDigit(r[i-1]) {
			bonus += 6
		}
		run++
		score += bonus + run*2
		qi++
	}
	if qi < len(q) {
		return -1
	}
	return score*10 - len(r)
}

// words scores how well query finds words, a list of them: each term of the
// query the start of one.
func words(query, list string) int {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return -1
	}
	score := 0
	for _, t := range terms {
		best := -1
		for _, w := range strings.Fields(strings.ToLower(list)) {
			if strings.HasPrefix(w, t) {
				best = max(best, fuzzy(t, w))
			}
		}
		if best < 0 {
			return -1
		}
		score += best
	}
	return score
}
