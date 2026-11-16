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
	a.about = aboutOf(self, aboutCopy{
		mygo:     mygo.Version,
		portable: sys.Portable(),
		console:  a.console,
		settings: sys.SettingsPath,
		log:      sys.LogPath,
	})
	a.about.Project = project
	a.redraw()
}

// aboutCopy is the part of the about dialog that is not the build: which copy
// this is, where it keeps its files and which console its terminals are made
// in.
//
// A struct rather than five arguments, two of which are a lookup that can fail:
// a test fills it in without a machine to ask.
type aboutCopy struct {
	mygo     string
	portable bool
	console  consoleSource
	settings func() (string, error)
	log      func() (string, error)
}

// aboutOf is the dialog about self, in the copy it gives, apart from the window
// it is shown in.
func aboutOf(self cli.Self, copy aboutCopy) *ui.About {
	return &ui.About{
		Open:         true,
		Version:      self.Label(),
		Revision:     self.Revision,
		Mygo:         copy.mygo,
		Go:           self.Go,
		Platform:     self.Platform,
		Portable:     copy.portable,
		Terminal:     copy.console.kind,
		TerminalPath: copy.console.path,
		Settings:     pathOf(copy.settings),
		Log:          pathOf(copy.log),
	}
}

// consoleSource is where the terminals' console comes from: the kind the about
// dialog words, and the dll it is, "" for the console the machine comes with.
type consoleSource struct {
	kind string
	path string
}

// consoleOf reads what sys.ConPTYLibrary answers as the dialog words it.
func consoleOf(dll string, fromEnv bool, err error) consoleSource {
	switch {
	case err != nil:
		return consoleSource{kind: ui.TerminalInherent}
	case fromEnv:
		return consoleSource{kind: ui.TerminalEnv, path: dll}
	}
	return consoleSource{kind: ui.TerminalApp, path: dll}
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
