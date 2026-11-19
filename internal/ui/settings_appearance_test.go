package ui

import (
	"fmt"
	myui "github.com/egoist/mygo/ui"
	"strings"
	"testing"
	"wtgo/internal/store"
	"wtgo/locales"
)

// TestAppearancePreviewKeepsTheHead covers the preview the appearance page
// draws: the setting that changes it is often just under it, so the preview
// keeps the head of the page while the fields scroll past, rather than leaving
// a person to scroll up after every change.
func TestAppearancePreviewKeepsTheHead(t *testing.T) {
	s := store.Default()
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageAppearance}, &navActions{})
	}, 900, 700)
	tt.Frame()
	line, ok := tt.Find(sampleLines()[1])
	if !ok {
		t.Fatalf("the appearance page draws no preview: %q", tt.Texts())
	}
	size, ok := tt.Find(locales.FontSize())
	if !ok {
		t.Fatalf("the appearance page lists no font size: %q", tt.Texts())
	}
	// The wheel turns over the fields, which is where a person's pointer is.
	tt.Scroll(600, 400, 0, 400)
	tt.Frame()
	held, ok := tt.Find(sampleLines()[1])
	if !ok {
		t.Fatalf("the preview is gone once the page scrolls: %q", tt.Texts())
	}
	if held.Y < 0 {
		t.Fatalf("the preview line at %v is above the page it should hold", held.Y)
	}
	if held.Y >= line.Y {
		t.Fatalf("the preview line went from %v to %v: it scrolled with the fields", line.Y, held.Y)
	}
	// The fields under it did scroll, so what held the preview is not the page
	// having nothing to scroll.
	if moved, ok := tt.Find(locales.FontSize()); ok && moved.Y >= size.Y {
		t.Fatalf("the font size went from %v to %v: the page did not scroll", size.Y, moved.Y)
	}
	// It paints above the fields that roll under it, and the pointer finds it
	// there rather than the field behind it.
	if err := tt.Click(sampleLines()[1]); err != nil {
		t.Fatalf("the preview cannot be clicked where it holds the head: %v", err)
	}
}

// TestFontChoicesNameTheGeneric covers the last choice of the font setting's
// menu: it stands for the generic family, and is named as the alias it is rather
// than as a family this machine is supposed to have.
// TestWeightOptions covers the weights the page offers: the nine a family may
// have, each shown as its number and the face it stands for, of the language
// the window speaks.
func TestWeightOptions(t *testing.T) {
	opts := weightOptions()
	if len(opts) != 9 {
		t.Fatalf("the page offers %d weights, want 9", len(opts))
	}
	for i, o := range opts {
		want := fmt.Sprintf("%d", 100*(i+1))
		if o.Value != want {
			t.Errorf("choice %d stands for %q, want %q", i, o.Value, want)
		}
		if o.Label == want || !strings.Contains(o.Label, want) {
			t.Errorf("choice %d is labelled %q, want the number and the face", i, o.Label)
		}
	}
}

// TestWeightIsKept covers the page's weight setting: a weight the file names
// but no choice does is drawn as the number it is, and stays that number
// instead of being read back as a face the terminal would not draw with.
func TestWeightIsKept(t *testing.T) {
	for _, in := range []int{400, 350} {
		s := store.Default()
		s.Font.Weight = in
		tt := myui.NewTester(func(c *myui.Context) {
			fieldBlock(c, "font.weight", locales.FontWeight(), locales.FontWeightNote(), func() {
				choiceOf(c, &s.Font.Weight, weightOptions())
			}, nil)
		}, 600, 200)
		tt.Frame()
		want := in
		if s.Font.Weight != want {
			t.Errorf("weight %d became %d, want %d", in, s.Font.Weight, want)
		}
		if _, ok := tt.Find(locales.FontWeight()); !ok {
			t.Error("the weight setting has no label")
		}
	}
}

func TestFontChoicesNameTheGeneric(t *testing.T) {
	opts := fontChoices([]string{"Cascadia Mono", "Consolas"})
	if len(opts) != 3 {
		t.Fatalf("the font setting offers %d choices, want 3", len(opts))
	}
	last := opts[len(opts)-1]
	if last.Value != "monospace" {
		t.Fatalf("the last choice stands for %q, want \"monospace\"", last.Value)
	}
	if last.Label == "monospace" {
		t.Fatalf("the generic family is named as a family: %q", last.Label)
	}
	if want := (option{"Cascadia Mono", "Cascadia Mono"}); opts[0] != want {
		t.Fatalf("the first choice is %+v, want %+v", opts[0], want)
	}
}
