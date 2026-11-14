package ui

import (
	myui "github.com/egoist/mygo/ui"
	"sync"
)

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
