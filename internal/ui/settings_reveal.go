package ui

import (
	"time"

	myui "github.com/egoist/mygo/ui"
)

// RevealFade is how long the mark the search leaves on a setting lasts. Frames
// are not a clock: a window draws them as fast as it can, so a fade counted in
// them was over in a tenth of a second on a machine that draws quickly, before
// the eye -- still travelling down the page to the setting -- had arrived. The
// mark is the answer to the search, so it lasts long enough to be read.
const RevealFade = 700 * time.Millisecond

// Reveal is the setting the search found, and how far the mark over it has
// faded. The window holds it while the mark lasts: the field that draws the
// setting scrolls into view on the first frame, and the color over it fades
// from the time the search found it, so that what the search found is what the
// eye lands on without being left with a row lit for good.
type Reveal struct {
	// Key names the setting the way the settings file does, so that the field
	// that draws that setting is the one the mark is over: the search knows the
	// words a setting answers to, and the field knows which setting it is. ""
	// marks nothing, which is a setting the page holds rather than a row of it.
	Key string
	// At is when the search found the setting, which is where the fade starts.
	At time.Time
	// Frame counts the frames since the search found it, from 0.
	Frame int
}

// shade is how strong the mark is on this frame, and whether it shows at all.
func (r *Reveal) shade(key string, now time.Time) (float32, bool) {
	if r == nil || r.Key == "" || r.Key != key {
		return 0, false
	}
	left := RevealFade - now.Sub(r.At)
	if left <= 0 {
		return 0, false
	}
	return float32(left) / float32(RevealFade), true
}

// mark puts the search's mark over the field of a setting: the row scrolls into
// view on the first frame, and the color fades over the time after it. A field
// the reveal does not name is drawn as it is.
//
// It is the page that draws the mark rather than the app that scrolls to a box,
// because the box a setting has is only known once the page is laid out: the
// field is where both the setting and its place are, and it is the one thing
// that knows them together.
func mark(c *myui.Context, el myui.Element, key string, reveal *Reveal) myui.Element {
	shade, on := reveal.shade(key, c.Now())
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
