package ui

import (
	myui "github.com/egoist/mygo/ui"
	"wtgo/internal/icons"
	"wtgo/locales"
)

// Boundary: the buttons in the strip -- the two that open a tab, the menu of
// profiles beside them, the arrows that scroll the tabs, and the list of them
// all, with the measurements they share. The window's own buttons, and the room
// they take, are window_buttons.go.

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

// The room the buttons beside the tabs take: a button's square, the line beside
// it with the room it clears on both sides, and the gap between two children.
const (
	stripButtonSize    float32 = 28
	stripDividerGutter float32 = 4
	stripDividerRoom   float32 = 1 + 2*stripDividerGutter
	stripButtonsGap    float32 = 2
)

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

// allTabsButton lists every tab, which the strip shows only while it has more
// tabs than room: the arrow beside it moves the strip a tab at a time, and the
// list is where a tab is found by the name it has.
func allTabsButton(c *myui.Context, act Actions) myui.Element {
	base := myui.Box(c).
		Size(stripButtonSize, stripButtonSize).
		Radius(6).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Label(locales.AllTabs()).
		Tooltip(locales.AllTabs()).
		OnClick(func() { act.OpenTabs() })
	return stripButton(c, base).Children(func() {
		// The same chevron the menu of the profiles wears: a mark that opens a
		// list is the same mark wherever the list comes from, and the two are
		// on opposite sides of the button that opens a tab.
		myui.Icon(c, icons.Chevron.SVG).Size(14, 14).TextColor(c.Theme().Text)
	})
}

// newTabButton opens a tab that runs the default profile.
func newTabButton(c *myui.Context, act Actions) myui.Element {
	base := myui.Box(c).
		Size(stripButtonSize, stripButtonSize).
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
