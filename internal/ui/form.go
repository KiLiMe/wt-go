package ui

import (
	myui "github.com/egoist/mygo/ui"
)

// Boundary: the shape of a settings page -- a field and the words and control
// it holds, the cards those sit in, and the sections and headings above them.
// The controls themselves are form_controls.go.

// field is one setting: what it is called and does on the left, and the
// control that sets it on the right, in a row of the card its section draws.
// The label names the control, so that assistive technology reads the two
// together, as a click on a form's label focuses its control.
//
// key is the setting the row holds, named the way the settings file names it:
// it is what the search over the settings marks when it finds this one, so the
// row is drawn from the same name the file writes. reveal is the mark to put
// over that row, and nil when nothing was found.
func field(c *myui.Context, key, title, description string, control func(), reveal *Reveal) {
	fieldRow(c, key, title, description, control, false, reveal)
}

// fieldBlock is a field whose control goes under the words rather than beside
// them: one that needs the whole width, as a command line does.
func fieldBlock(c *myui.Context, key, title, description string, control func(), reveal *Reveal) {
	fieldRow(c, key, title, description, control, true, reveal)
}

// fieldRow is field and fieldBlock: the words of a setting, and its control
// beside or under them.
func fieldRow(c *myui.Context, key, title, description string, control func(), block bool, reveal *Reveal) {
	t := c.Theme()
	words := func() {
		myui.Column(c).Gap(3).MinWidth(0).Children(func() {
			myui.Text(c, title).FontSize(13).FontWeight(500)
			if description != "" {
				myui.Text(c, description).FontSize(12).TextColor(t.TextMuted).Wrap()
			}
		})
	}
	cell := myui.Column(c)
	if block {
		cell.Gap(10).Padding(14, 16).Children(func() {
			words()
			myui.Row(c).Children(control)
		})
	} else {
		// The words take the room the control leaves, so that the controls of a
		// page line up on the right edge however wide they are.
		cell.Padding(14, 16).Children(func() {
			myui.Row(c).Gap(24).AlignItems(myui.Center).Children(func() {
				myui.Column(c).Grow(1).MinWidth(0).Children(words)
				myui.Row(c).Shrink(0).Justify(myui.End).AlignItems(myui.Center).Children(control)
			})
		})
	}
	// A row of words and a control is one group to a reader, as mygo names
	// the field of a form: the label names the group.
	cell.Role(myui.RoleGroup).Label(title)
	mark(c, cell, key, reveal)
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
// heading and its card. No line under the band: the card closes itself, and a
// border on the band draws a second edge in the room below it.
func pinnedSection(c *myui.Context, title string, fields func()) {
	t := c.Theme()
	myui.Column(c).Sticky().Top(0).Background(t.Background).
		Padding(10, 0, 12, 0).Children(func() {
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

// cardRadius is what a card's corners are rounded by, the ones a page draws
// and the two the profiles page builds of its own.
const cardRadius float32 = 10

// card is a surface between the window's background and its controls', which
// what it holds is drawn on. It draws the lines between its rows itself, in
// the room the gap gives them and inside its border, so that no row ends in a
// line of its own: the last would show as a second edge a pixel under the
// card's bottom, square where its corner is round, and every one reached
// across the whole card, past the corners.
func card(c *myui.Context, face, line myui.Color, body func()) {
	myui.Column(c).Background(face).BorderWidth(1).BorderColor(line).Radius(cardRadius).
		Gap(1).Dividers(1, line).Children(body)
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
	myui.Row(c).Padding(10, 16).AlignItems(myui.Center).Children(control)
}
