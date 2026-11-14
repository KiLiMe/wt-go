package ui

import (
	"sync"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
	"wtgo/internal/icons"
	"wtgo/locales"
)

// TitleBarHeight is the height of the tab strip, which is the title bar: the
// window keeps it clear of its window controls, and the strip runs under
// them. It is Windows Terminal's 30 DIP, so that a window beside one keeps
// the two bars level and its pane starts where the other one's does.
const TitleBarHeight = 30

// stripGutter is the room the strip keeps between the window's left edge and
// the first tab, for the same reason.
const stripGutter float32 = 8

// nameBold is the weight a bold name is drawn in, as myui's Bold names it.
const nameBold = 700

// tabChrome is the room a tab takes around its name: the padding on both sides,
// the mark before the name and the button that closes it, with the gaps between
// them.
const tabChrome float32 = 68

// The room a tab takes: Windows Terminal's widest, which a tab with a long name
// and room to spare reaches, and the narrowest that still holds the mark, the
// close button and a name -- below it a tab says nothing, and the strip scrolls
// instead of squeezing the tabs further.
const (
	minTabWidth float32 = tabChrome + minTabName
	maxTabWidth float32 = 240
)

// The room a tab's name may take: as little as a name that is a single
// letter, as much as fits the widest tab.
const (
	minTabName float32 = 48
	maxTabName float32 = maxTabWidth - tabChrome
)

// mouseMiddle is the button that closes the tab it points at, as it does in
// Windows Terminal.
const mouseMiddle = 2

// The room the strip keeps whatever the tabs do: the buttons that move them,
// and the part of the title bar that moves the window. A strip of many tabs
// never takes the last of either.
const (
	stepWidth float32 = 24
	dragRoom  float32 = 56
)

// The room the buttons beside the tabs take: a button's square, the line beside
// it with the room it clears on both sides, and the gap between two children.
const (
	stripButtonSize    float32 = 28
	stripDividerGutter float32 = 4
	stripDividerRoom   float32 = 1 + 2*stripDividerGutter
	stripButtonsGap    float32 = 2
)

// The room a window button takes, the gap between two of them, and the room
// the buttons keep from the corner of the window.
const (
	windowButtonWidth float32 = 34
	windowButtonGap   float32 = 2
	windowButtonRoom  float32 = 6
)

// windowIcon is how wide the icon of a window button is: a square that fills
// the icon area of the desktop's own buttons.
const windowIcon float32 = 10

// tabStep is how far the buttons beside the tabs move what the strip shows:
// about one tab, so that a press brings the next tabs into view.
const tabStep float32 = 176

// MinWindowWidth is the narrowest the window may be and still show a tab: the
// room of one tab at its narrowest, the buttons beside the tabs, the room that
// moves the window, the two buttons that scroll the tabs, and the desktop's
// own window buttons -- three of them at the widest a desktop draws them, which
// is Windows' 46 DIP each, so that macOS' traffic lights fit too. Narrower than
// this the strip shows no tab at all, only the buttons that move tabs that are
// not there. It is a variable because stripButtonsWidth is a function.
var MinWindowWidth = stripGutter + minTabWidth + dragRoom +
	stripButtonsWidth() + 2*stepWidth + 3*46

// tabDrag is a tab on its way to another place in the strip.
type tabDrag struct {
	from int
}

// tabMove is a tab dropped on another place: the tab at from belongs at to.
type tabMove struct {
	from, to int
}

// MinWindowHeight is the shortest the window may be: the strip tall enough for
// the title bar, and a pane below it tall enough for a prompt and the two lines
// above it at the default font and padding -- three rows of 16 DIPs and the
// padding of 8 above and below. Short of it the pane holds no line of output at
// all, and the window has nothing to show.
const MinWindowHeight float32 = TitleBarHeight + 3*16 + 2*8

