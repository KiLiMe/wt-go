// Package app is the window: it reads the settings and the shells of the
// machine, keeps the tabs, and hands what it has to ui. It is the only
// package that starts processes, and the only one that reads the machine.
package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/transfer"
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/cli"
	"wtgo/internal/fault"
	"wtgo/internal/i18n"
	"wtgo/internal/layout"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/internal/theme"
	"wtgo/internal/ui"
)

// App is the window and what it holds.
type App struct {
	win        *mygo.Window
	closed     bool
	settings   store.Settings
	profiles   []profile.Profile
	found      []profile.Found
	panes      pane.Set
	keys       []binding
	renaming   *ui.Rename
	confirm    *ask
	palette    *ui.Palette
	find       *ui.Find
	findTerm   *terminal.Terminal
	shownTitle string
	// content is the room the panes took in the last frame, which the keys
	// that split and move between panes measure.
	content layout.Rect

	// saved is the settings the file holds, and savedData the bytes it held,
	// which the window watches for changes; applied is what it gave the
	// window last, so that it knows when to give it more.
	saved     store.Settings
	savedData []byte
	applied   store.Settings
	// badData is file content the watcher already reported as not settings,
	// so that it is kept once and not once a second. file is what the watcher
	// has already handed to the UI, so an unchanged file does not wake it.
	badData []byte
	file    fileState
	// watchStop stops the goroutine that watches the file when the window
	// closes.
	watchStop chan struct{}

	// settingsTab is the tab that edits the settings, nil while it is not
	// open, and settingsPage the page it shows.
	settingsTab  *pane.Tab
	settingsPage string
	// recording is the action that waits for a key, "" while none does.
	recording string
	// note is what the last action did; configErr is why the file is not
	// settings and stays until the file is settings again.
	note      string
	configErr string
	// filter is what the color schemes page searches for.
	filter string
	// editing is the profile the profile page edits, and editingFrom what it
	// held when the page opened. editingArgs is the arguments as the page types
	// them, one on a line, so that an empty line stays while it is typed.
	editing     *store.Profile
	editingFrom store.Profile
	editingArgs string

	// frames counts the frames drawn; measure is the step of measuring a
	// cell for the size the settings start the window with, and sized says
	// whether the window took it.
	frames  int
	measure int
	base    reading
	last    reading
	cellW   float32
	cellH   float32
	sized   bool
}

// binding is a key, and the action it does.
type binding struct {
	mods   myui.Modifiers
	key    myui.Key
	action string
}

// Run opens the window and runs until it closes.
func Run() (err error) {
	// The commands that answer without a window: `wtgo diag` and `wtgo mcp`.
	if handled, code := cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); handled {
		if code != 0 {
			os.Exit(code)
		}
		return nil
	}
	// A window started from Explorer has no console: what it says goes to
	// wt-go.log as well, so that `wtgo diag` can read it back later.
	if f, ferr := sys.OpenLog(); ferr == nil {
		log.SetOutput(io.MultiWriter(os.Stderr, f))
	}
	defer func() {
		if r := recover(); r != nil {
			reportCrash(r)
			panic(r)
		}
	}()
	if err := sys.HoldChildren(); err != nil {
		log.Printf("wt-go: %v", err)
	}
	mygo.App.WhenReady(func() {
		a := newApp()
		a.appearance()
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:           "wt-go",
			Width:           960,
			Height:          600,
			TitleBarStyle:   titleBarStyle(),
			TitleBarHeight:  ui.TitleBarHeight,
			StateKey:        "main",
			BackgroundColor: theme.Hex(theme.Terminal(a.settings).Background),
			Opacity:         opacity(a.settings.BackgroundOpacity),
			Content:         myui.View(a.view),
		})
		a.win.OnClosed(a.onClosed)
		a.openFirst()
		if a.settingsNote() != "" {
			a.OpenSettings()
		}
		go a.watch(a.watchStop)
	})
	return mygo.App.Run()
}

