package app

import (
	"time"

	"wtgo/internal/ui"
)

// frameWindow is how many frames the rate the readout shows is taken over: four
// seconds at the rate a window runs at, long enough that one slow frame still
// shows, and short enough that the number follows what the window is doing now.
const frameWindow = 240

// frameMeter times the frames the window draws, for the readout that says what
// the rate is. A frame is drawn when the window asks for one, so the time
// between two of them is the frame the window took: the longest of them is the
// spike, and the frames it made are the rate.
type frameMeter struct {
	last  time.Time
	times [frameWindow]time.Duration
	next  int
	count int
}

// tick takes the frame drawn at now, and times it against the one before.
func (m *frameMeter) tick(now time.Time) {
	if !m.last.IsZero() {
		m.times[m.next] = now.Sub(m.last)
		m.next = (m.next + 1) % frameWindow
		if m.count < frameWindow {
			m.count++
		}
	}
	m.last = now
}

// frames is the readout's numbers, nil until a frame has been timed: one frame
// alone is no rate.
func (m *frameMeter) frames() *ui.Frames {
	if m.count == 0 {
		return nil
	}
	return rate(m.times[:m.count])
}

// rate is what a run of frame times says about the rate: the average frame in
// it, the longest one, and the rates those two frames are. A time of nothing or
// less is no frame, and is left out.
func rate(times []time.Duration) *ui.Frames {
	var sum, longest time.Duration
	var n int
	for _, d := range times {
		if d <= 0 {
			continue
		}
		n++
		sum += d
		if d > longest {
			longest = d
		}
	}
	if n == 0 {
		return nil
	}
	average := sum / time.Duration(n)
	return &ui.Frames{
		Average: seconds(average),
		Lowest:  seconds(longest),
		Longest: longest,
		Count:   n,
	}
}

// seconds is how many frames a second a frame of that time is.
func seconds(d time.Duration) float64 {
	if d <= 0 {
		return 0
	}
	return float64(time.Second) / float64(d)
}