// strip is the title bar: the window's buttons where the window draws them,
// the tabs with the buttons that open one, and the room that moves the window.
func strip(c *myui.Context, st State, act Actions) {
	t := c.Theme()
	bar := c.TitleBar()
	myui.Row(c).
		Height(TitleBarHeight).
		Background(t.Surface).
		Padding(0, bar.Right, 0, bar.Left+stripGutter).
		DragWindow().
		Children(func() {
			if b := st.WindowButtons; b != nil && len(b.Start) > 0 {
				windowButtonGroup(c, b, act, b.Start, false)
			}
			stripTabs(c, st, act)
			myui.Box(c).Width(dragRoom).Height(TitleBarHeight)
			if b := st.WindowButtons; b != nil && len(b.End) > 0 {
				windowButtonGroup(c, b, act, b.End, true)
			}
		})
}

// windowButtonGroup draws the window's buttons the desktop puts on a side of
// the title bar, at the end of it when end is set, and nothing where it puts
// none: a desktop may name some, and a tiling window manager names none.
func windowButtonGroup(c *myui.Context, b *WindowButtons, act Actions, names []string, end bool) myui.Element {
	row := myui.Row(c).
		Height(TitleBarHeight).
		Gap(windowButtonGap).
		AlignItems(myui.Center)
	if end {
		row = row.Padding(0, windowButtonRoom, 0, 0)
	} else {
		row = row.Padding(0, 0, 0, windowButtonRoom)
	}
	return row.Children(func() {
		for _, name := range names {
			windowButton(c, b, act, name)
		}
	})
}

// windowButton is one of the window's buttons. It draws each mark itself, in
// 1px lines at the size the title bar asks: the glyphs for these are missing
// from some fonts, or wide in the fonts that draw them, and the strokes of an
// icon set come out bold at this size.
func windowButton(c *myui.Context, b *WindowButtons, act Actions, name string) {
	t := c.Theme()
	label := ""
	var icon func(color myui.Color, bg myui.Color)
	switch name {
	case "minimize":
		label = locales.Minimize()
		icon = func(color myui.Color, bg myui.Color) {
			myui.Box(c).Width(windowIcon).Height(1).Background(color)
		}
	case "maximize":
		label = locales.Maximize()
		if b.Maximized {
			label = locales.Restore()
			icon = func(color myui.Color, bg myui.Color) {
				myui.Box(c).Size(windowIcon+1, windowIcon+1).Children(func() {
					myui.Box(c).Absolute().Left(4).Top(0).Size(windowIcon-3, windowIcon-3).Radius(1).Border(1, color)
					myui.Box(c).Absolute().Left(0).Top(4).Size(windowIcon-3, windowIcon-3).Radius(1).Border(1, color).Background(bg)
				})
			}
		} else {
			icon = func(color myui.Color, bg myui.Color) {
				myui.Box(c).Size(windowIcon-1, windowIcon-1).Radius(1).Border(1, color)
			}
		}
	case "close":
		label = locales.CloseWindow()
		icon = func(color myui.Color, bg myui.Color) {
			// A cross of 1px lines, as Windows draws it: the icon set's close
			// is a 2-unit stroke on a 24 grid, which comes out bold beside a
			// 1px minus and box.
			myui.Box(c).Size(windowIcon, windowIcon).Draw(func(p *myui.Painter, r myui.Rect) {
				const in = 0.5
				// Painter.Line paints only horizontal or vertical lines: given a
				// slant it paints the top edge of the box between the points, which
				// showed this cross as a second minus. StrokePolyline draws the slant.
				p.StrokePolyline([][2]float32{{r.X + in, r.Y + in}, {r.X + r.W - in, r.Y + r.H - in}}, 1, color)
				p.StrokePolyline([][2]float32{{r.X + r.W - in, r.Y + in}, {r.X + in, r.Y + r.H - in}}, 1, color)
			})
		}
	default:
		return
	}
	bg := t.Surface
	color := t.TextMuted
	el := myui.Box(c).
		Size(windowButtonWidth, 28).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Label(label).
		Tooltip(label)
	if el.Hovered() {
		bg = t.SurfaceHover
		color = t.Text
		if name == "close" {
			// The button that takes the window away turns the red Windows
			// gives it: the one button of the three that does something the
			// others do not is the one worth a color of its own.
			bg = t.Danger
		}
		el = el.Background(bg)
	}
	el.OnClick(func() {
		switch name {
		case "minimize":
			act.MinimizeWindow()
		case "maximize":
			act.ToggleMaximizeWindow()
		case "close":
			act.CloseWindow()
		}
	}).Children(func() {
		icon(color, bg)
	})
}

