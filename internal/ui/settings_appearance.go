package ui

import (
	myui "github.com/egoist/mygo/ui"
	"wtgo/internal/theme"
	"wtgo/locales"
)

// appearancePage is how the window and the terminals look: a terminal drawn as
// the draft draws it, held at the head of the page while the face, the
// typesetting, the cursor, and the colors it is drawn with scroll under it.
func appearancePage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	pinnedSection(c, locales.AppearancePreview(), func() {
		preview(c, s)
	})
	section(c, locales.AppearanceFace(), func() {
		field(c, locales.Theme(), locales.ThemeNote(), func() {
			options := []option{
				{"system", locales.FollowSystem()},
				{"light", locales.Light()},
				{"dark", locales.Dark()},
			}
			for _, name := range theme.UIThemes() {
				options = append(options, option{name, name})
			}
			choice(c, &s.Theme, options)
		})
		field(c, locales.BackgroundOpacity(), locales.BackgroundOpacityNote(), func() {
			ratio(c, &s.BackgroundOpacity, 0.3, 1, 0.05)
		})
	})
	section(c, locales.AppearanceType(), func() {
		fieldBlock(c, locales.Font(), locales.FontNote(), func() {
			myui.Column(c).Gap(8).MinWidth(0).Children(func() {
				fontPicker(c, &s.Font.Family, st.FontFamilies)
				if text, danger := fontNote(s.Font.Family, st.Fonts); text != "" {
					color := c.Theme().TextMuted
					if danger {
						color = c.Theme().Danger
					}
					myui.Text(c, text).FontSize(12).TextColor(color).Wrap()
				}
			})
		})
		field(c, locales.FontWeight(), locales.FontWeightNote(), func() {
			choiceOf(c, &s.Font.Weight, weightOptions())
		})
		field(c, locales.FontSize(), locales.FontSizeNote(), func() {
			measure(c, &s.Font.Size, 6, 48, 1)
		})
		field(c, locales.LineHeight(), locales.LineHeightNote(), func() {
			measure(c, &s.Font.LineHeight, 0.8, 2.4, 0.1)
		})
		field(c, locales.Padding(), locales.PaddingNote(), func() {
			myui.Row(c).AlignItems(myui.Center).Gap(8).Children(func() {
				measure(c, &s.Padding.X, 0, 64, 1)
				myui.Text(c, "\u00d7").TextColor(c.Theme().TextMuted)
				measure(c, &s.Padding.Y, 0, 64, 1)
			})
		})
		field(c, locales.PaddingColor(), locales.PaddingColorNote(), func() {
			choice(c, &s.PaddingColor, []option{
				{"background", locales.WindowBackground()},
				{"extend", locales.NearbyCells()},
				{"always", locales.AlwaysCells()},
			})
		})
	})
	section(c, locales.AppearanceCursor(), func() {
		field(c, locales.Cursor(), locales.CursorNote(), func() {
			myui.Row(c).AlignItems(myui.Center).Gap(12).Children(func() {
				choice(c, &s.CursorShape, []option{
					{"bar", locales.Bar()},
					{"block", locales.Block()},
					{"underline", locales.Underline()},
				})
				myui.Switch(c, &s.CursorBlink)
				myui.Text(c, locales.Blink()).TextColor(c.Theme().TextMuted)
			})
		})
		field(c, locales.MinimumContrast(), locales.MinimumContrastNote(), func() {
			measure(c, &s.MinimumContrast, 0, 21, 0.5)
		})
	})
	section(c, locales.ColorSchemes(), func() {
		schemePicker(c, s, st, act)
	})
}
