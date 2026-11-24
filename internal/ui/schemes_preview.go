package ui

import (
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
	"wtgo/internal/theme"
	"wtgo/locales"
)

// Boundary: the preview at the head of the appearance page -- a terminal drawn
// from the draft settings, with the sample lines it says and the cursor of the
// draft. The settings it reads are edited by settings.go.

// preview is the preview at the head of the appearance page: a terminal drawn
// as the draft would draw it, so that a change to the face, the typesetting,
// the cursor or the scheme shows as it is made rather than after a save.
func preview(c *myui.Context, s *store.Settings) {
	th := theme.Terminal(*s)
	t := c.Theme()
	// The window is translucent over what is behind it, which the preview has
	// no window behind: the scheme's own background at the same share of the
	// face the page paints on is what a person sees of it.
	bg := th.Background.MixOklab(t.Background, 1-float32(clampOpacity(s.BackgroundOpacity)))
	myui.Box(c).BorderWidth(1).BorderColor(cardLine(t)).Radius(8).Background(bg).
		Padding(previewPadY(s), previewPadX(s)).Children(func() {
		// The sample the preview is read off: the prompt with the cursor on it,
		// and the lines under it, which start where the prompt leaves the words.
		// A monospace face puts a letter, a digit and a symbol on the same
		// advance, so the columns of the three lines line up; a face that is not
		// monospace, or a fallback that draws the letters of a script at another
		// width, shows as the columns that do not, which is what the sample is
		// for. The words are the language's own, from its catalog.
		for i, line := range sampleLines() {
			myui.Row(c).AlignItems(myui.Center).Gap(6).Children(func() {
				myui.Text(c, line).TextColor(th.Foreground).Font(textFont(s)).
					FontSize(previewFontSize(s)).LineHeight(previewLineHeight(s))
				if i == 0 {
					cursorBlock(c, s, th)
				}
			})
		}
		// The sixteen colors a program names, in the two rows a terminal keeps
		// them in: the eight it names, and the bright eight below them, so that a
		// color stands over its bright twin. The room between them keeps two
		// neighbors from reading as one band, and the line around each keeps the
		// black of a scheme, which is its background's own color, from vanishing
		// into the page it is drawn on.
		myui.Column(c).Padding(8, 0, 0, 0).Gap(6).Children(func() {
			for _, row := range [][]myui.Color{th.Palette[:8], th.Palette[8:]} {
				myui.Row(c).Gap(6).Children(func() {
					for _, col := range row {
						myui.Box(c).Size(16, 16).Radius(3).Background(col).
							BorderWidth(1).BorderColor(cardLine(t))
					}
				})
			}
		})
	})
}

// sampleLines is what the preview says a terminal says: the prompt with the
// language's first line, and the two lines under it, drawn with the room the
// prompt takes so that the columns of the three line up. That is what the
// sample is for: a face that is not monospace, or a fallback that draws a
// script at another advance, puts them out of line where a person can see it.
func sampleLines() []string {
	return []string{
		"$ " + locales.LineOne(),
		"  " + locales.LineTwo(),
		"  " + locales.LineThree(),
	}
}

// clampOpacity holds a share to what the window can show, since the draft is
// drawn while it is typed and may name one the file would refuse.
func clampOpacity(v float64) float64 {
	return min(max(v, 0), 1)
}

// previewFontSize is the size the preview draws the terminal's lines with: the
// draft's own, held to a range that leaves the preview its room, so that a size
// typed into the file does not make the preview a card of lines alone.
func previewFontSize(s *store.Settings) float32 {
	return min(max(s.Font.Size, previewFontMin), previewFontMax)
}

// previewFontMin and previewFontMax are how small and how large the preview
// draws a line.
const (
	previewFontMin float32 = 9
	previewFontMax float32 = 20
)

// previewLineHeight is the draft's line height, held to what the control that
// sets it offers.
func previewLineHeight(s *store.Settings) float32 {
	return min(max(s.Font.LineHeight, LineHeightMin), LineHeightMax)
}

// textFont is the family the preview draws with: the draft's, or the default
// families where the draft names none, which is what the terminal draws with
// then and what the field is filled in with.
func textFont(s *store.Settings) string {
	if strings.TrimSpace(s.Font.Family) == "" {
		return store.DefaultFontFamily()
	}
	return s.Font.Family
}

// previewPadX and previewPadY are the room the preview keeps around its lines,
// held to what the preview has to give: a terminal's padding is measured for a
// window, and the preview is a card of one.
func previewPadX(s *store.Settings) float32 { return min(max(s.Padding.X, 0), 48) }

func previewPadY(s *store.Settings) float32 { return min(max(s.Padding.Y, 0), 24) }

// cursorBlock is the cursor of the draft, in its shape and the scheme's cursor
// color, beside the text it would sit after.
func cursorBlock(c *myui.Context, s *store.Settings, th *terminal.Theme) {
	// The line is as tall as the letters of the preview, so the cursor is too.
	h := previewFontSize(s) + 4
	shape := myui.Box(c).Height(h).Background(th.Cursor)
	switch s.CursorShape {
	case "block":
		shape.Width(previewFontSize(s))
	case "underline":
		shape.Width(previewFontSize(s)).Height(2)
	default:
		shape.Width(2)
	}
	// A blink is a moment, and the preview is a frame: the cursor is drawn
	// either way, so the switch says what is drawn and only the shape shows.
}
