package ui

import (
	"image"
	"image/color"
	"slices"
	"strings"
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
	"wtgo/internal/theme"
	"wtgo/locales"
)

// schemeActions is the window as the color schemes page sees it: it records the
// schemes the page asks to remove, and the ones it asks to pick.
type schemeActions struct {
	Actions
	removed []string
	picked  []string
}

func (a *schemeActions) RemoveScheme(name string) { a.removed = append(a.removed, name) }
func (a *schemeActions) PickScheme(name string)   { a.picked = append(a.picked, name) }

// schemesOf draws the color schemes of the appearance page over the settings
// and schemes it gives.
func schemesOf(t *testing.T, s *store.Settings, schemes []Scheme) (*myui.Tester, *schemeActions) {
	t.Helper()
	return schemesOfFilter(t, s, schemes, nil)
}

// schemesOfFilter draws the color schemes of the appearance page over the
// settings and schemes it gives, with filter as the words the search box edits
// and writes back into. The page is the appearance page, since the schemes are
// its last section: it is drawn tall enough to show that section without
// scrolling.
func schemesOfFilter(t *testing.T, s *store.Settings, schemes []Scheme, filter *string) (*myui.Tester, *schemeActions) {
	t.Helper()
	act := &schemeActions{}
	tt := myui.NewTester(func(c *myui.Context) {
		schemePicker(c, s, Settings{Draft: s, Schemes: schemes, Filter: filter}, act)
	}, 900, 700)
	tt.Frame()
	return tt, act
}

// TestSchemeRowRemoves covers the button a scheme the settings write carries:
// it shows while the pointer is over its row, and a click on it asks to remove
// that scheme alone, leaving the pick where it was.
func TestSchemeRowRemoves(t *testing.T) {
	s := store.Default()
	s.ColorScheme = "Campbell"
	s.Schemes = []store.Scheme{{
		Name:       "Paper",
		Background: "#fafafa",
		Foreground: "#111111",
		Palette:    []string{"#000000"},
	}}
	schemes := []Scheme{
		{Name: "Paper", Custom: true},
		{Name: "Campbell"},
	}
	tt, act := schemesOf(t, &s, schemes)

	label := locales.SchemeRemove()
	if _, ok := tt.Find(label); ok {
		t.Fatalf("the button shows while the pointer is elsewhere: %q", tt.Texts())
	}

	name, ok := tt.Find("Paper")
	if !ok {
		t.Fatalf("the list does not show the scheme: %q", tt.Texts())
	}
	tt.Move(name.X+name.W/2, name.Y+name.H/2)
	tt.Frame()
	if err := tt.Click(label); err != nil {
		t.Fatalf("the row under the pointer offers no way to remove it: %v", err)
	}
	if want := []string{"Paper"}; !slices.Equal(act.removed, want) {
		t.Errorf("removed %v, want %v", act.removed, want)
	}
	if s.ColorScheme != "Campbell" {
		t.Errorf("scheme = %q, want the pick left at %q", s.ColorScheme, "Campbell")
	}
	if len(act.picked) != 0 {
		t.Errorf("picked %v, want none: the button removes, it does not pick", act.picked)
	}
}

