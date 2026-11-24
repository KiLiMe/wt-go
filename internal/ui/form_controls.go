package ui

import (
	"math"
	"strconv"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/theme"
	"wtgo/locales"
)

// Boundary: the controls a settings page offers -- one of a few values, a whole
// number, a measurement, a share, a color -- and what each holds. Where they sit
// is form.go.

// option is one choice of a control: the value the settings hold, and what the
// control calls it.
type option struct {
	Value string
	Label string
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

// choiceOf is choice for a setting the file writes as a number: the choice is
// shown as the number it stands for, and a weight the file holds but no choice
// names shows as it is, which is what choice already does for a word.
func choiceOf(c *myui.Context, value *int, options []option) myui.Element {
	shown := strconv.Itoa(*value)
	el := choice(c, &shown, options)
	if n, err := strconv.Atoi(shown); err == nil {
		*value = n
	}
	return el
}

// weightOptions are the weights a family may have, as the number and the name
// of the face: the nine mygo clamps a font to, and no others.
func weightOptions() []option {
	names := []string{
		locales.WeightThin(), locales.WeightExtraLight(), locales.WeightLight(),
		locales.WeightRegular(), locales.WeightMedium(), locales.WeightSemiBold(),
		locales.WeightBold(), locales.WeightExtraBold(), locales.WeightBlack(),
	}
	out := make([]option, 0, len(names))
	for i, name := range names {
		w := strconv.Itoa(100 * (i + 1))
		out = append(out, option{w, w + " · " + name})
	}
	return out
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

// deviceLabel names the shell following the device, in the lists that offer
// it: as the shell it runs while the machine has one to name, and as following
// the device alone when it has none.
func deviceLabel(device string) string {
	if device == "" {
		return locales.SystemProfile()
	}
	return locales.SystemProfileOf(device)
}
