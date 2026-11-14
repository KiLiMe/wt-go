package ui

import (
	myui "github.com/egoist/mygo/ui"
	"wtgo/internal/icons"
)

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