// TestSchemeSearchKeepsWhatIsTyped covers the search box: what is typed into
// it goes into the string the page is handed, and the frame drawn after that
// still finds it there, both in the box and in the list it narrows.
func TestSchemeSearchKeepsWhatIsTyped(t *testing.T) {
	s := store.Default()
	s.ColorScheme = "Campbell"
	filter := ""
	tt, _ := schemesOfFilter(t, &s, []Scheme{{Name: "Paper"}, {Name: "Campbell"}}, &filter)
	// The pick above the list shows the chosen scheme too, so counting the
	// name is what tells the list apart: it shows twice before the search.
	count := func() int { return countTexts(tt.Texts(), "Campbell") }
	if got := count(); got != 2 {
		t.Fatalf("%q shows %d times, want the pick and the list: %q", "Campbell", got, tt.Texts())
	}

	// A text input shows nothing to find by while it is empty, so it is
	// reached where its field is drawn.
	box, ok := tt.AccessByRole(myui.RoleTextField, "")
	if !ok {
		t.Fatalf("the page shows no search box: %q", tt.Texts())
	}
	x, y := box.Bounds.X+box.Bounds.W/2, box.Bounds.Y+box.Bounds.H/2
	tt.Press(x, y)
	tt.Release(x, y)
	tt.Frame()
	if box, ok = tt.AccessByRole(myui.RoleTextField, ""); !ok || !box.Focused {
		t.Fatalf("a click at (%v, %v) leaves the search box without the keys", x, y)
	}
	tt.Type("Pap")
	if filter != "Pap" {
		t.Fatalf("filter = %q, want what was typed %q", filter, "Pap")
	}

	// A page that gave the box a string of its own every frame would show
	// the words for the frame they were typed in and lose them after that,
	// leaving the scheme the search does not ask for in the list.
	tt.Frame()
	tt.Frame()
	if filter != "Pap" {
		t.Errorf("filter = %q, want the next frames kept what was typed %q", filter, "Pap")
	}
	if got := count(); got != 1 {
		t.Errorf("%q shows %d times, want only the pick: the list was not narrowed: %q", "Campbell", got, tt.Texts())
	}
	// The value of a text field is what it holds, not its name: a field's
	// name is its label, and this one is labeled by nothing on screen.
	box, ok = tt.AccessByRole(myui.RoleTextField, "")
	if !ok || !strings.Contains(box.Value, "Pap") {
		t.Errorf("the search box reads %q, want the typed %q", box.Value, "Pap")
	}
}

// TestSchemePicksOnlyFromTheList covers what the page offers to pick a scheme
// with: the list alone. A control above it that picked too offered the same
// hundreds of names a second time, without the search that makes them findable.
func TestSchemePicksOnlyFromTheList(t *testing.T) {
	s := store.Default()
	s.ColorScheme = "Campbell"
	filter := ""
	tt, _ := schemesOfFilter(t, &s, []Scheme{{Name: "Paper"}, {Name: "Campbell"}}, &filter)
	if _, ok := tt.AccessByRole(myui.RolePopUpButton, ""); ok {
		t.Errorf("the page offers a second way to pick a scheme: %q", tt.Texts())
	}
	if !tt.HasText("Paper") {
		t.Errorf("the list is not there to pick from: %q", tt.Texts())
	}
}

// TestSchemeSearchEnterPicks covers the keys that reach the list: a person who
// has typed a name has found the scheme the search puts first, and Enter takes
// it, so that picking a scheme is not a thing the pointer alone does.
func TestSchemeSearchEnterPicks(t *testing.T) {
	s := store.Default()
	s.ColorScheme = "Campbell"
	filter := ""
	tt, act := schemesOfFilter(t, &s, []Scheme{{Name: "Paper"}, {Name: "Campbell"}}, &filter)

	box, ok := tt.AccessByRole(myui.RoleTextField, "")
	if !ok {
		t.Fatalf("the page shows no search box: %q", tt.Texts())
	}
	x, y := box.Bounds.X+box.Bounds.W/2, box.Bounds.Y+box.Bounds.H/2
	tt.Press(x, y)
	tt.Release(x, y)
	tt.Frame()
	tt.Type("Pap")
	tt.Frame()
	if got := len(act.picked); got != 0 {
		t.Fatalf("picked %v while typing, want none", act.picked)
	}

	tt.Key(0, myui.KeyEnter)
	if want := []string{"Paper"}; !slices.Equal(act.picked, want) {
		t.Errorf("Enter picked %v, want %v", act.picked, want)
	}
	if s.ColorScheme != "Campbell" {
		t.Errorf("scheme = %q, want the draft left as it was: %q", s.ColorScheme, "Campbell")
	}
}