// stripTabs is the room the tabs take: the buttons that move them, and the
// tabs themselves, which scroll sideways when they outgrow the room.
func stripTabs(c *myui.Context, st State, act Actions) {
	if st.TabsScroll == nil {
		room := tabsRoom(c, st, false)
		myui.Row(c).Height(TitleBarHeight).Grow(1).AlignItems(myui.Center).Children(func() {
			tabs(c, st, act, room)
			stripButtons(c, st, act, true)
		})
		return
	}
	// The buttons take room from the tabs, so they only show once the tabs
	// outgrow it. The viewport is measured every frame whether they show or
	// not, so the frame after a tab no longer fits has them, and a frame
	// they are in keeps them: what they take is what makes the tabs
	// outgrow the strip.
	scrolls := st.TabsScroll.MaxX > 0
	room := tabsRoom(c, st, scrolls)
	myui.Row(c).Height(TitleBarHeight).Grow(1).AlignItems(myui.Center).Children(func() {
		if scrolls {
			stepButton(c, st, act, -1)
		}
		viewport(c, st, act, room)
		if scrolls {
			stepButton(c, st, act, 1)
		}
		// The buttons that open a tab sit past the tabs rather than in them:
		// however many tabs there are, a tab is a press away rather than a
		// scroll to the end away.
		stripButtons(c, st, act, !scrolls)
	})
}

// tabsRoom is the room the tabs have: the window, less the strip's own padding,
// the room that moves the window and the window's buttons, the buttons beside
// the tabs and, when they show, the buttons that move the tabs.
func tabsRoom(c *myui.Context, st State, scrolls bool) float32 {
	w, _ := c.Size()
	bar := c.TitleBar()
	room := w - bar.Left - stripGutter - bar.Right - dragRoom - stripButtonsWidth()
	if b := st.WindowButtons; b != nil {
		room -= windowGroupWidth(len(b.Start)) + windowGroupWidth(len(b.End))
	}
	if scrolls {
		room -= 2 * stepWidth
	}
	return max(room, 0)
}

// stripButtonsWidth is the room the buttons beside the tabs take, as
// stripButtons lays them out: the one that opens a tab, the menu beside it, and
// the line on either side of the first. The line on its left is the tab's, so
// it is not drawn beside the tab in front -- but it keeps its room either way,
// and this is why the width takes no argument: what the strip gives the tabs,
// and where the buttons after them stand, are the same whichever tab is in
// front. Windows Terminal holds that too, measured on 2026-11-14: the buttons
// stand in the same place in its strip with either of two tabs in front, and
// only that line comes and goes.
func stripButtonsWidth() float32 {
	return 2*stripButtonSize + 2*stripDividerRoom + 3*stripButtonsGap
}

// windowGroupWidth is the room a group of n window's buttons takes.
func windowGroupWidth(n int) float32 {
	if n <= 0 {
		return 0
	}
	return float32(n)*windowButtonWidth + float32(n-1)*windowButtonGap + windowButtonRoom
}

// viewport is what the strip shows of the tabs: it scrolls them sideways when
// they outgrow it, which the wheel over it does too.
func viewport(c *myui.Context, st State, act Actions, room float32) {
	myui.ScrollHorizontal(c).
		Height(TitleBarHeight).
		Scrollbars(false).
		TrackScroll(st.TabsScroll).
		HandleInput(func(ev myui.InputEvent) bool {
			if ev.Kind != myui.InputScroll {
				return false
			}
			by := ev.DX
			if by == 0 {
				by = ev.DY
			}
			if by == 0 {
				return false
			}
			// The strip has no height to scroll in, so the wheel goes
			// sideways: it is the only way to reach the tabs past the
			// edge with a wheel.
			act.ScrollTabs(by)
			return true
		}).
		Children(func() { tabs(c, st, act, room) })
}

