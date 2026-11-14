// Package ui draws the window: the tab strip, which is also the title bar,
// and the panes of the tab in front. It draws what the window hands it and
// asks the window to act through Actions: it reads no settings and starts no
// process.
package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/layout"
	"wtgo/internal/pane"
	"wtgo/internal/store"
)

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

// WindowButtons are the window's buttons where the window draws them itself:
// "minimize", "maximize" and "close", in the desktop's order, at the start
// and at the end of the title bar. The desktop's layout names them, and a
// desktop that names none, as tiling window managers do, gets none.
type WindowButtons struct {
	Start, End []string
	// Maximized is whether the window is maximized, which the button that
	// maximizes draws as the one that restores it.
	Maximized bool
}

// Tab is a tab as the window draws it.
type Tab struct {
	// Name is what the strip calls the tab.
	Name string
	// Color is the color of the dot of the profile the tab runs, when it
	// has one.
	Color string
	// Icon is the mark the tab wears: a preset such as "pwsh7:official",
	// one character, or "" for the plain terminal mark in the profile's
	// color.
	Icon string
	// Panes are the panes the tab holds.
	Panes *pane.Tab
	// Mount shows the terminal of a pane. A terminal shows in one view at a
	// time, so only the tab in front mounts its own.
	Mount func(c *myui.Context, n *pane.Node) myui.Element
	// Page draws a tab that is not a terminal, the settings. It is nil for
	// the tabs that hold panes.
	Page func(c *myui.Context) myui.Element
}

// Profile is a profile in the menu that opens tabs.
type Profile struct {
	ID    string
	Name  string
	Color string
	// Icon is the mark the tab of this profile wears, as Tab.Icon has it.
	Icon string
	// Discovered is whether the window found this profile on the machine
	// rather than a person writing it, which is what the settings page calls
	// a built-in one.
	Discovered bool
	// Hidden is whether the settings page hides it, which is all a removal
	// can do to a profile the window finds again.
	Hidden bool
}

// Scheme is a color scheme the color schemes page lists: its name, whether the
// settings write it, which is what lets the page remove it, and the colors it
// paints a terminal with, which the page shows as a chip beside the name.
type Scheme struct {
	Name   string
	Custom bool
	Bg     myui.Color
	Fg     myui.Color
}

// Rename is the dialog that renames a tab.
type Rename struct {
	// Open is whether the dialog shows; the dialog closes it.
	Open bool
	// Tab is the tab the dialog renames.
	Tab int
	// Text is what the field holds.
	Text string
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
	// FontFamilies are the installed families the terminal can draw with,
	// which the font setting lists.
	FontFamilies []string
	// Search is what the box over the column of pages searches the settings
	// for. A page drawn without it lists the pages alone rather than give the
	// box nowhere to write what is typed.
	Search *string
}

// About is the dialog that says what the program is: the version it runs, what
// it was built from, what it runs on and the files it leaves, and what it is
// licensed under.
type About struct {
	// Open is whether the dialog shows; the dialog closes it.
	Open bool
	// Version is the build as `wt-go version` says it, such as
	// "v0.3.1-dev (packaged 0.3.0)", and Revision the commit it was built
	// from, "" in a build with no history.
	Version  string
	Revision string
	// Go and Platform are what it runs on.
	Go       string
	Platform string
	// Mygo is the mygo version the window and the terminals are built on:
	// which build of the framework a bug report is about matters as much as
	// which build of the program.
	Mygo string
	// Settings and Log are where settings.json and wt-go.log are, the first
	// thing `wt-go diag` is asked about: a person sent to look at a file by a
	// bug report should not have to open a console to find it.
	Settings string
	Log      string
	// Project is the address of the program's page, which the dialog opens.
	Project string
	// Note is what went wrong the last time the dialog opened the page, which
	// it says in place of the silent button a desktop without a browser
	// leaves.
	Note string
}

// Confirm is the question the window asks before it does something it cannot
// take back, as closing the settings with changes it has not saved.
type Confirm struct {
	// Open is false once the dialog is dismissed, which the window takes
	// as no.
	Open  bool
	Title string
	Text  string
	Yes   string
	No    string
}

// Face is the surface a pane paints: the color behind the tab in front and the
// one text reads on there. The strip around that tab is the window's, in the
// theme's colors, so a window whose theme is not the terminal's still joins the
// tab that is in front to the pane below it.
type Face struct {
	Background myui.Color
	Foreground myui.Color
}

// State is what the window shows this frame.
type State struct {
	// Tabs are the tabs, in order.
	Tabs []Tab
	// Active is the tab in front.
	Active int
	// Pane is the face of the pane below the strip, which the tab in front
	// merges with. The zero Face takes the theme's own background instead,
	// for a window that draws no pane there.
	Pane Face
	// Profiles are the ones the menu that opens tabs lists.
	Profiles []Profile
	// Device is the shell following the device runs, by the name the table
	// gives it, and "" when the machine has no shell.
	Device string
	// Rename is the rename dialog, nil while it is closed.
	Rename *Rename
	// Settings is the settings tab, nil until it opens.
	Settings *Settings
	// Confirm is the question the window asks, nil while it asks none.
	Confirm *Confirm
	// About is the dialog that says what the program is, nil while it is
	// closed.
	About *About
	// Palette is the command palette, nil while it is not open.
	Palette *Palette
	// Find is the bar that asks the terminal for text, nil while it is
	// closed.
	Find *Find
	// EqualTabs makes every tab as wide as the others, rather than as wide
	// as its name.
	EqualTabs bool
	// TabsScroll is where the strip shows its tabs, which the strip scrolls
	// when they outgrow the title bar. nil shows them all, unscrolled.
	TabsScroll *myui.ScrollState
	// RevealTab asks the strip to bring the tab in front into view in this
	// frame, which the window asks once as it changes that tab.
	RevealTab bool
	// WindowButtons are the buttons the strip draws for a window that hides
	// the desktop's, nil while the desktop draws them, or draws a title bar
	// of its own.
	WindowButtons *WindowButtons
	// Shortcut is the key the settings bind an action to, as a keybinding
	// writes it, and "" when none binds it. nil shows no shortcut.
	Shortcut func(action string) string
}

// Content is the room the panes take: the window, less the strip it draws
// above them.
func Content(c *myui.Context) layout.Rect {
	w, h := c.Size()
	return layout.Rect{Y: TitleBarHeight, W: w, H: max(h-TitleBarHeight, 0)}
}

// View is the window: the tab strip, which is also the title bar, and the
// panes of the tab in front.
func View(c *myui.Context, st State, act Actions) {
	myui.Column(c).Fill().Children(func() {
		strip(c, st, act)
		content(c, st, act, Content(c))
	})
	rename(c, st, act)
	confirm(c, st, act)
	about(c, st, act)
	findBar(c, st, act)
	palette(c, st, act)
}
