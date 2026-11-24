package app

import (
	"strconv"
	"strings"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/ui"
	"wtgo/locales"
)

// Boundary: changing the file's list of profiles from the settings page --
// adding one, writing it back, removing it, hiding it, copying it, and the id a
// new one gets. Which profile the page edits is settings_profiles.go.

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