// TestSwatchOf covers the colors the page shows a scheme by: the ones the list
// carries, and the window's own muted pair for a name the list has none of.
func TestSwatchOf(t *testing.T) {
	bg, fg := myui.RGB(1, 2, 3), myui.RGB(4, 5, 6)
	schemes := []Scheme{{Name: "Paper", Bg: bg, Fg: fg}}
	if gotBg, gotFg := swatchOf(schemes, "Paper", myui.DarkTheme()); gotBg != bg || gotFg != fg {
		t.Errorf("swatchOf(Paper) = %v, %v, want the scheme's own %v, %v", gotBg, gotFg, bg, fg)
	}
	th := myui.DarkTheme()
	if gotBg, gotFg := swatchOf(schemes, "None", th); gotBg != th.Surface || gotFg != th.TextMuted {
		t.Errorf("swatchOf(None) = %v, %v, want the window's muted pair", gotBg, gotFg)
	}
}

// countTexts is how many of the texts are s.
func countTexts(texts []string, s string) int {
	n := 0
	for _, text := range texts {
		if text == s {
			n++
		}
	}
	return n
}

// TestSchemesWithoutFilterDraws covers a page drawn with no string to search
// with: the box would have nowhere to write what is typed, so it is left out,
// and the page still draws and still picks schemes.
func TestSchemesWithoutFilterDraws(t *testing.T) {
	s := store.Default()
	s.ColorScheme = "Campbell"
	tt, act := schemesOf(t, &s, []Scheme{{Name: "Paper"}, {Name: "Campbell"}})

	// The search box can only be there if it has a string to edit: one made
	// up for the frame would take what is typed down with it.
	if _, ok := tt.AccessByRole(myui.RoleTextField, ""); ok {
		t.Error("the page offers a search box with no string to search with")
	}
	name, ok := tt.Find("Paper")
	if !ok {
		t.Fatalf("the list does not show the scheme: %q", tt.Texts())
	}
	tt.ClickAt(name.X+name.W/2, name.Y+name.H/2)
	if want := []string{"Paper"}; !slices.Equal(act.picked, want) {
		t.Errorf("picked %v, want %v", act.picked, want)
	}
	if s.ColorScheme != "Campbell" {
		t.Errorf("scheme = %q, want the draft left as it was: %q", s.ColorScheme, "Campbell")
	}
	if len(act.removed) != 0 {
		t.Errorf("removed %v, want none", act.removed)
	}
}

// TestSchemeRowPicks covers the rest of a row: a click on the name picks the
// scheme, and a scheme the settings do not write has no button to remove it.
func TestSchemeRowPicks(t *testing.T) {
	s := store.Default()
	s.ColorScheme = "Campbell"
	tt, act := schemesOf(t, &s, []Scheme{{Name: "Paper"}, {Name: "Campbell"}})

	name, ok := tt.Find("Paper")
	if !ok {
		t.Fatalf("the list does not show the scheme: %q", tt.Texts())
	}
	tt.Move(name.X+name.W/2, name.Y+name.H/2)
	tt.Frame()
	if _, ok := tt.Find(locales.SchemeRemove()); ok {
		t.Error("a scheme the settings do not write offers to be removed")
	}
	if err := tt.Click("Paper"); err != nil {
		t.Fatal(err)
	}
	if want := []string{"Paper"}; !slices.Equal(act.picked, want) {
		t.Errorf("picked %v, want %v", act.picked, want)
	}
	if s.ColorScheme != "Campbell" {
		t.Errorf("scheme = %q, want the draft left as it was: %q", s.ColorScheme, "Campbell")
	}
	if len(act.removed) != 0 {
		t.Errorf("removed %v, want none", act.removed)
	}
}

