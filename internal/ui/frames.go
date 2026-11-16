package ui

import (
	"fmt"
	"time"

	myui "github.com/egoist/mygo/ui"

	"wtgo/locales"
)

// Frames is what the window's frame rate is over the frames it last drew: the
// average rate, the lowest of them, which is the spike, and the longest frame
// that made it. The readout shows it while it is open, and the window measures
// it from its own frames, so a frame the window did not draw in time is one the
// numbers show.
type Frames struct {
	// Average is the rate the frames made, in frames a second.
	Average float64
	// Lowest is the rate of the longest frame, which is how slow the window
	// got at its worst.
	Lowest float64
	// Longest is that frame, in time.
	Longest time.Duration
	// Count is how many frames the numbers are over.
	Count int
}

// The readout is small and sits in the corner the strips leave alone: wide
// enough for a label and a number at the sizes below, and as tall as its lines.
const (
	framesWidth  = 150
	framesHeight = 74
	framesMargin = 12
)

// frames draws the readout of the rate the window draws at: the average, the
// lowest and the longest frame. It is the window's own cost that it says, which
// is what a person looking at it is asking about.
func frames(c *myui.Context, st State) {
	f := st.Frames
	if f == nil {
		return
	}
	t := c.Theme()
	room := Content(c)
	left := max(room.W-framesWidth-framesMargin, 0)
	top := max(room.Y+room.H-framesHeight-framesMargin, 0)
	myui.Overlay(c, func() {
		myui.Column(c).
			Absolute().
			Left(left).
			Top(top).
			Width(framesWidth).
			Height(framesHeight).
			Padding(8, 10).
			Gap(2).
			Radius(8).
			Background(t.Surface).
			BorderWidth(1).
			BorderColor(t.Border).
			Children(func() {
				frameLine(c, locales.FramesAverage(), rateText(f.Average), t)
				frameLine(c, locales.FramesLowest(), rateText(f.Lowest), t)
				frameLine(c, locales.FramesLongest(), frameText(f.Longest), t)
			})
	})
}

// frameLine is one line of the readout: what it is, and what it is now.
func frameLine(c *myui.Context, what, value string, t *myui.Theme) {
	myui.Row(c).AlignItems(myui.Center).Children(func() {
		myui.Text(c, what).FontSize(11).TextColor(t.TextMuted).Grow(1)
		myui.Text(c, value).FontSize(11).TextColor(t.Text)
	})
}

// rateText is a rate as the readout writes it: whole frames a second, since a
// tenth of one is no difference a person can see.
func rateText(rate float64) string {
	if rate < 0 {
		rate = 0
	}
	return fmt.Sprintf("%.0f fps", rate)
}

// frameText is a frame as the readout writes it: milliseconds, the unit a
// person times a frame in.
func frameText(d time.Duration) string {
	return fmt.Sprintf("%.1f ms", float64(d)/float64(time.Millisecond))
}
