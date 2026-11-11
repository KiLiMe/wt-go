package app

import (
	"bytes"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/i18n"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/internal/theme"
	"wtgo/internal/ui"
)

// OpenSettings shows the settings, which one tab holds: it opens that tab, or
// a new one.
func (a *App) OpenSettings() {
	if a.settingsTab == nil {
		a.settingsTab = &pane.Tab{Name: i18n.T("Settings")}
		a.settingsPage = ui.PageStartup
		a.recording = ""
		a.insertTab(a.settingsTab)
	}
	for i, tab := range a.panes.Tabs {
		if tab == a.settingsTab {
			a.panes.Active = i
		}
	}
	a.redraw()
}

// insertTab puts a tab where the settings say new tabs go, and shows it.
func (a *App) insertTab(tab *pane.Tab) {
	if a.settings.NewTabPosition == "first" {
		a.panes.Tabs = append([]*pane.Tab{tab}, a.panes.Tabs...)
		a.panes.Active = 0
		return
	}
	a.panes.Add(tab)
}

// settingsState is the settings tab as the window draws it.
func (a *App) settingsState() ui.Settings {
	return ui.Settings{
		Page:         a.settingsPage,
		Draft:        &a.settings,
		Dirty:        a.dirty(),
		Note:         a.settingsNote(),
		Recording:    a.recording,
		Filter:       &a.filter,
		Profile:      a.editing,
		Args:         &a.editingArgs,
		Profiles:     fileProfiles(a.settings),
		Known:        knownProfiles(a.profiles),
		Schemes:      theme.Schemes(a.settings),
		FontFamilies: a.families,
	}
}

// fileProfiles are the profiles the file holds, as the settings list them.
func fileProfiles(s store.Settings) []ui.Profile {
	out := make([]ui.Profile, 0, len(s.Profiles))
	for _, p := range s.Profiles {
		out = append(out, ui.Profile{ID: p.ID, Name: p.Name, Color: p.Color})
	}
	return out
}

// knownProfiles are the profiles the window runs, as the settings list them.
func knownProfiles(profiles []profile.Profile) []ui.Profile {
	out := make([]ui.Profile, 0, len(profiles))
	for _, p := range profiles {
		out = append(out, ui.Profile{ID: p.ID, Name: p.Name, Color: p.Color})
	}
	return out
}

// SelectSettingsPage shows a page of the settings.
func (a *App) SelectSettingsPage(page string) {
	if a.settingsTab == nil {
		return
	}
	a.settingsPage = page
	a.recording = ""
	a.note = ""
	a.editNone()
	if id, ok := ui.ProfileID(page); ok {
		a.editProfile(id)
	}
	a.redraw()
}

// editProfile opens a copy of a profile for the profile page to edit: the page
// leaves the file alone until the settings are saved, so that looking at a
// profile writes nothing.
func (a *App) editProfile(id string) {
	p, ok := profileByID(a.settings, id)
	if !ok {
		a.editNone()
		return
	}
	a.editing, a.editingFrom = &p, p
	a.editingArgs = strings.Join(p.Args, "\n")
}

// editNone closes the profile page's copy: nothing is being edited.
func (a *App) editNone() {
	a.editing, a.editingFrom, a.editingArgs = nil, store.Profile{}, ""
}

// splitArgs is the arguments a text holds, one on each line.
func splitArgs(text string) []string {
	var args []string
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			args = append(args, line)
		}
	}
	return args
}

// profileByID is the profile the file holds by an id.
func profileByID(s store.Settings, id string) (store.Profile, bool) {
	for _, p := range s.Profiles {
		if p.ID == id {
			return p, true
		}
	}
	return store.Profile{}, false
}

