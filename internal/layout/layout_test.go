package layout

import "testing"

func TestSplitSideBySide(t *testing.T) {
	got := Split(Rect{X: 10, Y: 20, W: 204, H: 100}, SideBySide, 0.5)
	want := Divide{
		First:  Rect{X: 10, Y: 20, W: 100, H: 100},
		Line:   Rect{X: 110, Y: 20, W: 4, H: 100},
		Second: Rect{X: 114, Y: 20, W: 100, H: 100},
	}
	if got != want {
		t.Fatalf("Split: got %+v, want %+v", got, want)
	}
}

func TestSplitStacked(t *testing.T) {
	got := Split(Rect{X: 0, Y: 0, W: 100, H: 204}, Stacked, 0.25)
	want := Divide{
		First:  Rect{X: 0, Y: 0, W: 100, H: 50},
		Line:   Rect{X: 0, Y: 50, W: 100, H: 4},
		Second: Rect{X: 0, Y: 54, W: 100, H: 150},
	}
	if got != want {
		t.Fatalf("Split: got %+v, want %+v", got, want)
	}
}

func TestSplitKeepsPanesAtLeastMin(t *testing.T) {
	r := Rect{W: 204, H: 204}
	for _, dir := range []Dir{SideBySide, Stacked} {
		for _, share := range []float32{-1, 0, 0.01, 0.99, 1, 2} {
			got := Split(r, dir, share)
			first, second := got.First.W, got.Second.W
			if dir == Stacked {
				first, second = got.First.H, got.Second.H
			}
			if first < Min || second < Min {
				t.Fatalf("Split %v share %v: panes %v and %v are narrower than %v", dir, share, first, second, Min)
			}
			if first+second+Divider != r.W {
				t.Fatalf("Split %v share %v: panes and line take %v, want %v", dir, share, first+second+Divider, r.W)
			}
		}
	}
}

func TestSplitTooSmallForTwoPanes(t *testing.T) {
	got := Split(Rect{W: 50, H: 30}, SideBySide, 0.9)
	if got.First.W != 23 || got.Second.W != 23 {
		t.Fatalf("Split: panes %v and %v, want half of the room each", got.First.W, got.Second.W)
	}
}

func TestRoom(t *testing.T) {
	r := Rect{W: 204, H: 100}
	if got := Room(r, SideBySide); got != 200 {
		t.Fatalf("Room side by side: got %v, want 200", got)
	}
	if got := Room(r, Stacked); got != 96 {
		t.Fatalf("Room stacked: got %v, want 96", got)
	}
	if got := Room(Rect{W: 2}, SideBySide); got != 0 {
		t.Fatalf("Room of no room: got %v, want 0", got)
	}
}

func TestShare(t *testing.T) {
	r := Rect{W: 204}
	if got := Share(r, SideBySide, 0.5, 50); got != 0.75 {
		t.Fatalf("Share moved 50: got %v, want 0.75", got)
	}
	if got := Share(r, SideBySide, 0.5, -500); got != Min/200 {
		t.Fatalf("Share moved far back: got %v, want %v", got, Min/200)
	}
	if got := Share(Rect{W: 50}, SideBySide, 0.5, 10); got != 0.5 {
		t.Fatalf("Share of no room: got %v, want the share it was", got)
	}
}
