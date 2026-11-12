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
	"sync"

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
	"wtgo/locales"
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
	// shells is the process each terminal runs, when the window learned it,
	// which is how it tells a shell at its prompt from a program running.
	shells map[*terminal.Terminal]sys.ShellProc
	// spawnMu guards taken, the process each terminal took: two shells that
	// start at once would each be found by the other's search, and a shell
	// that a launcher started is two processes of the same terminal.
	spawnMu sync.Mutex
	taken   map[int]*terminal.Terminal
	// tabsScroll is where the strip shows its tabs, which the wheel over the
	// strip and the buttons that move it change; revealTabs counts the frames
	// left in which the strip brings the tab in front into view.
	tabsScroll myui.ScrollState
	revealTabs int
	// buttons is the desktop's window button layout, which the strip draws
	// where the window hides the desktop's buttons, and maximized whether the
	// window is maximized, which the button that maximizes shows.
	buttons   mygo.WindowButtons
	maximized bool
	// fontWarned is the family the log already said is not monospace, so that
	// a settings file naming one leaves a line behind and not one a tab.
	fontWarned string
	// uiTheme is the theme the window draws with: the one the settings name,
	// else the terminal's own colors, so that the strip and the panes below it
	// are one surface. nil until the settings are read.
	uiTheme *myui.Theme
	// fonts are the machine's monospace families and the measures the window
	// made of them; the settings page lists the families, so they are asked for
	// when that page first shows.
	fonts fontCensus
	// schemes are the color schemes the settings page lists, built once per
	// settings the window applied: the list is a directory of Ghostty's names
	// and a sort of several hundred of them, which are not worth a frame. nil
	// says the settings changed since it was built.
	schemes []ui.Scheme
	// settingsShown is the settings tab as the frame being drawn shows it: the
	// frame and the page both need it, and building it twice costs a directory
	// read and a sort of the color scheme names.
	settingsShown ui.Settings
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
	note string
	// relaunch starts the program again, which the advanced page does after
	// it deletes the settings; the tests replace it.
	relaunch  func() error
	configErr string
	// filter is what the color schemes page searches for.
	filter string
	// showHidden is whether the profiles page lists the profiles it hides.
	showHidden bool
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
	// The commands that answer without a window: `wt-go diag` and `wt-go mcp`.
	if handled, code := cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); handled {
		if code != 0 {
			os.Exit(code)
		}
		return nil
	}
	// A window started from Explorer has no console: what it says goes to
	// wt-go.log as well, so that `wt-go diag` can read it back later.
	if f, ferr := sys.OpenLog(); ferr == nil {
		// The file comes first: a window started from Explorer has no console,
		// and writing to one that is not there must not lose the line.
		log.SetOutput(io.MultiWriter(f, console{os.Stderr}))
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
		// Before the first layout: a glyph drawn after it is already the wrong
		// color, and the fix only reaches glyphs drawn later.
		if !sys.FixTextColor() {
			log.Print(fault.Line(fault.NoFontconfig, ""))
		}
		a := newApp()
		a.appearance()
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:             "wt-go",
			Width:             960,
			Height:            600,
			TitleBarStyle:     titleBarStyle(),
			TitleBarHeight:    ui.TitleBarHeight,
			HideWindowButtons: hideWindowButtons(),
			StateKey:          "main",
			BackgroundColor:   theme.Hex(a.background()),
			Opacity:           opacity(a.settings.BackgroundOpacity),
			Content:           myui.View(a.view),
		})
		a.win.OnClosed(a.onClosed)
		a.watchWindowButtons()
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

// locale makes the window say its messages in the language the settings name,
// "" for the system's. A file written by hand may hold a tag that names no
// language, which MyGo panics on: the window shows the system's instead of
// refusing to start.
func locale(tag string) {
	if !i18n.Valid(tag) {
		tag = ""
	}
	mygo.App.SetLocale(tag)
}

// newApp reads the settings and finds the shells, and says what went wrong
// when it could not: the window opens with the defaults.
func newApp() *App {
	a := &App{settings: store.Default(), shells: map[*terminal.Terminal]sys.ShellProc{}, watchStop: make(chan struct{}), relaunch: sys.Relaunch}
	settings, data, fresh, err := load()
	if err == nil {
		a.settings = settings
	} else {
		log.Printf("wt-go: %s", fault.Line(faultCode(err), err.Error()))
	}
	locale(a.settings.Language)
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
	// The shells of the machine are probed when the window writes its first
	// settings file. After that the file is the list, and a probe is what the
	// advanced page asks for.
	if fresh && err == nil {
		a.found = found()
		if adopt(&a.settings, a.found, exists) {
			a.adoptWrite()
		}
	}
	a.profiles = profile.Resolve(a.settings.Profiles, a.found)
	a.keys = bindings(a.settings.Keybindings)
	return a
}

