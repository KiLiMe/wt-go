package app

import (
	"wtgo/internal/action"
	"wtgo/internal/ui"
	"wtgo/locales"
)

// OpenPalette shows the command palette, which lists what the window does.
// Whatever other overlay was open goes: the palette holds the keyboard, and
// two of them would leave the keys, Escape among them, to whichever was built
// last.
func (a *App) OpenPalette() {
	a.openOverlay(overlayPalette)
	a.redraw()
}

// Run does the action the palette picked.
func (a *App) Run(id string) {
	a.closeOverlay(overlayPalette)
	a.act(id)
	a.redraw()
}

// paletteState is the palette as the window draws it, with what it offers: the
// tabs, the profiles, and the actions the window has, each with its key.
func (a *App) paletteState() *ui.Palette {
	if a.palette == nil {
		return nil
	}
	a.palette.Commands = a.paletteCommands()
	return a.palette
}

// paletteCommands are what the palette offers.
func (a *App) paletteCommands() []ui.Command {
	var out []ui.Command
	for _, cmd := range action.All() {
		// The numbered actions are listed by what they open, below, which
		// names the tab or the profile.
		if _, ok := action.GotoIndex(cmd.ID); ok {
			continue
		}
		if _, ok := action.ProfileIndex(cmd.ID); ok {
			continue
		}
		out = append(out, ui.Command{
			Action: cmd.ID,
			Title:  action.Title(cmd.ID),
			Group:  action.GroupTitle(cmd.GroupKey),
			Key:    a.shortcut(cmd.ID),
		})
	}
	for i, tab := range a.panes.Tabs {
		if i >= action.MaxNumber {
			break
		}
		id := action.GotoAction(i + 1)
		out = append(out, ui.Command{
			Action: id,
			Title:  a.tabName(tab),
			Group:  action.GroupTitle(action.GroupTab),
			Key:    a.shortcut(id),
		})
	}
	for i, p := range a.menuProfiles() {
		if i >= action.MaxNumber {
			break
		}
		id := action.ProfileAction(i + 1)
		out = append(out, ui.Command{
			Action: id,
			Title:  locales.PaletteNewTabWith(p.Name),
			Group:  action.GroupTitle(action.GroupProfiles),
			Key:    a.shortcut(id),
		})
	}
	return out
}
