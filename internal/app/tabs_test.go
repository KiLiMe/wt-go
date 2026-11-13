package app

import (
	"testing"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/action"
	"wtgo/internal/i18n"
	"wtgo/internal/layout"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/store"
)

// TestLabel covers what a tab is called.
func TestLabel(t *testing.T) {
	for _, tc := range []struct {
		name, title, profile, runs, want string
	}{
		{"build", "whatever", "PowerShell 7", "pwsh", "build"},
		{"", "vim main.go", "PowerShell 7", "pwsh", "vim main.go"},
		{"", `D:\Applications\pwsh\pwsh.exe`, "PowerShell 7", "pwsh", "PowerShell 7"},
		{"", `C:\Windows\System32\cmd.exe`, "Command Prompt", "pwsh", "Command Prompt"},
		{"", "pwsh.exe", "PowerShell 7", "pwsh", "PowerShell 7"},
		{"", "/usr/bin/zsh", "Shell", "pwsh", "Shell"},
		{"", "", "PowerShell 7", "pwsh", "PowerShell 7"},
		// A tab that runs no named profile runs the system's own shell, and
		// says which one it got when that shell left its path as the title.
		{"", `D:\Applications\pwsh\pwsh.exe`, "", "bash", "pwsh"},
		{"", `C:\Windows\System32\cmd.exe`, "", "bash", "cmd"},
		{"", "/usr/bin/zsh", "", "bash", "zsh"},
		// Before any title, the tab says which shell it started: an empty
		// strip with a close button on it is what the word below was for.
		{"", "", "", "bash", "bash"},
		{"", "", "", "pwsh", "pwsh"},
		{"", "", "", "", i18n.T("systemProfile")},
	} {
		if got := label(tc.name, tc.title, tc.profile, tc.runs); got != tc.want {
			t.Errorf("label(%q, %q, %q, %q) = %q, want %q",
				tc.name, tc.title, tc.profile, tc.runs, got, tc.want)
		}
	}
}

// TestTabNamesTheShellItRuns covers the tab itself: it says the shell before
// that shell has a title, and the title once it has one.
func TestTabNamesTheShellItRuns(t *testing.T) {
	a := &App{}
	a.settings.TitleFromActiveTerminal = true
	tab := pane.New(nil)
	tab.Runs = "bash"
	if got := a.tabName(tab); got != "bash" {
		t.Errorf("tabName() = %q, want bash", got)
	}
	tab.ProfileName = "Ubuntu"
	if got := a.tabName(tab); got != "Ubuntu" {
		t.Errorf("tabName() = %q with a profile name, want Ubuntu", got)
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
	a.recording = action.TabClose
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
	if a.recording != "" {
		t.Errorf("recording = %q, want none", a.recording)
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
	if a.editing != nil || a.recording != "" || a.note != "" {
		t.Errorf("editing = %v, recording = %q, note = %q, want none", a.editing, a.recording, a.note)
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

// TestTabProfile covers what a tab opened without a profile of its own runs:
// the default profile, the way the click on the strip's '+' and the keybinding
// for a new tab both ask for it, and the shell of the machine when that names
// nothing the machine has.
func TestTabProfile(t *testing.T) {
	profiles := []profile.Profile{
		{ID: "pwsh7", Name: "PowerShell 7"},
		{ID: "zsh", Name: "zsh"},
		{ID: "gone", Name: "Gone", Hidden: true},
	}
	found := []profile.Found{{ID: "pwsh7", Command: `C:\Program Files\PowerShell\7\pwsh.exe`}}
	for _, tc := range []struct {
		name, def, id, want string
	}{
		{"theDefaultProfile", "pwsh7", "", "pwsh7"},
		{"theProfileItNames", "pwsh7", "zsh", "zsh"},
		{"noDefaultProfile", "", "", "pwsh7"},
		{"anIdTheMachineLacks", "pwsh7", "nope", "pwsh7"},
		{"aProfileSinceHidden", "gone", "", "pwsh7"},
	} {
		a := &App{profiles: profiles, device: profile.Device(found)}
		a.settings.DefaultProfile = tc.def
		if got := a.tabProfile(tc.id); got.ID != tc.want {
			t.Errorf("%s: tabProfile(%q) = %q, want %q", tc.name, tc.id, got.ID, tc.want)
		}
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
