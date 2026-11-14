package app

import (
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/transfer"
	"log"
	"os"
	"strings"
	"wtgo/internal/fault"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
)

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

// openFirst opens the tab the settings start with; with no profile named it
// runs the shell of this machine. `startOn: new` says the window starts on a
// new tab: the shell of this machine, whatever defaultProfile names. Session
// restore is not part of that choice yet, so the two words differ only here.
func (a *App) openFirst() {
	a.openTab(a.firstTab(), "")
}

// firstTab is the profile of the tab the window opens with: the default one,
// or the device's for `startOn: new`.
func (a *App) firstTab() profile.Profile {
	if a.settings.StartOn == store.StartNew {
		return profile.Default(a.profiles, "", a.device)
	}
	return profile.Default(a.profiles, a.settings.DefaultProfile, a.device)
}

// startDirectory is where a tab starts when neither it nor its profile names
// a directory: what -d said for this run, else the setting.
func (a *App) startDirectory() string {
	if a.startDir != "" {
		return a.startDir
	}
	return a.settings.StartDirectory
}

// openTab opens a tab that runs p, in dir when dir is not "", else where p,
// the command line or the settings, and this process say, in that order.
func (a *App) openTab(p profile.Profile, dir string) {
	if dir == "" {
		dir = p.Dir
	}
	if dir == "" {
		dir = a.startDirectory()
	}
	// Either source may name the directory with a word: "current" is where we
	// were started, which is what a file manager's "open a terminal here"
	// passes and what mygo would otherwise replace with the home directory;
	// "home" is the user's. A path goes on to ExpandPath as it is.
	dir = sys.StartDir(dir)
	dir = sys.ExpandPath(dir)
	var term *terminal.Terminal
	opts := a.termOptions(p, dir)
	// The paste that waits for an answer needs the terminal that takes it,
	// which exists once the options are read.
	opts.OnPaste = func(d transfer.Data) bool { return a.askPaste(term, d) }
	term, err := a.newTerminal(opts)
	if err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.ChildFailed, startFailed(p, err)))
		return
	}
	tab := pane.New(term)
	tab.ProfileID, tab.ProfileName, tab.Color, tab.Icon = p.ID, p.Name, p.Color, p.Icon
	// Only a profile that names no program leaves the shell to the terminal,
	// and only then does the strip have to say which shell that was.
	if len(p.CommandLine()) == 0 {
		tab.Runs = programName(a.device.Command)
	}
	a.insertTab(tab)
	a.revealTab()
	a.syncTitle()
	a.redraw()
}

// startFailed is what a log line says about a tab that could not start: the
// profile it was to run, or just the error when it names none, because the
// program that failed was the one the terminal picked.
func startFailed(p profile.Profile, err error) string {
	if p.Name == "" {
		return err.Error()
	}
	return p.Name + ": " + err.Error()
}
