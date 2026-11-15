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
	"wtgo/internal/fault"
	"wtgo/internal/sys"
)

// packagedVersion is the version mygo.json announces, passed in by the build
// scripts with -ldflags -X. It is not the same number as the one the toolchain
// stamps into the binary: that one says which build answered, this one says
// which release it is meant to be.
var packagedVersion string

// Report is what wt-go knows about itself from outside: a fresh process
// reading the files the window reads.
type Report struct {
	Program    string          `json:"program"`
	Version    string          `json:"version"`
	Revision   string          `json:"revision,omitempty"`
	Packaged   string          `json:"packaged,omitempty"`
	Go         string          `json:"go"`
	Platform   string          `json:"platform"`
	Executable string          `json:"executable,omitempty"`
	Settings   SettingsReport  `json:"settings"`
	Native     NativeReport    `json:"native"`
	ConPTY     ConPTYReport    `json:"conpty"`
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

// ConPTYReport is the conpty.dll the terminals would be made in, empty when
// Windows' own console is: the window still works, and what a program asks the
// terminal for does not come back.
type ConPTYReport struct {
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
	r.Packaged = packagedVersion
	if exe, err := sys.Executable(); err == nil {
		r.Executable = exe
	}
	r.Settings = settingsReport()
	r.Native = nativeReport()
	r.ConPTY = conptyReport()
	r.Shells = shells()
	r.Profiles = profiles(r.Shells, r.Settings.Error == "")
	r.Crashes = crashes()
	r.Log = logReport()
	r.Checks = checks(r)
	return r
}

// version is what `wt-go version` and the report call this build: which one
// answered, and, in a build a script stamped, the release it was packaged as.
func (r Report) version() string { return versionLabel(r.Version, r.Packaged) }

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
	line("wt-go %s %s %s", r.version(), r.Revision, r.Platform)
	line("exe:      %s", or(r.Executable, "unknown"))
	line("settings: %s", or(r.Settings.Path, "not available"))
	if r.Settings.Error != "" {
		line("          %s", r.Settings.Error)
	}
	line("native:   %s", or(r.Native.Path, "not found"))
	if strings.HasPrefix(r.Platform, "windows/") {
		line("conpty:   %s", or(r.ConPTY.Path, "Windows' own console, which drops what a program asks"))
	}
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
// each failed check with what its code means, then where the details are
// written — once for the report, not once per line.
func (r Report) Explain() string {
	var b strings.Builder
	r.Text(&b)
	failed := r.Failed()
	for _, c := range failed {
		fmt.Fprintf(&b, "\n%s\n", fault.Line(c.Code, c.Detail))
	}
	if len(failed) > 0 {
		fmt.Fprintf(&b, "\nsee %s\n", fault.Doc())
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
