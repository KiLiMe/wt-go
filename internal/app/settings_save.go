package app

import (
	"bytes"
	"strings"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/locales"
)

// Boundary: writing settings.json from the settings page, and whether it is
// worth writing at all -- a draft that says what the file already says is not.

// SaveSettings writes the settings to the file, and keeps the copy the window
// compares against to know what is unsaved. It refuses to write over a file
// that changed on disk since the window last read or saved it, and refuses a
// draft no window can run with, which it says where the page is looking.
func (a *App) SaveSettings() {
	if a.editing != nil {
		a.editing.Args = splitArgs(a.editingArgs)
		// The box is the mark while it holds a character, and clearing it is
		// a profile that wears none: a shape it kept, or the one its command
		// asks for.
		if g := profile.Glyph(a.editingGlyph); g != "" {
			a.editing.Icon = g
		} else if profile.IconKey(a.editing.Icon) == "" {
			a.editing.Icon = ""
		}
		a.upsertProfile(*a.editing)
	}
	// A family the person emptied is the default families, which is what the
	// window draws with: the save says so in the file rather than in the window
	// alone.
	a.settings.FillEmptyFont()
	if err := a.settings.Validate(); err != nil {
		// The draft stays as it is: what is wrong with it is something the
		// person corrects on the page, frame after frame.
		a.validation = err.Error()
		a.setNote(err.Error())
		return
	}
	a.validation = ""
	data, err := store.Marshal(a.settings)
	if err != nil {
		a.setNote(err.Error())
		return
	}
	unchanged, err := a.settingsFileUnchanged()
	if err != nil {
		a.setNote(locales.ErrSettings(err.Error()))
		return
	}
	if !unchanged {
		a.setNote(locales.SettingsChangedNote())
		return
	}
	if err := sys.WriteSettings(data); err != nil {
		a.setNote(err.Error())
		return
	}
	a.saved, a.savedData, a.badData = a.settings.Clone(), data, nil
	if a.editing != nil {
		a.editingFrom = *a.editing
	}
	a.configErr = ""
	a.file.markSeen(data)
	a.setNote(locales.Saved())
	a.redraw()
}

// settingsFileUnchanged reports whether settings.json still holds what the
// window last loaded or saved, or the bad content it last reported. A file
// that is gone is not in the way; a file with other content is.
func (a *App) settingsFileUnchanged() (bool, error) {
	data, ok, err := sys.ReadSettings()
	if err != nil {
		return false, err
	}
	if !ok {
		return true, nil
	}
	if a.savedData != nil && bytes.Equal(data, a.savedData) {
		return true, nil
	}
	if a.badData != nil && bytes.Equal(data, a.badData) {
		return true, nil
	}
	return false, nil
}

// dirty reports whether the settings hold something the file does not: a
// setting it has not seen, or a profile the page changed.
func (a *App) dirty() bool {
	if !store.Equal(a.settings, a.saved) {
		return true
	}
	if a.editing == nil {
		return false
	}
	if a.editingArgs != strings.Join(a.editingFrom.Args, "\n") {
		return true
	}
	if profile.Glyph(a.editingGlyph) != profile.Glyph(a.editingFrom.Icon) {
		return true
	}
	return !store.EqualProfile(*a.editing, a.editingFrom)
}
