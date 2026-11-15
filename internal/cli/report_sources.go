package cli

import (
	"strings"
	"time"
	"wtgo/internal/fault"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
)

func settingsReport() SettingsReport {
	out := SettingsReport{}
	file, err := sys.SettingsFileInfo()
	out.Path = file.Path
	if err != nil {
		out.Error = err.Error()
		return out
	}
	out.Exists, out.Size = file.Exists, file.Size
	if file.Exists {
		out.ModTime = file.ModTime.Format(time.RFC3339)
	}
	settings, err := readSettings()
	if err != nil {
		out.Error = err.Error()
		return out
	}
	out.Language = settings.Language
	out.Theme = settings.Theme
	out.ColorScheme = settings.ColorScheme
	out.DefaultProfile = settings.DefaultProfile
	out.Keybindings = len(settings.Keybindings)
	return out
}

// readSettings is settings.json parsed, with the defaults standing in for a
// file that is not there: a machine that never ran the window still answers
// with the settings the window would use.
func readSettings() (store.Settings, error) {
	data, ok, err := sys.ReadSettings()
	if err != nil {
		return store.Default(), err
	}
	if !ok {
		return store.Default(), nil
	}
	return store.Parse(data)
}

func nativeReport() NativeReport {
	if p := strings.TrimSpace(sys.Getenv("MYGO_GHOSTTY_VT")); p != "" {
		return NativeReport{Path: p, Env: true, Found: true}
	}
	path, err := sys.NativeLibrary()
	if err != nil {
		return NativeReport{}
	}
	return NativeReport{Path: path, Found: true}
}

// conptyReport is the conpty.dll the terminals would be made in. Windows' own
// console is no file, and this is where a copy of the exe on its own shows: the
// terminals run, and programs that ask for the colors and the mouse hear
// nothing back.
func conptyReport() ConPTYReport {
	named := strings.TrimSpace(sys.Getenv("MYGO_CONPTY"))
	path, err := sys.ConPTYLibrary()
	if err != nil {
		return ConPTYReport{}
	}
	return ConPTYReport{Path: path, Env: named != "" && named != "system", Found: true}
}

func shells() []ShellReport {
	found := sys.Discover()
	out := make([]ShellReport, 0, len(found))
	for _, s := range found {
		out = append(out, ShellReport{ID: s.ID, Name: profile.Name(s.ID, s.Command), Command: s.Command})
	}
	return out
}

// profiles merges the configured profiles with the found shells the way the
// window does, and asks the machine whether each one can start. A file that
// did not parse leaves the found shells alone, as it does in the window.
func profiles(found []ShellReport, settingsOK bool) []ProfileReport {
	known := make([]profile.Found, 0, len(found))
	for _, s := range found {
		known = append(known, profile.Found{ID: s.ID, Command: s.Command})
	}
	var configured []store.Profile
	if settingsOK {
		if s, err := readSettings(); err == nil {
			configured = s.Profiles
		}
	}
	merged := profile.Resolve(configured, known)
	out := make([]ProfileReport, 0, len(merged))
	for _, p := range merged {
		source := "configured"
		if p.Discovered {
			source = "found"
		}
		_, ok := sys.LookCommand(p.Command)
		out = append(out, ProfileReport{ID: p.ID, Name: p.Name, Command: p.Command, Source: source, Available: ok})
	}
	return out
}

// checks are the questions a diagnostic asks, in the order the window meets
// them.
func crashes() []CrashReport {
	paths, err := sys.CrashLogs()
	if err != nil {
		return nil
	}
	out := make([]CrashReport, 0, len(paths))
	for _, p := range paths {
		c := CrashReport{Path: p}
		if text, _, err := sys.Tail(p, 64<<10); err == nil {
			if line, ok := fault.In(text); ok {
				c.Fault = line
			}
		}
		out = append(out, c)
	}
	return out
}

// logReport is the last fault the window wrote to wt-go.log: the failures
// that do not stop the program and so leave no crash log.
func logReport() LogReport {
	path, err := sys.LogPath()
	if err != nil {
		return LogReport{}
	}
	out := LogReport{Path: path}
	if text, _, err := sys.Tail(path, 64<<10); err == nil {
		if line, ok := fault.In(text); ok {
			out.Fault = line
		}
	}
	return out
}