// background is the color the window fills what the page has not painted yet:
// the theme the window draws with, which is the terminal's own unless the
// settings name a theme.
func (a *App) background() myui.Color {
	if a.uiTheme != nil {
		return a.uiTheme.Background
	}
	return theme.Terminal(a.settings).Background
}

// appearance gives the window the appearance the settings ask for: dark and
// light are the settings', system is the machine's, and a theme of its own is
// the one that theme is.
func (a *App) appearance() {
	if t, ok := theme.UITheme(a.settings, a.settings.Theme); ok {
		a.uiTheme = t
		if t.Dark {
			mygo.Theme.SetSource(mygo.ThemeDark)
		} else {
			mygo.Theme.SetSource(mygo.ThemeLight)
		}
		return
	}
	// No theme of its own: the window still draws in the terminal's colors, so
	// that the tab in front carries the pane's background up into the strip.
	// What the settings name as the appearance is what the desktop's own
	// window, menus and dialogs follow.
	a.uiTheme = theme.Chrome(a.settings)
	switch a.settings.Theme {
	case "light":
		mygo.Theme.SetSource(mygo.ThemeLight)
	case "dark":
		mygo.Theme.SetSource(mygo.ThemeDark)
	default:
		mygo.Theme.SetSource(mygo.ThemeSystem)
	}
}

// windowButtons are the buttons the strip draws where the window hides the
// desktop's: the ones the desktop's layout names, at the sides it puts them,
// and whether the window is maximized. It is nil where mygo draws them, or
// where the compositor draws a title bar of its own.
func (a *App) windowButtons() *ui.WindowButtons {
	if !hideWindowButtons() {
		return nil
	}
	// A strip that draws none is the same as a strip that hides them, and nil
	// says so without building one every frame.
	if len(a.buttons.Start) == 0 && len(a.buttons.End) == 0 {
		return nil
	}
	return &ui.WindowButtons{Start: a.buttons.Start, End: a.buttons.End, Maximized: a.maximized}
}

// watchWindowButtons keeps the strip's copy of the desktop's buttons in step:
// the layout changes when the user picks another one, and the window maximizes
// and restores as the user asks it to.
func (a *App) watchWindowButtons() {
	a.buttons = mygo.Theme.WindowButtons()
	a.maximized = a.win.IsMaximized()
	mygo.Theme.OnUpdated(func() {
		a.Do(func() {
			a.buttons = mygo.Theme.WindowButtons()
			a.redraw()
		})
	})
	a.win.OnMaximize(func() {
		a.Do(func() {
			a.maximized = true
			a.redraw()
		})
	})
	a.win.OnUnmaximize(func() {
		a.Do(func() {
			a.maximized = false
			a.redraw()
		})
	})
}

// MinimizeWindow minimizes the window, as its own button would.
func (a *App) MinimizeWindow() {
	a.Do(func() { a.win.Minimize() })
}

// ToggleMaximizeWindow maximizes the window, or restores it, as its own button
// would.
func (a *App) ToggleMaximizeWindow() {
	a.Do(func() { a.win.ToggleMaximize() })
}

// CloseWindow closes the window, as its own close button would, so that the
// shells the window runs still close with it.
func (a *App) CloseWindow() {
	a.Do(func() { a.win.Close() })
}

// load reads the settings, and writes the defaults where there are none. It
// answers with the bytes the file holds, which the window watches.
// load reads settings.json, writing the defaults when there is no file, and
// reports whether it wrote them: that first file is the only one the window
// probes the machine for.
func load() (store.Settings, []byte, bool, error) {
	data, ok, err := sys.ReadSettings()
	if err != nil {
		return store.Settings{}, nil, false, err
	}
	if !ok {
		settings := store.Default()
		data, err := store.Marshal(settings)
		if err != nil {
			return settings, nil, true, err
		}
		if err := sys.WriteSettings(data); err != nil {
			return settings, nil, true, err
		}
		return settings, data, true, nil
	}
	settings, err := store.Parse(data)
	return settings, data, false, err
}

// found is the shells of the machine, as profile resolves them.
func found() []profile.Found {
	shells := sys.Discover()
	out := make([]profile.Found, 0, len(shells))
	for _, s := range shells {
		out = append(out, profile.Found{ID: s.ID, Command: s.Command, Args: s.Args})
	}
	return out
}

