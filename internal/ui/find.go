package ui

import (
	"fmt"

	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/i18n"
	"wtgo/internal/icons"
)

// The find bar's size.
const (
	findWidth  float32 = 320
	findHeight float32 = 38
	findMargin float32 = 8
)

// Find is the bar that asks the terminal in front for text, as Windows
// Terminal's does.
type Find struct {
	// Text is what the field holds.
	Text string
	// Status is what the terminal says of the matches, which the window
	// reads for each frame.
	Status terminal.SearchStatus
}

// findBar draws the bar over the terminal in front, at the top of the tab, and
// searches as the field is typed in.
func findBar(c *myui.Context, st State, act Actions) {
	f := st.Find
	if f == nil {
		return
	}
	t := c.Theme()
	room := Content(c)
	myui.Overlay(c, func() {
		el := myui.Row(c).
			Absolute().
			Left(max(room.W-findWidth-findMargin, 0)).
			Top(room.Y+findMargin).
			Width(min(findWidth, max(room.W-2*findMargin, 0))).
			Height(findHeight).
			Padding(0, 8).
			Gap(6).
			AlignItems(myui.Center).
			Radius(8).
			Background(t.Surface).
			BorderWidth(1).
			BorderColor(t.Border)
		el.Children(func() {
			in := myui.TextInputBase(c.Key("find"), &f.Text).
				Grow(1).
				Placeholder(i18n.T("Find"))
			// The bar holds the keyboard while it is open: a click on the
			// pane or on a button would otherwise take the focus and leave
			// the field deaf to the next key.
			in.Focus()
			if in.Changed() {
				act.Find(f.Text)
			}
			myui.Text(c, status(f.Status)).FontSize(11).TextColor(t.TextMuted).SingleLine()
			findButton(c, icons.Up, i18n.T("Previous match"), act.FindPrevious)
			findButton(c, icons.Chevron, i18n.T("Next match"), act.FindNext)
			findButton(c, icons.Close, i18n.T("Close"), act.EndFind)
			// Enter picks the next match, Shift+Enter the one before it. The
			// field takes Enter as its submit, and the bar cannot see the key
			// itself, so the modifiers held say which one it was.
			switch {
			case in.Submitted() && c.Modifiers()&myui.Shift != 0:
				act.FindPrevious()
			case in.Submitted():
				act.FindNext()
			}
			if el.OverlayShortcut(0, myui.KeyEscape) {
				act.EndFind()
			}
		})
	})
}

// findButton is one of the bar's buttons: an icon, and what a click does.
func findButton(c *myui.Context, ic *icons.Icon, tooltip string, onClick func()) {
	myui.Box(c).
		Size(22, 22).
		Radius(4).
		AlignItems(myui.Center).
		Justify(myui.Center).
		Cursor(myui.CursorPointer).
		Tooltip(tooltip).
		OnClick(onClick).
		Children(func() {
			myui.Icon(c, ic.SVG).Size(12, 12).TextColor(c.Theme().TextMuted)
		})
}

// status is what the bar says of the matches: where the selected one is among
// them, that there is none yet, or nothing while the field is empty.
func status(s terminal.SearchStatus) string {
	switch {
	case s.Text == "":
		return ""
	case s.Total == 0:
		return i18n.T("No matches")
	case s.Selected == 0:
		return fmt.Sprintf(i18n.T("%d found"), s.Total)
	}
	return fmt.Sprintf(i18n.T("%d of %d"), s.Selected, s.Total)
}
