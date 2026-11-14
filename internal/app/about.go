package app

import (
	"log"

	"github.com/egoist/mygo"

	"wtgo/internal/cli"
	"wtgo/internal/fault"
	"wtgo/internal/sys"
	"wtgo/internal/ui"
	"wtgo/locales"
)

// project is the page of the program, which the about dialog opens for a person
// who came to it looking for where to say what went wrong.
const project = "https://github.com/KiLiMe/wt-go"

// OpenAbout shows the dialog that says what the program is. Whatever other
// overlay was open goes: the dialog holds the keyboard, and two of them would
// leave the keys to the last one built.
func (a *App) OpenAbout() {
	self := cli.About()
	a.openOverlay(overlayAbout)
	a.about = aboutOf(self, mygo.Version, sys.SettingsPath, sys.LogPath)
	a.about.Project = project
	a.redraw()
}

// aboutOf is the dialog about self, built on mygo version and the two paths,
// apart from the window it is shown in: a test can say what the dialog shows
// without a machine to ask.
func aboutOf(self cli.Self, mygoVersion string, settings, logPath func() (string, error)) *ui.About {
	return &ui.About{
		Open:     true,
		Version:  self.Label(),
		Revision: self.Revision,
		Mygo:     mygoVersion,
		Go:       self.Go,
		Platform: self.Platform,
		Settings: pathOf(settings),
		Log:      pathOf(logPath),
	}
}

// pathOf is a path the dialog shows, or "" when the folder it comes from
// cannot be named: the dialog says less rather than a half path (E103).
func pathOf(path func() (string, error)) string {
	p, err := path()
	if err != nil {
		return ""
	}
	return p
}

// CloseAbout closes the about dialog.
func (a *App) CloseAbout() {
	a.closeOverlay(overlayAbout)
	a.redraw()
}

// OpenProjectPage opens the program's page in the desktop's browser, and says
// in the dialog when the desktop has none: a button that does nothing leaves a
// person clicking it.
func (a *App) OpenProjectPage() {
	if err := sys.OpenPath(project); err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.OpenLink, err.Error()))
		if a.about != nil {
			a.about.Note = locales.ErrOpenLink(fault.Line(fault.OpenLink, err.Error()))
		}
	}
	a.redraw()
}