// TestPreviewPalette covers the colors at the foot of the preview: the sixteen
// a program names, in the two rows a terminal keeps them in, the bright eight
// under the eight they brighten. A person reads a scheme's colors off this, so
// two colors that run into one another, and a color that runs into the page it
// is drawn on, are what it catches.
func TestPreviewPalette(t *testing.T) {
	s := store.Default()
	// A window that shows the scheme's background as it is: that is where the
	// black of a scheme and the page are one color.
	s.BackgroundOpacity = 1
	th := theme.Terminal(s)
	tt := myui.NewTester(func(c *myui.Context) { preview(c, &s) }, 400, 240)
	tt.Frame()
	img := tt.Image()
	px := func(x, y int) color.RGBA { return img.RGBAAt(x, y) }
	rgb := func(col myui.Color) color.RGBA { return color.RGBA{R: col.R, G: col.G, B: col.B, A: 0xff} }
	inside := boxesOf(img)

	// A chip's colors fill a box 16 DIP square with a line around it, so what
	// one color fills is 14 by 14.
	const chipSize, chipGap = 16, 6
	const pitch = chipSize + chipGap
	// A row is read from its second box, since its first is the scheme's black
	// and the page is black too: where the row starts is one box and one gap to
	// the left of it.
	seat, ok := inside[rgb(th.Palette[1])]
	if !ok {
		t.Fatalf("the preview draws no box of color 1")
	}
	origin := image.Pt(seat.Min.X-pitch, seat.Min.Y)
	// The colors sit in the order a terminal holds them: the eight a program
	// names, then the bright eight below them, a box and a gap apart, so that a
	// color stands over its bright twin.
	for i := 1; i < 16; i++ {
		want := image.Pt(origin.X+i%8*pitch, origin.Y+i/8*pitch)
		box, ok := inside[rgb(th.Palette[i])]
		if !ok {
			t.Errorf("the preview draws no box of color %d", i)
			continue
		}
		if box.Dx() != chipSize-2 || box.Dy() != chipSize-2 {
			t.Errorf("color %d fills %v by %v DIP, want %v inside a %v box", i, box.Dx(), box.Dy(), chipSize-2, chipSize)
		}
		if box.Min != want {
			t.Errorf("color %d sits at %v, want %v", i, box.Min, want)
		}
		if got := px(box.Min.X+(chipSize-2)/2, box.Min.Y+(chipSize-2)/2); got != rgb(th.Palette[i]) {
			t.Errorf("color %d is drawn %v, want %v", i, got, rgb(th.Palette[i]))
		}
	}

	// The black of a scheme is the page's own color, so the line around its box
	// is all there is of it: a box with no line is a chip a person cannot see.
	middle := origin.X + chipSize/2
	fill := px(middle, origin.Y+chipSize/2)
	if want := rgb(th.Palette[0]); fill != want {
		t.Errorf("the box of the scheme's black is drawn %v, want the scheme's own %v", fill, want)
	}
	if got := px(middle, origin.Y-1); got == fill {
		t.Errorf("the box of the scheme's black draws %v along its edge, its own color, so the chip is not there", got)
	}
}

// TestPreviewSpeaksTheLanguage covers the sample the preview is drawn with: it
// is the language's own sample, the one in its catalog, so the terminal the
// preview shows shows the words the window is in. That is what a person wants
// the face and the typesetting read off: their own words, in their own script,
// rather than an English one that says nothing about either.
func TestPreviewSpeaksTheLanguage(t *testing.T) {
	for _, tc := range []struct {
		locale string
		lines  []string
	}{
		{"en", []string{"$ Hello! AaBbZz 001lI8B", "  1234567890123456 ABCD", "  0123456789 !@#$%^&*()"}},
		{"zh-Hans", []string{"$ 你好 AaBbZz 001lI8B", "  一二三四五六七八 中文测试", "  0123456789 !@#$%^&*()"}},
	} {
		t.Run(tc.locale, func(t *testing.T) {
			speakIn(t, tc.locale)
			s := store.Default()
			tt := myui.NewTester(func(c *myui.Context) { preview(c, &s) }, 400, 240)
			tt.Frame()
			// The two lines are all the preview says: the colors below them are
			// boxes, not words.
			if got := tt.Texts(); !slices.Equal(got, tc.lines) {
				t.Errorf("the preview says %q, want %q", got, tc.lines)
			}
		})
	}
}

// boxesOf is where each color of the picture fills a box: the runs of one color
// 14 DIP square, which is a chip's inside a 16 DIP box with a line around it.
func boxesOf(img *image.RGBA) map[color.RGBA]image.Rectangle {
	b := img.Bounds()
	seen := make(map[image.Point]bool, b.Dx()*b.Dy())
	out := map[color.RGBA]image.Rectangle{}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			start := image.Pt(x, y)
			if seen[start] {
				continue
			}
			col := img.RGBAAt(x, y)
			box := image.Rect(x, y, x+1, y+1)
			stack := []image.Point{start}
			seen[start] = true
			for len(stack) > 0 {
				p := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				box = box.Union(image.Rect(p.X, p.Y, p.X+1, p.Y+1))
				for _, d := range []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					q := p.Add(d)
					if !q.In(b) || seen[q] || img.RGBAAt(q.X, q.Y) != col {
						continue
					}
					seen[q] = true
					stack = append(stack, q)
				}
			}
			if box.Dx() == chipDot && box.Dy() == chipDot {
				out[col] = box
			}
		}
	}
	return out
}

