package ui

import (
	"math"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/theme"
	"wtgo/locales"
)

// option is one choice of a control: the value the settings hold, and what the
// control calls it.
type option struct {
	Value string
	Label string
}

// field is one setting: what it is called and does on the left, and the
// control that sets it on the right, in a row of the card its section draws.
// The label names the control, so that assistive technology reads the two
// together, as a click on a form's label focuses its control.
func field(c *myui.Context, title, description string, control func()) {
	fieldRow(c, title, description, control, false)
}

// fieldBlock is a field whose control goes under the words rather than beside
// them: one that needs the whole width, as a command line does.
func fieldBlock(c *myui.Context, title, description string, control func()) {
	fieldRow(c, title, description, control, true)
}

// fieldRow is field and fieldBlock: the words of a setting, and its control
// beside or under them.
func fieldRow(c *myui.Context, title, description string, control func(), block bool) {
	t := c.Theme()
	words := func() {
		myui.Column(c).Gap(3).MinWidth(0).Children(func() {
			myui.Text(c, title).FontSize(13).FontWeight(500)
			if description != "" {
				myui.Text(c, description).FontSize(12).TextColor(t.TextMuted).Wrap()
			}
		})
	}
	line := myui.Column(c).BorderWidth(0, 0, 1, 0).BorderColor(cardLine(t))
	if block {
		line.Gap(10).Padding(14, 16).Children(func() {
			words()
			myui.Row(c).Children(control)
		})
	} else {
		// The words take the room the control leaves, so that the controls of a
		// page line up on the right edge however wide they are.
		line.Padding(14, 16).Children(func() {
			myui.Row(c).Gap(24).AlignItems(myui.Center).Children(func() {
				myui.Column(c).Grow(1).MinWidth(0).Children(words)
				myui.Row(c).Shrink(0).Justify(myui.End).AlignItems(myui.Center).Children(control)
			})
		})
	}
	// A row of words and a control is one group to a reader, as mygo names
	// the field of a form: the label names the group.
	line.Role(myui.RoleGroup).Label(title)
}

// controlWidth is as wide as the control of a field grows: past it the words
// beside it would be a line a word long.
const controlWidth = 320

// section is a heading and the fields under it, which the page draws as one
// card: a surface between the window's own and its controls', so that a
// button in it still reads as a button.
func section(c *myui.Context, title string, fields func()) {
	cardSection(c, title, false, fields)
}

// dangerSection is a section in the theme's danger color: what a person does
// once and cannot take back.
func dangerSection(c *myui.Context, title string, fields func()) {
	cardSection(c, title, true, fields)
}

// pinnedSection is a section that keeps the head of the page while the fields
// under it scroll: the appearance page draws the terminal the settings make,
// and the setting that changes it is often under the thumb, so the preview is
// worth more at the top than wherever the scroll leaves it. The band paints
// the page's own background, since a sticky element draws above its siblings
// and what rolls behind it would otherwise show through the gap between its
// heading and its card.
func pinnedSection(c *myui.Context, title string, fields func()) {
	t := c.Theme()
	myui.Column(c).Sticky().Top(0).Background(t.Background).
		Padding(10, 0, 12, 0).BorderWidth(0, 0, 1, 0).BorderColor(cardLine(t)).Children(func() {
		section(c, title, fields)
	})
}

// cardSection is section and dangerSection: a heading, and a card of rows.
func cardSection(c *myui.Context, title string, danger bool, fields func()) {
	t := c.Theme()
	face, line := cardFace(t), cardLine(t)
	if danger {
		face, line = t.Background.MixOklab(t.Danger, 0.1), t.Danger.Alpha(0.45)
	}
	myui.Column(c).Gap(8).Children(func() {
		if title != "" {
			myui.Text(c, title).FontSize(12).FontWeight(600).TextColor(t.TextMuted)
		}
		card(c, face, line, fields)
	})
}

