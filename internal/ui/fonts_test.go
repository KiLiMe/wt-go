package ui

import (
	"slices"
	"strings"
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
	"wtgo/locales"
)

func TestFontNote(t *testing.T) {
	speakIn(t, "en")
	for _, tc := range []struct {
		list   string
		font   FontDiagnostic
		want   string
		danger bool
	}{
		{list: "Grid"},
		{font: FontDiagnostic{Primary: "monospace", SystemFallback: true}},
		{list: "Grid", font: FontDiagnostic{Primary: "Grid"}},
		{
			list: "Missing, Grid", font: FontDiagnostic{Primary: "Grid", Missing: []string{"Missing"}},
			want: locales.FontMissing("Missing", "Grid"),
		},
		{
			list: "Wide, Grid", font: FontDiagnostic{Primary: "Grid", Proportional: "Wide"},
			want: locales.FontWideFirst("Wide", "Grid"), danger: true,
		},
		{
			list: "Missing", font: FontDiagnostic{Primary: "monospace", SystemFallback: true},
			want: locales.FontNoneUsable(), danger: true,
		},
	} {
		if text, danger := fontNote(tc.list, tc.font); text != tc.want || danger != tc.danger {
			t.Errorf("fontNote(%q, %+v) = %q, %v; want %q, %v", tc.list, tc.font, text, danger, tc.want, tc.danger)
		}
	}
}

func TestFontNoteSpeaksTheLanguage(t *testing.T) {
	speakIn(t, "zh-Hans")
	text, danger := fontNote("Arial, Consolas", FontDiagnostic{Primary: "Consolas", Proportional: "Arial"})
	if !danger || !strings.Contains(text, "比例字体") || !strings.Contains(text, "Consolas") {
		t.Errorf("fontNote() = %q, %v; want a Chinese warning naming the actual fallback", text, danger)
	}
}

// TestPickFamily covers the menu's choice: the family it names goes to the
// front of the list, where the terminal looks first, and a list already holding
// it moves the copy rather than adding one.
func TestPickFamily(t *testing.T) {
	cases := []struct {
		list string
		name string
		want string
	}{
		{"", "Cascadia Mono", "Cascadia Mono"},
		{"Consolas", "Cascadia Mono", "Cascadia Mono, Consolas"},
		{"Consolas, Cascadia Mono", "Cascadia Mono", "Cascadia Mono, Consolas"},
		{"consolas, Cascadia Mono", "Consolas", "Consolas, Cascadia Mono"},
		{"monospace", "monospace", "monospace"},
	}
	for _, tc := range cases {
		list := tc.list
		pickFamily(&list, tc.name)
		if list != tc.want {
			t.Errorf("pickFamily(%q, %q) = %q, want %q", tc.list, tc.name, list, tc.want)
		}
	}
}

// TestFillFamily covers a submitted field that holds nothing: the default
// families are what the terminal draws with then, and the file says them.
func TestFillFamily(t *testing.T) {
	for _, list := range []string{"", "   "} {
		got := list
		fillFamily(&got)
		if got != store.DefaultFontFamily() {
			t.Errorf("fillFamily(%q) = %q, want %q", list, got, store.DefaultFontFamily())
		}
	}
	if got := "Consolas"; func() string { fillFamily(&got); return got }() != "Consolas" {
		t.Error("a list that names a family was emptied")
	}
}

// TestFontChoices covers the menu: the families the machine can draw with, in
// the order they are listed, then the generic name, which is not a family of
// the machine and is called as the alias it is.
func TestFontChoices(t *testing.T) {
	opts := fontChoices([]string{"Cascadia Mono", "Consolas"})
	if len(opts) != 3 {
		t.Fatalf("the menu offers %d choices, want 3", len(opts))
	}
	if want := (option{"Cascadia Mono", "Cascadia Mono"}); opts[0] != want {
		t.Errorf("the first choice is %+v, want %+v", opts[0], want)
	}
	if last := opts[2]; last.Value != "monospace" || last.Label == last.Value {
		t.Errorf("the generic choice is %+v, want the alias named as one", last)
	}
}

// TestFontPickerOpensTheMenu covers the field beside the menu: a press on the
// plus lists the families this machine has, and a choice writes the list.
func TestFontPickerOpensTheMenu(t *testing.T) {
	list := "Consolas"
	tt := myui.NewTester(func(c *myui.Context) {
		fontPicker(c, &list, []string{"Cascadia Mono", "Consolas"})
	}, 420, 40)
	tt.Frame()
	if err := tt.Click("\uff0b"); err != nil {
		t.Fatalf("the menu did not open: %v", err)
	}
	menu := tt.Menu()
	for _, want := range []string{"Cascadia Mono", "Consolas", locales.FontGeneric()} {
		if !slices.Contains(menu, want) {
			t.Errorf("the menu is %q, want %q in it", menu, want)
		}
	}
	if err := tt.ChooseMenuItem("Cascadia Mono"); err != nil {
		t.Fatalf("the choice did not take: %v", err)
	}
	tt.Frame()
	if list != "Cascadia Mono, Consolas" {
		t.Errorf("the list is %q, want the chosen family first", list)
	}
}
