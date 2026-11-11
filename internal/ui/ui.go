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
	// DiscardSettings goes back to the settings the file holds.
	DiscardSettings()
	// OpenSettingsFile opens the file in the program that edits it.
	OpenSettingsFile()
	// RecordKeybinding waits for a key of the action with an id.
	RecordKeybinding(action string)
	// SetKeybinding gives the action with an id a key; "" stops recording.
	SetKeybinding(action, key string)
	// RemoveKeybinding drops the keybindings of the action with an id.
	RemoveKeybinding(action string)
	// ResetKeybindings goes back to the keybindings the window starts with.
	ResetKeybindings()
	// CopyScheme makes a color scheme of the settings from the one the
	// terminal paints with, and picks it.
	CopyScheme()
	// ConfirmYes does what the question asks.
	ConfirmYes()
	// ConfirmNo drops the question, doing nothing.
	ConfirmNo()
	// AddProfile adds a profile, and shows its page.
	AddProfile()
	// RemoveProfile drops the profile with an id.
	RemoveProfile(id string)
	// DuplicateProfile copies the profile with an id, and shows its page.
	DuplicateProfile(id string)
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

// Tab is a tab as the window draws it.
type Tab struct {
	// Name is what the strip calls the tab.
	Name string
	// Color is the color of the dot of the profile the tab runs, when it
	// has one.
	Color string
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
	// Draft is the settings the form edits, and the window applies as it
	// goes.
	Draft *store.Settings
	// Dirty is whether the draft holds something the file does not.
	Dirty bool
	// Note says what the last save did, or why the file is not the
	// settings.
	Note string
	// Recording is the action the window waits for a key of, "" while it
	// waits for none.
	Recording string
	// Filter is what the color schemes page searches for.
	Filter *string
	// Profiles are the profiles the file holds, which the column lists.
	Profiles []Profile
	// Known are the profiles a setting may name, the file's and the ones
	// the window found.
	Known []Profile
	// Profile is the profile page's copy of the profile it edits, nil on
	// the other pages, and Args its arguments as the page types them, one
	// on a line.
	Profile *store.Profile
	Args    *string
	// Schemes are the color schemes the color schemes page lists.
	Schemes []string
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

// State is what the window shows this frame.
type State struct {
	// Tabs are the tabs, in order.
	Tabs []Tab
	// Active is the tab in front.
	Active int
	// Profiles are the ones the menu that opens tabs lists.
	Profiles []Profile
	// Rename is the rename dialog, nil while it is closed.
	Rename *Rename
	// Settings is the settings tab, nil until it opens.
	Settings *Settings
	// Confirm is the question the window asks, nil while it asks none.
	Confirm *Confirm
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
	findBar(c, st, act)
	palette(c, st, act)
}