// card is a surface between the window's background and its controls', which
// what it holds is drawn on.
func card(c *myui.Context, face, line myui.Color, body func()) {
	myui.Column(c).Background(face).BorderWidth(1).BorderColor(line).Radius(10).Children(body)
}

// cardFace is the surface of a card: a shade of the text above the window's
// background, under the surfaces of the controls in it. The window's Surface
// would hide them, being the face of a button itself.
func cardFace(t *myui.Theme) myui.Color {
	return t.Background.MixOklab(t.Text, 0.07)
}

// cardLine is what a card is outlined with, and the line between two of its
// rows: a hairline, where Border is a control's edge.
func cardLine(t *myui.Theme) myui.Color {
	return t.Background.MixOklab(t.Text, 0.14)
}

// intro is the name of a page and what it is for, above its cards.
func intro(c *myui.Context, title, note string) {
	t := c.Theme()
	myui.Column(c).Gap(4).Children(func() {
		myui.Text(c, title).FontSize(20).FontWeight(600)
		if note != "" {
			myui.Text(c, note).FontSize(12).TextColor(t.TextMuted).Wrap()
		}
	})
}

// row is a line of a card that no label names, such as a button under the
// fields.
func row(c *myui.Context, control func()) {
	myui.Row(c).Padding(10, 16).AlignItems(myui.Center).
		BorderWidth(0, 0, 1, 0).BorderColor(cardLine(c.Theme())).Children(control)
}

// deviceLabel names the shell following the device, in the lists that offer
// it: as the shell it runs while the machine has one to name, and as following
// the device alone when it has none.
func deviceLabel(device string) string {
	if device == "" {
		return locales.SystemProfile()
	}
	return locales.SystemProfileOf(device)
}

// choice is a control that picks one of a few values: the string settings
// hold, and the named types over one, such as store.Confirm. myui.Select
// picks among the labels it shows, so it opens on the label of the value in
// force, and a chosen label is written back as the value it stands for.
func choice[T ~string](c *myui.Context, value *T, options []option) myui.Element {
	labels := make([]string, len(options))
	shown := ""
	for i, o := range options {
		labels[i] = o.Label
		if o.Value == string(*value) {
			shown = o.Label
		}
	}
	if shown == "" {
		// A value no option names, such as a font family list typed into
		// the file, shows as it is.
		shown = string(*value)
	}
	label := shown
	el := myui.Select(c, &label, labels)
	if label != shown {
		for _, o := range options {
			if o.Label == label {
				*value = T(o.Value)
			}
		}
	}
	return el
}

// count is a control that picks a whole number.
func count(c *myui.Context, value *int, lo, hi int) myui.Element {
	v := float64(*value)
	el := myui.NumberInput(c, &v, float64(lo), float64(hi), 1)
	*value = int(v)
	return el
}

// measure is a control that picks a measurement, in steps.
func measure(c *myui.Context, value *float32, lo, hi, step float32) myui.Element {
	v := float64(*value)
	el := myui.NumberInput(c, &v, float64(lo), float64(hi), whole(step))
	*value = float32(v)
	return el
}

// whole is a step as it was meant, without the noise of a float32: 0.1 as a
// float32 is 0.10000000149011612, which asks the control for seventeen
// decimals.
func whole(step float32) float64 {
	const places = 1e6
	return math.Round(float64(step)*places) / places
}

// ratio is a control that picks a number from zero to one, in steps.
func ratio(c *myui.Context, value *float64, lo, hi, step float64) myui.Element {
	return myui.NumberInput(c, value, lo, hi, step)
}

// colorField is a field that holds a color, and the color it holds.
func colorField(c *myui.Context, value *string) myui.Element {
	t := c.Theme()
	return myui.Row(c).AlignItems(myui.Center).Gap(8).Children(func() {
		swatch := myui.Box(c).Size(18, 18).Radius(4).BorderWidth(1).BorderColor(t.Border)
		if col, ok := theme.Color(*value); ok {
			swatch.Background(col)
		}
		myui.TextInput(c, value).Width(160).Placeholder("#rrggbb")
	})
}
