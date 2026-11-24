package ui

import (
	"wtgo/internal/pane"
)

// Boundary: the whole of what the window does, as ui asks it. Nothing here
// draws; app implements it, and this is the only way the two meet.

// Actions are what the window does, as the interface asks it to.
type Actions interface {
	// NewTab opens a tab that runs a profile; "" runs the default one.
	NewTab(profileID string)
	// NewSystemTab opens a tab that runs the shell of the machine, which is
	// what a profile with no program of its own is (ADR-44).
	NewSystemTab()
	// CloseTab closes the tab at i.
	CloseTab(i int)
	// DuplicateTab opens a tab like the one at i.
	DuplicateTab(i int)
	// CloseOtherTabs closes every tab but the one at i.
	CloseOtherTabs(i int)
	// MoveTab moves the tab at from to to.
	MoveTab(from, to int)
	// FocusTab brings the tab at i in front.
	FocusTab(i int)
	// ScrollTabs moves the tabs the strip shows by by DIPs.
	ScrollTabs(by float32)
	// OpenTabs lists every tab, which is how the strip names the ones it has
	// no room to show.
	OpenTabs()
	// MinimizeWindow minimizes the window.
	MinimizeWindow()
	// ToggleMaximizeWindow maximizes the window, or restores it.
	ToggleMaximizeWindow()
	// CloseWindow closes the window, as its close button would.
	CloseWindow()
	// BeginRename opens the dialog that renames the tab at i.
	BeginRename(i int)
	// RenameTab gives the tab at i a name; "" keeps the one it had.
	RenameTab(i int, name string)
	// CancelRename closes the rename dialog.
	CancelRename()
	// FocusPane gives the keyboard to a pane of the tab at i.
	FocusPane(i int, n *pane.Node)
	// ResizePane moves the line of a split of the tab at i by delta DIPs.
	ResizePane(i int, d pane.Divide, delta float32)
	// OpenSettings shows the settings tab.
	OpenSettings()
	// SelectSettingsPage shows a page of the settings.
	SelectSettingsPage(page string)
	// RevealSetting shows the page a search found a setting on, and marks the
	// setting itself, which the page scrolls into view: a search over the
	// settings is a way to a setting, not a way to the page it lives on.
	RevealSetting(page, key string)
	// SaveSettings writes the settings to the file.
	SaveSettings()
	// SetDefaultProfile opens the profile with an id in the tab the window
	// opens first, or the shell of the machine for "": the choice stands for
	// the profile the file names that the page no longer offers.
	SetDefaultProfile(id string)
	// DiscardSettings goes back to the settings the file holds.
	DiscardSettings()
	// OpenSettingsFolder opens the folder settings.json lives in.
	OpenSettingsFolder()
	// OpenSettingsFile opens settings.json itself, writing it first when the
	// window has none to show.
	OpenSettingsFile()
	// OpenProgramFolder opens the folder the program runs from.
	OpenProgramFolder()
	// OpenAbout shows the dialog that says what the program is.
	OpenAbout()
	// CloseAbout closes the about dialog.
	CloseAbout()
	// OpenProjectPage opens the page of the program in the browser.
	OpenProjectPage()
	// AskAdvanced asks the question a button of the advanced page needs
	// answered, which is AskRediscover or AskReset.
	AskAdvanced(kind string)
	// RecordKeybinding waits for a key of the action with an id.
	RecordKeybinding(action string)
	// SetKeybinding gives the action with an id a key; "" stops recording.
	SetKeybinding(action, key string)
	// RemoveKeybinding drops the keybindings of the action with an id.
	RemoveKeybinding(action string)
	// ResetKeybinding gives the action with an id the keys the catalog
	// starts it on.
	ResetKeybinding(action string)
	// ToggleChangedKeys lists the actions whose keys are not the catalog's,
	// or every action again.
	ToggleChangedKeys()
	// ToggleNumberedKeys unfolds the numbered actions, or folds them.
	ToggleNumberedKeys()
	// ResetKeybindings goes back to the keybindings the window starts with.
	ResetKeybindings()
	// PickScheme paints with the color scheme the settings page names. The
	// page asks for it rather than writing the draft itself, so that the
	// window applies the new scheme as it does any other change.
	PickScheme(name string)
	// CopyScheme makes a color scheme of the settings from the one the
	// terminal paints with, and picks it.
	CopyScheme()
	// RemoveScheme drops the color scheme the settings write by a name.
	RemoveScheme(name string)
	// ConfirmYes does what the question asks.
	ConfirmYes()
	// ConfirmNo drops the question, doing nothing.
	ConfirmNo()
	// AddProfile adds a profile, and shows its page.
	AddProfile()
	// RemoveProfile drops the profile with an id, or hides the one the window
	// found on this machine: the next start adds it back, so hiding is all a
	// person can do to a shell that is there.
	RemoveProfile(id string)
	// HideProfile hides the profile with an id, through the draft the page
	// edits: the list goes on listing it, marked, and saving writes it.
	HideProfile(id string)
	// DuplicateProfile copies the profile with an id, and shows its page.
	DuplicateProfile(id string)
	// ResetProfile puts the profile with an id back the way this machine has
	// it now: the shell the probe finds, and the name and dot of the
	// built-in table.
	ResetProfile(id string)
	// SetProfileIcon gives the profile the page edits a mark: a preset such
	// as "pwsh7:official", "" for the one its command asks for, which is what
	// the character box clears.
	SetProfileIcon(mark string)
	// Run does the action with an id, as the palette picked it.
	Run(action string)
	// Find searches the terminal in front for text, as it is typed.
	Find(text string)
	// FindNext picks the match after the one selected.
	FindNext()
	// FindPrevious picks the match before the one selected.
	FindPrevious()
	// EndFind closes the find bar, and stops the search.
	EndFind()
}
