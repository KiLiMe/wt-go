// Package app is the window: it reads the settings and the shells of the
// machine, keeps the tabs, and hands what it has to ui. It is the only
// package that starts processes, and the only one that reads the machine.
package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/terminal"
	myui "github.com/egoist/mygo/ui"
	"log"
	"os"
	"sync"
	"time"
	"wtgo/internal/fault"
	"wtgo/internal/layout"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/internal/ui"
)

// App is the window and what it holds.
type App struct {
	win    *mygo.Window
	closed bool
	// closing is the window already on its way out: the question about
	// closing it is answered, so the window that hears the close does not
	// ask again.
	closing  bool
	settings store.Settings

	// startDir is the directory the command line's -d named: it stands in for
	// the setting with the same job, for this run only. A profile's own dir is
	// still the more specific answer and still wins.
	startDir string
	// device is the shell of this machine, as following the device runs it: the
	// first shell the probe found, probed at every start whatever the file says.
	// The dropdown names it, and the strip names a tab that runs no program after
	// it.
	device     profile.Profile
	profiles   []profile.Profile
	found      []profile.Found
	panes      pane.Set
	keys       []binding
	renaming   *ui.Rename
	confirm    *ask
	about      *ui.About
	palette    *ui.Palette
	find       *ui.Find
	findPane   *pane.Node
	tabsList   *ui.TabsList
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
	// else mygo's own for the appearance they ask for. nil until the settings
	// are read.
	uiTheme *myui.Theme
	// pane is the face of the pane below the strip: the terminal's colors,
	// which the tab in front carries up into it. The window keeps it because
	// asking for it parses a scheme, which no frame may pay for.
	pane ui.Face
	// fonts are the machine's monospace families and the measures the window
	// made of them; the settings page lists the families, so they are asked for
	// when that page first shows.
	fonts fontCensus
	// schemes are the color schemes the settings page lists, built once per set
	// of schemes the settings write: the list is a directory of Ghostty's names
	// and a sort of several hundred of them, which are not worth a frame, and
	// nothing but the schemes in the file can change it. schemesFrom is the
	// schemes it was built from, which a new draft is compared against.
	schemes     []ui.Scheme
	schemesFrom []store.Scheme
	// swatches are the colors the page shows those schemes by, the ones the
	// settings do not write kept from one build of the list to the next.
	swatches schemeSwatches
	// settingsShown is the settings tab as the frame being drawn shows it: the
	// frame and the page both need it, and building it twice costs a directory
	// read and a sort of the color scheme names.
	settingsShown ui.Settings
	// keysDerived and schemesDerived are what the actions page and the color
	// schemes page of the settings derived from the file, which the pages
	// keep from frame to frame: a frame that changes neither the file nor the
	// switches that change the page's rows draws it from these.
	keysDerived    ui.PageKeybindings
	schemesDerived ui.SchemesCache
	// content is the room the panes took in the last frame, which the keys
	// that split and move between panes measure.
	content layout.Rect
	// console is where the terminals' console comes from, as the about dialog
	// words it: read once at startup, since mygo picks a ConPTY once and the
	// dialog must name the one the terminals in front were made in.
	console consoleSource

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

	// faceSent is whether the terminals have been given a look at all: the
	// first apply gives them all of it, and later ones only what changed.
	faceSent bool

	// meter times the frames the window draws, and showFrames is whether the
	// readout of the rate it makes is open: the window is what a person
	// asking what the frame rate is is asking about.
	meter      frameMeter
	showFrames bool

	// capture is what the actions page listens for a new key with: the keys a
	// recording needs are made once here, not on every frame of a recording.
	capture ui.Capture

	// settingsTab is the tab that edits the settings, nil while it is not
	// open, and settingsPage the page it shows.
	settingsTab  *pane.Tab
	settingsPage string
	// recording is the action that waits for a key, "" while none does.
	recording string
	// note is what the last action did; configErr is why the file is not
	// settings and stays until the file is settings again; validation is why
	// the last save refused the draft, cleared by the next save or discard.
	note       string
	validation string
	// relaunch starts the program again, which the advanced page does after
	// it deletes the settings; the tests replace it.
	relaunch  func() error
	configErr string
	// filter is what the color schemes page searches for, and actionSearch
	// what the actions page does, and profileSearch what the profiles page
	// searches its list of profiles for. settingsSearch is what the box over
	// the column of pages searches the settings for.
	filter         string
	actionSearch   string
	profileSearch  string
	settingsSearch string
	// reveal is the setting the search over the column of pages found, and how
	// far the mark over it has faded: the page scrolls to the setting on the
	// first frame of it, and the mark goes over the frames after.
	reveal ui.Reveal
	// onlyChangedKeys is whether the actions page lists the actions the file
	// binds on keys of its own alone, and numberedKeys whether it shows the
	// numbered actions it folds away.
	onlyChangedKeys bool
	numberedKeys    bool
	// editing is the profile the profile page edits, and editingFrom what it
	// held when the page opened. editingArgs is the arguments as the page types
	// them, one on a line, so that an empty line stays while it is typed.
	editing     *store.Profile
	editingFrom store.Profile
	editingArgs string
	// editingGlyph is the character the mark box holds, and "" while the mark
	// is a shape or a shape's absence: the box shows one kind of mark, while
	// the file holds both in one field.
	editingGlyph string
}

