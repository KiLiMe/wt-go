package app

import (
	"bytes"
	"log"
	"strconv"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/internal/theme"
	"wtgo/internal/ui"
	"wtgo/locales"
)

// OpenSettings shows the settings, which one tab holds: it opens that tab, or
// a new one. The terminal's overlays go: the settings tab is not a terminal, so
// a bar that asks the one in front for text has nothing left to ask.
func (a *App) OpenSettings() {
	a.closeOverlays()
	if a.settingsTab == nil {
		a.settingsTab = &pane.Tab{Name: locales.Settings()}
		a.settingsPage = ui.PageGeneral
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
	// The list is a menu of the machine's fonts, so it is asked for when the
	// page that shows it opens, not before the window does. The appearance
	// page shows it too, since the font setting is part of how the window looks.
	if a.settingsPage == ui.PageProfiles || a.settingsPage == ui.PageAppearance {
		a.ensureFontFamilies(a.settings.Font.Size)
	}
	return ui.Settings{
		Page:          a.settingsPage,
		KeysCache:     &a.keysDerived,
		SchemesCache:  &a.schemesDerived,
		Draft:         &a.settings,
		Dirty:         a.dirty(),
		Note:          a.settingsNote(),
		Recording:     a.recording,
		Validation:    a.validation,
		Filter:        &a.filter,
		ActionFilter:  &a.actionSearch,
		OnlyChanged:   a.onlyChangedKeys,
		KeysExpanded:  a.numberedKeys,
		Profile:       a.editing,
		Args:          &a.editingArgs,
		Glyph:         &a.editingGlyph,
		Profiles:      fileProfiles(a.settings),
		ProfileFilter: &a.profileSearch,
		Search:        &a.settingsSearch,
		Known:         knownProfiles(a.profiles),
		Device:        a.device.Name,
		Schemes:       a.schemeList(),
		FontFamilies:  a.fonts.families,
	}
}

// schemeList is the color schemes the settings page lists, built once per
// settings the window applied and kept until they change.
func (a *App) schemeList() []ui.Scheme {
	if a.schemes == nil {
		a.schemes = uiSchemes(a.settings)
	}
	return a.schemes
}

// uiSchemes are the color schemes the color schemes page lists, each saying
// whether the settings write it: the page offers to remove those alone.
func uiSchemes(s store.Settings) []ui.Scheme {
	custom := make(map[string]bool, len(s.Schemes))
	for _, sc := range s.Schemes {
		custom[strings.ToLower(sc.Name)] = true
	}
	names := theme.Schemes(s)
	out := make([]ui.Scheme, 0, len(names))
	for _, name := range names {
		out = append(out, ui.Scheme{Name: name, Custom: custom[strings.ToLower(name)]})
	}
	return out
}

// fileProfiles are the profiles the file holds, as the settings list them: the
// hidden ones too, since the page edits hiding and is where it is undone.
func fileProfiles(s store.Settings) []ui.Profile {
	out := make([]ui.Profile, 0, len(s.Profiles))
	for _, p := range s.Profiles {
		out = append(out, ui.Profile{ID: p.ID, Name: p.Name, Color: p.Color, Icon: profile.IconFor(p), Discovered: p.Discovered, Hidden: p.Hidden})
	}
	return out
}

// knownProfiles are the profiles the window runs, as the settings list them. A
// hidden one is not one of them: a setting that names it runs the system shell,
// so offering it would be a choice that does not hold.
func knownProfiles(profiles []profile.Profile) []ui.Profile {
	out := make([]ui.Profile, 0, len(profiles))
	for _, p := range profiles {
		if p.Hidden {
			continue
		}
		out = append(out, ui.Profile{ID: p.ID, Name: p.Name, Color: p.Color, Icon: p.Icon, Discovered: p.Discovered})
	}
	return out
}

// SelectSettingsPage shows a page of the settings. The profiles are one page
// with the profile in front beside its list, so opening it with none in front
// opens the first one there is: the page is never a list with nothing to edit.
func (a *App) SelectSettingsPage(page string) {
	if a.settingsTab == nil {
		return
	}
	a.settingsPage = page
	a.settingsSearch = ""
	a.recording = ""
	a.editNone()
	if id, ok := ui.ProfileID(page); ok {
		a.editProfile(id)
		a.profilePageOpened()
		return
	}
	if page == ui.PageProfiles {
		a.openFirstProfile()
	}
	a.redraw()
}

// openFirstProfile opens the first profile the page lists, which is what the
// profiles page shows when it opens with none in front.
func (a *App) openFirstProfile() {
	for _, p := range fileProfiles(a.settings) {
		a.editProfile(p.ID)
		if a.editing != nil {
			a.settingsPage = ui.ProfilePage(p.ID)
		}
		return
	}
}

// profilePageOpened shows the page of the profile in front, or the list alone
// when the window has no such profile: a page that names a profile the file
// does not hold edits nothing.
func (a *App) profilePageOpened() {
	if a.editing == nil {
		a.settingsPage = ui.PageProfiles
		// A profile the list holds while none can be edited, such as one the
		// search box keeps, is not the page's to open.
		a.openFirstProfile()
	}
	a.redraw()
}

// SetDefaultProfile names the profile the window opens first; "" follows the
// shell of the machine. It is what the select of the page asks for while the
// profile the file names is not one the page offers: the select shows the
// device, and a click on it writes the choice down instead of the page doing it.
func (a *App) SetDefaultProfile(id string) {
	a.settings.DefaultProfile = id
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
	a.editingGlyph = profile.Glyph(p.Icon)
}

// editNone closes the profile page's copy: nothing is being edited.
func (a *App) editNone() {
	a.editing, a.editingFrom, a.editingArgs, a.editingGlyph = nil, store.Profile{}, "", ""
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
// that changed on disk since the window last read or saved it, and refuses a
// draft no window can run with, which it says where the page is looking.
func (a *App) SaveSettings() {
	if a.editing != nil {
		a.editing.Args = splitArgs(a.editingArgs)
		// The box is the mark while it holds a character, and clearing it is
		// a profile that wears none: a shape it kept, or the one its command
		// asks for.
		if g := profile.Glyph(a.editingGlyph); g != "" {
			a.editing.Icon = g
		} else if profile.IconKey(a.editing.Icon) == "" {
			a.editing.Icon = ""
		}
		a.upsertProfile(*a.editing)
	}
	if err := a.settings.Validate(); err != nil {
		// The draft stays as it is: what is wrong with it is something the
		// person corrects on the page, frame after frame.
		a.validation = err.Error()
		a.setNote(err.Error())
		return
	}
	a.validation = ""
	data, err := store.Marshal(a.settings)
	if err != nil {
		a.setNote(err.Error())
		return
	}
	unchanged, err := a.settingsFileUnchanged()
	if err != nil {
		a.setNote(locales.ErrSettings(err.Error()))
		return
	}
	if !unchanged {
		a.setNote(locales.SettingsChangedNote())
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
	a.setNote(locales.Saved())
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
	a.validation = ""
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
	a.validation = ""
	a.recording = ""
	if a.editing != nil {
		a.editProfile(a.editing.ID)
		if a.editing == nil {
			// The profile was never saved, so there is no page to edit: the
			// list is where the others are.
			a.settingsPage = ui.PageProfiles
		}
	}
	a.refreshSettings()
	a.redraw()
}

// refreshSettings reads settings.json now, as discarding a draft does, because
// the watcher may have seen a change while the draft was in the way.
func (a *App) refreshSettings() {
	data, ok, err := sys.ReadSettings()
	if err != nil {
		a.setNote(locales.ErrSettings(err.Error()))
		return
	}
	if !ok {
		a.forgetBadFile()
		return
	}
	a.reloadSafely(data)
}

// forgetBadFile drops what the settings page says about a file that is not
// settings, for a file that is gone: the window wrote none until the first
// launch, so a file the reader deleted is not something to keep explaining.
// The reason is the page's own, so nothing is shown while there is none to
// drop, which is the first launch on a machine that has no file yet.
func (a *App) forgetBadFile() bool {
	if a.configErr == "" && a.badData == nil {
		return false
	}
	a.configErr, a.badData = "", nil
	// The content the watcher remembers goes too, so that a file written back
	// to what it held before it was deleted is taken as a new file.
	a.file.forgetSeen()
	a.redraw()
	return true
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
// replaces the keys the action had with that one, or adds the keybinding it did
// not have. An action runs on one key, so the keys it had go: leaving one
// behind would keep running it, invisibly, on a key the page shows as free.
func (a *App) SetKeybinding(action, key string) {
	if key != "" {
		out := make([]store.Keybinding, 0, len(a.settings.Keybindings)+1)
		found := false
		for _, kb := range a.settings.Keybindings {
			if kb.Action != action {
				out = append(out, kb)
				continue
			}
			if !found {
				// The first keybinding of the action keeps its place in the
				// file, so that the list does not reorder as keys change.
				out = append(out, store.Keybinding{Key: key, Action: action})
				found = true
			}
		}
		if !found {
			out = append(out, store.Keybinding{Key: key, Action: action})
		}
		a.settings.Keybindings = out
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

// ResetKeybinding gives an action the keys the catalog starts it on, in the
// place of the first key it ran on: one row at a time, so that the rest of the
// file, and where the action sat in it, stay as they are.
func (a *App) ResetKeybinding(action string) {
	rest := make([]store.Keybinding, 0, len(a.settings.Keybindings))
	at := -1
	for _, kb := range a.settings.Keybindings {
		if kb.Action == action {
			if at < 0 {
				at = len(rest)
			}
			continue
		}
		rest = append(rest, kb)
	}
	if at < 0 {
		at = len(rest)
	}
	out := make([]store.Keybinding, 0, len(rest)+1)
	out = append(out, rest[:at]...)
	for _, kb := range store.DefaultKeybindings() {
		if kb.Action == action {
			out = append(out, kb)
		}
	}
	out = append(out, rest[at:]...)
	a.settings.Keybindings = out
	a.recording = ""
	a.redraw()
}

// ToggleChangedKeys narrows the actions page to the actions the file binds on
// keys of its own, or lists them all again.
func (a *App) ToggleChangedKeys() {
	a.onlyChangedKeys = !a.onlyChangedKeys
	a.redraw()
}

// ToggleNumberedKeys unfolds the numbered actions the actions page folds away,
// or folds them again.
func (a *App) ToggleNumberedKeys() {
	a.numberedKeys = !a.numberedKeys
	a.redraw()
}

// PickScheme paints with a color scheme, as a click on the color schemes page
// asks: the choice of the page edits the draft in place, so the window applies
// the scheme as it does any other change.
func (a *App) PickScheme(name string) {
	a.settings.ColorScheme = name
	a.redraw()
}

// CopyScheme makes a color scheme of the settings from the one the terminal
// paints with, and picks it.
func (a *App) CopyScheme() {
	th, ok := theme.Scheme(a.settings, a.settings.ColorScheme)
	if !ok {
		return
	}
	scheme := theme.Stored(th)
	scheme.Name = a.customSchemeName(schemeBaseName(a.settings.ColorScheme))
	a.settings.Schemes = append(a.settings.Schemes, scheme)
	a.settings.ColorScheme = scheme.Name
	a.note = locales.SchemeCopied()
	a.redraw()
}

// customSuffix starts the mark a copied scheme carries on its name.
const customSuffix = " (custom"

// schemeBaseName is the name a copy starts from: the name the copies before it
// added their marks to, without them. A copy copies the scheme in front, which
// may be a copy itself, so the marks nest unless they come off again.
func schemeBaseName(name string) string {
	for {
		i := strings.LastIndex(name, customSuffix)
		if i <= 0 {
			return name
		}
		// What follows the mark is ")" or a number and ")", as
		// customSchemeName writes them.
		n := strings.TrimSpace(strings.TrimSuffix(name[i+len(customSuffix):], ")"))
		if n != "" && !digits(n) {
			return name
		}
		name = name[:i]
	}
}

// digits reports whether s is a run of decimal digits.
func digits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// customSchemeName is a name no scheme holds yet for a copy of the scheme
// named base: its own name with a mark, numbered once that is taken.
func (a *App) customSchemeName(base string) string {
	name := base + customSuffix + ")"
	for i := 2; a.schemeNamed(name); i++ {
		name = base + customSuffix + " " + strconv.Itoa(i) + ")"
	}
	return name
}

// RemoveScheme drops a color scheme the settings write, which is the only kind
// the color schemes page offers to remove: the terminal keeps painting with the
// scheme the one removed was a copy of, while that is still there.
func (a *App) RemoveScheme(name string) {
	out := a.settings.Schemes[:0]
	removed := false
	for _, s := range a.settings.Schemes {
		if strings.EqualFold(s.Name, name) {
			removed = true
			continue
		}
		out = append(out, s)
	}
	if !removed {
		return
	}
	a.settings.Schemes = out
	if strings.EqualFold(a.settings.ColorScheme, name) {
		a.settings.ColorScheme = a.schemeAfterRemoving(name)
	}
	a.note = locales.SchemeRemoved()
	a.redraw()
}

// schemeAfterRemoving is the scheme to paint with once one is gone: the one it
// was a copy of, while that is still there, else the one the defaults name.
func (a *App) schemeAfterRemoving(name string) string {
	for _, candidate := range []string{schemeBaseName(name), store.Default().ColorScheme} {
		if _, ok := theme.Scheme(a.settings, candidate); ok {
			return candidate
		}
	}
	return ""
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

// AddProfile adds a profile to the file's list, opens it in the fields of the
// profiles page, and keeps the search box of that page from hiding it.
func (a *App) AddProfile() {
	p := store.Profile{ID: a.newProfileID(), Name: locales.NewProfile()}
	a.settings.Profiles = append(a.settings.Profiles, p)
	a.profileSearch = ""
	a.settingsPage = ui.ProfilePage(p.ID)
	a.editing, a.editingFrom = &p, p
	// The page shows the arguments it holds, so a line left from the profile
	// that was open before would become this profile's arguments on save.
	a.editingArgs = ""
	a.editingGlyph = ""
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
	if p, ok := profileByID(a.settings, id); ok && p.Discovered {
		// The window finds this shell again on the next start, so hiding it
		// is the only removal that holds: dropping it would put it back.
		p.Hidden = true
		a.upsertProfile(p)
		a.settingsPage = ui.PageProfiles
		a.editNone()
		a.nextProfile(id)
		a.recording = ""
		a.redraw()
		return
	}
	out := make([]store.Profile, 0, len(a.settings.Profiles))
	for _, p := range a.settings.Profiles {
		if p.ID != id {
			out = append(out, p)
		}
	}
	a.settings.Profiles = out
	a.settingsPage = ui.PageProfiles
	a.editNone()
	a.nextProfile(id)
	a.recording = ""
	a.redraw()
}

// nextProfile opens the profile the list now starts at, once the one that was
// in front is gone or hidden from the list: the fields beside the list show
// another profile rather than go empty. A profile that moved out of a list
// the search box keeps is the box's to need, not the page's.
func (a *App) nextProfile(gone string) {
	a.openFirstProfile()
	if a.editing != nil && a.editing.ID == gone {
		a.editNone()
		a.settingsPage = ui.PageProfiles
	}
}

// HideProfile hides the profile with an id, through the draft the page edits:
// the list goes on listing it, marked, and saving writes it to the file.
func (a *App) HideProfile(id string) {
	p, ok := profileByID(a.settings, id)
	if !ok || p.Hidden {
		return
	}
	p.Hidden = true
	a.upsertProfile(p)
	if a.editing != nil && a.editing.ID == id {
		a.editing.Hidden = true
	}
	a.redraw()
}

// ResetProfile puts the profile with an id back the way this machine has it:
// the probe gives it the program and the arguments of the shell it is, and the
// built-in table gives it back the name and the dot. A profile the machine no
// longer has the program for is hidden, since that is all the menu can hold,
// and a profile that is hidden comes back, since the one thing a reset can say
// about a program it found is that it is there.
func (a *App) ResetProfile(id string) {
	p, ok := profileByID(a.settings, id)
	if !ok || !p.Discovered {
		return
	}
	f, present := foundByID(found(), id)
	a.resetProfile(p, f, present)
}

// resetProfile is the reset a probe of this machine has already answered.
func (a *App) resetProfile(p store.Profile, f profile.Found, present bool) {
	if !present {
		p.Hidden = true
		a.upsertProfile(p)
		if a.editing != nil && a.editing.ID == p.ID {
			a.editing.Hidden = true
		}
		a.setNote(locales.ResetProfileGone(p.Name))
		return
	}
	p = profile.Reset(p, f)
	a.upsertProfile(p)
	if a.editing != nil && a.editing.ID == p.ID {
		*a.editing = p
		a.editingArgs = strings.Join(p.Args, "\n")
		a.editingGlyph = profile.Glyph(p.Icon)
	}
	a.setNote(locales.ResetProfileDone(p.Name))
}

// foundByID is the shell a probe found, by the id a profile has.
func foundByID(shells []profile.Found, id string) (profile.Found, bool) {
	for _, f := range shells {
		if f.ID == id {
			return f, true
		}
	}
	return profile.Found{}, false
}

// DuplicateProfile copies a profile, and shows the copy's page.
func (a *App) DuplicateProfile(id string) {
	p, ok := profileByID(a.settings, id)
	if !ok {
		return
	}
	p.ID = a.newProfileID()
	p.Name = p.Name + " " + locales.CopySuffix()
	p.Discovered = false
	// Args is a slice, so a copy that kept it would share its lines with the
	// profile it copies, and the page's edits would write both.
	p.Args = append([]string(nil), p.Args...)
	a.settings.Profiles = append(a.settings.Profiles, p)
	a.profileSearch = ""
	a.settingsPage = ui.ProfilePage(p.ID)
	a.editing, a.editingFrom = &p, p
	a.editingArgs = strings.Join(p.Args, "\n")
	a.editingGlyph = profile.Glyph(p.Icon)
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
	if profile.Glyph(a.editingGlyph) != profile.Glyph(a.editingFrom.Icon) {
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
	locale(s.Language)
	a.appearance()
	a.keys = bindings(s.Keybindings)
	a.profiles = profile.Resolve(s.Profiles, a.found)
	font := a.terminalFont()
	colors := theme.Terminal(s)
	a.pane = ui.Face{Background: colors.Background, Foreground: colors.Foreground}
	copyOnSelect := terminal.CopyOnSelectNone
	if s.CopyOnSelect {
		copyOnSelect = terminal.CopyOnSelectClipboard
	}
	multiplier := scrollMultiplier(s.ScrollMultiplier)
	for _, tab := range a.panes.Tabs {
		for _, term := range tab.Terminals() {
			term.SetFont(font)
			term.SetTheme(colors, nil)
			term.SetCopyOnSelect(copyOnSelect)
			term.SetScrollMultiplier(multiplier)
			term.SetMinimumContrast(s.MinimumContrast)
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
	// What the settings page lists is built from the settings, so a new set is
	// a new list: it is built again when the page asks.
	a.schemes = nil
	a.redraw()
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
		return locales.ErrSettings(err.Error())
	}
	path, aerr := sys.ArchiveSettings(data)
	if aerr != nil {
		return locales.ErrSettingsInvalidKeep(err.Error(), aerr.Error())
	}
	return locales.ErrSettingsInvalid(err.Error(), path)
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

// SetProfileIcon gives the profile the page edits a mark: a preset, or "" for
// the one its command asks for. The character box is cleared, because the
// marks of a profile are one setting, and a shape and a character cannot both
// be the one it holds.
func (a *App) SetProfileIcon(mark string) {
	if a.editing == nil {
		return
	}
	a.editing.Icon = mark
	a.editingGlyph = ""
	a.redraw()
}
