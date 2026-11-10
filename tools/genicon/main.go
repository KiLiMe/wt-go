// Command genicon draws the app icon: a rounded dark square with a green
// chevron and a cursor bar, at the sizes Windows, macOS and Linux ask for.
// The PNG it writes is what `go tool mygo build` embeds as the app icon.
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

const size = 1024

func main() {
	out := filepath.Join("resources", "icon.png")
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	img := draw(size)
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		fail(err)
	}
	f, err := os.Create(out)
	if err != nil {
		fail(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fail(err)
	}
}

func fail(err error) {
	os.Stderr.WriteString("genicon: " + err.Error() + "\n")
	os.Exit(1)
}

// draw renders the icon at n pixels square, with four samples per axis: the
// shapes are few and the edges are the whole design.
func draw(n int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	scale := float64(n) / size
	const samples = 4
	for y := range n {
		for x := range n {
			var r, g, b, a float64
			for sy := range samples {
				for sx := range samples {
					px := (float64(x) + (float64(sx)+0.5)/samples) / scale
					py := (float64(y) + (float64(sy)+0.5)/samples) / scale
					cr, cg, cb, ca := at(px, py)
					r, g, b, a = r+cr*ca, g+cg*ca, b+cb*ca, a+ca
				}
			}
			if a == 0 {
				continue
			}
			n4 := float64(samples * samples)
			img.SetNRGBA(x, y, color.NRGBA{
				R: clamp8(r / a),
				G: clamp8(g / a),
				B: clamp8(b / a),
				A: clamp8(a / n4 * 255),
			})
		}
	}
	return img
}

// at is the color at a point of the 1024 square, its alpha the coverage.
func at(x, y float64) (float64, float64, float64, float64) {
	const (
		inset  = 24.0
		radius = 200.0
	)
	// The plate: a rounded square, its color a slight top-to-bottom gradient
	// so the icon reads as a surface rather than a flat sticker.
	in := sdRoundRect(x, y, 1024-2*inset, 1024-2*inset, radius) - inset
	cover := coverage(in)
	if cover <= 0 {
		return 0, 0, 0, 0
	}
	t := clamp01(y / 1024)
	bgR := lerp(0x23, 0x0c, t)
	bgG := lerp(0x26, 0x0d, t)
	bgB := lerp(0x2e, 0x12, t)

	// The prompt: a chevron in the terminal's own green, and a cursor bar of
	// near white beside it.
	green := [3]float64{0x16, 0xc6, 0x0c}
	white := [3]float64{0xe8, 0xe8, 0xe8}
	const stroke = 44.0
	chevron := math.Min(
		sdSegment(x, y, 300, 340, 560, 512)-stroke,
		sdSegment(x, y, 560, 512, 300, 684)-stroke,
	)
	bar := sdRoundRect(x-694, y-664, 186, 68, 34)

	r, g, b := bgR, bgG, bgB
	if c := coverage(chevron); c > 0 {
		r, g, b = mix(r, g, b, green[0], green[1], green[2], c)
	}
	if c := coverage(bar); c > 0 {
		r, g, b = mix(r, g, b, white[0], white[1], white[2], c)
	}
	return r, g, b, cover
}

// sdRoundRect is the signed distance to a w x h rounded rectangle whose top
// left corner is (0, 0).
func sdRoundRect(x, y, w, h, r float64) float64 {
	dx := math.Abs(x-w/2) - (w/2 - r)
	dy := math.Abs(y-h/2) - (h/2 - r)
	ox, oy := math.Max(dx, 0), math.Max(dy, 0)
	return math.Sqrt(ox*ox+oy*oy) + math.Min(math.Max(dx, dy), 0) - r
}

// sdSegment is the distance to the segment (x0, y0)-(x1, y1).
func sdSegment(x, y, x0, y0, x1, y1 float64) float64 {
	ex, ey := x1-x0, y1-y0
	px, py := x-x0, y-y0
	t := clamp01((px*ex + py*ey) / (ex*ex + ey*ey))
	dx, dy := px-t*ex, py-t*ey
	return math.Sqrt(dx*dx + dy*dy)
}

// coverage turns a signed distance into how much of the pixel the shape
// covers: no more than one pixel of feathering.
func coverage(d float64) float64 { return clamp01(0.5 - d) }

func mix(r0, g0, b0, r1, g1, b1, t float64) (float64, float64, float64) {
	return r0 + (r1-r0)*t, g0 + (g1-g0)*t, b0 + (b1-b0)*t
}

func lerp(a, b byte, t float64) float64 { return float64(a) + (float64(b)-float64(a))*t }

func clamp01(v float64) float64 { return math.Min(1, math.Max(0, v)) }

func clamp8(v float64) uint8 { return uint8(math.Min(255, math.Max(0, v+0.5))) }