// faultCode is the code a settings error answers for: a field out of range
// is a different fix from a file that is not JSON at all.
func faultCode(err error) fault.Code {
	if strings.Contains(err.Error(), " must be ") {
		return fault.BadField
	}
	return fault.BadSettings
}

// reportCrash writes what stopped the program next to the settings, so that a
// window that goes away leaves something to read.
func reportCrash(r any) {
	path, err := sys.WriteCrashLog(fmt.Sprint(r), debug.Stack())
	if err != nil {
		log.Printf("wt-go: stopped: %v (%v)", r, err)
		return
	}
	log.Printf("wt-go: stopped: %v (written to %s)", r, path)
}

// opacity is the opacity of the window the settings ask for, between the 0.3
// that keeps the text readable and the 1 that hides nothing.
func opacity(v float64) float64 {
	switch {
	case v <= 0, v > 1:
		return 1
	case v < 0.3:
		return 0.3
	}
	return v
}

// newApp reads the settings and finds the shells, and says what went wrong
// when it could not: the window opens with the defaults.
func newApp() *App {
	a := &App{settings: store.Default(), watchStop: make(chan struct{})}
	settings, data, err := load()
	if err == nil {
		a.settings = settings
	} else {
		log.Printf("wt-go: %s", fault.Line(faultCode(err), err.Error()))
	}
	i18n.Set(a.settings.Language)
	a.saved = a.settings.Clone()
	if data != nil {
		a.file.markSeen(data)
	}
	if err != nil {
		a.configErr = faultCode(err).String() + ": " + configError(data, err)
		a.badData = data
	} else {
		a.savedData = data
	}
	a.found = found()
	a.profiles = profile.Resolve(a.settings.Profiles, a.found)
	a.keys = bindings(a.settings.Keybindings)
	return a
}

// appearance gives the window the appearance the settings ask for: dark and
// light are the settings', system is the machine's.
func (a *App) appearance() {
	switch a.settings.Theme {
	case "light":
		mygo.Theme.SetSource(mygo.ThemeLight)
	case "dark":
		mygo.Theme.SetSource(mygo.ThemeDark)
	default:
		mygo.Theme.SetSource(mygo.ThemeSystem)
	}
}

// load reads the settings, and writes the defaults where there are none. It
// answers with the bytes the file holds, which the window watches.
func load() (store.Settings, []byte, error) {
	data, ok, err := sys.ReadSettings()
	if err != nil {
		return store.Settings{}, nil, err
	}
	if !ok {
		settings := store.Default()
		data, err := store.Marshal(settings)
		if err != nil {
			return settings, nil, err
		}
		if err := sys.WriteSettings(data); err != nil {
			return settings, nil, err
		}
		return settings, data, nil
	}
	settings, err := store.Parse(data)
	return settings, data, err
}

// found is the shells of the machine, as profile resolves them.
func found() []profile.Found {
	shells := sys.Discover()
	out := make([]profile.Found, 0, len(shells))
	for _, s := range shells {
		out = append(out, profile.Found{ID: s.ID, Name: s.Name, Command: s.Command, Args: s.Args})
	}
	return out
}

// bindings are the keybindings the settings name, as the keys they bind.
func bindings(keybindings []store.Keybinding) []binding {
	out := make([]binding, 0, len(keybindings))
	for _, kb := range keybindings {
		mods, key, ok := ui.ParseKey(kb.Key)
		if !ok {
			log.Printf("wt-go: %s", fault.Line(fault.BadKeybinding, fmt.Sprintf("keybinding %q", kb.Key)))
			continue
		}
		out = append(out, binding{mods: mods, key: key, action: kb.Action})
	}
	return out
}

// shortcut is the key the settings bind to an action, as a keybinding
// writes it, and "" when none binds it.
func (a *App) shortcut(action string) string {
	for _, b := range a.keys {
		if b.action != action {
			continue
		}
		key, ok := ui.FormatKey(b.mods, b.key)
		if !ok {
			return ""
		}
		return key
	}
	return ""
}

