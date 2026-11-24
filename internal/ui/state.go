package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/pane"
)

// Boundary: what a frame is made of, and nothing about how it is drawn.
// Each part is drawn by the file named after it (tabs.go, panes.go, ...).

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
	// TabsList is the list of every tab the strip opens when it has no room
	// for them, nil while it is not open.
	TabsList *TabsList
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
	// Frames is the readout of the rate the window draws at, nil while it is
	// closed.
	Frames *Frames
}
