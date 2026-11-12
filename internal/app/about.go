package app

import (
	"log"

	"wtgo/internal/cli"
	"wtgo/internal/fault"
	"wtgo/internal/i18n"
	"wtgo/internal/sys"
	"wtgo/internal/ui"
)

// project is the page of the program, which the about dialog opens for a person
// who came to it looking for where to say what went wrong.
const project = "https://github.com/KiLiMe/wt-go"

// OpenAbout shows the dialog that says what the program is. The palette goes:
// the dialog is modal, and two open bars would leave the keys to the last one
// built.
func (a *App) OpenAbout() {
	self := cli.About()
	a.palette = nil
	a.about = &ui.About{
		Open:     true,
		Version:  self.Label(),
		Revision: self.Revision,
		Go:       self.Go,
		Platform: self.Platform,
		Project:  project,
	}
	a.redraw()
}

// CloseAbout closes the about dialog.
func (a *App) CloseAbout() {
	a.about = nil
	a.redraw()
}

// OpenProjectPage opens the program's page in the desktop's browser, and says
// in the dialog when the desktop has none: a button that does nothing leaves a
// person clicking it.
func (a *App) OpenProjectPage() {
	if err := sys.OpenPath(project); err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.OpenLink, err.Error()))
		if a.about != nil {
			a.about.Note = i18n.F("errOpenLink", map[string]any{"error": fault.Line(fault.OpenLink, err.Error())})
		}
	}
	a.redraw()
}
