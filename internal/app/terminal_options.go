package app

import (
	"github.com/egoist/mygo/plugins/terminal"
	"wtgo/internal/profile"
	"wtgo/internal/theme"
)

// termOptions is the terminal of a tab that runs p in dir.
func (a *App) termOptions(p profile.Profile, dir string) terminal.Options {
	s := a.settings
	return terminal.Options{
		Command:    p.CommandLine(),
		Dir:        dir,
		Font:       a.terminalFont(),
		Theme:      theme.Terminal(s),
		Cursor:     cursor(s.CursorShape),
		NoBlink:    !s.CursorBlink,
		Scrollback: s.ScrollbackMB << 20,
		Padding:    terminal.Padding{X: s.Padding.X, Y: s.Padding.Y},
		RightClick: rightClick(s.RightClick),
		// What the grid is surrounded by, and the read floor of its text,
		// are fixed for the life of a terminal: they are set here, not in
		// apply, and a change shows in the next tab.
		PaddingColor:     paddingColor(s.PaddingColor),
		MinimumContrast:  s.MinimumContrast,
		ScrollMultiplier: scrollMultiplier(s.ScrollMultiplier),
		ClipboardRead:    clipboardRead(s.ClipboardRead),
		NoClipboardWrite: s.ClipboardWrite == "deny",
		CopyOnSelect: func() terminal.CopyOnSelect {
			if s.CopyOnSelect {
				return terminal.CopyOnSelectClipboard
			}
			return terminal.CopyOnSelectNone
		}(),
		OnTitle: func(string) { a.Do(a.syncTitle) },
		OnExit:  func(int) { a.Do(a.syncTitle) },
		// A path printed in a terminal is a link only because this window
		// opens it, from the directory the program started in.
		OpenLink: a.openLink(dir),
	}
}

// rightClick is what a secondary click does, as the settings name it.
func rightClick(action string) terminal.RightClickAction {
	switch action {
	case "paste":
		return terminal.RightClickPaste
	case "copy":
		return terminal.RightClickCopy
	case "menu":
		return terminal.RightClickContextMenu
	case "nothing":
		return terminal.RightClickIgnore
	}
	return terminal.RightClickCopyOrPaste
}

// cursor is the shape of the cursor the settings name, and a bar, the
// default, for a name the terminal does not know.
func cursor(shape string) terminal.CursorStyle {
	switch shape {
	case "block":
		return terminal.CursorBlock
	case "underline":
		return terminal.CursorUnderline
	}
	return terminal.CursorBar
}

// paddingColor is what the room around a terminal's grid shows, as the
// settings name it: the window's background unless they name the cells.
func paddingColor(name string) terminal.PaddingColor {
	switch name {
	case "extend":
		return terminal.PaddingExtend
	case "always":
		return terminal.PaddingExtendAlways
	}
	return terminal.PaddingBackground
}

// scrollMultiplier is how far a notch of the wheel and a touchpad's scrolling
// go, as the settings name it. The wheel scrolls the rows its own setting
// names at 3, the default, so the multiplier scales that too: 2 scrolls twice
// as far, at the wheel and under the fingers.
func scrollMultiplier(m float32) terminal.ScrollMultiplier {
	if m <= 0 {
		m = 1
	}
	return terminal.ScrollMultiplier{Precision: m, Discrete: 3 * m}
}

// clipboardRead is what a program reading the clipboard gets, as the settings
// name it: refused unless they say otherwise.
func clipboardRead(name string) terminal.ClipboardAccess {
	switch name {
	case "allow":
		return terminal.ClipboardAllow
	case "ask":
		return terminal.ClipboardAsk
	}
	return terminal.ClipboardDeny
}