// chipDot is how wide a chip's inside is: its box, less the line around it.
const chipDot = 14

// TestSchemesCacheKeepsTheRows covers the rows the page keeps from frame to
// frame: the same list and the same words are answered with the rows already
// found, and a list the window built again counts as the same while it names
// the same schemes, since moving the pointer is what draws the frames.
func TestSchemesCacheKeepsTheRows(t *testing.T) {
	all := []Scheme{{Name: "Paper"}, {Name: "Campbell"}}
	cache := &SchemesCache{}
	rows := searchSchemes(cache, all, "")
	if !slices.Equal(rows, all) {
		t.Fatalf("the rows of an empty search are %v, want %v", rows, all)
	}
	rebuilt := []Scheme{{Name: "Paper"}, {Name: "Campbell"}}
	if again := searchSchemes(cache, rebuilt, ""); &again[0] != &rows[0] {
		t.Error("the list was searched again for a list that names the same schemes")
	}
	typed := searchSchemes(cache, all, "pap")
	if len(typed) != 1 || typed[0].Name != "Paper" {
		t.Fatalf("the rows of \"pap\" are %v, want Paper alone", typed)
	}
	if again := searchSchemes(cache, all, "pap"); &again[0] != &typed[0] {
		t.Error("the words were searched again for the words they were searched for")
	}
	// A list that names another scheme is another list, and the rows kept for
	// the one before it are not its answer.
	renamed := []Scheme{{Name: "Paper"}, {Name: "Campbells"}}
	if again := searchSchemes(cache, renamed, ""); !slices.Equal(again, renamed) || &again[0] == &typed[0] {
		t.Error("a list that names another scheme was answered with the rows of the one before it")
	}
	// A page drawn without a window to keep a cache for it keeps none: it asks
	// the list again each time.
	if a, b := searchSchemes(nil, all, ""), searchSchemes(nil, all, ""); &a[0] == &b[0] {
		t.Error("a search with no cache to keep it in kept the rows anyway")
	}
}

// TestSchemesCacheReset covers emptying the cache: what it kept is the answer to
// no search after it.
func TestSchemesCacheReset(t *testing.T) {
	all := []Scheme{{Name: "Paper"}, {Name: "Campbell"}}
	cache := &SchemesCache{}
	rows := searchSchemes(cache, all, "")
	cache.reset()
	if cache.All != nil || cache.Rows != nil || cache.Key != "" {
		t.Errorf("the cache keeps %v, %q, %v after it was emptied", cache.All, cache.Key, cache.Rows)
	}
	if again := searchSchemes(cache, all, ""); &again[0] == &rows[0] {
		t.Error("the rows emptied out of the cache are the answer still")
	}
}

// TestSameSchemes covers telling one list of schemes from another: the length,
// each name and each pair of colors count; the slice the window built is not
// what is compared.
func TestSameSchemes(t *testing.T) {
	paper := Scheme{Name: "Paper", Bg: myui.RGB(1, 2, 3)}
	for _, tc := range []struct {
		name string
		a, b []Scheme
		want bool
	}{
		{"none", nil, nil, true},
		{"one", []Scheme{paper}, []Scheme{paper}, true},
		{"another name", []Scheme{paper}, []Scheme{{Name: "Campbell", Bg: paper.Bg}}, false},
		{"another color", []Scheme{paper}, []Scheme{{Name: "Paper", Bg: myui.RGB(9, 9, 9)}}, false},
		{"one more", []Scheme{paper}, []Scheme{paper, {Name: "Campbell"}}, false},
	} {
		if got := sameSchemes(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: sameSchemes = %v, want %v", tc.name, got, tc.want)
		}
	}
}
