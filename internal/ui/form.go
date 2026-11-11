package ui

import (
	"math"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/theme"
)

// option is one choice of a control: the value the settings hold, and what the
// control calls it.
type option struct {
	Value string
	Label string
}

// field is one setting: what it is called, what it does, and the control that
// sets it. The label names the control, so a click on it focuses the control
// and assistive technology reads the two together.
func field(c *myui.Context, title, description string, control func()) {
	myui.Field(c, title, control).Description(description)
}

// section is a heading and the fields under it, which a form lines up: the
// labels in a column of their own, as wide as the widest, level with the
// first line of the control beside them.
func section(c *myui.Context, title string, fields func()) {
	myui.Column(c).Gap(12).Children(func() {
		myui.Text(c, title).FontSize(13).TextColor(c.Theme().TextMuted)
		myui.Form(c, fields)
	})
}

// row is a line of a form that no label names, such as a button under the
// fields, which lines up with the controls above it.
func row(c *myui.Context, control func()) {
	myui.Field(c, "", control)
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