// binding is a key, and the action it does.
type binding struct {
	mods   myui.Modifiers
	key    myui.Key
	action string
	// repeats says the action takes the key again as the keyboard repeats it
	// while it stays held, as a key that steps through the tabs does.
	repeats bool
}

// newApp reads the settings and finds the shells, and says what went wrong
// when it could not: the window opens with the defaults. startDir is where the
// command line said new tabs start, "" when it said nothing.
func newApp(startDir string) *App {
	// The machine is probed once, before the file is read: the device is a
	// fact about the machine, and the dropdown names it even when the file
	// already says which profiles to offer.
	shells := found()
	a := &App{
		settings:  store.Default(),
		startDir:  startDir,
		device:    profile.Device(shells),
		shells:    map[*terminal.Terminal]sys.ShellProc{},
		watchStop: make(chan struct{}),
		relaunch:  sys.Relaunch,
	}
	// The console a terminal is made in is a fact about the program's folder
	// and the environment, and it is read before anything is drawn: a read
	// later would name a dll the terminals in front were not made in.
	a.console = consoleOf(sys.ConPTYLibrary())
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
	// The shells of the machine are written into the settings when the window
	// writes its first settings file. After that the file is the list, and a
	// probe is what the advanced page asks for.
	if fresh && err == nil {
		a.found = shells
		if adopt(&a.settings, a.found, exists) {
			a.adoptWrite()
		}
	}
	a.profiles = profile.Resolve(a.settings.Profiles, a.found)
	// A profile the file names twice, or not at all, is one the window leaves
	// out: it says which one here, since the list on the profiles page shows
	// only what survived and would leave the person looking for it.
	for _, problem := range store.ProfileIDs(a.settings.Profiles) {
		log.Printf("wt-go: %s", fault.Line(fault.BadProfileID, problem))
	}
	a.keys = bindings(a.settings.Keybindings)
	return a
}

// view is the window's frame.
func (a *App) view(c *myui.Context) {
	if a.uiTheme != nil {
		c.SetTheme(a.uiTheme)
	}
	// The popup menus are drawn in the window with the theme's look rather
	// than by the desktop: a menu of the desktop's beside the window's own
	// strip reads as another program's.
	c.SetMenuPresentation(myui.PopoverMenus)
	if a.dismissed() {
		a.dropClosed()
	}
	a.content = ui.Content(c)
	a.apply()
	a.bind(c)
	st := ui.State{
		Active:     a.panes.Active,
		Pane:       a.pane,
		Profiles:   a.menuProfiles(),
		Device:     a.device.Name,
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
	st.About = a.about
	st.Palette = a.paletteState()
	st.TabsList = a.tabsList
	st.Find = a.findState()
	if a.settingsTab != nil {
		a.settingsShown = a.settingsState()
		st.Settings = &a.settingsShown
	}
	for _, tab := range a.panes.Tabs {
		st.Tabs = append(st.Tabs, a.uiTab(tab))
	}
	if a.showFrames {
		a.meter.tick(time.Now())
		st.Frames = a.meter.frames()
	}
	ui.View(c, st, a)
	// The mark over the setting the search found is one frame older than it
	// was drawn: the frame that scrolls to the setting is the first one, and
	// the fade is what the frames after it show.
	a.revealFrame()
	a.syncTitle()
}

// uiTab is a tab as the window draws it: the settings tab draws its pages,
// the others their panes.
func (a *App) uiTab(tab *pane.Tab) ui.Tab {
	out := ui.Tab{Name: a.tabName(tab), Color: tab.Color, Icon: tab.Icon, Panes: tab, Mount: a.mount}
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
// key of their own or an overlay has the keyboard. A key the keyboard repeats
// while it stays held runs an action that steps through something, and is left
// to the program for the rest: a held Ctrl+T makes one tab, not fifty.
func (a *App) bind(c *myui.Context) {
	if a.recording != "" || a.top() != overlayNone {
		return
	}
	for _, b := range a.keys {
		b := b
		c.OnShortcut(b.mods, b.key, func() {
			if c.ShortcutRepeated() && !b.repeats {
				return
			}
			a.run(b.action)
		})
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
		out = append(out, ui.Profile{ID: p.ID, Name: p.Name, Color: p.Color, Icon: p.Icon})
	}
	return out
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
