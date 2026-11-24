package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
)

// Boundary: what the pages draw from -- the settings page, its color
// schemes, its font diagnostics. The drawing is settings.go, schemes.go and
// fonts.go.

// Scheme is a color scheme the color schemes page lists: its name, whether the
// settings write it, which is what lets the page remove it, and the colors it
// paints a terminal with, which the page shows as a chip beside the name.
type Scheme struct {
	Name   string
	Custom bool
	Bg     myui.Color
	Fg     myui.Color
}

// FontDiagnostic describes how the terminal resolves the draft font list.
type FontDiagnostic struct {
	Primary        string
	Proportional   string
	Missing        []string
	SystemFallback bool
}

// Settings is the tab that edits the settings.
type Settings struct {
	// Page is the page the left column shows.
	Page string
	// KeysCache is what the actions page derived from the file, and
	// SchemesCache what the color schemes page did: the window keeps one of
	// each from frame to frame, and a page drawn without one keeps nothing
	// past its frame. They hold no settings, only the answers to a file that
	// has not changed.
	KeysCache    *PageKeybindings
	SchemesCache *SchemesCache
	// Draft is the settings the form edits, and the window applies as it
	// goes.
	Draft *store.Settings
	// Dirty is whether the draft holds something the file does not.
	Dirty bool
	// Note says what the last save did, or why the file is not the
	// settings.
	Note string
	// Validation is why the last save refused the draft, "" while the
	// draft is one the window can run with. It stays while the note goes,
	// so that what a person clicks away from does not hide what is wrong.
	Validation string
	// Recording is the action the window waits for a key of, "" while it
	// waits for none.
	Recording string
	// Capture is where the actions page keeps the keys it listens for while an
	// action is being recorded, so that they are made once instead of on every
	// frame of the recording. A page drawn without one makes them for its own
	// frame.
	Capture *Capture
	// Filter is what the color schemes page searches for.
	Filter *string
	// ActionFilter is what the actions page searches for: an action's name,
	// one of its keys, or its id.
	ActionFilter *string
	// OnlyChanged is whether the actions page lists only the actions whose
	// keys are not the ones the catalog starts them on.
	OnlyChanged bool
	// KeysExpanded is whether the actions page lists the numbered actions
	// past the third, which it folds away to keep the list short.
	KeysExpanded bool
	// Profiles are the profiles the file holds, which the column lists, the
	// hidden ones too: a list that leaves them out leaves out the only place
	// they come back from.
	Profiles []Profile
	// ProfileFilter is what the profiles page searches its list of profiles
	// for. A page drawn without it leaves the box out rather than give it
	// nowhere to write what is typed.
	ProfileFilter *string
	// Known are the profiles a setting may name, the file's and the ones
	// the window found.
	Known []Profile
	// Device is the shell of this machine that following the device runs, by
	// the name the table gives it, and "" when the machine has no shell.
	Device string
	// Profile is the profile page's copy of the profile it edits, nil on
	// the other pages, and Args its arguments as the page types them, one
	// on a line.
	Profile *store.Profile
	Args    *string
	// Glyph is the character the profile page's mark box holds, which is the
	// profile's mark while it is a character and "" while it is a preset: the
	// box shows one kind of mark, and the marks are one field of the file.
	Glyph *string
	// Schemes are the color schemes the color schemes page lists.
	Schemes []Scheme
	// FontFamilies are the families the font setting offers: what the machine
	// advertises as monospaced, narrowed to the ones that measure as it once the
	// census has run.
	FontFamilies   []string
	FontDiagnostic FontDiagnostic
	// Search is what the box over the column of pages searches the settings
	// for. A page drawn without it lists the pages alone rather than give the
	// box nowhere to write what is typed.
	Search *string
	// Pick is the row of what the search found that the arrows are on, which
	// Enter opens: the window keeps it, so that it stays while the box has the
	// keys and a click on a row leaves it where the click put it.
	Pick *int
	// Reveal is the setting the search found and how far the mark over it has
	// faded, nil when the search found nothing to mark.
	Reveal *Reveal
}