// SaveSettings writes the settings to the file, and keeps the copy the window
// compares against to know what is unsaved. It refuses to write over a file
// that changed on disk since the window last read or saved it.
func (a *App) SaveSettings() {
	if a.editing != nil {
		a.editing.Args = splitArgs(a.editingArgs)
		a.upsertProfile(*a.editing)
	}
	if err := a.settings.Validate(); err != nil {
		a.setNote(err.Error())
		return
	}
	data, err := store.Marshal(a.settings)
	if err != nil {
		a.setNote(err.Error())
		return
	}
	unchanged, err := a.settingsFileUnchanged()
	if err != nil {
		a.setNote(fmt.Sprintf(i18n.T("settings.json: %v"), err))
		return
	}
	if !unchanged {
		a.setNote(i18n.T("settings.json changed on disk; discard your changes to take it."))
		return
	}
	if err := sys.WriteSettings(data); err != nil {
		a.setNote(err.Error())
		return
	}
	a.saved, a.savedData, a.badData = a.settings.Clone(), data, nil
	if a.editing != nil {
		a.editingFrom = *a.editing
	}
	a.configErr = ""
	a.file.markSeen(data)
	a.setNote(i18n.T("Saved"))
	a.redraw()
}

// settingsFileUnchanged reports whether settings.json still holds what the
// window last loaded or saved, or the bad content it last reported. A file
// that is gone is not in the way; a file with other content is.
func (a *App) settingsFileUnchanged() (bool, error) {
	data, ok, err := sys.ReadSettings()
	if err != nil {
		return false, err
	}
	if !ok {
		return true, nil
	}
	if a.savedData != nil && bytes.Equal(data, a.savedData) {
		return true, nil
	}
	if a.badData != nil && bytes.Equal(data, a.badData) {
		return true, nil
	}
	return false, nil
}

// upsertProfile writes a profile into the file's list: where it is, or at the
// end.
func (a *App) upsertProfile(p store.Profile) {
	for i, q := range a.settings.Profiles {
		if q.ID == p.ID {
			a.settings.Profiles[i] = p
			return
		}
	}
	a.settings.Profiles = append(a.settings.Profiles, p)
}

// closeSettingsTab forgets the settings tab, which is going away: the settings
// go back to what the file holds, and what the tab was doing -- editing a
// profile, waiting for a key, saying what the last action did -- goes with it.
// The file is read again, as discarding a draft does, so that a change the
// watcher held back while the draft was dirty is taken now. The tab stays set
// until the read is done, so a file that is still not settings does not open a
// second tab behind the caller.
func (a *App) closeSettingsTab() {
	a.settings = a.saved.Clone()
	a.note = ""
	a.recording = ""
	a.editNone()
	a.refreshSettings()
	a.settingsTab = nil
}

// DiscardSettings goes back to the settings the file holds, and reads the file
// again: a change the watcher held back while the draft was dirty is taken now.
func (a *App) DiscardSettings() {
	a.settings = a.saved.Clone()
	a.note = ""
	a.recording = ""
	if a.editing != nil {
		a.editProfile(a.editing.ID)
	}
	a.refreshSettings()
	a.redraw()
}

// refreshSettings reads settings.json now, as discarding a draft does, because
// the watcher may have seen a change while the draft was in the way.
func (a *App) refreshSettings() {
	data, ok, err := sys.ReadSettings()
	if err != nil {
		a.setNote(fmt.Sprintf(i18n.T("settings.json: %v"), err))
		return
	}
	if !ok {
		return
	}
	a.reloadSafely(data)
}

// OpenSettingsFile opens settings.json in the program that edits it.
func (a *App) OpenSettingsFile() {
	path, err := sys.SettingsPath()
	if err != nil {
		a.note = err.Error()
		a.redraw()
		return
	}
	if err := sys.OpenPath(path); err != nil {
		a.note = err.Error()
		a.redraw()
	}
}

// RecordKeybinding waits for a key of an action; asking again for the one it
// waits for stops it.
func (a *App) RecordKeybinding(action string) {
	if a.recording == action {
		a.recording = ""
	} else {
		a.recording = action
	}
	a.redraw()
}

