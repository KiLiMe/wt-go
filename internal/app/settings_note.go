package app

import (
	"wtgo/internal/sys"
	"wtgo/locales"
)

// Boundary: what the settings page says about the file -- the note beside it,
// why it is there, and when it stops being true. The write it speaks of is
// settings_save.go.

// settingsNote is what the settings page says: what the last action did, and
// why the file is not settings while that is true. The action comes first, so a
// save, a write or a read that just failed is what the reader sees; the file's
// reason follows instead of hiding it, and stays once the action note is gone.
func (a *App) settingsNote() string {
	switch {
	case a.note == "":
		return a.configErr
	case a.configErr == "":
		return a.note
	}
	return a.note + " — " + a.configErr
}

// clearNote drops a note that is still the one that was shown.
func (a *App) clearNote(note string) {
	if a.note == note {
		a.setNote("")
	}
}

// configError is the note the settings page shows when the file is not the
// settings: what is wrong, and where the content that was wrong is kept.
func configError(data []byte, err error) string {
	if data == nil {
		return locales.ErrSettings(err.Error())
	}
	path, aerr := sys.ArchiveSettings(data)
	if aerr != nil {
		return locales.ErrSettingsInvalidKeep(err.Error(), aerr.Error())
	}
	return locales.ErrSettingsInvalid(err.Error(), path)
}

// setNote says what the settings page says, and redraws when it is something
// new.
func (a *App) setNote(note string) {
	if a.note == note {
		return
	}
	a.note = note
	a.redraw()
}

// forgetBadFile drops what the settings page says about a file that is not
// settings, for a file that is gone: the window wrote none until the first
// launch, so a file the reader deleted is not something to keep explaining.
// The reason is the page's own, so nothing is shown while there is none to
// drop, which is the first launch on a machine that has no file yet.
func (a *App) forgetBadFile() {
	if a.configErr == "" && a.badData == nil {
		return
	}
	a.configErr, a.badData = "", nil
	// The content the watcher remembers goes too, so that a file written back
	// to what it held before it was deleted is taken as a new file.
	a.file.forgetSeen()
	a.redraw()
}
