package theme

import (
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
)

// TestColor covers the colors a scheme may write.
func TestColor(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want myui.Color
		ok   bool
	}{
		{"#0c0c0c", myui.RGB(0x0c, 0x0c, 0x0c), true},
		{"  #CCC  ", myui.RGB(0xcc, 0xcc, 0xcc), true},
		{"#264f78ff", myui.Hex("#264f78ff"), true},
		{"", myui.Color{}, false},
		{"0c0c0c", myui.Color{}, false},
		{"#12345", myui.Color{}, false},
		{"#gggggg", myui.Color{}, false},
	} {
		got, ok := Color(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("Color(%q) = %v, %v, want %v, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

// TestScheme covers the order a name resolves in: the settings' own schemes,
// then the built in ones, then Ghostty's.
func TestScheme(t *testing.T) {
	s := store.Default()
	s.Theme = "dark"

	campbell, ok := Scheme(s, "Campbell")
	if !ok {
		t.Fatal(`Scheme(s, "Campbell") is no scheme`)
	}
	if want := myui.Hex("#0c0c0c"); campbell.Background != want {
		t.Errorf("Campbell background = %v, want %v", campbell.Background, want)
	}
	if want := myui.Hex("#c50f1f"); campbell.Palette[1] != want {
		t.Errorf("Campbell red = %v, want %v", campbell.Palette[1], want)
	}

	if _, ok := Scheme(s, "One Half Dark"); !ok {
		t.Error(`Scheme(s, "One Half Dark") is no scheme`)
	}
	if _, ok := Scheme(s, "No Such Scheme"); ok {
		t.Error(`Scheme(s, "No Such Scheme") is a scheme`)
	}
	if _, ok := Scheme(s, ""); ok {
		t.Error(`Scheme(s, "") is a scheme`)
	}

	// A scheme the settings write wins over a built in one of the same name,
	// and over Ghostty's.
	s.Schemes = []store.Scheme{{
		Name:       "Campbell",
		Background: "#101010",
		Foreground: "#eeeeee",
		Palette:    []string{"#000000", "#ff0000"},
	}}
	own, ok := Scheme(s, "Campbell")
	if !ok {
		t.Fatal(`Scheme(s, "Campbell") is no scheme`)
	}
	if want := myui.Hex("#101010"); own.Background != want {
		t.Errorf("own Campbell background = %v, want %v", own.Background, want)
	}
	if want := myui.Hex("#ff0000"); own.Palette[1] != want {
		t.Errorf("own Campbell red = %v, want %v", own.Palette[1], want)
	}
	// The colors the scheme does not write stay the ones of the window's
	// theme.
	if want := myui.Hex("#13a10e"); own.Palette[2] != want {
		t.Errorf("own Campbell green = %v, want %v", own.Palette[2], want)
	}
}

// TestTerminal covers the colors the terminal draws with: the scheme, or the
// window's theme for a name that is no scheme.
func TestTerminal(t *testing.T) {
	s := store.Default()
	if got, want := Terminal(s).Background, myui.Hex("#0c0c0c"); got != want {
		t.Errorf("Terminal(dark).Background = %v, want %v", got, want)
	}

	s.ColorScheme = "No Such Scheme"
	s.Theme = "light"
	if got, want := Terminal(s).Background, myui.Hex("#ffffff"); got != want {
		t.Errorf("Terminal(light).Background = %v, want %v", got, want)
	}
}

// TestHex covers a color as a window's background.
func TestHex(t *testing.T) {
	if got, want := Hex(myui.RGB(0x0c, 0x0c, 0x0c)), "#0c0c0c"; got != want {
		t.Errorf("Hex = %q, want %q", got, want)
	}
}
