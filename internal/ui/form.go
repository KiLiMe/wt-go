package ui

import (
	"math"
	"strconv"

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
// sets it.
func field(c *myui.Context, title, description string, control func()) {
	t := c.Theme()
	myui.Row(c).AlignItems(myui.Start).Gap(16).Children(func() {
		myui.Column(c).Width(200).Gap(2).Children(func() {
			myui.Text(c, title).TextColor(t.Text)
			if description != "" {
				myui.Text(c, description).FontSize(11).TextColor(t.TextMuted).SingleLine()
			}
		})
		myui.Column(c).Grow(1).Children(control)
	})
}

// section is a heading and the fields under it.
func section(c *myui.Context, title string, fields func()) {
	myui.Column(c).Gap(12).Children(func() {
		myui.Text(c, title).FontSize(13).TextColor(c.Theme().TextMuted)
		fields()
	})
}

// choice is a control that picks one of a few values.
func choice(c *myui.Context, value *string, options []option) myui.Element {
	label := *value
	for _, o := range options {
		if o.Value == *value {
			label = o.Label
		}
	}
	return myui.MenuButton(c, label, func(m *myui.Menu) {
		for _, o := range options {
			if m.Item(o.Label).Checked(o.Value == *value).Chosen() {
				*value = o.Value
			}
		}
	})
}

// count is a control that picks a whole number.
func count(c *myui.Context, value *int, lo, hi int) myui.Element {
	v := float64(*value)
	el := number(c, &v, float64(lo), float64(hi), 1)
	*value = int(v)
	return el
}

// measure is a control that picks a measurement, in steps.
func measure(c *myui.Context, value *float32, lo, hi, step float32) myui.Element {
	v := float64(*value)
	el := number(c, &v, float64(lo), float64(hi), float64(step))
	*value = float32(v)
	return el
}

// ratio is a control that picks a number from zero to one, in steps.
func ratio(c *myui.Context, value *float64, lo, hi, step float64) myui.Element {
	return number(c, value, lo, hi, step)
}

// number is a control that picks a number between lo and hi in steps: the
// number, and the two buttons that step it.
func number(c *myui.Context, value *float64, lo, hi, step float64) myui.Element {
	nudge := func(delta float64) {
		n := math.Round((*value+delta-lo)/step)*step + lo
		*value = math.Min(hi, math.Max(lo, n))
	}
	return myui.Row(c).Gap(4).Children(func() {
		myui.Button(c, "-").OnClick(func() { nudge(-step) })
		myui.Text(c, formatNumber(*value, step)).Width(56).TextColor(c.Theme().Text)
		myui.Button(c, "+").OnClick(func() { nudge(step) })
	})
}

// formatNumber writes a number with the decimals its step asks for.
func formatNumber(v, step float64) string {
	switch {
	case step >= 1:
		return strconv.FormatFloat(v, 'f', 0, 64)
	case step >= 0.1:
		return strconv.FormatFloat(v, 'f', 1, 64)
	default:
		return strconv.FormatFloat(v, 'f', 2, 64)
	}
}

// colorField is a field that holds a color, and the color it holds.
func colorField(c *myui.Context, value *string) myui.Element {
	t := c.Theme()
	return myui.Row(c).AlignItems(myui.Center).Gap(8).Children(func() {
		swatch := myui.Box(c).Size(18, 18).Radius(4).BorderWidth(1).BorderColor(t.Border)
		if col, ok := theme.Color(*value); ok {
			swatch.Background(col)
		}
		myui.TextInput(c, value).Width(120).Placeholder("#rrggbb")
	})
}
