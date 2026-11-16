package ui

import (
	"slices"
	"testing"
	"time"

	myui "github.com/egoist/mygo/ui"

	"wtgo/locales"
)

// TestFramesReadout covers the readout of the rate the window draws at: it shows
// the average, the lowest and the longest frame in the corner the panes leave
// alone, and nothing at all while it is closed. A readout placed outside the
// window draws nowhere, so where it lands is part of what this covers.
func TestFramesReadout(t *testing.T) {
	readout := &Frames{Average: 66.6, Lowest: 43.29, Longest: 23 * time.Millisecond}
	tt := myui.NewTester(func(c *myui.Context) {
		frames(c, State{Frames: readout})
	}, 900, 700)
	tt.Frame()

	for _, want := range []string{locales.FramesAverage(), "67 fps", locales.FramesLowest(), "43 fps", locales.FramesLongest(), "23.0 ms"} {
		if !slices.Contains(tt.Texts(), want) {
			t.Errorf("the readout does not show %q: %v", want, tt.Texts())
		}
	}

	at, ok := tt.Find(locales.FramesAverage())
	if !ok {
		t.Fatalf("the readout has no place in the window: %v", tt.Texts())
	}
	if at.X+at.W > 900 || at.Y+at.H > 700 {
		t.Errorf("the readout runs off the window at %+v, in a 900x700 one", at)
	}
	if at.X < 900/2 || at.Y < 700/2 {
		t.Errorf("the readout sits at %+v, want the bottom right corner", at)
	}

	closed := myui.NewTester(func(c *myui.Context) {
		frames(c, State{})
	}, 900, 700)
	closed.Frame()
	if texts := closed.Texts(); slices.Contains(texts, locales.FramesAverage()) {
		t.Errorf("a closed readout shows %q", texts)
	}
}
