package ui

import (
	myui "github.com/egoist/mygo/ui"
)

// Boundary: the composition of the strip -- the title bar, the room the tabs
// have and how it is shared, and the scrolling of the tabs past it. One tab is
// drawn in tabs_item.go, the buttons in tabs_buttons.go, and every DIP the
// strip measures with is in tabs_geometry.go.

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
		ContextMenu(func(m *myui.Menu) { titleBarMenu(m, st, act) }).
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
			// The tabs past the edge are reached by scrolling or by name: the
			// list names them, which is the way to a tab whose place in the
			// strip a person does not know.
			dividerSlot(c, true, stripDividerRoom)
			allTabsButton(c, act)
		}
		// The buttons that open a tab sit past the tabs rather than in them:
		// however many tabs there are, a tab is a press away rather than a
		// scroll to the end away.
		stripButtons(c, st, act, !scrolls)
	})
}

// tabsRoom is the room the tabs have: the window, less the strip's own padding,
// the room that moves the window and the window's buttons, the buttons beside
// the tabs and, when they show, the buttons that move the tabs and the one that
// lists every tab.
func tabsRoom(c *myui.Context, st State, scrolls bool) float32 {
	w, _ := c.Size()
	bar := c.TitleBar()
	room := w - bar.Left - stripGutter - bar.Right - dragRoom - stripButtonsWidth()
	if b := st.WindowButtons; b != nil {
		room -= windowGroupWidth(len(b.Start)) + windowGroupWidth(len(b.End))
	}
	if scrolls {
		room -= 2*stepWidth + stripDividerRoom + stripButtonSize
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

// tabDrag is a tab on its way to another place in the strip.
type tabDrag struct {
	from int
}

// tabMove is a tab dropped on another place: the tab at from belongs at to.
type tabMove struct {
	from, to int
}
