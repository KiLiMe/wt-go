package app

import (
	"math"
	"testing"
	"time"

	"wtgo/internal/ui"
)

// sameRate reports whether two rates are the same number, as far as a test
// cares: a rate is a division, and the last digit of one is not what it is
// about.
func sameRate(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

// TestRate covers what a run of frame times says about the rate: the average
// frame in it, the longest one, and the rates those two frames are. A time of
// nothing is no frame, and says nothing.
func TestRate(t *testing.T) {
	ms := time.Millisecond
	cases := []struct {
		name  string
		times []time.Duration
		want  *ui.Frames
	}{
		{"no frames", nil, nil},
		{"frames of no time", []time.Duration{0, -ms}, nil},
		{
			"frames of a sixtieth, one of them a little late",
			[]time.Duration{16 * ms, 17 * ms, 16 * ms, 17 * ms},
			&ui.Frames{Average: 1000 / 16.5, Lowest: 1000.0 / 17, Longest: 17 * ms, Count: 4},
		},
		{
			"a frame of no time is left out",
			[]time.Duration{0, 20 * ms},
			&ui.Frames{Average: 50, Lowest: 50, Longest: 20 * ms, Count: 1},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rate(tc.times)
			switch {
			case tc.want == nil:
				if got != nil {
					t.Fatalf("rate(%v) = %v, want nil", tc.times, got)
				}
			case got == nil:
				t.Fatalf("rate(%v) = nil, want %v", tc.times, tc.want)
			default:
				if !sameRate(got.Average, tc.want.Average) {
					t.Errorf("average = %v fps, want %v", got.Average, tc.want.Average)
				}
				if !sameRate(got.Lowest, tc.want.Lowest) {
					t.Errorf("lowest = %v fps, want %v", got.Lowest, tc.want.Lowest)
				}
				if got.Longest != tc.want.Longest {
					t.Errorf("longest = %v, want %v", got.Longest, tc.want.Longest)
				}
				if got.Count != tc.want.Count {
					t.Errorf("count = %d, want %d", got.Count, tc.want.Count)
				}
			}
		})
	}
}

// TestFrameMeter covers the meter over the frames the window draws: the first
// frame is no rate, the frames after it are, and the window keeps only the last
// few, so an old spike falls out of the numbers.
func TestFrameMeter(t *testing.T) {
	var m frameMeter
	at := time.Unix(0, 0)
	if got := m.frames(); got != nil {
		t.Fatalf("a meter with no frames says %v, want nil", got)
	}
	m.tick(at)
	if got := m.frames(); got != nil {
		t.Fatalf("one frame is no rate, said %v", got)
	}
	for i := 0; i < frameWindow+60; i++ {
		at = at.Add(10 * time.Millisecond)
		m.tick(at)
	}
	got := m.frames()
	if got == nil {
		t.Fatal("a meter of many frames says nothing")
	}
	if !sameRate(got.Average, 100) || !sameRate(got.Lowest, 100) || got.Longest != 10*time.Millisecond {
		t.Errorf("rate = %v, want 100 fps over 10 ms frames", got)
	}
	if got.Count != frameWindow {
		t.Errorf("count = %d, want the meter to keep %d frames", got.Count, frameWindow)
	}
}