// stepButton moves what the strip shows by a tab's worth, toward the start
// with -1 and toward the end with 1. It fades when there is nothing that way,
// rather than going Disabled, which would take the press with it: the press
// is harmless, the window keeps the offset in the strip.
func stepButton(c *myui.Context, st State, act Actions, dir int) myui.Element {
	t := c.Theme()
	// The arrows come from the icon set: the angle quotes that would stand
	// for them read as specks at this size, and some fonts have no such
	// glyph at all.
	ic, tip := icons.Right, locales.ScrollTabsRight()
	if dir < 0 {
		ic, tip = icons.Left, locales.ScrollTabsLeft()
	}
	at := st.TabsScroll
	color := t.Text
	if at == nil || dir < 0 && at.X <= 0 || dir > 0 && at.X >= at.MaxX {
		color = t.TextMuted
		color.A = uint8(float32(color.A) * 0.45)
	}
	el := myui.Box(c).
		Size(stepWidth, 28).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Label(tip).
		Tooltip(tip).
		OnClick(func() { act.ScrollTabs(float32(dir) * tabStep) })
	if el.Hovered() {
		el = el.Background(t.SurfaceHover)
	}
	return el.Children(func() {
		myui.Icon(c, ic.SVG).Size(16, 16).TextColor(color)
	})
}

// tabs are the tabs, in order, each a place another tab may be dropped on, and
// the buttons that open a tab after the last of them. room is the room they
// have, which they split.
func tabs(c *myui.Context, st State, act Actions, room float32) {
	share := tabsShare(st, room)
	var move *tabMove
	myui.Row(c).Height(TitleBarHeight).Children(func() {
		for i := range st.Tabs {
			// Windows Terminal draws a line between two tabs that are both
			// behind, and none beside the tab in front: that tab is the dark
			// rounded one, which is the line the others are given. The line
			// keeps its room either way, so that a tab is as wide with it in
			// front as with the tab next to it in front.
			if i > 0 {
				dividerSlot(c, i != st.Active && i-1 != st.Active, tabLineRoom)
			}
			el := tab(c, st, act, i, share)
			if d, ok := myui.Drop[tabDrag](el); ok && move == nil {
				move = &tabMove{from: d.from, to: i}
			}
		}
	})
	if move != nil {
		act.MoveTab(move.from, move.to)
	}
}

// tabLineRoom is the room a line between two tabs takes: the line itself, which
// sits flush against both of them.
const tabLineRoom float32 = 1

// tabsShare is the room one tab may take: the room the tabs have, less the lines
// between them, split among them. Every place a line can take keeps one, drawn
// or not -- the line beside the tab in front is left out -- so that the tabs keep
// their width as another tab comes to the front. A tab takes no more than this,
// so that tabs grow narrower as more of them come, before the strip starts to
// scroll.
func tabsShare(st State, room float32) float32 {
	n := len(st.Tabs)
	if n == 0 {
		return 0
	}
	return max((room-float32(n-1)*tabLineRoom)/float32(n), 0)
}

// stripButtons are the buttons after the last tab: the one that opens a tab
// and the menu that picks the profile it runs, with the lines Windows Terminal
// draws on both sides of the button that opens a tab. They sit right after the
// tabs rather than at the far end of a wide window, so that a tab is a press
// away whatever the strip shows. beside says the buttons touch the last tab
// rather than standing past the arrows that scroll it.
func stripButtons(c *myui.Context, st State, act Actions, beside bool) myui.Element {
	// The line on the left belongs to the tab it separates from, and a tab in
	// front is given none: beside the tab in front, that line is left out, as
	// Windows Terminal leaves it out between two of its own. It keeps its room
	// even then: a strip whose buttons moved as another tab came to the front
	// would make the same press land on a different button.
	front := len(st.Tabs) > 0 && st.Active == len(st.Tabs)-1
	return myui.Row(c).Height(TitleBarHeight).AlignItems(myui.Center).Gap(stripButtonsGap).Children(func() {
		dividerSlot(c, !beside || !front, stripDividerRoom)
		newTabButton(c, act)
		dividerSlot(c, true, stripDividerRoom)
		profilesMenu(c, st, act)
	})
}

