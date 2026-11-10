package app

import (
	"testing"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/layout"
	"wtgo/internal/pane"
	"wtgo/internal/store"
)

// TestLabel covers what a tab is called.
func TestLabel(t *testing.T) {
	for _, tc := range []struct {
		name, title, profile, want string
	}{
		{"build", "whatever", "PowerShell 7", "build"},
		{"", "vim main.go", "PowerShell 7", "vim main.go"},
		{"", `D:\Applications\pwsh\pwsh.exe`, "PowerShell 7", "PowerShell 7"},
		{"", `C:\Windows\System32\cmd.exe`, "Command Prompt", "Command Prompt"},
		{"", "pwsh.exe", "PowerShell 7", "PowerShell 7"},
		{"", "/usr/bin/zsh", "Shell", "Shell"},
		{"", "", "PowerShell 7", "PowerShell 7"},
	} {
		if got := label(tc.name, tc.title, tc.profile); got != tc.want {
			t.Errorf("label(%q, %q, %q) = %q, want %q", tc.name, tc.title, tc.profile, got, tc.want)
		}
	}
}

// settingsWindow is a window with a terminal tab and a settings tab holding a
// draft: a theme change, a profile being edited, a keybinding waiting for a
// key, and a note from the last action.
func settingsWindow(t *testing.T) (*App, *pane.Tab) {
	t.Helper()
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	a.panes.Add(pane.New(&terminal.Terminal{}))
	tab := &pane.Tab{Name: "Settings"}
	a.insertTab(tab)
	a.settingsTab = tab
	a.settings.Theme = "light"
	p := store.Profile{ID: "p1", Name: "P", Args: []string{"-NoLogo"}}
	a.editing, a.editingFrom = &p, p
	a.editingArgs = "-NoLogo\n--dotfiles"
	a.recording = 0
	a.note = "Saved"
	return a, tab
}

// TestCloseSettingsTabRollsBackTheDraft covers closing the settings tab: the
// question says the changes are dropped, so answering it drops the draft and
// everything the tab was doing, and cancelling keeps them.
func TestCloseSettingsTabRollsBackTheDraft(t *testing.T) {
	a, settingsTab := settingsWindow(t)

	a.CloseTab(a.panes.Index(settingsTab))
	if a.confirm == nil {
		t.Fatal("no question about the unsaved settings")
	}
	a.ConfirmNo()
	if a.settingsTab != settingsTab || a.panes.Index(settingsTab) < 0 {
		t.Fatal("the settings tab closed without an answer")
	}
	if !a.dirty() {
		t.Fatal("the draft is gone after cancelling")
	}

	a.CloseTab(a.panes.Index(settingsTab))
	if a.confirm == nil {
		t.Fatal("no question about the unsaved settings")
	}
	a.ConfirmYes()
	if a.settingsTab != nil {
		t.Error("settingsTab still set after closing it")
	}
	if a.panes.Index(settingsTab) >= 0 {
		t.Error("the settings tab is still in the strip")
	}
	if a.dirty() {
		t.Error("the settings are still dirty after closing the tab")
	}
	if a.settings.Theme != a.saved.Theme {
		t.Errorf("theme = %q, want the file's %q", a.settings.Theme, a.saved.Theme)
	}
	if a.editing != nil {
		t.Error("a profile is still open for editing")
	}
	if a.recording != -1 {
		t.Errorf("recording = %d, want none", a.recording)
	}
	if a.note != "" {
		t.Errorf("note = %q, want nothing", a.note)
	}
}

// TestCloseOtherTabsDropsTheSettingsTab covers Keep dropping the settings tab:
// the window must forget the tab, roll its draft back, and be able to open a
// clean settings tab again.
func TestCloseOtherTabsDropsTheSettingsTab(t *testing.T) {
	a, settingsTab := settingsWindow(t)
	keep := a.panes.Tabs[0]

	a.CloseOtherTabs(0)
	if a.confirm != nil {
		t.Fatal("asked about a settings tab that runs nothing")
	}
	if a.settingsTab != nil {
		t.Error("settingsTab still set after the tab was dropped")
	}
	if a.panes.Index(settingsTab) >= 0 {
		t.Error("the settings tab is still in the strip")
	}
	if len(a.panes.Tabs) != 1 || a.panes.Tabs[0] != keep {
		t.Errorf("tabs = %d, want the one kept", len(a.panes.Tabs))
	}
	if a.dirty() {
		t.Error("the settings are still dirty after the tab was dropped")
	}
	if a.editing != nil || a.recording != -1 || a.note != "" {
		t.Errorf("editing = %v, recording = %d, note = %q, want none", a.editing, a.recording, a.note)
	}

	a.OpenSettings()
	if a.settingsTab == nil {
		t.Fatal("the settings tab did not open again")
	}
	if a.settingsTab == settingsTab {
		t.Error("opened the tab that was dropped")
	}
	if a.panes.Index(a.settingsTab) < 0 {
		t.Error("the settings tab opened is not in the strip")
	}
	if a.dirty() {
		t.Error("the settings tab opened dirty")
	}
}

// TestCloseOthers covers closing every pane but the one with the keyboard:
// that pane keeps its room and its keyboard, and the others come back for
// the caller to close.
func TestCloseOthers(t *testing.T) {
	tab := pane.New(&terminal.Terminal{})
	tab.Split(layout.SideBySide, &terminal.Terminal{})
	tab.Split(layout.Stacked, &terminal.Terminal{})
	keep := tab.Focus
	dropped := closeOthers(tab)
	if len(dropped) != 2 {
		t.Fatalf("dropped %d terminals, want 2", len(dropped))
	}
	if tab.Count() != 1 || tab.Root != keep || tab.Focus != keep {
		t.Errorf("count = %d, root = %p, focus = %p, want the pane kept: %p",
			tab.Count(), tab.Root, tab.Focus, keep)
	}
	if !tab.Focus.Pane() || tab.Focus.Term == nil {
		t.Error("the pane kept has no terminal")
	}
}
