package ui

import (
	"slices"

	myui "github.com/egoist/mygo/ui"

	"wtgo/locales"
)

// Boundary: the palette overlay -- its rows, its filter box and the list of
// what the window offers. Ranking a command against what was typed is match.go.

// The palette's size: eight rows of the list, and the padding above and
// below it.
const (
	paletteRowHeight float32 = 30
	palettePadding   float32 = 6

	// PaletteHeight is how tall the list is before it scrolls.
	PaletteHeight float32 = 8*paletteRowHeight + 2*palettePadding
)

// Palette is the command palette: open while the window lists what it does,
// and what the user typed and picked. The window keeps it from frame to frame.
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
	// matched is what filtering Commands by Filter last found, kept by its
	// two inputs so that the pointer moving over the list does not run the
	// filter and the sort again for each frame it draws.
	matched []Command
	filter  string
	cmds    []Command
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
	items := p.matchedCommands()
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
							Placeholder(locales.PalettePlaceholder()).
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
							myui.Text(c, locales.PaletteNoMatch()).
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

// matchedCommands are what the palette lists: the commands that answer the
// filter, which is what Match finds. The window builds the commands for the
// frame, the same ones until the tabs, the profiles or the keys change, so
// they are compared with what the answer was found among rather than sorted
// again. A palette is drawn on the goroutine that draws its window.
func (p *Palette) matchedCommands() []Command {
	if p.matched != nil && p.filter == p.Filter && slices.Equal(p.cmds, p.Commands) {
		return p.matched
	}
	items := Match(p.Filter, p.Commands)
	p.matched = items
	p.filter = p.Filter
	p.cmds = append(p.cmds[:0], p.Commands...)
	return items
}