// SetKeybinding gives an action a key, which ends the wait for one: it
// replaces the key the action had, or adds the keybinding it did not have.
func (a *App) SetKeybinding(action, key string) {
	if key != "" {
		found := false
		for i := range a.settings.Keybindings {
			if a.settings.Keybindings[i].Action == action {
				a.settings.Keybindings[i].Key = key
				found = true
				break
			}
		}
		if !found {
			a.settings.Keybindings = append(a.settings.Keybindings, store.Keybinding{Key: key, Action: action})
		}
	}
	a.recording = ""
	a.redraw()
}

// RemoveKeybinding drops the keybindings of an action, so that it runs on no
// key at all.
func (a *App) RemoveKeybinding(action string) {
	out := a.settings.Keybindings[:0]
	for _, kb := range a.settings.Keybindings {
		if kb.Action != action {
			out = append(out, kb)
		}
	}
	a.settings.Keybindings = out
	a.recording = ""
	a.redraw()
}

// CopyScheme makes a color scheme of the settings from the one the terminal
// paints with, and picks it.
func (a *App) CopyScheme() {
	th, ok := theme.Scheme(a.settings, a.settings.ColorScheme)
	if !ok {
		return
	}
	name := a.settings.ColorScheme + " (custom)"
	for i := 2; a.schemeNamed(name); i++ {
		name = a.settings.ColorScheme + " (custom " + strconv.Itoa(i) + ")"
	}
	scheme := theme.Stored(th)
	scheme.Name = name
	a.settings.Schemes = append(a.settings.Schemes, scheme)
	a.settings.ColorScheme = name
	a.note = i18n.T("The scheme is copied; save to keep it.")
	a.redraw()
}

// schemeNamed reports whether the settings hold a color scheme of that name.
func (a *App) schemeNamed(name string) bool {
	for _, s := range a.settings.Schemes {
		if strings.EqualFold(s.Name, name) {
			return true
		}
	}
	return false
}

// ResetKeybindings goes back to the keybindings the window starts with.
func (a *App) ResetKeybindings() {
	a.settings.Keybindings = store.DefaultKeybindings()
	a.recording = ""
	a.redraw()
}

// AddProfile adds a profile to the file's list, and shows its page.
func (a *App) AddProfile() {
	p := store.Profile{ID: a.newProfileID(), Name: i18n.T("New profile")}
	a.settings.Profiles = append(a.settings.Profiles, p)
	a.settingsPage = ui.ProfilePage(p.ID)
	a.editing, a.editingFrom = &p, p
	a.redraw()
}

// newProfileID is an id no profile the file holds has.
func (a *App) newProfileID() string {
	for n := 1; ; n++ {
		id := "profile-" + strconv.Itoa(n)
		if _, ok := profileByID(a.settings, id); !ok {
			return id
		}
	}
}

// RemoveProfile drops a profile from the file's list.
func (a *App) RemoveProfile(id string) {
	out := make([]store.Profile, 0, len(a.settings.Profiles))
	for _, p := range a.settings.Profiles {
		if p.ID != id {
			out = append(out, p)
		}
	}
	a.settings.Profiles = out
	a.settingsPage = ui.PageProfiles
	a.editNone()
	a.recording = ""
	a.redraw()
}

// DuplicateProfile copies a profile, and shows the copy's page.
func (a *App) DuplicateProfile(id string) {
	p, ok := profileByID(a.settings, id)
	if !ok {
		return
	}
	p.ID = a.newProfileID()
	p.Name = p.Name + " " + i18n.T("copy")
	p.Discovered = false
	a.settings.Profiles = append(a.settings.Profiles, p)
	a.settingsPage = ui.ProfilePage(p.ID)
	a.editing, a.editingFrom = &p, p
	a.redraw()
}

// dirty reports whether the settings hold something the file does not: a
// setting it has not seen, or a profile the page changed.
func (a *App) dirty() bool {
	if !store.Equal(a.settings, a.saved) {
		return true
	}
	if a.editing == nil {
		return false
	}
	if a.editingArgs != strings.Join(a.editingFrom.Args, "\n") {
		return true
	}
	return !store.EqualProfile(*a.editing, a.editingFrom)
}

