package store

import (
	"strconv"
	"strings"
)

// The actions a keybinding names.
const (
	ActionTabNew       = "tab.new"
	ActionTabClose     = "tab.close"
	ActionTabNext      = "tab.next"
	ActionTabPrev      = "tab.prev"
	ActionTabDuplicate = "tab.duplicate"
	// ActionTabGoto is the prefix of the actions that bring the nth tab in
	// front: "tab.goto.1" to "tab.goto.9".
	ActionTabGoto = "tab.goto."

	ActionPaneSplitAuto   = "pane.split.auto"
	ActionPaneSplitRight  = "pane.split.right"
	ActionPaneSplitDown   = "pane.split.down"
	ActionPaneFocusLeft   = "pane.focus.left"
	ActionPaneFocusRight  = "pane.focus.right"
	ActionPaneFocusUp     = "pane.focus.up"
	ActionPaneFocusDown   = "pane.focus.down"
	ActionPaneClose       = "pane.close"
	ActionPaneCloseOthers = "pane.closeOthers"

	// ActionSettings opens the settings tab.
	ActionSettings = "settings.open"
	// ActionProfile is the prefix of the actions that open a tab that runs
	// the nth profile: "profile.1" to "profile.9".
	ActionProfile = "profile."
)

// ProfileAction is the action that opens a tab running the nth profile,
// counting from one.
func ProfileAction(n int) string { return ActionProfile + strconv.Itoa(n) }

// ProfileIndex reads the profile an action opens, counting from zero.
func ProfileIndex(action string) (int, bool) {
	rest, ok := strings.CutPrefix(action, ActionProfile)
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
func GotoAction(n int) string { return ActionTabGoto + strconv.Itoa(n) }

// GotoIndex reads the tab an action brings in front, counting from zero.
func GotoIndex(action string) (int, bool) {
	rest, ok := strings.CutPrefix(action, ActionTabGoto)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	if err != nil || n < 1 {
		return 0, false
	}
	return n - 1, true
}