// tab is one tab: the dot of the profile it runs, its name, and the button
// that closes it.
func tab(c *myui.Context, st State, act Actions, i int, share float32) myui.Element {
	t := c.Theme()
	tb := st.Tabs[i]
	active := i == st.Active
	menu := func(m *myui.Menu) { tabMenu(m, st, act, i) }
	el := myui.Row(c).
		Height(TitleBarHeight).
		AlignItems(myui.Center).
		Gap(8).
		Padding(0, 8, 0, 12).
		Cursor(myui.CursorPointer).
		OnClick(func() { act.FocusTab(i) }).
		Drag(tabDrag{from: i}).
		ContextMenu(menu)
	nameWeight := 0
	if active {
		// The tab in front is the one the eye looks for first: it names itself
		// in bold, as Windows Terminal's does.
		nameWeight = nameBold
	}
	name := measuredName(c, tb.Name, nameWeight)
	// The tab in front carries the pane's colors up into the strip; a tab that
	// draws a page of its own, the settings, has no pane of the terminal's, and
	// neither has a window that draws no pane below the strip. Those take the
	// theme's own colors.
	face := st.Pane
	if face == (Face{}) || (active && tb.Page != nil) {
		face = Face{Background: t.Background, Foreground: t.Text}
	}
	// A tab is as wide as the share of the strip it has, but not wider than
	// its name asks for -- unless the settings make every tab alike -- and not
	// narrower than the room its parts take: past that the strip scrolls
	// rather than squeezing the name and the close button away.
	natural := name + tabChrome
	w := min(max(share, minTabWidth), maxTabWidth)
	if !st.EqualTabs {
		w = min(w, natural)
	}
	el = el.Width(w)
	name = w - tabChrome
	// A tab keeps its width when the strip is short of room: it is the strip
	// that scrolls, not the name and the close button that go.
	el = el.Shrink(0)
	if active {
		// The tab in front carries the pane's background up into the strip,
		// rounded at the top the way Windows Terminal does: the tab joins the
		// pane below it, and the strip is what tells the two apart. The color
		// is the pane's own, not a shade the theme derives: a window whose
		// theme is not the terminal's would otherwise seam at that edge.
		el = el.Background(face.Background).Radius(8, 8, 0, 0)
	} else if el.Hovered() {
		el = el.Background(t.SurfaceHover)
	}
	if active && st.RevealTab {
		// The strip shows the tab in front when it changes, once: asking
		// every frame would keep the user from scrolling it away.
		el = el.ScrollIntoView()
	}
	el = el.ObserveInput(func(ev myui.InputEvent) {
		if ev.Kind == myui.InputPointerDown && ev.Button == mouseMiddle {
			act.CloseTab(i)
		}
	})
	el.Children(func() {
		tabIcon(c, tb, active)
		nameColor := t.TextMuted
		if active {
			nameColor = face.Foreground
		}
		// The name is a label, not text to select: pressing it starts a drag
		// of the tab, the way pressing the rest of the tab does. It carries
		// the tab's menu itself: an unselectable element ends the search for
		// one above it, so a menu only on the row would not open over the
		// name, the widest part of the tab.
		text := myui.Text(c, tb.Name).
			SingleLine().MaxWidth(name).Ellipsis("…").TextColor(nameColor).
			Unselectable().ContextMenu(menu)
		// The name takes the room the tab gives it, not only what its letters
		// need: a short name would otherwise leave the close button a gap short
		// of the tab's edge, where nothing is drawn.
		text = text.Grow(1)
		if nameWeight != 0 {
			// Bold sets the weight on the element already drawing: the call
			// mutates it, so its result is not needed.
			text.Bold()
		}
		if active || el.Hovered() {
			closeButton(c, act, i, active, face)
		} else {
			// The button keeps its room, so that a tab does not change width
			// as the pointer comes over it.
			myui.Box(c).Size(16, 16)
		}
	})
	return el
}

