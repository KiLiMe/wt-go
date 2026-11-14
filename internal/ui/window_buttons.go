package ui

import (
	myui "github.com/egoist/mygo/ui"
	"wtgo/locales"
)

// windowIcon is how wide the icon of a window button is: a square that fills
// the icon area of the desktop's own buttons.
const windowIcon float32 = 10

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

// windowGroupWidth is the room a group of n window's buttons takes.
func windowGroupWidth(n int) float32 {
	if n <= 0 {
		return 0
	}
	return float32(n)*windowButtonWidth + float32(n-1)*windowButtonGap + windowButtonRoom
}
