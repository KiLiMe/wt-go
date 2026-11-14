package app

import (
	"bytes"
	"strings"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/internal/ui"
	"wtgo/locales"
)

// OpenSettings shows the settings, which one tab holds: it opens that tab, or
// a new one. The terminal's overlays go: the settings tab is not a terminal, so
// a bar that asks the one in front for text has nothing left to ask.
func (a *App) OpenSettings() {
	a.closeOverlays()
	if a.settingsTab == nil {
		a.settingsTab = &pane.Tab{Name: locales.Settings()}
		a.settingsPage = ui.PageGeneral
		a.recording = ""
		a.insertTab(a.settingsTab)
	}
	for i, tab := range a.panes.Tabs {
		if tab == a.settingsTab {
			a.panes.Active = i
		}
	}
	a.redraw()
}

// insertTab puts a tab where the settings say new tabs go, and shows it.
func (a *App) insertTab(tab *pane.Tab) {
	if a.settings.NewTabPosition == "first" {
		a.panes.Tabs = append([]*pane.Tab{tab}, a.panes.Tabs...)
		a.panes.Active = 0
		return
	}
	a.panes.Add(tab)
}

// settingsState is the settings tab as the window draws it.
func (a *App) settingsState() ui.Settings {
	// The list is a menu of the machine's fonts, so it is asked for when the
	// page that shows it opens, not before the window does. The appearance
	// page shows it too, since the font setting is part of how the window looks.
	if a.settingsPage == ui.PageProfiles || a.settingsPage == ui.PageAppearance {
		a.ensureFontFamilies(a.settings.Font.Size)
	}
	return ui.Settings{
		Page:          a.settingsPage,
		KeysCache:     &a.keysDerived,
		SchemesCache:  &a.schemesDerived,
		Draft:         &a.settings,
		Dirty:         a.dirty(),
		Note:          a.settingsNote(),
		Recording:     a.recording,
		Validation:    a.validation,
		Filter:        &a.filter,
		ActionFilter:  &a.actionSearch,
		OnlyChanged:   a.onlyChangedKeys,
		KeysExpanded:  a.numberedKeys,
		Profile:       a.editing,
		Args:          &a.editingArgs,
		Glyph:         &a.editingGlyph,
		Profiles:      fileProfiles(a.settings),
		ProfileFilter: &a.profileSearch,
		Search:        &a.settingsSearch,
		Known:         knownProfiles(a.profiles),
		Device:        a.device.Name,
		Schemes:       a.schemeList(),
		FontFamilies:  a.fonts.families,
		Fonts:         a.fonts.installed,
	}
}

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
	// A family the person emptied is the generic one, which is what the window
	// draws with: the save says so in the file rather than in the window alone.
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

// closeSettingsTab forgets the settings tab, which is going away: the settings
// go back to what the file holds, and what the tab was doing -- editing a
// profile, waiting for a key, saying what the last action did -- goes with it.
// The file is read again, as discarding a draft does, so that a change the
// watcher held back while the draft was dirty is taken now. The tab stays set
// until the read is done, so a file that is still not settings does not open a
// second tab behind the caller.
func (a *App) closeSettingsTab() {
	a.settings = a.saved.Clone()
	a.note = ""
	a.validation = ""
	a.recording = ""
	a.editNone()
	a.refreshSettings()
	a.settingsTab = nil
}

// DiscardSettings goes back to the settings the file holds, and reads the file
// again: a change the watcher held back while the draft was dirty is taken now.
func (a *App) DiscardSettings() {
	a.settings = a.saved.Clone()
	a.note = ""
	a.validation = ""
	a.recording = ""
	if a.editing != nil {
		a.editProfile(a.editing.ID)
		if a.editing == nil {
			// The profile was never saved, so there is no page to edit: the
			// list is where the others are.
			a.settingsPage = ui.PageProfiles
		}
	}
	a.refreshSettings()
	a.redraw()
}

// refreshSettings reads settings.json now, as discarding a draft does, because
// the watcher may have seen a change while the draft was in the way.
func (a *App) refreshSettings() {
	data, ok, err := sys.ReadSettings()
	if err != nil {
		a.setNote(locales.ErrSettings(err.Error()))
		return
	}
	if !ok {
		a.forgetBadFile()
		return
	}
	a.reloadSafely(data)
}

// forgetBadFile drops what the settings page says about a file that is not
// settings, for a file that is gone: the window wrote none until the first
// launch, so a file the reader deleted is not something to keep explaining.
// The reason is the page's own, so nothing is shown while there is none to
// drop, which is the first launch on a machine that has no file yet.
func (a *App) forgetBadFile() bool {
	if a.configErr == "" && a.badData == nil {
		return false
	}
	a.configErr, a.badData = "", nil
	// The content the watcher remembers goes too, so that a file written back
	// to what it held before it was deleted is taken as a new file.
	a.file.forgetSeen()
	a.redraw()
	return true
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
// new. It answers whether the note changed.
func (a *App) setNote(note string) bool {
	if a.note == note {
		return false
	}
	a.note = note
	a.redraw()
	return true
}