// nameMeasures is bounded so that a window that has opened many tabs over a
// day does not keep a measurement for each of them: a name is stable, so the
// cache is dropped whole rather than aged out one at a time.
const nameMeasures = 256

// nameMeasure is a key of the tab-name cache: what the width of a name
// depends on, since the same name is a different width in another font.
type nameMeasure struct {
	name, font string
	size       float32
	weight     int
}

// nameMeasureCache keeps the widths of tab names. A window draws its strip
// on frames it did not build anew, as the pointer moves over it, so a name
// measured every frame costs a layout each frame; the names of the tabs
// change far less often. The mutex is for the window that draws on another
// goroutine than the one that reads the settings: the window is the only
// writer of its names either way.
var nameMeasureCache = struct {
	sync.Mutex
	widths map[nameMeasure]float32
}{widths: map[nameMeasure]float32{}}

// forTest, called for its side effects, empties the caches this package keeps
// across frames, so that a test draws what a window draws and sees nothing of
// the frame before it. Those caches are kept by the frame a window builds it
// view from, and a test builds its own views, so a cache that did not empty
// would answer a test with the answer of another one.
var forTest = func() {}

// resetCaches empties every cache the package keeps across frames.
func resetCaches() {
	nameMeasureCache.Lock()
	clear(nameMeasureCache.widths)
	nameMeasureCache.Unlock()
}

// measuredName is how wide a tab draws its name: as wide as the name takes
// in the theme's font, within the room a tab may take. Weight is the weight
// the name is drawn in, so that a bold name is measured bold: a regular
// width would ellipsize it in the tab it is widest in.
func measuredName(c *myui.Context, name string, weight int) float32 {
	t := c.Theme()
	key := nameMeasure{name: name, font: t.Font, size: t.FontSize, weight: weight}
	nameMeasureCache.Lock()
	w, ok := nameMeasureCache.widths[key]
	nameMeasureCache.Unlock()
	if !ok {
		w, _ = c.MeasureText(0, myui.Span{Text: name, Font: t.Font, Size: t.FontSize, Weight: weight})
		nameMeasureCache.Lock()
		// A full cache is dropped rather than aged: 256 names is nothing to
		// measure again, and an eviction order would cost more than that.
		if len(nameMeasureCache.widths) >= nameMeasures {
			clear(nameMeasureCache.widths)
		}
		nameMeasureCache.widths[key] = w
		nameMeasureCache.Unlock()
	}
	return nameWidth(w)
}

// nameWidth keeps a measured name inside the room a tab may take.
func nameWidth(w float32) float32 {
	return min(max(w, minTabName), maxTabName)
}

// tabMenu is the tab's context menu. The tab that is not a terminal, the
// settings, has no name of its own to keep and nothing to open a second time:
// its menu offers only closing it, so it opens without the line that would
// separate what it does not offer.
func tabMenu(m *myui.Menu, st State, act Actions, i int) {
	if i >= 0 && i < len(st.Tabs) && st.Tabs[i].Page == nil {
		if m.Item(locales.Rename()).Chosen() {
			act.BeginRename(i)
		}
		if m.Item(locales.DuplicateTab()).Chosen() {
			act.DuplicateTab(i)
		}
		m.Separator()
	}
	if m.Item(locales.CloseTab()).Chosen() {
		act.CloseTab(i)
	}
	if m.Item(locales.CloseOtherTabs()).Disabled(len(st.Tabs) < 2).Chosen() {
		act.CloseOtherTabs(i)
	}
}

