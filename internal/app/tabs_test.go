package app

import (
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/action"
	"wtgo/internal/layout"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/locales"
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
		{"", "", "", "", locales.SystemProfile()},
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

// TestTabNameNilRoot covers a tab whose tree holds no pane while the window
// names tabs by the first pane: there is nothing to read a title from, so the
// tab keeps the name it was given rather than panicking.
func TestTabNameNilRoot(t *testing.T) {
	a := &App{}
	a.settings.TitleFromActiveTerminal = false
	// This test says what a tab is called, so it says it in English whatever
	// the machine's language is set to by the tests before it.
	mygo.App.SetLocale("en")
	for _, tc := range []struct {
		name string
		tab  *pane.Tab
		want string
	}{
		{"a tab of no panes", &pane.Tab{Focus: &pane.Node{}}, locales.SystemProfile()},
		// A tab that lost its tree still has a name, and says it.
		{"a tab of no panes with a name", &pane.Tab{Name: "build", Focus: &pane.Node{}}, "build"},
		{"a tab of no panes with a profile", &pane.Tab{ProfileName: "Ubuntu", Focus: &pane.Node{}}, "Ubuntu"},
		{"a tab of no panes that runs a shell", &pane.Tab{Runs: "bash", Focus: &pane.Node{}}, "bash"},
		// The root is there but holds no terminal, so First finds no pane to
		// take a title from either.
		{"a tab whose root holds no terminal", &pane.Tab{Root: &pane.Node{}, Focus: &pane.Node{}}, locales.SystemProfile()},
	} {
		if got := a.tabName(tc.tab); got != tc.want {
			t.Errorf("%s: tabName() = %q, want %q", tc.name, got, tc.want)
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

// headlessTerminal is a terminal connected to nothing, which the close of a
// tab can close without a shell of the machine behind it.
func headlessTerminal(t *testing.T) *terminal.Terminal {
	t.Helper()
	term, err := terminal.New(terminal.Options{Conn: sink{}, NoExitMessage: true})
	if err != nil {
		t.Fatal(err)
	}
	return term
}

// sink is the other end of a headless terminal: it never sends anything, and
// takes what the terminal writes.
type sink struct{}

func (sink) Read([]byte) (int, error)    { select {} }
func (sink) Write(p []byte) (int, error) { return len(p), nil }
func (sink) Close() error                { return nil }

// TestCloseOtherTabsForgetsTerminals covers Keep forgetting the processes of
// the tabs it drops: the maps of terminals to processes and of processes to
// terminals do not grow with every batch of tabs a window ever closed.
func TestCloseOtherTabsForgetsTerminals(t *testing.T) {
	if err := terminal.Load(); err != nil {
		t.Skip(err)
	}
	a := &App{
		shells: map[*terminal.Terminal]sys.ShellProc{},
		taken:  map[int]*terminal.Terminal{},
	}
	var tabs []*pane.Tab
	for i := range 3 {
		tab := pane.New(headlessTerminal(t))
		tab.Split(layout.SideBySide, headlessTerminal(t))
		a.panes.Add(tab)
		for n, term := range tab.Terminals() {
			pid := 100*(n+1) + i
			a.shells[term] = sys.ShellProc{Pid: pid}
			a.taken[pid] = term
		}
		tabs = append(tabs, tab)
	}
	keep := tabs[1]
	kept := make(map[*terminal.Terminal]bool)
	for _, term := range keep.Terminals() {
		kept[term] = true
	}

	a.closeOtherTabs(keep)

	for _, term := range keep.Terminals() {
		if _, ok := a.shells[term]; !ok {
			t.Errorf("the kept tab lost its shell process %p", term)
		}
	}
	for pid, term := range a.taken {
		if !kept[term] {
			t.Errorf("taken[%d] still belongs to a closed tab", pid)
		}
	}
	for _, tab := range []*pane.Tab{tabs[0], tabs[2]} {
		for _, term := range tab.Terminals() {
			if _, ok := a.shells[term]; ok {
				t.Errorf("a closed tab kept its shell process %p", term)
			}
		}
	}
	if got, want := len(a.shells), 2; got != want {
		t.Errorf("shells = %d, want the %d of the kept tab", got, want)
	}
	// The kept tab's terminals are the only owners left, one pid each.
	if got, want := len(a.taken), 2; got != want {
		t.Errorf("taken = %d, want the %d of the kept tab", got, want)
	}
}

// TestNewTerminalDoesNotLeakAfterClose covers the write-back of a process to a
// terminal the user closed while the window was still looking for it: neither
// map keeps the terminal, so they do not grow with every tab that came and
// went before its shell showed.
func TestNewTerminalDoesNotLeakAfterClose(t *testing.T) {
	if err := terminal.Load(); err != nil {
		t.Skip(err)
	}
	for _, tc := range []struct {
		name string
		// keep is whether the strip still holds the terminal, closed says
		// whether its program ended.
		keep, closed bool
	}{
		{"a closed terminal in no tab", false, true},
		{"a live terminal in no tab", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &App{
				shells: map[*terminal.Terminal]sys.ShellProc{},
				taken:  map[int]*terminal.Terminal{},
			}
			term := headlessTerminal(t)
			if tc.keep {
				a.panes.Add(pane.New(term))
			}
			// The search took the process before the tab went away.
			const pid = 4242
			if !a.claim(pid, term) {
				t.Fatal("the process was already taken")
			}
			a.taken[pid+1] = term
			if tc.closed {
				term.Close()
			}

			shell := sys.ShellProc{Pid: pid}
			if a.recordShell(term, shell) {
				t.Error("recordShell kept a terminal no tab shows")
			}
			a.moveShell(term, pid-1, sys.ShellProc{Pid: pid + 1})

			if _, ok := a.shells[term]; ok {
				t.Error("the terminal was written back into shells")
			}
			if got := len(a.taken); got != 0 {
				t.Errorf("taken = %d, want the processes of the closed terminal dropped", got)
			}
			if got, want := a.hasTerminal(term), tc.keep; got != want {
				t.Errorf("hasTerminal() = %v, want %v", got, want)
			}
		})
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

// TestCloseLastTabClosesTheWindow closes the last tab of a window that says
// so: the window goes with it. With one tab there is nothing to ask about, so
// the close asks nothing even with the question turned on.
func TestCloseLastTabClosesTheWindow(t *testing.T) {
	a := &App{}
	a.settings.CloseOnLastTab = true
	a.settings.ConfirmCloseWindow = true
	a.panes.Add(&pane.Tab{Name: "bash"})

	a.CloseTab(0)
	if a.confirm != nil {
		t.Fatalf("asked %q about a window with one tab", a.confirm.Title)
	}
	if !a.closing {
		t.Fatal("the last tab left the window open")
	}
}

// TestWindowAsksBeforeItCloses covers the close the settings ask about, as the
// window's own button and Alt+F4 raise it: the question says how many tabs go
// with it, cancelling leaves everything as it was, and the answer goes on to
// close the window without a second question.
func TestWindowAsksBeforeItCloses(t *testing.T) {
	a := &App{}
	a.settings.ConfirmCloseWindow = true
	a.panes.Add(&pane.Tab{Name: "bash"})
	a.panes.Add(&pane.Tab{Name: "pwsh"})
	a.panes.Add(&pane.Tab{Name: "zsh"})

	e := &mygo.CloseEvent{}
	a.onClose(e)
	if !e.DefaultPrevented() {
		t.Fatal("the window closed without asking")
	}
	if a.confirm == nil {
		t.Fatal("the window kept itself open without a question")
	}
	if got := a.confirm.Title; got != locales.ConfirmCloseWindowTitle() {
		t.Errorf("question = %q, want %q", got, locales.ConfirmCloseWindowTitle())
	}
	if got, want := a.confirm.Text, locales.ConfirmCloseWindowText(3); got != want {
		t.Errorf("question says %q, want %q", got, want)
	}
	a.ConfirmNo()
	if a.closing {
		t.Error("cancelling left the window on its way out")
	}
	if len(a.panes.Tabs) != 3 {
		t.Errorf("tabs = %d after cancelling, want the three kept", len(a.panes.Tabs))
	}
	e = &mygo.CloseEvent{}
	a.onClose(e)
	if !e.DefaultPrevented() || a.confirm == nil {
		t.Fatal("the second try closed without asking")
	}
	a.ConfirmYes()
	if !a.closing {
		t.Fatal("saying yes left the window open")
	}
	again := &mygo.CloseEvent{}
	a.onClose(again)
	if again.DefaultPrevented() {
		t.Error("the window asked a second time about the close it agreed to")
	}
}

// TestTabMarksFollowTheProfile covers the mark of a tab that is already open:
// saving a profile's icon, color or name reaches the tab that runs it, without
// restarting the shell in it. The settings tab runs no profile and keeps what
// it has, and so does a tab whose profile the settings no longer offer.
func TestTabMarksFollowTheProfile(t *testing.T) {
	settingsDir(t)
	a := newApp("")
	a.settings.Profiles = []store.Profile{{
		ID:      "pwsh7",
		Name:    "PowerShell 7",
		Command: `C:\Windows\pwsh.exe`,
		Color:   "#3b78ff",
		Icon:    "pwsh7:official",
	}}
	a.profiles = profile.Resolve(a.settings.Profiles, a.found)
	tab := &pane.Tab{ProfileID: "pwsh7", ProfileName: "PowerShell 7", Color: "#3b78ff", Icon: "pwsh7:official"}
	renamed := &pane.Tab{Name: "build", ProfileID: "pwsh7", ProfileName: "PowerShell 7", Color: "#3b78ff", Icon: "pwsh7:official"}
	gone := &pane.Tab{ProfileID: "nope", ProfileName: "Gone", Color: "#111111", Icon: "cmd:classic"}
	settings := &pane.Tab{Name: locales.Settings()}
	a.panes.Add(tab)
	a.panes.Add(renamed)
	a.panes.Add(gone)
	a.panes.Add(settings)

	edited := a.settings.Clone()
	edited.Profiles[0].Name = "Pwsh"
	edited.Profiles[0].Icon = "\U0001f44c"
	edited.Profiles[0].Color = "#ff0000"
	a.settings = edited
	// apply() is what saves this: it resolves the profiles and hands the tabs
	// what they now name. It draws for the desktop too, which a test has none
	// of, so the two steps it takes for the tabs are taken here.
	a.profiles = profile.Resolve(edited.Profiles, a.found)
	a.syncTabMarks()

	for _, tb := range []*pane.Tab{tab, renamed} {
		if tb.Icon != "\U0001f44c" || tb.Color != "#ff0000" || tb.ProfileName != "Pwsh" {
			t.Errorf("tab %q = %q, %q, %q, want the profile's new mark", tb.Name, tb.Icon, tb.Color, tb.ProfileName)
		}
	}
	if renamed.Name != "build" {
		t.Errorf("the name a person gave the tab became %q, want it left alone", renamed.Name)
	}
	if gone.Icon != "cmd:classic" || gone.Color != "#111111" {
		t.Errorf("a tab whose profile is gone = %q, %q, want the mark it opened with", gone.Icon, gone.Color)
	}
	if settings.Name != locales.Settings() || settings.Icon != "" || settings.Color != "" {
		t.Errorf("the settings tab = %q, %q, %q, want it left alone", settings.Name, settings.Icon, settings.Color)
	}
}

// TestWindowDoesNotAskWhenTheSettingsDoNot covers the same close with the
// settings off: the window goes without a word.
func TestWindowDoesNotAskWhenTheSettingsDoNot(t *testing.T) {
	a := &App{}
	a.panes.Add(&pane.Tab{Name: "bash"})
	a.panes.Add(&pane.Tab{Name: "pwsh"})
	e := &mygo.CloseEvent{}
	a.onClose(e)
	if e.DefaultPrevented() || a.confirm != nil {
		t.Fatal("the window asked while the settings do not")
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