// apply gives the window what the settings ask for, and does nothing while
// they are what it gave it last time: the settings tab edits the settings the
// window shows, so that a change shows as it is made.
func (a *App) apply() {
	if store.Equal(a.settings, a.applied) {
		return
	}
	s := a.settings
	i18n.Set(s.Language)
	a.appearance()
	a.keys = bindings(s.Keybindings)
	a.profiles = profile.Resolve(s.Profiles, a.found)
	font := a.terminalFont()
	colors := theme.Terminal(s)
	copyOnSelect := terminal.CopyOnSelectNone
	if s.CopyOnSelect {
		copyOnSelect = terminal.CopyOnSelectClipboard
	}
	for _, tab := range a.panes.Tabs {
		for _, term := range tab.Terminals() {
			term.SetFont(font)
			term.SetTheme(colors, nil)
			term.SetCopyOnSelect(copyOnSelect)
		}
	}
	if a.win != nil {
		if err := a.win.SetBackgroundColor(theme.Hex(colors.Background)); err != nil {
			log.Printf("wt-go: background: %v", err)
		}
		a.win.SetOpacity(opacity(s.BackgroundOpacity))
		a.win.SetAlwaysOnTop(s.AlwaysOnTop)
	}
	a.applied = s.Clone()
	a.redraw()
}

// watch takes the settings the file holds when another program writes them,
// once a second, as Windows Terminal does, until stop closes.
func (a *App) watch(stop <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			a.poll(a.Do)
		}
	}
}

// poll reads the file once and hands a change to dispatch, which runs it on the
// UI thread. It reads the same file without dispatching anything while the
// content is unchanged. A panic in one read is logged and the watcher goes on:
// a bad setting is not a reason to stop watching the rest.
func (a *App) poll(dispatch func(func())) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("wt-go: settings watcher: %v\n%s", r, debug.Stack())
		}
	}()
	data, ok, err := sys.ReadSettings()
	if err != nil {
		note := fmt.Sprintf(i18n.T("settings.json: %v"), err)
		if a.file.setReadError(note) {
			dispatch(func() { a.setNote(note) })
		}
		return
	}
	if note, had := a.file.clearReadError(); had {
		dispatch(func() { a.clearNote(note) })
	}
	if !ok {
		return
	}
	if !a.file.changed(data) {
		a.file.clearBad()
		return
	}
	if _, err := store.Parse(data); err != nil && !a.file.settledBad(data) {
		return
	}
	a.file.markSeen(data)
	dispatch(func() { a.reloadSafely(data) })
}

// reload takes the settings the file holds now, when they are settings: a file
// that is not keeps the ones the window shows, and says why.
func (a *App) reload(data []byte) {
	a.file.markSeen(data)
	if a.savedData != nil && bytes.Equal(data, a.savedData) {
		if a.badData != nil {
			a.badData, a.configErr = nil, ""
			a.setNote("")
		}
		if a.note == i18n.T("settings.json changed on disk; discard your changes to take it.") {
			a.setNote("")
		}
		return
	}
	if a.badData != nil && bytes.Equal(data, a.badData) {
		return
	}
	if a.badData != nil {
		a.badData, a.configErr = nil, ""
	}
	if a.dirty() {
		a.setNote(i18n.T("settings.json changed on disk; discard your changes to take it."))
		return
	}
	settings, err := store.Parse(data)
	if err != nil {
		a.badData = data
		a.reject(data, err)
		return
	}
	a.settings, a.saved, a.savedData, a.badData = settings, settings.Clone(), data, nil
	a.configErr = ""
	if a.editing != nil {
		a.editProfile(a.editing.ID)
	}
	a.setNote("")
	a.redraw()
}

