package app

import (
	"strings"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/locales"
)

// Boundary: putting a profile back the way a probe of this machine found it,
// and the probe's shell for a profile. The rest of the page is
// settings_profiles.go.

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