// tabIcon is the mark before a tab's name: the mark of the profile it runs, or
// a gear for the tab that holds the settings.
func tabIcon(c *myui.Context, tb Tab, active bool) {
	t := c.Theme()
	if tb.Page != nil {
		iconTile(c, icons.Settings, t.Accent, inkOn(t.Accent), active)
		return
	}
	markIcon(c, tb.Icon, tb.Color, 16, active)
}

// markIcon is the mark of a profile at the size of the room it has: a preset
// draws its own face and its own mark, a character is drawn in the square the
// terminal mark would have taken, and a profile that names neither wears the
// terminal mark on a square of its color.
//
// A profile without a color gets the quiet square, since a color the user never
// chose should not be the loudest thing in the strip. The quiet square still
// needs a mark that reads on it: TextMuted on Border is a smudge at this size,
// and a profile without a color is the one every tab of a fresh install wears.
func markIcon(c *myui.Context, icon, color string, size float32, active bool) {
	t := c.Theme()
	if preset := icons.Preset(icon); preset != nil {
		// The face is the preset's, not the profile's color, so only the
		// brightness says whether the tab is in front of the window.
		el := myui.Box(c).Size(size, size).Shrink(0).Center().Children(func() {
			// Image, not Icon: a preset is a picture with its own colors,
			// and Icon paints every shape it draws in one color.
			myui.Image(c, preset.SVG).Size(size, size)
		})
		if !active {
			el.Opacity(0.85)
		}
		return
	}
	face := colorOf(color)
	switch {
	case icon != "" && face.A > 0:
		glyphTile(c, icon, face, inkOn(face), active, size)
	case icon != "":
		glyphTile(c, icon, t.Border, t.Text, active, size)
	case face.A > 0:
		iconTile(c, icons.Terminal, face, inkOn(face), active)
	default:
		iconTile(c, icons.Terminal, t.Border, t.Text, active)
	}
}

// iconTile is a square of a profile's color with a mark on it, for the marks
// that are drawn rather than written: the plain terminal, and the gear of the
// settings tab.
func iconTile(c *myui.Context, ic *icons.Icon, face, mark myui.Color, active bool) {
	const size float32 = 16
	if !active {
		face = face.Alpha(0.85)
	}
	myui.Box(c).Size(size, size).Shrink(0).Radius(size * 0.3).Background(face).Center().Children(func() {
		myui.Icon(c, ic.SVG).Size(size*0.62, size*0.62).TextColor(mark)
	})
}

// glyphTile is the square a character mark is drawn on: the character in place
// of the terminal mark, on the face of the profile's color. A color glyph, as
// an emoji is, keeps the colors it has.
func glyphTile(c *myui.Context, glyph string, face, mark myui.Color, active bool, size float32) {
	if !active {
		face = face.Alpha(0.85)
	}
	myui.Box(c).Size(size, size).Shrink(0).Radius(size * 0.3).Background(face).Center().Children(func() {
		myui.Text(c, glyph).FontSize(size * 0.72).SingleLine().TextColor(mark)
	})
}

// closeButton closes a tab. It takes the surface under the pointer, so that
// it reads as the button it is rather than as a mark on the tab, and it keeps
// the text color of the tab it sits on: Windows Terminal's close button is as
// bright on the tab behind as on the tab in front.
func closeButton(c *myui.Context, act Actions, i int, active bool, face Face) myui.Element {
	t := c.Theme()
	text := t.Text
	el := myui.Box(c).
		Size(16, 16).
		Radius(4).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Tooltip(locales.CloseTab()).
		OnClick(func() { act.CloseTab(i) })
	if active {
		// The button sits on the pane's background, which the theme knows
		// nothing about, so its hover takes that background rather than the
		// theme's surface.
		text = face.Foreground
		if el.Hovered() {
			el = el.Background(face.Foreground.Alpha(0.16))
		}
	} else if el.Hovered() {
		el = el.Background(t.SurfaceHover)
	}
	return el.Children(func() {
		myui.Icon(c, icons.Close.SVG).Size(14, 14).TextColor(text)
	})
}

