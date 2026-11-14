package app

import (
	"strconv"
	"strings"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/ui"
	"wtgo/locales"
)

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
	for _, p := range fileProfiles(a.settings) {
		a.editProfile(p.ID)
		if a.editing != nil {
			a.settingsPage = ui.ProfilePage(p.ID)
		}
		return
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

// splitArgs is the arguments a text holds, one on each line.
func splitArgs(text string) []string {
	var args []string
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			args = append(args, line)
		}
	}
	return args
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

// upsertProfile writes a profile into the file's list: where it is, or at the
// end.
func (a *App) upsertProfile(p store.Profile) {
	for i, q := range a.settings.Profiles {
		if q.ID == p.ID {
			a.settings.Profiles[i] = p
			return
		}
	}
	a.settings.Profiles = append(a.settings.Profiles, p)
}

// AddProfile adds a profile to the file's list, opens it in the fields of the
// profiles page, and keeps the search box of that page from hiding it.
func (a *App) AddProfile() {
	p := store.Profile{ID: a.newProfileID(), Name: locales.NewProfile()}
	a.settings.Profiles = append(a.settings.Profiles, p)
	a.profileSearch = ""
	a.settingsPage = ui.ProfilePage(p.ID)
	a.editing, a.editingFrom = &p, p
	// The page shows the arguments it holds, so a line left from the profile
	// that was open before would become this profile's arguments on save.
	a.editingArgs = ""
	a.editingGlyph = ""
	a.redraw()
}

// newProfileID is an id no profile the file holds has.
func (a *App) newProfileID() string {
	for n := 1; ; n++ {
		id := "profile-" + strconv.Itoa(n)
		if _, ok := profileByID(a.settings, id); !ok {
			return id
		}
	}
}

// RemoveProfile drops a profile from the file's list.
func (a *App) RemoveProfile(id string) {
	if p, ok := profileByID(a.settings, id); ok && p.Discovered {
		// The window finds this shell again on the next start, so hiding it
		// is the only removal that holds: dropping it would put it back.
		p.Hidden = true
		a.upsertProfile(p)
		a.settingsPage = ui.PageProfiles
		a.editNone()
		a.nextProfile(id)
		a.recording = ""
		a.redraw()
		return
	}
	out := make([]store.Profile, 0, len(a.settings.Profiles))
	for _, p := range a.settings.Profiles {
		if p.ID != id {
			out = append(out, p)
		}
	}
	a.settings.Profiles = out
	a.settingsPage = ui.PageProfiles
	a.editNone()
	a.nextProfile(id)
	a.recording = ""
	a.redraw()
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

// HideProfile hides the profile with an id, through the draft the page edits:
// the list goes on listing it, marked, and saving writes it to the file.
func (a *App) HideProfile(id string) {
	p, ok := profileByID(a.settings, id)
	if !ok || p.Hidden {
		return
	}
	p.Hidden = true
	a.upsertProfile(p)
	if a.editing != nil && a.editing.ID == id {
		a.editing.Hidden = true
	}
	a.redraw()
}

// ResetProfile puts the profile with an id back the way this machine has it:
// the probe gives it the program and the arguments of the shell it is, and the
// built-in table gives it back the name and the dot. A profile the machine no
// longer has the program for is hidden, since that is all the menu can hold,
// and a profile that is hidden comes back, since the one thing a reset can say
// about a program it found is that it is there.
func (a *App) ResetProfile(id string) {
	p, ok := profileByID(a.settings, id)
	if !ok || !p.Discovered {
		return
	}
	f, present := foundByID(found(), id)
	a.resetProfile(p, f, present)
}

// resetProfile is the reset a probe of this machine has already answered.
func (a *App) resetProfile(p store.Profile, f profile.Found, present bool) {
	if !present {
		p.Hidden = true
		a.upsertProfile(p)
		if a.editing != nil && a.editing.ID == p.ID {
			a.editing.Hidden = true
		}
		a.setNote(locales.ResetProfileGone(p.Name))
		return
	}
	p = profile.Reset(p, f)
	a.upsertProfile(p)
	if a.editing != nil && a.editing.ID == p.ID {
		*a.editing = p
		a.editingArgs = strings.Join(p.Args, "\n")
		a.editingGlyph = profile.Glyph(p.Icon)
	}
	a.setNote(locales.ResetProfileDone(p.Name))
}

// foundByID is the shell a probe found, by the id a profile has.
func foundByID(shells []profile.Found, id string) (profile.Found, bool) {
	for _, f := range shells {
		if f.ID == id {
			return f, true
		}
	}
	return profile.Found{}, false
}

// DuplicateProfile copies a profile, and shows the copy's page.
func (a *App) DuplicateProfile(id string) {
	p, ok := profileByID(a.settings, id)
	if !ok {
		return
	}
	p.ID = a.newProfileID()
	p.Name = p.Name + " " + locales.CopySuffix()
	p.Discovered = false
	// Args is a slice, so a copy that kept it would share its lines with the
	// profile it copies, and the page's edits would write both.
	p.Args = append([]string(nil), p.Args...)
	a.settings.Profiles = append(a.settings.Profiles, p)
	a.profileSearch = ""
	a.settingsPage = ui.ProfilePage(p.ID)
	a.editing, a.editingFrom = &p, p
	a.editingArgs = strings.Join(p.Args, "\n")
	a.editingGlyph = profile.Glyph(p.Icon)
	a.redraw()
}

// SetProfileIcon gives the profile the page edits a mark: a preset, or "" for
// the one its command asks for. The character box is cleared, because the
// marks of a profile are one setting, and a shape and a character cannot both
// be the one it holds.
func (a *App) SetProfileIcon(mark string) {
	if a.editing == nil {
		return
	}
	a.editing.Icon = mark
	a.editingGlyph = ""
	a.redraw()
}
