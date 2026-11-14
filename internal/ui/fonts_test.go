package ui

import (
	"slices"
	"strings"
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/locales"
)

// machine is what the tests say this machine has: two families a terminal can
// draw a grid with, one it cannot, and nothing else.
func machine() []InstalledFont {
	return []InstalledFont{
		{Name: "Cascadia Mono", Mono: true},
		{Name: "Consolas", Mono: true},
		{Name: "Arial", Mono: false},
	}
}

// TestFontNames covers the names a family list holds: the engine splits the
// list on commas and takes off the spaces and the quotes a name may wear, and
// the field holds whatever the person typed, so the page reads it the same way.
func TestFontNames(t *testing.T) {
	cases := []struct {
		list string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"Cascadia Mono", []string{"Cascadia Mono"}},
		{"Cascadia Mono,monospace", []string{"Cascadia Mono", "monospace"}},
		{" Cascadia Mono , monospace ", []string{"Cascadia Mono", "monospace"}},
		{`"Cascadia Mono", 'monospace'`, []string{"Cascadia Mono", "monospace"}},
		{"Cascadia Mono,,monospace,", []string{"Cascadia Mono", "monospace"}},
	}
	for _, tc := range cases {
		if got := fontNames(tc.list); !slices.Equal(got, tc.want) {
			t.Errorf("fontNames(%q) = %v, want %v", tc.list, got, tc.want)
		}
	}
}

// TestFontNoteSaysNothingWhenItDraws covers the quiet case: every name is one
// the machine has and can draw a grid with, so the field needs no line under it.
func TestFontNoteSaysNothingWhenItDraws(t *testing.T) {
	speakIn(t, "en")
	for _, list := range []string{"Cascadia Mono", "Cascadia Mono, monospace", "monospace", "'Consolas'"} {
		if text, danger := fontNote(list, machine()); text != "" || danger {
			t.Errorf("fontNote(%q) = %q, %v; want nothing", list, text, danger)
		}
	}
	// An empty list is the placeholder's business, and the window fills it in.
	if text, danger := fontNote("", machine()); text != "" || danger {
		t.Errorf("fontNote(\"\") = %q, %v; want nothing", text, danger)
	}
}

// TestFontNoteSkipsTheNamesTheMachineLacks covers a list written on another
// machine: the names this one does not have are skipped, which is worth a quiet
// line, and the first one it has is what the terminal draws with.
func TestFontNoteSkipsTheNamesTheMachineLacks(t *testing.T) {
	speakIn(t, "en")
	text, danger := fontNote("Monaco, no-such-font-at-all, Consolas", machine())
	if danger {
		t.Errorf("a list whose first usable family is monospace is not a problem: %q", text)
	}
	for _, name := range []string{"Monaco", "no-such-font-at-all"} {
		if !strings.Contains(text, name) {
			t.Errorf("the line %q does not say %q is missing", text, name)
		}
	}
	if strings.Contains(text, "Consolas") {
		t.Errorf("the line %q names a family the machine has", text)
	}
}

// TestFontNoteWarnsAboutAFirstFamilyThatIsNotMonospace is the case the line
// exists for: the terminal draws with the first family the machine has, and a
// proportional one spreads the letters out.
func TestFontNoteWarnsAboutAFirstFamilyThatIsNotMonospace(t *testing.T) {
	speakIn(t, "en")
	text, danger := fontNote("Monaco, Arial, Cascadia Mono", machine())
	if !danger {
		t.Fatalf("a first usable family that is not monospace is a problem: %q", text)
	}
	if !strings.Contains(text, "Arial") {
		t.Errorf("the line %q does not say which family it is", text)
	}
}

// TestFontNoteWarnsWhenTheMachineHasNone covers a list of names this machine
// does not have at all: the terminal draws with the generic family, which is
// what the line says.
func TestFontNoteWarnsWhenTheMachineHasNone(t *testing.T) {
	speakIn(t, "en")
	text, danger := fontNote("Monaco, no-such-font-at-all", machine())
	if !danger || text != locales.FontNoneUsable() {
		t.Fatalf("fontNote() = %q, %v; want %q", text, danger, locales.FontNoneUsable())
	}
}

// TestFontNoteSpeaksTheLanguage covers the line in the language the window
// speaks: the words are checked against the catalog, as a person reads them.
func TestFontNoteSpeaksTheLanguage(t *testing.T) {
	speakIn(t, "zh-Hans")
	text, _ := fontNote("Monaco, Arial", machine())
	if want := locales.FontWideFirst("Arial"); text != want {
		t.Fatalf("fontNote() = %q, want %q", text, want)
	}
	if !strings.Contains(text, "\u5b83\u4e0d\u662f\u7b49\u5bbd") {
		t.Errorf("the line is not the one the catalog holds: %q", text)
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

// TestFillFamily covers a submitted field that holds nothing: the generic
// family is what the terminal draws with then, and the file says it.
func TestFillFamily(t *testing.T) {
	for _, list := range []string{"", "   "} {
		got := list
		fillFamily(&got)
		if got != "monospace" {
			t.Errorf("fillFamily(%q) = %q, want %q", list, got, "monospace")
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
