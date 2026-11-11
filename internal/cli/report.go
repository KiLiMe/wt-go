// Package cli is wt-go without a window: the report `wt-go diag` prints and
// the MCP server `wt-go mcp` runs for an editor or an agent. It holds no OS
// knowledge of its own — every path and file comes from internal/sys — so one
// report answers a terminal, a JSON reader and a tool call alike.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"wtgo/internal/fault"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
)

// Report is what wt-go knows about itself from outside: a fresh process
// reading the files the window reads.
type Report struct {
	Program    string          `json:"program"`
	Version    string          `json:"version"`
	Revision   string          `json:"revision,omitempty"`
	Go         string          `json:"go"`
	Platform   string          `json:"platform"`
	Executable string          `json:"executable,omitempty"`
	Settings   SettingsReport  `json:"settings"`
	Native     NativeReport    `json:"native"`
	Shells     []ShellReport   `json:"shells"`
	Profiles   []ProfileReport `json:"profiles"`
	Crashes    []CrashReport   `json:"crash_logs"`
	Log        LogReport       `json:"log"`
	Checks     []Check         `json:"checks"`
}

// SettingsReport is the file the settings live in, and what it holds.
type SettingsReport struct {
	Path    string `json:"path"`
	Exists  bool   `json:"exists"`
	Size    int64  `json:"size,omitempty"`
	ModTime string `json:"mod_time,omitempty"`
	// Error is why the window would ignore the file, empty when it parses.
	Error          string `json:"error,omitempty"`
	Language       string `json:"language,omitempty"`
	Theme          string `json:"theme,omitempty"`
	ColorScheme    string `json:"color_scheme,omitempty"`
	DefaultProfile string `json:"default_profile,omitempty"`
	Keybindings    int    `json:"keybindings,omitempty"`
}

// NativeReport is the libghostty-vt the terminal plugin would load.
type NativeReport struct {
	Path  string `json:"path,omitempty"`
	Env   bool   `json:"from_env,omitempty"`
	Found bool   `json:"found"`
}

// ShellReport is one shell the machine has.
type ShellReport struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Command string `json:"command"`
}

// ProfileReport is a profile after the configured and found ones were merged,
// with whether this machine can run it.
type ProfileReport struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Command   string `json:"command"`
	Source    string `json:"source"`
	Available bool   `json:"available"`
}

// CrashReport is one crash log, newest first.
type CrashReport struct {
	Path  string `json:"path"`
	Fault string `json:"fault,omitempty"`
}

// LogReport is what the window said to its log, and the last fault in it: a
// keybinding that did not parse or a shell that would not start is written
// there once and read back here.
type LogReport struct {
	Path  string `json:"path,omitempty"`
	Fault string `json:"fault,omitempty"`
}

// Check is one thing a diagnostic asks, with the code that answers for it
// when it fails.
type Check struct {
	Code   fault.Code `json:"code"`
	OK     bool       `json:"ok"`
	Detail string     `json:"detail"`
}

// Collect is wt-go as the files on disk say it is.
func Collect() Report {
	r := Report{
		Program:  "wt-go",
		Go:       runtime.Version(),
		Platform: runtime.GOOS + "/" + runtime.GOARCH,
	}
	r.Version, r.Revision = buildInfo()
	if exe, err := sys.Executable(); err == nil {
		r.Executable = exe
	}
	r.Settings = settingsReport()
	r.Native = nativeReport()
	r.Shells = shells()
	r.Profiles = profiles(r.Shells, r.Settings.Error == "")
	r.Crashes = crashes()
	r.Log = logReport()
	r.Checks = checks(r)
	return r
}

// buildInfo is the version the toolchain stamped into the exe and the commit
// it was built from: a report says which build answered.
func buildInfo() (version, revision string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown", ""
	}
	version = info.Main.Version
	if version == "" || version == "(devel)" {
		version = "devel"
	}
	version = tidyVersion(version)
	modified := false
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if len(revision) > 12 {
		revision = revision[:12]
	}
	if modified && revision != "" {
		revision += "+dirty"
	}
	return version, revision
}

// pseudo is the version `go build` stamps into a build made from a commit
// that no tag names: v0.1.1-0.20261110153946-e9f374cbc6b0. A report keeps the
// part that says something and drops the rest, the commit being printed
// beside it already.
var pseudo = regexp.MustCompile(`^(v[0-9][^-]*)-(0\.[0-9]{14}|[0-9]{14})-[0-9a-f]+`)

// tidyVersion turns a pseudo-version into v0.1.1-dev.
func tidyVersion(v string) string {
	if m := pseudo.FindStringSubmatch(v); m != nil {
		return m[1] + "-dev"
	}
	return v
}

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

