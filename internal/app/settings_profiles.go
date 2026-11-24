package app

import (
	"strings"

	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/ui"
)

// Boundary: the profiles of the settings page -- the list the page shows, the
// profile it edits, and the copy of that profile the draft holds. Adding,
// removing and marking one is settings_profile_edit.go; putting one back the
// way this machine has it is settings_profile_reset.go.

// fileProfiles are the profiles the file holds, as the settings list them: the
// hidden ones too, since the page edits hiding and is where it is undone.
func fileProfiles(s store.Settings) []ui.Profile {
	out := make([]ui.Profile, 0, len(s.Profiles))
	for _, p := range s.Profiles {
		out = append(out, ui.Profile{ID: p.ID, Name: p.Name, Color: p.Color, Icon: profile.IconFor(p), Discovered: p.Discovered, Hidden: p.Hidden})
	}
	return out
}

// knownProfiles are the profiles the window runs, as the settings list them. A
// hidden one is not one of them: a setting that names it runs the system shell,
// so offering it would be a choice that does not hold.
func knownProfiles(profiles []profile.Profile) []ui.Profile {
	out := make([]ui.Profile, 0, len(profiles))
	for _, p := range profiles {
		if p.Hidden {
			continue
		}
		out = append(out, ui.Profile{ID: p.ID, Name: p.Name, Color: p.Color, Icon: p.Icon, Discovered: p.Discovered})
	}
	return out
}

// SelectSettingsPage shows a page of the settings. The profiles are one page
// with the profile in front beside its list, so opening it with none in front
// opens the first one there is: the page is never a list with nothing to edit.
func (a *App) SelectSettingsPage(page string) {
	if a.settingsTab == nil {
		return
	}
	a.settingsPage = page
	a.settingsSearch = ""
	a.recording = ""
	a.editNone()
	if id, ok := ui.ProfileID(page); ok {
		a.editProfile(id)
		a.profilePageOpened()
		return
	}
	if page == ui.PageProfiles {
		a.openFirstProfile()
	}
	a.redraw()
}

// openFirstProfile opens the first profile the page lists, which is what the
// profiles page shows when it opens with none in front.
func (a *App) openFirstProfile() {
	first := fileProfiles(a.settings)
	if len(first) == 0 {
		return
	}
	a.editProfile(first[0].ID)
	if a.editing != nil {
		a.settingsPage = ui.ProfilePage(first[0].ID)
	}
}

// profilePageOpened shows the page of the profile in front, or the list alone
// when the window has no such profile: a page that names a profile the file
// does not hold edits nothing.
func (a *App) profilePageOpened() {
	if a.editing == nil {
		a.settingsPage = ui.PageProfiles
		// A profile the list holds while none can be edited, such as one the
		// search box keeps, is not the page's to open.
		a.openFirstProfile()
	}
	a.redraw()
}

// SetDefaultProfile names the profile the window opens first; "" follows the
// shell of the machine. It is what the select of the page asks for while the
// profile the file names is not one the page offers: the select shows the
// device, and a click on it writes the choice down instead of the page doing it.
func (a *App) SetDefaultProfile(id string) {
	a.settings.DefaultProfile = id
	a.redraw()
}

// editProfile opens a copy of a profile for the profile page to edit: the page
// leaves the file alone until the settings are saved, so that looking at a
// profile writes nothing.
func (a *App) editProfile(id string) {
	p, ok := profileByID(a.settings, id)
	if !ok {
		a.editNone()
		return
	}
	a.editing, a.editingFrom = &p, p
	a.editingArgs = strings.Join(p.Args, "\n")
	a.editingGlyph = profile.Glyph(p.Icon)
}

// editNone closes the profile page's copy: nothing is being edited.
func (a *App) editNone() {
	a.editing, a.editingFrom, a.editingArgs, a.editingGlyph = nil, store.Profile{}, "", ""
}

// profileByID is the profile the file holds by an id.
func profileByID(s store.Settings, id string) (store.Profile, bool) {
	for _, p := range s.Profiles {
		if p.ID == id {
			return p, true
		}
	}
	return store.Profile{}, false
}

// nextProfile opens the profile the list now starts at, once the one that was
// in front is gone or hidden from the list: the fields beside the list show
// another profile rather than go empty. A profile that moved out of a list
// the search box keeps is the box's to need, not the page's.
func (a *App) nextProfile(gone string) {
	a.openFirstProfile()
	if a.editing != nil && a.editing.ID == gone {
		a.editNone()
		a.settingsPage = ui.PageProfiles
	}
}
