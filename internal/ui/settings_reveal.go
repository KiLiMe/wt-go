package ui

import (
	myui "github.com/egoist/mygo/ui"
)

// RevealFrames is how many frames the mark the search leaves on a setting
// shows: a page's own scroll can travel most of the window, and the mark has to
// still be there when the setting it names lands where it is read.
const RevealFrames = 12

// Reveal is the setting the search found, and how far the mark over it has
// faded. The window holds it while the mark lasts: the field that draws the
// setting scrolls into view on the first frame, and the color over it fades
// over the rest, so that what the search found is what the eye lands on without
// being left with a row lit for good.
type Reveal struct {
	// Key names the setting the way the settings file does, so that the field
	// that draws that setting is the one the mark is over: the search knows the
	// words a setting answers to, and the field knows which setting it is. ""
	// marks nothing, which is a setting the page holds rather than a row of it.
	Key string
	// Frame counts the frames since the search found the setting, from 0.
	Frame int
}

// shade is how strong the mark is on this frame, and whether it shows at all.
func (r *Reveal) shade(key string) (float32, bool) {
	if r == nil || r.Key == "" || r.Key != key || r.Frame > RevealFrames {
		return 0, false
	}
	return float32(RevealFrames-r.Frame) / RevealFrames, true
}

// mark puts the search's mark over the field of a setting: the row scrolls into
// view on the first frame, and the color fades over the frames after it. A field
// the reveal does not name is drawn as it is.
//
// It is the page that draws the mark rather than the app that scrolls to a box,
// because the box a setting has is only known once the page is laid out: the
// field is where both the setting and its place are, and it is the one thing
// that knows them together.
func mark(c *myui.Context, el myui.Element, key string, reveal *Reveal) myui.Element {
	shade, on := reveal.shade(key)
	if !on {
		return el
	}
	if reveal.Frame == 0 {
		// Scrolling is the first frame's work: the frames after it show the
		// mark where it has come to rest.
		el = el.ScrollIntoView()
	}
	// The fade needs frames the window would otherwise not draw, because
	// nothing else about the settings is moving.
	c.AnimationFrame()
	return el.Background(c.Theme().Accent.Alpha(0.18 * shade))
}
