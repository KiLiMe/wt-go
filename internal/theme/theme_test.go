package theme

import (
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
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

// TestUITheme covers a theme of the window's own: every theme the settings page
// offers is one, the appearance is none, and the window takes the scheme's
// background and text with the shades a window needs derived between them.
func TestUITheme(t *testing.T) {
	s := store.Default()
	for _, name := range UIThemes() {
		th, ok := Scheme(s, name)
		if !ok {
			t.Fatalf("%q is no scheme, so the page offers a theme that is none", name)
		}
		ui, ok := UITheme(s, name)
		if !ok {
			t.Fatalf("UITheme(s, %q) is no theme", name)
		}
		if ui.Background != th.Background || ui.Text != th.Foreground {
			t.Errorf("%q draws %v on %v, want %v on %v", name, ui.Text, ui.Background, th.Foreground, th.Background)
		}
		if want := luminance(th.Background) < 0.5; ui.Dark != want {
			t.Errorf("%q is dark = %v, want %v", name, ui.Dark, want)
		}
		if ui.Surface == ui.Background || ui.Border == ui.Background {
			t.Errorf("%q draws its surfaces and borders in its background", name)
		}
		if c := contrast(ui.AccentText, ui.Accent); c < 3 {
			t.Errorf("%q writes text on its accent at a contrast of %.1f", name, c)
		}
	}
	for _, name := range []string{"", "system", "light", "dark", "No Such Scheme"} {
		if _, ok := UITheme(s, name); ok {
			t.Errorf("UITheme(s, %q) is a theme", name)
		}
	}
}

// TestChrome covers the window's theme where the settings name none of their
// own: the terminal's colors, so that the tab in front carries the pane's
// background up into the strip, with shades derived between them.
func TestChrome(t *testing.T) {
	s := store.Default()
	for _, appearance := range []string{"", "system", "light", "dark"} {
		s.Theme = appearance
		th := Terminal(s)
		ui := Chrome(s)
		if ui == nil {
			t.Fatalf("Chrome with the %q appearance is no theme", appearance)
		}
		if ui.Background != th.Background || ui.Text != th.Foreground {
			t.Errorf("the %q appearance draws %v on %v, want the terminal's %v on %v",
				appearance, ui.Text, ui.Background, th.Foreground, th.Background)
		}
		if want := luminance(th.Background) < 0.5; ui.Dark != want {
			t.Errorf("the %q appearance is dark = %v, want %v", appearance, ui.Dark, want)
		}
		// The strip stands behind a tab that carries the background itself, so
		// the surfaces and the borders have to read as such.
		if ui.Surface == ui.Background || ui.Border == ui.Background || ui.Border == ui.Surface {
			t.Errorf("the %q appearance draws %v on %v with %v between them",
				appearance, ui.Surface, ui.Background, ui.Border)
		}
	}

	// A colorScheme of its own is what the terminal, and so the window, draws
	// with.
	s = store.Default()
	s.Theme = "system"
	s.ColorScheme = "Catppuccin Latte"
	th, ok := Scheme(s, s.ColorScheme)
	if !ok {
		t.Fatalf("%q is no scheme", s.ColorScheme)
	}
	if got := Chrome(s).Background; got != th.Background {
		t.Errorf("the window draws on %v, want the scheme's %v", got, th.Background)
	}
}

// TestUILightTheme covers the colors the terminal falls back to: those of a
// light theme of the settings' own, where they name one.
func TestUILightTheme(t *testing.T) {
	s := store.Default()
	s.ColorScheme = ""
	light := terminal.LightTheme().Background
	dark := terminal.DarkTheme().Background

	s.Theme = "Catppuccin Latte"
	if got := Terminal(s).Background; got != light {
		t.Errorf("a light theme's terminal draws on %v, want %v", got, light)
	}
	s.Theme = "Nord"
	if got := Terminal(s).Background; got != dark {
		t.Errorf("a dark theme's terminal draws on %v, want %v", got, dark)
	}
	s.Theme = "system"
	if got := Terminal(s).Background; got != dark {
		t.Errorf("the system's terminal draws on %v, want %v", got, dark)
	}

	// A scheme of the settings' own decides it by its own background.
	s.Schemes = []store.Scheme{{Name: "Paper", Background: "#fafafa"}}
	s.Theme = "Paper"
	if got := Terminal(s).Background; got != light {
		t.Errorf("a light scheme of its own draws on %v, want %v", got, light)
	}
}
