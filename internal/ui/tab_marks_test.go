package ui

import (
	myui "github.com/egoist/mygo/ui"
	"image/color"
	"strings"
	"testing"
)

// TestMarkIcon covers the three kinds of mark a tab wears: a shape drawn as it
// is, a character in the square of the profile's color, and the plain terminal
// mark, with and without a color, in front and behind.
func TestMarkIcon(t *testing.T) {
	tt := myui.NewTester(func(c *myui.Context) {
		myui.Row(c).Children(func() {
			markIcon(c, "bash:arch", "#4eaa25", 16, true)
			markIcon(c, "\U0001F92C", "#3b78ff", 16, true)
			markIcon(c, "\U0001F92C", "", 16, false)
			markIcon(c, "", "#cccccc", 16, true)
			markIcon(c, "", "", 16, false)
		})
	}, 200, 24)
	tt.Frame()
	// The character is written rather than drawn, so the frame says it twice:
	// once on the profile's color and once on the quiet square.
	if got := strings.Count(strings.Join(tt.Texts(), ""), "\U0001F92C"); got != 2 {
		t.Fatalf("the character is drawn %d times, want 2: %q", got, tt.Texts())
	}
}

// TestMarkIconPreset paints a preset on its own: it is a picture, so the frame
// holds the tones it was drawn with, not the one flat color a glyph would be.
func TestMarkIconPreset(t *testing.T) {
	tt := myui.NewTester(func(c *myui.Context) {
		markIcon(c, "bash:ubuntu", "", 32, true)
	}, 32, 32)
	tt.Frame()
	tones := map[color.RGBA]struct{}{}
	img := tt.Image()
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			tones[color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}] = struct{}{}
		}
	}
	if len(tones) < 4 {
		t.Fatalf("the preset drew %d colors, want its own", len(tones))
	}
}
