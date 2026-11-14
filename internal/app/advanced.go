package app

import (
	"log"
	"path/filepath"
	"slices"

	"wtgo/internal/fault"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/internal/ui"
	"wtgo/locales"
)

// AskAdvanced asks the question one of the advanced page's buttons needs
// answered, doing what it says when the answer is yes.
func (a *App) AskAdvanced(kind string) {
	switch kind {
	case "rediscover":
		a.ask(ui.Confirm{
			Title: locales.ConfirmProbeTitle(),
			Text:  locales.ConfirmProbeText(),
			Yes:   locales.Probe(),
		}, a.rediscover, nil)
	case "reset":
		a.ask(ui.Confirm{
			Title: locales.ConfirmDeleteTitle(),
			Text:  locales.ConfirmDeleteText(),
			Yes:   locales.DeleteAndRestart(),
		}, a.reset, nil)
	}
}

// rediscover probes the machine and writes what it finds into the settings,
// which is what the first start did before the file said what to start.
func (a *App) rediscover() {
	f := found()
	changed := adoptAgain(&a.settings, f)
	a.found = f
	a.device = profile.Device(f)
	a.profiles = profile.Resolve(a.settings.Profiles, a.found)
	if changed {
		a.adoptWrite()
		a.note = locales.ProbedNote()
	} else {
		a.note = locales.ProbeSameNote()
	}
	a.redraw()
}

// reset drops settings.json and starts the window again, which writes the
// defaults and probes the machine as the first start did. A window that cannot
// start again stays open, so that the settings it lost are not a dead end.
func (a *App) reset() {
	if err := sys.RemoveSettings(); err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.SettingsGone, err.Error()))
		a.note = locales.ErrDeleteSettings(fault.Line(fault.SettingsGone, err.Error()))
		a.redraw()
		return
	}
	if err := a.relaunch(); err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.Relaunch, err.Error()))
		a.note = locales.ErrRelaunch(fault.Line(fault.Relaunch, err.Error()))
		a.redraw()
		return
	}
	// The settings are gone and the window that writes them is on its way:
	// the question reset asked was the one about closing, so this close does
	// not ask another.
	a.closeWindow()
}

// OpenSettingsFolder opens the directory settings.json lives in. The file
// itself needs a program the desktop has a handler for, which a Linux desktop
// need not have; a directory opens everywhere.
func (a *App) OpenSettingsFolder() {
	dir, err := sys.Dir()
	if err != nil {
		a.note = err.Error()
		a.redraw()
		return
	}
	a.openFolder(dir)
}

// OpenSettingsFile opens settings.json itself, which is what a person wants to
// edit rather than to look at the folder. A window that never wrote the file
// writes it first: a button that opens nothing would be a dead end, and the
// draft on the page is the best answer to what the file should hold.
func (a *App) OpenSettingsFile() {
	path, err := sys.SettingsPath()
	if err != nil {
		a.setNote(err.Error())
		return
	}
	if err := a.ensureSettingsFile(); err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.SettingsWrite, err.Error()))
		a.setNote(locales.ErrSettings(fault.Line(fault.SettingsWrite, err.Error())))
		return
	}
	if err := sys.OpenPath(path); err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.OpenFile, err.Error()))
		a.setNote(locales.ErrOpenFile(fault.Line(fault.OpenFile, err.Error())))
		a.redraw()
	}
}

// ensureSettingsFile writes the draft to settings.json when the window has no
// file yet, and leaves a file that is there alone: what the person opens is
// what the window shows, and a save is what writes changes.
func (a *App) ensureSettingsFile() error {
	if _, ok, err := sys.ReadSettings(); err != nil {
		return err
	} else if ok {
		return nil
	}
	data, err := store.Marshal(a.settings)
	if err != nil {
		return err
	}
	if err := sys.WriteSettings(data); err != nil {
		return err
	}
	a.saved, a.savedData = a.settings.Clone(), data
	a.file.markSeen(data)
	return nil
}

// OpenProgramFolder opens the directory this program runs from.
func (a *App) OpenProgramFolder() {
	path, err := sys.Executable()
	if err != nil {
		a.note = err.Error()
		a.redraw()
		return
	}
	a.openFolder(filepath.Dir(path))
}

// openFolder opens dir in the file manager, and says so on the page when the
// desktop has none.
func (a *App) openFolder(dir string) {
	if err := sys.OpenPath(dir); err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.OpenFolder, err.Error()))
		a.note = locales.ErrOpenFolder(fault.Line(fault.OpenFolder, err.Error()))
	}
	a.redraw()
}

// adoptAgain makes the settings name the shells the machine has now: a
// discovered profile takes the command the probe found, one whose shell is
// gone is hidden, and one that is back is shown. What a person chose, which is
// the name, the color and the directory, stays.
func adoptAgain(s *store.Settings, found []profile.Found) bool {
	byID := make(map[string]profile.Found, len(found))
	for _, f := range found {
		byID[f.ID] = f
	}
	changed := false
	for i := range s.Profiles {
		p := &s.Profiles[i]
		if !p.Discovered {
			continue
		}
		f, ok := byID[p.ID]
		if !ok {
			if !p.Hidden {
				p.Hidden = true
				changed = true
			}
			continue
		}
		if p.Command != f.Command || !slices.Equal(p.Args, f.Args) || p.Hidden {
			p.Command, p.Args, p.Hidden = f.Command, f.Args, false
			changed = true
		}
	}
	// The shells the file does not name are added; the probe, not a stat, is
	// what tells whether a program is there this time.
	if adopt(s, found, func(string) bool { return true }) {
		changed = true
	}
	return changed
}
