// Package action names what the window does: the actions a keybinding binds,
// the menus offer and the palette lists, their titles, the group the settings
// show them under, and the keys they start on. It is the one place the actions
// are spelled, so that a keybinding, a menu and the palette cannot drift apart.
package action

import (
	"fmt"
	"strconv"
	"strings"

	"wtgo/internal/i18n"
)

// The actions.
const (
	TabNew       = "tab.new"
	TabClose     = "tab.close"
	TabNext      = "tab.next"
	TabPrev      = "tab.prev"
	TabDuplicate = "tab.duplicate"
	// TabGoto is the prefix of the actions that bring the nth tab in front:
	// "tab.goto.1" to "tab.goto.9".
	TabGoto = "tab.goto."

	PaneSplitAuto   = "pane.split.auto"
	PaneSplitRight  = "pane.split.right"
	PaneSplitDown   = "pane.split.down"
	PaneFocusLeft   = "pane.focus.left"
	PaneFocusRight  = "pane.focus.right"
	PaneFocusUp     = "pane.focus.up"
	PaneFocusDown   = "pane.focus.down"
	PaneClose       = "pane.close"
	PaneCloseOthers = "pane.closeOthers"

	// Find asks a terminal for text.
	Find = "find.open"
	// Palette lists the actions.
	Palette = "palette.open"
	// Settings opens the settings tab.
	Settings = "settings.open"
	// Profile is the prefix of the actions that open a tab that runs the
	// nth profile: "profile.1" to "profile.9".
	Profile = "profile."
)

// The groups the settings page lists the actions under.
const (
	GroupTab      = "Tab"
	GroupPane     = "Pane"
	GroupProfiles = "Profiles"
	GroupFind     = "Find"
	GroupWindow   = "Window"
)

// MaxNumber is the highest number a numbered action names.
const MaxNumber = 9

// Action is one thing the window does.
type Action struct {
	// ID is what a keybinding names and the palette runs.
	ID string
	// Title is what the action is called, with %d where a numbered action
	// takes its number; Title reads it in the window's language.
	Title string
	// Group is what the settings page lists the action under.
	Group string
	// Keys are the keys the action starts on, the first the one the menus
	// show. A keybinding the file holds replaces them.
	Keys []string
}

// fixed are the actions that are not numbered, in the order the settings list
// them, and the keys they start on.
var fixed = []Action{
	{ID: TabNew, Title: "New Tab", Group: GroupTab, Keys: []string{"ctrl+shift+t", "ctrl+t"}},
	{ID: TabNext, Title: "Next Tab", Group: GroupTab, Keys: []string{"ctrl+tab"}},
	{ID: TabPrev, Title: "Previous Tab", Group: GroupTab, Keys: []string{"ctrl+shift+tab"}},
	{ID: TabDuplicate, Title: "Duplicate Tab", Group: GroupTab, Keys: []string{"ctrl+shift+d"}},
	{ID: TabClose, Title: "Close Tab", Group: GroupTab},

	{ID: PaneSplitAuto, Title: "Split Pane", Group: GroupPane, Keys: []string{"alt+shift+d"}},
	{ID: PaneSplitRight, Title: "Split Right", Group: GroupPane, Keys: []string{"alt+shift+right"}},
	{ID: PaneSplitDown, Title: "Split Down", Group: GroupPane, Keys: []string{"alt+shift+down"}},
	{ID: PaneFocusLeft, Title: "Focus Left Pane", Group: GroupPane, Keys: []string{"alt+left"}},
	{ID: PaneFocusRight, Title: "Focus Right Pane", Group: GroupPane, Keys: []string{"alt+right"}},
	{ID: PaneFocusUp, Title: "Focus Up Pane", Group: GroupPane, Keys: []string{"alt+up"}},
	{ID: PaneFocusDown, Title: "Focus Down Pane", Group: GroupPane, Keys: []string{"alt+down"}},
	{ID: PaneClose, Title: "Close Pane", Group: GroupPane, Keys: []string{"ctrl+shift+w"}},
	{ID: PaneCloseOthers, Title: "Close Other Panes", Group: GroupPane, Keys: []string{"alt+shift+w"}},

	{ID: Find, Title: "Find", Group: GroupFind, Keys: []string{"ctrl+shift+f"}},

	{ID: Settings, Title: "Settings", Group: GroupWindow, Keys: []string{"ctrl+comma"}},
	{ID: Palette, Title: "Command Palette", Group: GroupWindow, Keys: []string{"ctrl+shift+p"}},
}

// all is every action: the fixed ones, and the numbered ones that bring the
// nth tab in front or open a tab that runs the nth profile.
var all = func() []Action {
	out := append([]Action{}, fixed...)
	for n := 1; n <= MaxNumber; n++ {
		out = append(out, Action{
			ID:    GotoAction(n),
			Title: "Switch to Tab %d",
			Group: GroupTab,
			Keys:  []string{"ctrl+" + strconv.Itoa(n)},
		})
	}
	for n := 1; n <= MaxNumber; n++ {
		out = append(out, Action{
			ID:    ProfileAction(n),
			Title: "Open Profile %d",
			Group: GroupProfiles,
			Keys:  []string{"ctrl+shift+" + strconv.Itoa(n)},
		})
	}
	return out
}()

// All is every action the window has.
func All() []Action { return all }

// Lookup finds an action by its id, and reports whether the window has one.
func Lookup(id string) (Action, bool) {
	for _, a := range all {
		if a.ID == id {
			return a, true
		}
	}
	return Action{}, false
}

// Title is what an action is called in the window's language, and the id
// itself for an action the window does not know, which a hand-written
// keybinding may name.
func Title(id string) string {
	a, ok := Lookup(id)
	if !ok {
		return id
	}
	if n, ok := number(id); ok {
		return fmt.Sprintf(i18n.T(a.Title), n)
	}
	return i18n.T(a.Title)
}

// Group is what the settings page lists an action under, and "" for an action
// the window does not know, which it lists last.
func Group(id string) string {
	a, _ := Lookup(id)
	return a.Group
}

// number is the number an action that ends in one names, counting from one, as
// its title says it.
func number(id string) (int, bool) {
	if n, ok := GotoIndex(id); ok {
		return n + 1, true
	}
	if n, ok := ProfileIndex(id); ok {
		return n + 1, true
	}
	return 0, false
}

// ProfileAction is the action that opens a tab running the nth profile,
// counting from one.
func ProfileAction(n int) string { return Profile + strconv.Itoa(n) }

// ProfileIndex reads the profile an action opens, counting from zero.
func ProfileIndex(id string) (int, bool) {
	rest, ok := strings.CutPrefix(id, Profile)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	if err != nil || n < 1 {
		return 0, false
	}
	return n - 1, true
}

// GotoAction is the action that brings the nth tab in front, counting from
// one.
func GotoAction(n int) string { return TabGoto + strconv.Itoa(n) }

// GotoIndex reads the tab an action brings in front, counting from zero.
func GotoIndex(id string) (int, bool) {
	rest, ok := strings.CutPrefix(id, TabGoto)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	if err != nil || n < 1 {
		return 0, false
	}
	return n - 1, true
}