// stripProfiles are what the menu of the strip lists. With every profile
// deleted or hidden it lists one entry, following the device, which is the
// shell the window runs then anyway: a menu with nothing to open a tab on
// reads as a window that cannot open one. That entry names the device the way
// the settings page does, so the menu says which shell it opens.
func stripProfiles(st State) []Profile {
	if len(st.Profiles) > 0 {
		return st.Profiles
	}
	return []Profile{{Name: deviceLabel(st.Device)}}
}

// profilesMenu opens a tab that runs one of the profiles.
func profilesMenu(c *myui.Context, st State, act Actions) myui.Element {
	base := myui.Box(c).
		Size(stripButtonSize, stripButtonSize).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Label(locales.Profiles()).
		Tooltip(locales.Profiles()).Menu(func(m *myui.Menu) {
		for i, p := range stripProfiles(st) {
			it := m.Item(p.Name)
			if p.ID == "" {
				// Following the device is no numbered profile of the file.
				if it.Chosen() {
					act.NewSystemTab()
				}
				continue
			}
			if i < 9 {
				shortcut(it, st, action.ProfileAction(i+1))
			}
			if it.Chosen() {
				act.NewTab(p.ID)
			}
		}
		m.Separator()
		it := m.Item(locales.Settings())
		shortcut(it, st, action.Settings)
		if it.Chosen() {
			act.OpenSettings()
		}
		if about := m.Item(locales.About()); about.Chosen() {
			act.OpenAbout()
		}
	})
	return stripButton(c, base).Children(func() {
		myui.Icon(c, icons.Chevron.SVG).Size(14, 14).TextColor(c.Theme().Text)
	})
}

// stripDivider is the line the strip draws between two tabs, and beside the
// button that opens one: a DIP wide and as high as a tab's mark, centered in
// the strip, out of the border color between two surfaces. Windows Terminal
// draws the same line in both places.
func stripDivider(c *myui.Context) myui.Element {
	return myui.Box(c).Width(1).Height(16).Background(c.Theme().Border)
}

// stripButton is one of the buttons beside the tabs: a square in the room of a
// glyph and a half, which takes the surface under the pointer, as the window's
// own buttons do, so that a mark in the strip reads as a button. Its mark takes
// the text color, as Windows Terminal's own + and ▾ do: at 14 DIP a muted mark
// reads as a disabled control, and the surface is enough to say it is a button.
func stripButton(c *myui.Context, el myui.Element) myui.Element {
	if el.Hovered() {
		return el.Background(c.Theme().SurfaceHover)
	}
	return el
}

// shortcut shows an action's key by a menu item, when the window knows one.
func shortcut(it *myui.MenuItem, st State, action string) {
	if st.Shortcut == nil {
		return
	}
	mods, key, ok := ParseKey(st.Shortcut(action))
	if !ok {
		return
	}
	it.Shortcut(mods, key)
}

// newTabButton opens a tab that runs the default profile.
func newTabButton(c *myui.Context, act Actions) myui.Element {
	base := myui.Box(c).
		Size(stripButtonSize, stripButtonSize).
// dividerSlot is one of the strip's lines -- the line itself when drawn is set,
// and the room it would take when it is not. The strip leaves a line out beside
// the tab in front, and lays itself out around it either way: a strip that moved
// as another tab came to the front would make the same press land on another
// button, or a tab change width under the pointer.
func dividerSlot(c *myui.Context, drawn bool, room float32) {
	if !drawn {
		myui.Box(c).Width(room).Height(16)
		return
	}
	stripDivider(c).Margin(0, (room-1)/2)
}

		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Label(locales.NewTab()).
		Tooltip(locales.NewTab()).
		OnClick(func() { act.NewTab("") })
	return stripButton(c, base).Children(func() {
		myui.Icon(c, icons.Add.SVG).Size(14, 14).TextColor(c.Theme().Text)
	})
}
