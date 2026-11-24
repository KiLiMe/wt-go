package app

import (
	"wtgo/internal/pane"
	"wtgo/internal/sys"
	"wtgo/internal/ui"
	"wtgo/locales"
)

// Boundary: the settings tab itself -- opening it, what the page draws from,
// and closing it back to what the file holds. Writing the file is
// settings_save.go; what the page says about it is settings_note.go.

// OpenSettings shows the settings, which one tab holds: it opens that tab, or
// a new one. The terminal's overlays go: the settings tab is not a terminal, so
// a bar that asks the one in front for text has nothing left to ask.
func (a *App) OpenSettings() {
	a.closeOpen()
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
	var fontDiagnostic ui.FontDiagnostic
	if a.settingsPage == ui.PageAppearance {
		a.ensureFontFamilies()
		fontDiagnostic = a.selectFont(a.settings.Font.Family, a.settings.Font.Size).note
	}
	// The settings the search found are another list once the words change, so
	// the row the arrows are on goes back to the top of it.
	if a.settingsSeen != a.settingsSearch {
		a.settingsSeen = a.settingsSearch
		a.settingsPick = 0
	}
	return ui.Settings{
		Page:           a.settingsPage,
		KeysCache:      &a.keysDerived,
		SchemesCache:   &a.schemesDerived,
		Draft:          &a.settings,
		Dirty:          a.dirty(),
		Note:           a.settingsNote(),
		Recording:      a.recording,
		Capture:        &a.capture,
		Validation:     a.validation,
		Filter:         &a.filter,
		ActionFilter:   &a.actionSearch,
		OnlyChanged:    a.onlyChangedKeys,
		KeysExpanded:   a.numberedKeys,
		Profile:        a.editing,
		Args:           &a.editingArgs,
		Glyph:          &a.editingGlyph,
		Profiles:       fileProfiles(a.settings),
		ProfileFilter:  &a.profileSearch,
		Search:         &a.settingsSearch,
		Pick:           &a.settingsPick,
		Reveal:         &a.reveal,
		Known:          knownProfiles(a.profiles),
		Device:         a.device.Name,
		Schemes:        a.schemeList(),
		FontFamilies:   a.fonts.families,
		FontDiagnostic: fontDiagnostic,
	}
}

// closeSettingsTab forgets the settings tab, which is going away: the settings
// go back to what the file holds, and what the tab was doing -- editing a
// profile, waiting for a key, saying what the last action did -- goes with it.
// The file is read again, as discarding a draft does, so that a change the
// watcher held back while the draft was dirty is taken now. The tab stays set
// until the read is done, so a file that is still not settings does not open a
// second tab behind the caller.
func (a *App) closeSettingsTab() {
	a.resetDraft()
	a.editNone()
	a.refreshSettings()
	a.settingsTab = nil
}

// DiscardSettings goes back to the settings the file holds, and reads the file
// again: a change the watcher held back while the draft was dirty is taken now.
func (a *App) DiscardSettings() {
	a.resetDraft()
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

// resetDraft goes back to the settings the file holds, and drops what the page
// was saying about the draft: what the last action did, which field it
// complained about, and the key it was waiting to record.
func (a *App) resetDraft() {
	a.settings = a.saved.Clone()
	a.note = ""
	a.validation = ""
	a.recording = ""
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
