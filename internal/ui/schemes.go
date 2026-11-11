package ui

import (
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/i18n"
	"wtgo/internal/icons"
	"wtgo/internal/store"
	"wtgo/internal/theme"
)

// schemesPage lists the color schemes, and shows the one in front.
func schemesPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	t := c.Theme()
	section(c, i18n.T("Color scheme"), func() {
		field(c, i18n.T("Scheme"), i18n.T("The colors a terminal paints with unless its profile names another."), func() {
			choice(c, &s.ColorScheme, schemeOptions(st.Schemes))
		})
		row(c, func() {
			myui.Row(c).Children(func() {
				myui.Button(c, i18n.T("Copy as a custom scheme")).OnClick(act.CopyScheme)
			})
		})
	})
	names := search(st.Schemes, filter(st))
	myui.Row(c).AlignItems(myui.Start).Gap(24).Children(func() {
		myui.Column(c).Width(320).Gap(6).Children(func() {
			if st.Filter != nil {
				myui.TextInput(c, st.Filter).Placeholder(i18n.T("Search the schemes"))
			} else {
				myui.TextInput(c, new(string)).Placeholder(i18n.T("Search the schemes"))
			}
			if len(names) == 0 {
				myui.Text(c, i18n.T("No scheme by that name.")).FontSize(11).TextColor(t.TextMuted)
			}
			myui.Column(c).Gap(2).Children(func() {
				for _, sc := range names {
					schemeRow(c, sc, s, act)
				}
			})
		})
		if th, ok := theme.Scheme(*s, s.ColorScheme); ok {
			myui.Column(c).Grow(1).Gap(8).Children(func() {
				myui.Text(c, i18n.T("Preview")).FontSize(11).TextColor(t.TextMuted)
				preview(c, th)
			})
		}
	})
}

// filter is what the schemes are searched for.
func filter(st Settings) string {
	if st.Filter == nil {
		return ""
	}
	return *st.Filter
}

// schemeOptions are the schemes a control picks from.
func schemeOptions(schemes []Scheme) []option {
	out := make([]option, 0, len(schemes))
	for _, sc := range schemes {
		out = append(out, option{sc.Name, sc.Name})
	}
	return out
}

// schemeRow is one scheme in the list: its name, chosen by a click, and for a
// scheme the settings write the button that removes it -- shown while the row
// is in front, in room the row keeps either way, so that the names do not jump
// as the pointer moves.
func schemeRow(c *myui.Context, sc Scheme, s *store.Settings, act Actions) {
	el := myui.Row(c).Height(28).Padding(0, 4, 0, 8).AlignItems(myui.Center).Gap(4).Radius(6).
		Cursor(myui.CursorPointer).OnClick(func() { s.ColorScheme = sc.Name })
	if s.ColorScheme == sc.Name {
		el.Background(c.Theme().SurfaceHover)
	}
	shown := sc.Custom && (el.Hovered() || s.ColorScheme == sc.Name)
	el.Children(func() {
		myui.Text(c, sc.Name).Grow(1).Shrink(1).SingleLine()
		if !sc.Custom {
			return
		}
		remove := myui.Box(c).Size(20, 20).Radius(4).AlignItems(myui.Center).Justify(myui.Center)
		if !shown {
			// The box keeps its room, so that the names do not move as the
			// pointer goes from row to row, and nothing of the button is there
			// to click: a row that hides its button still picks the scheme.
			remove.Opacity(0).PassThrough()
		} else {
			remove.Cursor(myui.CursorPointer).Label(i18n.T("Remove the scheme")).
				Tooltip(i18n.T("Remove the scheme")).
				OnClick(func() { act.RemoveScheme(sc.Name) })
		}
		remove.Children(func() {
			myui.Icon(c, icons.Close.SVG).Size(12, 12).TextColor(c.Theme().TextMuted)
		})
	})
}

// search is the schemes that hold what the search asks for, at most the rows a
// page shows at once: the schemes number in the hundreds.
func search(schemes []Scheme, search string) []Scheme {
	search = strings.ToLower(strings.TrimSpace(search))
	out := make([]Scheme, 0, len(schemes))
	for _, sc := range schemes {
		if search != "" && !strings.Contains(strings.ToLower(sc.Name), search) {
			continue
		}
		out = append(out, sc)
		if len(out) == 60 {
			break
		}
	}
	return out
}

// preview shows the colors of a scheme: its background, foreground, cursor and
// selection, and the sixteen colors programs ask for.
func preview(c *myui.Context, th *terminal.Theme) {
	t := c.Theme()
	myui.Column(c).Gap(8).Children(func() {
		myui.Box(c).Padding(8).Radius(6).Background(th.Background).Children(func() {
			myui.Text(c, "The quick brown fox").TextColor(th.Foreground).SingleLine()
			myui.Text(c, "jumps over the lazy dog").TextColor(th.Foreground).SingleLine()
		})
		myui.Row(c).Gap(6).Children(func() {
			for _, col := range []myui.Color{th.Background, th.Foreground, th.Cursor, th.Selection} {
				myui.Box(c).Size(26, 26).Radius(4).Background(col).BorderWidth(1).BorderColor(t.Border)
			}
		})
		myui.Row(c).Gap(4).Children(func() {
			for _, col := range th.Palette {
				myui.Box(c).Size(22, 22).Radius(3).Background(col)
			}
		})
	})
}