// view is the window's frame.
func (a *App) view(c *myui.Context) {
	if a.renaming != nil && !a.renaming.Open {
		a.renaming = nil
	}
	if a.confirm != nil && !a.confirm.Open {
		a.confirm = nil
	}
	if a.palette != nil && !a.palette.Open {
		a.palette = nil
	}
	a.content = ui.Content(c)
	a.frames++
	a.apply()
	a.bind(c)
	st := ui.State{
		Active:    a.panes.Active,
		Profiles:  a.menuProfiles(),
		Rename:    a.renaming,
		EqualTabs: a.settings.TabWidth == "equal",
		Shortcut:  a.shortcut,
	}
	if a.confirm != nil {
		st.Confirm = &a.confirm.Confirm
	}
	st.Palette = a.paletteState()
	st.Find = a.findState()
	if a.settingsTab != nil {
		state := a.settingsState()
		st.Settings = &state
	}
	for _, tab := range a.panes.Tabs {
		st.Tabs = append(st.Tabs, a.uiTab(tab))
	}
	ui.View(c, st, a)
	a.startSize()
	a.syncTitle()
}

// uiTab is a tab as the window draws it: the settings tab draws its pages,
// the others their panes.
func (a *App) uiTab(tab *pane.Tab) ui.Tab {
	out := ui.Tab{Name: a.tabName(tab), Color: tab.Color, Panes: tab, Mount: a.mount}
	if tab == a.settingsTab {
		out.Page = a.settingsView
	}
	return out
}

// settingsView is the settings tab's page.
func (a *App) settingsView(c *myui.Context) myui.Element {
	return ui.SettingsTab(c, a.settingsState(), a)
}

// bind gives the window the keys of the tabs, unless the settings wait for a
// key of their own.
func (a *App) bind(c *myui.Context) {
	if a.recording != "" {
		return
	}
	for _, b := range a.keys {
		c.OnShortcut(b.mods, b.key, func() { a.act(b.action) })
	}
}

// mount shows the terminal of a pane.
func (a *App) mount(c *myui.Context, n *pane.Node) myui.Element {
	return terminal.View(c, n.Term)
}

// menuProfiles are the profiles the menu that opens tabs lists: the ones
// the settings show.
func (a *App) menuProfiles() []ui.Profile {
	out := make([]ui.Profile, 0, len(a.profiles))
	for _, p := range a.profiles {
		if p.Hidden {
			continue
		}
		out = append(out, ui.Profile{ID: p.ID, Name: p.Name, Color: p.Color})
	}
	return out
}

// title is the window's title: the name of the tab in front.
func (a *App) title() string {
	if tab, ok := a.panes.Front(); ok {
		return a.tabName(tab)
	}
	return "wt-go"
}

// tabName is what the strip calls a tab. A tab whose program ended says how
// it ended.
func (a *App) tabName(tab *pane.Tab) string {
	if tab.Focus == nil {
		return tab.Name
	}
	n := tab.Focus
	if !a.settings.TitleFromActiveTerminal {
		n = tab.First()
	}
	term := n.Term
	if term == nil {
		return label(tab.Name, "", tab.ProfileName)
	}
	name := label(tab.Name, term.Title(), tab.ProfileName)
	select {
	case <-term.Done():
		return name + " - " + i18n.T("exited") + " (" + strconv.Itoa(term.ExitCode()) + ")"
	default:
		return name
	}
}

// label is what a tab is called: the name it was given, else the title its
// program set, unless that title is the path of the program, which is what a
// shell sets when it has no title of its own; else the name of its profile.
func label(name, title, profileName string) string {
	switch {
	case name != "":
		return name
	case title != "" && !isProgramPath(title):
		return title
	default:
		return profileName
	}
}

// isProgramPath reports whether s is the path of a program rather than a
// title, which a shell with no title of its own sets.
func isProgramPath(s string) bool {
	if filepath.VolumeName(s) != "" {
		return true // D:\Applications\pwsh\pwsh.exe, \\server\share\pwsh.exe
	}
	if strings.HasPrefix(s, `\\`) || strings.HasPrefix(s, "/") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(s), ".exe")
}