// reject keeps what the file held that was not settings, and says on the
// settings page why the window is not showing it. The note is kept apart from
// what the last save said, so switching pages or discarding a draft does not
// clear it.
func (a *App) reject(data []byte, err error) {
	note := configError(data, err)
	if note == a.configErr {
		return
	}
	a.note = ""
	a.configErr = note
	a.redraw()
	if a.win != nil && !a.closed && a.settingsTab == nil {
		a.OpenSettings()
	}
}

// reloadSafely runs reload on the UI thread. A panic is logged and shown on the
// settings page instead of escaping: the watcher must keep running, and a
// reload that failed must be visible.
func (a *App) reloadSafely(data []byte) {
	a.safely(func() { a.reload(data) })
}

// safely runs fn, turning a panic into a note. It is the boundary between a
// background read and the window: nothing a reload does may take the watcher
// down with it.
func (a *App) safely(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("wt-go: settings reload: %v\n%s", r, debug.Stack())
			a.setNote(fmt.Sprintf(i18n.T("settings.json: %v"), r))
		}
	}()
	fn()
}

// fileState is what the watcher knows of settings.json: the content it has
// already handed to the UI, bad content it is waiting to see twice, and the
// read error it last showed. The mutex keeps the watcher goroutine and the UI
// thread from racing on it.
type fileState struct {
	mu      sync.Mutex
	seen    []byte
	bad     []byte
	readErr string
}

// changed reports whether data differs from the content the watcher last
// handed to the UI. It does not record data: only content that is dispatched
// is seen.
func (f *fileState) changed(data []byte) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return !bytes.Equal(data, f.seen)
}

// markSeen records the content the watcher handed to the UI.
func (f *fileState) markSeen(data []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seen = append([]byte(nil), data...)
	f.bad = nil
}

// clearBad forgets bad content that was seen once but never reported, as when
// an editor restores the file it was writing.
func (f *fileState) clearBad() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.bad = nil
}

// settledBad reports whether data is the same bad content the watcher saw on
// its last read. A file an editor is still writing changes from read to read;
// waiting for the same bytes twice keeps its half-written state from being
// archived once a second.
func (f *fileState) settledBad(data []byte) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !bytes.Equal(data, f.bad) {
		f.bad = append([]byte(nil), data...)
		return false
	}
	return true
}

// setReadError records a read error and reports whether it is new.
func (f *fileState) setReadError(note string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.readErr == note {
		return false
	}
	f.readErr = note
	return true
}

// clearReadError forgets a read error and answers with the note it showed.
func (f *fileState) clearReadError() (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.readErr == "" {
		return "", false
	}
	note := f.readErr
	f.readErr = ""
	return note, true
}

// settingsNote is what the settings page says: what the last action did, and
// why the file is not settings while that is true. The action comes first, so a
// save, a write or a read that just failed is what the reader sees; the file's
// reason follows instead of hiding it, and stays once the action note is gone.
func (a *App) settingsNote() string {
	switch {
	case a.note == "":
		return a.configErr
	case a.configErr == "":
		return a.note
	}
	return a.note + " — " + a.configErr
}

// clearNote drops a note that is still the one that was shown.
func (a *App) clearNote(note string) {
	if a.note == note {
		a.setNote("")
	}
}

// configError is the note the settings page shows when the file is not the
// settings: what is wrong, and where the content that was wrong is kept.
func configError(data []byte, err error) string {
	if data == nil {
		return fmt.Sprintf(i18n.T("settings.json: %v"), err)
	}
	path, aerr := sys.ArchiveSettings(data)
	if aerr != nil {
		return fmt.Sprintf(i18n.T("settings.json is not valid: %v; the original could not be kept: %v"), err, aerr)
	}
	return fmt.Sprintf(i18n.T("settings.json is not valid: %v; the original was kept at %s"), err, path)
}

// setNote says what the settings page says, and redraws when it is something
// new. It answers whether the note changed.
func (a *App) setNote(note string) bool {
	if a.note == note {
		return false
	}
	a.note = note
	a.redraw()
	return true
}