// checks are the questions a diagnostic asks, in the order the window meets
// them.
func checks(r Report) []Check {
	out := make([]Check, 0, 6)
	add := func(c fault.Code, ok bool, detail string) {
		out = append(out, Check{Code: c, OK: ok, Detail: detail})
	}
	dirOK := r.Settings.Path != ""
	add(fault.NoConfigDir, dirOK, r.Settings.Path)
	switch {
	case r.Settings.Error == "":
		add(fault.BadSettings, true, "settings.json parses"+keybindings(r.Settings.Keybindings))
	case strings.Contains(r.Settings.Error, " must be "):
		add(fault.BadField, false, r.Settings.Error)
	default:
		add(fault.BadSettings, false, r.Settings.Error)
	}
	add(fault.NativeLib, r.Native.Found, or(r.Native.Path, "no libghostty-vt in $MYGO_GHOSTTY_VT, next to the exe, or under the mygo natives cache"))
	add(fault.NoShell, len(r.Profiles) > 0, count(len(r.Profiles), "profile"))
	// One profile that cannot start is a tab that opens empty, which the user
	// asked for: the check fails and names it.
	why := missing(r.Profiles)
	if why == "" {
		why = count(len(r.Profiles), "profile") + " can start"
	}
	add(fault.ChildFailed, len(r.Profiles) > 0 && missing(r.Profiles) == "", why)
	if len(r.Crashes) > 0 {
		c := r.Crashes[0]
		add(fault.Crashed, c.Fault == "", or(c.Fault, c.Path))
	}
	return out
}

// Redact is the report with the machine's own paths replaced by the
// placeholders that name them, so it can be pasted somewhere public: the
// report still says which files matter without saying whose machine it is.
func (r Report) Redact(pairs [][2]string) Report {
	if len(pairs) == 0 {
		return r
	}
	s := func(v string) string { return redact(v, pairs) }
	r.Executable = s(r.Executable)
	r.Settings.Path = s(r.Settings.Path)
	r.Settings.Error = s(r.Settings.Error)
	r.Native.Path = s(r.Native.Path)
	profiles := make([]ProfileReport, len(r.Profiles))
	for i, p := range r.Profiles {
		p.Command = s(p.Command)
		profiles[i] = p
	}
	r.Profiles = profiles
	crashes := make([]CrashReport, len(r.Crashes))
	for i, c := range r.Crashes {
		c.Path, c.Fault = s(c.Path), s(c.Fault)
		crashes[i] = c
	}
	r.Crashes = crashes
	r.Log.Path = s(r.Log.Path)
	r.Log.Fault = s(r.Log.Fault)
	checks := make([]Check, len(r.Checks))
	for i, c := range r.Checks {
		c.Detail = s(c.Detail)
		checks[i] = c
	}
	r.Checks = checks
	return r
}

// redact replaces each path with its placeholder, longest first so a nested
// directory wins over the home it sits under.
func redact(s string, pairs [][2]string) string {
	for _, p := range pairs {
		s = strings.ReplaceAll(s, p[0], p[1])
	}
	return s
}

// Failed is the checks that did not pass.
func (r Report) Failed() []Check {
	out := make([]Check, 0, len(r.Checks))
	for _, c := range r.Checks {
		if !c.OK {
			out = append(out, c)
		}
	}
	return out
}

// JSON is the report as a tool or a script reads it.
func (r Report) JSON() (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Text writes the report as lines a person reads.
func (r Report) Text(w io.Writer) {
	line := func(format string, a ...any) { fmt.Fprintf(w, format+"\n", a...) }
	line("wt-go %s %s %s", r.Version, r.Revision, r.Platform)
	line("exe:      %s", or(r.Executable, "unknown"))
	line("settings: %s", or(r.Settings.Path, "not available"))
	if r.Settings.Error != "" {
		line("          %s", r.Settings.Error)
	}
	line("native:   %s", or(r.Native.Path, "not found"))
	for _, p := range r.Profiles {
		state := "ok"
		if !p.Available {
			state = "missing"
		}
		line("profile:  %-16s %-40s %s", p.ID, p.Command, state)
	}
	for _, c := range r.Crashes {
		line("crash:    %s", c.Path)
	}
	if r.Log.Path != "" {
		line("log:      %s", r.Log.Path)
	}
	if r.Log.Fault != "" {
		line("last fault: %s", r.Log.Fault)
	}
	if failed := r.Failed(); len(failed) > 0 {
		for _, c := range failed {
			line("FAILED    %s %s", c.Code, c.Detail)
		}
		return
	}
	line("checks:   all %d passed", len(r.Checks))
}

// Explain is the report as one block of text for an agent: the lines, then
// each failed check with what its code means and where the details are
// written.
func (r Report) Explain() string {
	var b strings.Builder
	r.Text(&b)
	for _, c := range r.Failed() {
		fmt.Fprintf(&b, "\n%s\n", fault.Line(c.Code, c.Detail))
	}
	return b.String()
}

func missing(ps []ProfileReport) string {
	var out []string
	for _, p := range ps {
		if !p.Available {
			out = append(out, p.ID+" ("+p.Command+")")
		}
	}
	if len(out) == 0 {
		return ""
	}
	return "not on this machine: " + strings.Join(out, ", ")
}

func count(n int, what string) string {
	if n == 1 {
		return "1 " + what
	}
	return fmt.Sprintf("%d %ss", n, what)
}

func keybindings(n int) string {
	if n == 0 {
		return ""
	}
	return ", " + count(n, "keybinding")
}

func or(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