// syncTitle gives the window the title of the tab in front, when it changed.
func (a *App) syncTitle() {
	if a.win == nil || a.closed {
		return
	}
	if title := a.title(); title != a.shownTitle {
		a.shownTitle = title
		a.win.SetTitle(title)
	}
}

// redraw asks the window for the next frame, which draws the state.
func (a *App) redraw() {
	if a.win == nil || a.closed {
		return
	}
	a.win.Invalidate()
}

// Do runs fn on the UI thread, where it may change what the window shows:
// a terminal reports its title and its exit from goroutines of its own.
func (a *App) Do(fn func()) {
	if a.win == nil {
		return
	}
	a.win.Update(func() {
		if a.closed {
			return
		}
		fn()
	})
}

// onClosed closes the terminals the window was showing.
func (a *App) onClosed() {
	a.closed = true
	if a.watchStop != nil {
		close(a.watchStop)
		a.watchStop = nil
	}
	for _, tab := range a.panes.Tabs {
		closeTerminals(tab)
	}
	a.panes.Tabs = nil
	a.settingsTab = nil
}

// openFirst opens the tab the settings start with.
func (a *App) openFirst() {
	p, ok := profile.Default(a.profiles, a.settings.DefaultProfile)
	if !ok {
		return
	}
	a.openTab(p, "")
}

// openTab opens a tab that runs p, in dir when dir is not "", else where p
// and the settings say.
func (a *App) openTab(p profile.Profile, dir string) {
	if dir == "" {
		dir = p.Dir
	}
	if dir == "" {
		dir = a.settings.StartDirectory
	}
	dir = sys.ExpandPath(dir)
	var term *terminal.Terminal
	opts := a.termOptions(p, dir)
	// The paste that waits for an answer needs the terminal that takes it,
	// which exists once the options are read.
	opts.OnPaste = func(d transfer.Data) bool { return a.askPaste(term, d) }
	term, err := terminal.New(opts)
	if err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.ChildFailed, p.Name+": "+err.Error()))
		return
	}
	tab := pane.New(term)
	tab.ProfileID, tab.ProfileName, tab.Color = p.ID, p.Name, p.Color
	a.insertTab(tab)
	a.syncTitle()
	a.redraw()
}

// termOptions is the terminal of a tab that runs p in dir.
func (a *App) termOptions(p profile.Profile, dir string) terminal.Options {
	s := a.settings
	return terminal.Options{
		Command:    p.CommandLine(),
		Dir:        dir,
		Font:       terminal.Font{Family: s.Font.Family, Size: s.Font.Size, LineHeight: s.Font.LineHeight},
		Theme:      theme.Terminal(s),
		Cursor:     cursor(s.CursorShape),
		NoBlink:    !s.CursorBlink,
		Scrollback: s.ScrollbackMB << 20,
		Padding:    terminal.Padding{X: s.Padding.X, Y: s.Padding.Y},
		RightClick: rightClick(s.RightClick),
		CopyOnSelect: func() terminal.CopyOnSelect {
			if s.CopyOnSelect {
				return terminal.CopyOnSelectClipboard
			}
			return terminal.CopyOnSelectNone
		}(),
		OnTitle: func(string) { a.Do(a.syncTitle) },
		OnExit:  func(int) { a.Do(a.syncTitle) },
	}
}

// rightClick is what a secondary click does, as the settings name it.
func rightClick(action string) terminal.RightClickAction {
	switch action {
	case "paste":
		return terminal.RightClickPaste
	case "copy":
		return terminal.RightClickCopy
	case "menu":
		return terminal.RightClickContextMenu
	case "nothing":
		return terminal.RightClickIgnore
	}
	return terminal.RightClickCopyOrPaste
}

// cursor is the shape of the cursor the settings name, and a bar, the
// default, for a name the terminal does not know.
func cursor(shape string) terminal.CursorStyle {
	switch shape {
	case "block":
		return terminal.CursorBlock
	case "underline":
		return terminal.CursorUnderline
	}
	return terminal.CursorBar
}