// adopt writes the shells the machine has into the settings, the way the
// settings page would: a shell the file does not name is added below the ones
// it does, so that the strip lists what can be started without a person
// typing a path, and editing the command line of a discovered profile
// afterwards keeps the edit. A discovered profile whose program is gone is
// hidden rather than dropped, so that installing the program again is all it
// takes to bring it back.
// adoptWrite puts the settings the window adopted into the file, and keeps the
// copy the window compares against.
func (a *App) adoptWrite() {
	data, err := store.Marshal(a.settings)
	if err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.BadSettings, err.Error()))
		return
	}
	if err := sys.WriteSettings(data); err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.SettingsWrite, err.Error()))
		return
	}
	a.saved, a.savedData = a.settings.Clone(), data
	a.file.markSeen(data)
}

// adopt adds the shells the machine has to the settings, and hides the found
// profiles whose program is gone, and reports whether it changed anything.
func adopt(s *store.Settings, found []profile.Found, exists func(string) bool) bool {
	at := 0
	have := make(map[string]bool, len(s.Profiles))
	for i, p := range s.Profiles {
		have[p.ID] = true
		if p.Discovered {
			at = i + 1
		}
	}
	var added []store.Profile
	for _, f := range found {
		if have[f.ID] {
			continue
		}
		have[f.ID] = true
		added = append(added, store.Profile{
			ID:         f.ID,
			Name:       profile.Name(f.ID, f.Command),
			Command:    f.Command,
			Args:       f.Args,
			Color:      profile.Color(f.ID),
			Discovered: true,
		})
	}
	changed := len(added) > 0
	if changed {
		list := make([]store.Profile, 0, len(s.Profiles)+len(added))
		list = append(list, s.Profiles[:at]...)
		list = append(list, added...)
		list = append(list, s.Profiles[at:]...)
		s.Profiles = list
	}
	for i := range s.Profiles {
		p := &s.Profiles[i]
		if p.Discovered && !p.Hidden && commandGone(p.Command, exists) {
			p.Hidden = true
			changed = true
		}
	}
	return changed
}

// exists reports whether a path is there.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// commandGone reports whether the program a profile runs is no longer where
// the profile says. A command that is a bare name is PATH's business, not this
// one's: it is left to fail when a tab runs it.
func commandGone(command string, exists func(string) bool) bool {
	if command == "" || !strings.ContainsAny(command, `\/:`) {
		return false
	}
	return !exists(command)
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
	if a.uiTheme != nil {
		c.SetTheme(a.uiTheme)
	}
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
		Active:     a.panes.Active,
		Profiles:   a.menuProfiles(),
		Rename:     a.renaming,
		EqualTabs:  a.settings.TabWidth == "equal",
		Shortcut:   a.shortcut,
		TabsScroll: &a.tabsScroll,
		RevealTab:  a.revealTabs > 0,
		// The buttons the strip draws for the window, where the window hides
		// the desktop's.
		WindowButtons: a.windowButtons(),
	}
	// The strip brings the tab in front into view for the frames the window
	// asked for, and no more: asking on every frame would keep the user from
	// scrolling the strip away from that tab. Two frames are what it takes,
	// because the buttons that scroll the strip take their room only on the
	// frame after they first appear, and the strip is narrower by then.
	if a.revealTabs > 0 {
		a.revealTabs--
	}
	if a.confirm != nil {
		st.Confirm = &a.confirm.Confirm
	}
	st.Palette = a.paletteState()
	st.Find = a.findState()
	if a.settingsTab != nil {
		a.settingsShown = a.settingsState()
		st.Settings = &a.settingsShown
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
	return ui.SettingsTab(c, a.settingsShown, a)
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
		return name + " - " + locales.TabExited() + " (" + strconv.Itoa(term.ExitCode()) + ")"
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

// console is stderr for a window that may not have one: the file keeps the
// lines either way, so a console that is not there is not an error.
type console struct{ f *os.File }

func (c console) Write(p []byte) (int, error) {
	_, _ = c.f.Write(p)
	return len(p), nil
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
	term, err := a.newTerminal(opts)
	if err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.ChildFailed, p.Name+": "+err.Error()))
		return
	}
	tab := pane.New(term)
	tab.ProfileID, tab.ProfileName, tab.Color = p.ID, p.Name, p.Color
	a.insertTab(tab)
	a.revealTab()
	a.syncTitle()
	a.redraw()
}

// termOptions is the terminal of a tab that runs p in dir.
func (a *App) termOptions(p profile.Profile, dir string) terminal.Options {
	s := a.settings
	return terminal.Options{
		Command:    p.CommandLine(),
		Dir:        dir,
		Font:       a.terminalFont(),
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
